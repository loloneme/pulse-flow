package in_memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/loloneme/pulse-flow/internal/infrastructure/logger"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
	"go.uber.org/zap"
)

type Bus struct {
	mu sync.RWMutex

	subs  map[messaging.EventType][]messaging.Subscriber
	queue chan messaging.Event
	stop  chan struct{}
	wg    sync.WaitGroup

	closed bool
}

func New() *Bus {
	bus := &Bus{
		subs:   make(map[messaging.EventType][]messaging.Subscriber),
		queue:  make(chan messaging.Event),
		stop:   make(chan struct{}),
		closed: false,
		wg:     sync.WaitGroup{},
	}

	bus.wg.Add(1)
	go bus.startWorker()

	return bus
}

func (b *Bus) Publish(ctx context.Context, event messaging.Event) error {
	b.mu.RLock()
	closed := b.closed
	b.mu.RUnlock()

	if closed {
		return fmt.Errorf("bus is closed")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case b.queue <- event:
		return nil
	}
}

func (b *Bus) Subscribe(eventType messaging.EventType, sub messaging.Subscriber) error {
	b.mu.Lock()
	closed := b.closed
	defer b.mu.Unlock()

	if closed {
		return fmt.Errorf("bus is closed")
	}

	b.subs[eventType] = append(b.subs[eventType], sub)
	return nil
}

func (b *Bus) startWorker() {
	defer b.wg.Done()

	for {
		select {
		case <-b.stop:
			return
		case event := <-b.queue:
			b.handleEvent(event)
		}
	}
}

func (b *Bus) handleEvent(event messaging.Event) {
	b.mu.RLock()
	subs := b.subs[event.Type()]
	b.mu.RUnlock()
	for _, sub := range subs {
		b.wg.Add(1)
		go func(sub messaging.Subscriber) {
			defer b.wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := sub.Handle(ctx, event); err != nil {
				logger.Log.Error("Error handling event",
					zap.String("event_type", string(event.Type())),
					zap.String("event_id", event.ID().String()),
					zap.Error(err),
				)
			}
		}(sub)
	}

}

func (b *Bus) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	b.mu.Unlock()

	close(b.stop)
	b.wg.Wait()
	close(b.queue)
	return nil
}

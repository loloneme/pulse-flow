package messaging

import "context"

type Subscriber interface {
	Handle(ctx context.Context, event Event) error
}

type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType EventType, sub Subscriber) error
}

package confirmation

import (
	"context"
	"fmt"
	"time"

	"github.com/loloneme/pulse-flow/internal/domain/events"
	"github.com/loloneme/pulse-flow/internal/infrastructure/logger"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
	"go.uber.org/zap"
)


type Worker struct {
	log       *logger.WorkerLogger
	eventBus  messaging.EventBus
	orderRepo OrderRepo
}

func New(log *logger.WorkerLogger, eventBus messaging.EventBus, orderRepo OrderRepo) *Worker {
	return &Worker{
		log:       log,
		eventBus:  eventBus,
		orderRepo: orderRepo,
	}
}

func (w *Worker) Handle(ctx context.Context, event messaging.Event) error {
	start := time.Now()
	payload, ok := event.Payload().(events.PaymentSucceededPayload)
	if !ok {
		return fmt.Errorf("invalid payload type: expected PaymentSucceededPayload, got %T", event.Payload())
	}

	eventLog := w.log.LogEventStart(ctx, string(event.Type()), event.ID(), payload.OrderID)
	ctx = logger.WithEventHandleState(ctx, logger.EventHandleState{Log: eventLog, Start: start})

	order, err := w.orderRepo.GetByID(ctx, payload.OrderID)
	if err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to get order: %w", err))
		return fmt.Errorf("failed to get order: %w", err)
	}

	if err := (&order).MarkAsConfirmed(); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to confirm order: %w", err))
		return fmt.Errorf("failed to confirm order: %w", err)
	}
	if err := w.orderRepo.Save(ctx, order); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to save order: %w", err))
		return fmt.Errorf("failed to save order: %w", err)
	}
	confirmedEvent := events.NewOrderConfirmed(events.OrderConfirmedPayload{OrderID: order.ID})
	if err := w.eventBus.Publish(ctx, confirmedEvent); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to publish OrderConfirmed: %w", err))
		return fmt.Errorf("failed to publish OrderConfirmed: %w", err)
	}
	w.log.Success(ctx, zap.String("order_id", order.ID.String()))
	return nil
}

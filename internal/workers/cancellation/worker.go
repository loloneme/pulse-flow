package cancellation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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

	var orderID uuid.UUID
	var reason string

	switch event.Type() {
	case messaging.ValidationFailed:
		payload, ok := event.Payload().(events.ValidationFailedPayload)
		if !ok {
			return fmt.Errorf("invalid payload type for ValidationFailed event")
		}
		orderID = payload.OrderID
		reason = payload.Reason

	case messaging.PaymentFailed:
		payload, ok := event.Payload().(events.PaymentFailedPayload)
		if !ok {
			return fmt.Errorf("invalid payload type for PaymentFailed event")
		}
		orderID = payload.OrderID
		reason = payload.Reason

	default:
		return fmt.Errorf("unsupported event type: %s", event.Type())
	}

	eventLog := w.log.LogEventStart(ctx, string(event.Type()), event.ID(), orderID)
	ctx = logger.WithEventHandleState(ctx, logger.EventHandleState{Log: eventLog, Start: start})

	order, err := w.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to get order: %w", err))
		return fmt.Errorf("failed to get order: %w", err)
	}

	if err := order.Cancel(); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to mark order as cancelled: %w", err))
		return fmt.Errorf("failed to mark order as cancelled: %w", err)
	}

	if err := w.orderRepo.Save(ctx, order); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to save order: %w", err))
		return fmt.Errorf("failed to save order: %w", err)
	}

	cancelledEvent := events.NewOrderCancelled(events.OrderCancelledPayload{
		OrderID: order.ID,
		Reason:  reason,
	})
	if err := w.eventBus.Publish(ctx, cancelledEvent); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to publish OrderCancelled: %w", err))
		return fmt.Errorf("failed to publish OrderCancelled: %w", err)
	}

	w.log.Success(ctx, zap.String("order_id", order.ID.String()), zap.String("reason", reason))
	return nil
}

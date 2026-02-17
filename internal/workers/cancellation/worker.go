package cancellation

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/loloneme/pulse-flow/internal/domain/events"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type Worker struct {
	eventBus  messaging.EventBus
	orderRepo OrderRepo
}

func New(eventBus messaging.EventBus, orderRepo OrderRepo) *Worker {
	return &Worker{
		eventBus:  eventBus,
		orderRepo: orderRepo,
	}
}

func (w *Worker) Handle(ctx context.Context, event messaging.Event) error {
	log.Printf("[Cancellation Worker] Processing event: %v", event.Type())

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
		log.Printf("[Cancellation Worker] Handling ValidationFailed for order %s: %s", orderID, reason)

	case messaging.PaymentFailed:
		payload, ok := event.Payload().(events.PaymentFailedPayload)
		if !ok {
			return fmt.Errorf("invalid payload type for PaymentFailed event")
		}
		orderID = payload.OrderID
		reason = payload.Reason
		log.Printf("[Cancellation Worker] Handling PaymentFailed for order %s: %s", orderID, reason)

	default:
		return fmt.Errorf("unsupported event type: %s", event.Type())
	}

	order, err := w.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if err := order.Cancel(); err != nil {
		return fmt.Errorf("failed to mark order as cancelled: %w", err)
	}

	if err := w.orderRepo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	cancelledEvent := events.NewOrderCancelled(events.OrderCancelledPayload{
		OrderID: order.ID,
		Reason:  reason,
	})
	if err := w.eventBus.Publish(ctx, cancelledEvent); err != nil {
		return fmt.Errorf("failed to publish OrderCancelled: %w", err)
	}

	log.Printf("[Cancellation Worker] Order %s cancelled due to: %s", order.ID, reason)
	return nil
}

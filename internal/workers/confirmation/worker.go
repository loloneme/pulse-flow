package confirmation

import (
	"context"
	"fmt"
	"log"

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
	log.Printf("[Confirmation Worker] Processing event: %v", event.Type())
	payload, ok := event.Payload().(events.PaymentSucceededPayload)
	if !ok {
		return fmt.Errorf("invalid payload type: expected PaymentSucceededPayload, got %T", event.Payload())
	}

	order, err := w.orderRepo.GetByID(ctx, payload.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if err := (&order).MarkAsConfirmed(); err != nil {
		return fmt.Errorf("failed to confirm order: %w", err)
	}
	if err := w.orderRepo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}
	confirmedEvent := events.NewOrderConfirmed(events.OrderConfirmedPayload{OrderID: order.ID})
	if err := w.eventBus.Publish(ctx, confirmedEvent); err != nil {
		return fmt.Errorf("failed to publish OrderConfirmed: %w", err)
	}
	log.Printf("[Confirmation Worker] Order %s confirmed", order.ID)
	return nil
}

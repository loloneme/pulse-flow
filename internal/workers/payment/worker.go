package payment

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/loloneme/pulse-flow/internal/domain/events"
	domain "github.com/loloneme/pulse-flow/internal/domain/order"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type Worker struct {
	eventBus       messaging.EventBus
	orderRepo      OrderRepo
	paymentService PaymentService
}

func New(eventBus messaging.EventBus, orderRepo OrderRepo, paymentService PaymentService) *Worker {
	return &Worker{
		eventBus:       eventBus,
		orderRepo:      orderRepo,
		paymentService: paymentService,
	}
}

func (w *Worker) Handle(ctx context.Context, event messaging.Event) error {
	log.Printf("[Payment Worker] Processing event: %v", event.Type())

	payload, ok := event.Payload().(events.PaymentStartedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for PaymentStarted event")
	}

	order, err := w.orderRepo.GetByID(ctx, payload.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if err := order.MarkAsPaymentPending(); err != nil {
		return fmt.Errorf("failed to mark order as payment pending: %w", err)
	}

	if err := w.orderRepo.Save(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	err = w.paymentService.ProcessPayment(ctx, &order)

	if err != nil {
		return w.handleFailure(ctx, &order, err)
	}

	return w.handleSuccess(ctx, &order)
}

func (w *Worker) handleSuccess(ctx context.Context, order *domain.Order) error {
	if err := order.MarkAsPaid(); err != nil {
		return fmt.Errorf("failed to mark order as paid: %w", err)
	}

	if err := w.orderRepo.Save(ctx, *order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	successEvent := events.NewPaymentSucceeded(events.PaymentSucceededPayload{
		OrderID:     order.ID,
		PaymentID:   uuid.New(),
		Amount:      order.Amount,
		ProcessedAt: time.Now().Unix(),
	})

	if err := w.eventBus.Publish(ctx, successEvent); err != nil {
		return fmt.Errorf("failed to publish PaymentSucceeded: %w", err)
	}

	log.Printf("[Payment Worker] Order %s payment succeeded", order.ID)
	return nil
}

func (w *Worker) handleFailure(ctx context.Context, order *domain.Order, paymentErr error) error {
	if err := order.MarkAsPaymentFailed(); err != nil {
		return fmt.Errorf("failed to mark order as payment failed: %w", err)
	}

	if err := w.orderRepo.Save(ctx, *order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	failedEvent := events.NewPaymentFailed(events.PaymentFailedPayload{
		OrderID: order.ID,
		Reason:  paymentErr.Error(),
	})

	if err := w.eventBus.Publish(ctx, failedEvent); err != nil {
		return fmt.Errorf("failed to publish PaymentFailed: %w", err)
	}

	log.Printf("[Payment Worker] Order %s payment failed: %s", order.ID, paymentErr.Error())
	return nil
}

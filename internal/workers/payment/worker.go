package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/loloneme/pulse-flow/internal/config"
	"github.com/loloneme/pulse-flow/internal/domain/events"
	domain "github.com/loloneme/pulse-flow/internal/domain/order"
	"github.com/loloneme/pulse-flow/internal/infrastructure/logger"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
	"github.com/loloneme/pulse-flow/internal/infrastructure/resilience"
	"go.uber.org/zap"
)

type Worker struct {
	config *config.WorkerConfig
	log    *logger.WorkerLogger

	PaymentCircuitBreaker *resilience.CircuitBreaker[struct{}]

	eventBus       messaging.EventBus
	orderRepo      OrderRepo
	paymentService PaymentService
}

func New(config *config.WorkerConfig,
	log *logger.WorkerLogger,
	eventBus messaging.EventBus,
	orderRepo OrderRepo,
	paymentService PaymentService) *Worker {
	return &Worker{
		config:                config,
		log:                   log,
		PaymentCircuitBreaker: resilience.NewCircuitBreaker[struct{}](config.CircuitBreakerConfig),
		eventBus:              eventBus,
		orderRepo:             orderRepo,
		paymentService:        paymentService,
	}
}

func (w *Worker) Handle(ctx context.Context, event messaging.Event) error {
	ctx, cancel := context.WithTimeout(ctx, w.config.ExternalServiceTimeout)
	defer cancel()

	start := time.Now()
	payload, ok := event.Payload().(events.OrderValidatedPayload)
	if !ok {
		return fmt.Errorf("invalid payload type for OrderValidated event")
	}

	eventLog := w.log.LogEventStart(ctx, string(event.Type()), event.ID(), payload.OrderID)
	ctx = logger.WithEventHandleState(ctx, logger.EventHandleState{Log: eventLog, Start: start})

	order, err := w.orderRepo.GetByID(ctx, payload.OrderID)
	if err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to get order: %w", err))
		return fmt.Errorf("failed to get order: %w", err)
	}

	if err := order.MarkAsPaymentPending(); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to mark order as payment pending: %w", err))
		return fmt.Errorf("failed to mark order as payment pending: %w", err)
	}

	if err := w.orderRepo.Save(ctx, order); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to save order: %w", err))
		return fmt.Errorf("failed to save order: %w", err)
	}

	_, err = resilience.WithResilience(ctx, w.PaymentCircuitBreaker, w.config.RetryConfig,
		func(ctx context.Context) (struct{}, error) {
			return struct{}{}, w.paymentService.ProcessPayment(ctx, &order)
		})

	if err != nil {
		return w.handleFailure(ctx, &order, err)
	}

	return w.handleSuccess(ctx, &order)
}

func (w *Worker) handleSuccess(ctx context.Context, order *domain.Order) error {
	if err := order.MarkAsPaid(); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to mark order as paid: %w", err))
		return fmt.Errorf("failed to mark order as paid: %w", err)
	}

	if err := w.orderRepo.Save(ctx, *order); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to save order: %w", err))
		return fmt.Errorf("failed to save order: %w", err)
	}

	successEvent := events.NewPaymentSucceeded(events.PaymentSucceededPayload{
		OrderID:     order.ID,
		PaymentID:   uuid.New(),
		Amount:      order.Amount,
		ProcessedAt: time.Now().Unix(),
	})

	if err := w.eventBus.Publish(ctx, successEvent); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to publish PaymentSucceeded: %w", err))
		return fmt.Errorf("failed to publish PaymentSucceeded: %w", err)
	}

	w.log.Success(ctx, zap.String("order_id", order.ID.String()))
	return nil
}

func (w *Worker) handleFailure(ctx context.Context, order *domain.Order, paymentErr error) error {
	if err := order.MarkAsPaymentFailed(); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to mark order as payment failed: %w", err))
		return fmt.Errorf("failed to mark order as payment failed: %w", err)
	}

	if err := w.orderRepo.Save(ctx, *order); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to save order: %w", err))
		return fmt.Errorf("failed to save order: %w", err)
	}

	failedEvent := events.NewPaymentFailed(events.PaymentFailedPayload{
		OrderID: order.ID,
		Reason:  paymentErr.Error(),
	})

	if err := w.eventBus.Publish(ctx, failedEvent); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to publish PaymentFailed: %w", err))
		return fmt.Errorf("failed to publish PaymentFailed: %w", err)
	}

	w.log.Error(ctx, paymentErr, zap.String("order_id", order.ID.String()))
	return nil
}

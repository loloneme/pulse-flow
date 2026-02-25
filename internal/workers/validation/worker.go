package validation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/loloneme/pulse-flow/internal/config"
	"github.com/loloneme/pulse-flow/internal/domain/events"
	domain "github.com/loloneme/pulse-flow/internal/domain/order"
	"github.com/loloneme/pulse-flow/internal/infrastructure/logger"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
	"github.com/loloneme/pulse-flow/internal/infrastructure/resilience"
	"github.com/loloneme/pulse-flow/internal/workers/validation/mocks"
	"go.uber.org/zap"
)

type check struct {
	name    string
	success bool
	reason  string
	err     error
}

type Services struct {
	WarehouseService WarehouseService
	AntiFraudService AntiFraudService
	UserService      UserService
}

type Worker struct {
	config *config.WorkerConfig
	log    *logger.WorkerLogger

	ValidationCircuitBreaker *resilience.CircuitBreaker[any]

	eventBus  messaging.EventBus
	orderRepo OrderRepo
	Services  Services
}

func New(
	config *config.WorkerConfig,
	log *logger.WorkerLogger,
	eventBus messaging.EventBus,
	orderRepo OrderRepo,
	services Services) *Worker {
	return &Worker{
		config:                   config,
		log:                      log,
		ValidationCircuitBreaker: resilience.NewCircuitBreaker[any](config.CircuitBreakerConfig),
		eventBus:                 eventBus,
		orderRepo:                orderRepo,
		Services:                 services,
	}
}

func (w *Worker) Handle(ctx context.Context, event messaging.Event) error {
	ctx, cancel := context.WithTimeout(ctx, w.config.ExternalServiceTimeout)
	defer cancel()

	payload, ok := event.Payload().(events.OrderCreatedPayload)
	if !ok {
		return errors.New("invalid payload")
	}

	start := time.Now()
	eventLog := w.log.LogEventStart(ctx, string(event.Type()), event.ID(), payload.OrderID)
	ctx = logger.WithEventHandleState(ctx, logger.EventHandleState{Log: eventLog, Start: start})

	order, err := w.orderRepo.GetByID(ctx, payload.OrderID)
	if err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to get order: %w", err))
		return fmt.Errorf("failed to get order: %w", err)
	}

	success, err := w.validateOrder(ctx, &order)
	if success {
		return w.handleSuccess(ctx, &order)
	}
	return w.handleFailure(ctx, &order, err.Error())
}

func (w *Worker) handleSuccess(ctx context.Context, order *domain.Order) error {
	if err := order.MarkAsValidated(); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to mark as validated: %w", err))
		return fmt.Errorf("failed to mark as validated: %w", err)
	}
	if err := w.orderRepo.Save(ctx, *order); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to save order: %w", err))
		return fmt.Errorf("failed to save order: %w", err)
	}
	validatedEvent := events.NewOrderValidated(events.OrderValidatedPayload{OrderID: order.ID})
	if err := w.eventBus.Publish(ctx, validatedEvent); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to publish OrderValidated: %w", err))
		return fmt.Errorf("failed to publish OrderValidated: %w", err)
	}
	w.log.Success(ctx, zap.String("order_id", order.ID.String()))
	return nil
}

func (w *Worker) handleFailure(ctx context.Context, order *domain.Order, reason string) error {
	validationFailedEvent := events.NewValidationFailed(events.ValidationFailedPayload{OrderID: order.ID, Reason: reason})
	if err := w.eventBus.Publish(ctx, validationFailedEvent); err != nil {
		w.log.Error(ctx, fmt.Errorf("failed to publish ValidationFailed: %w", err))
		return fmt.Errorf("failed to publish ValidationFailed: %w", err)
	}
	w.log.Error(ctx, fmt.Errorf("validation failed: %s", reason), zap.String("order_id", order.ID.String()), zap.String("reason", reason))
	return nil
}

func (w *Worker) validateOrder(ctx context.Context, order *domain.Order) (bool, error) {
	checks := make(chan check)
	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		w.checkWarehouseService(ctx, order, checks)
	}()

	go func() {
		defer wg.Done()
		w.checkAntiFraudCreditLimit(ctx, order, checks)
	}()

	go func() {
		defer wg.Done()
		w.checkAntiFraudOrder(ctx, order, checks)
	}()

	go func() {
		defer wg.Done()
		w.checkUserService(ctx, order, checks)
	}()

	go func() {
		wg.Wait()
		close(checks)
	}()

	for res := range checks {
		if res.err != nil || !res.success {
			return false, fmt.Errorf("validation failed: %s %v", res.name, res.err)
		}
	}
	return true, nil
}

func (w *Worker) checkWarehouseService(ctx context.Context, order *domain.Order, checks chan check) {
	success, err := resilience.WithResilienceAs(ctx, w.ValidationCircuitBreaker, w.config.RetryConfig,
		func(ctx context.Context) (bool, error) {
			return w.Services.WarehouseService.CheckProductAvailability(ctx, order.ProductID, order.Amount)
		})

	checks <- check{
		name:    "Warehouse Service Check",
		success: success,
		err:     err,
	}
}

func (w *Worker) checkAntiFraudCreditLimit(ctx context.Context, order *domain.Order, checks chan check) {
	success, err := resilience.WithResilienceAs(ctx, w.ValidationCircuitBreaker, w.config.RetryConfig,
		func(ctx context.Context) (bool, error) {
			return w.Services.AntiFraudService.CheckUserCreditLimit(ctx, order.UserID)
		})

	checks <- check{
		name:    "Anti Fraud Credit Limit Check",
		success: success,
		err:     err,
	}
}

func (w *Worker) checkAntiFraudOrder(ctx context.Context, order *domain.Order, checks chan check) {
	result, err := resilience.WithResilienceAs(ctx, w.ValidationCircuitBreaker, w.config.RetryConfig,
		func(ctx context.Context) (mocks.OrderCheckResult, error) {
			return w.Services.AntiFraudService.CheckOrder(ctx, order)
		})

	checks <- check{
		name:    "Anti Fraud Order Check",
		success: result.Success,
		reason:  result.Reason,
		err:     err,
	}
}

func (w *Worker) checkUserService(ctx context.Context, order *domain.Order, checks chan check) {
	success, err := resilience.WithResilienceAs(ctx, w.ValidationCircuitBreaker, w.config.RetryConfig,
		func(ctx context.Context) (bool, error) {
			return w.Services.UserService.CheckUserStatus(ctx, order.UserID)
		})

	checks <- check{
		name:    "User Service Check",
		success: success,
		err:     err,
	}
}

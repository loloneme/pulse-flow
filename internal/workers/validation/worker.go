package validation

import (
	"context"
	"errors"
	"fmt"
	"log"

	"sync"

	"github.com/loloneme/pulse-flow/internal/domain/events"
	domain "github.com/loloneme/pulse-flow/internal/domain/order"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type check struct {
	name    string
	success bool
	reason  string
	err     error
}

type Worker struct {
	eventBus         messaging.EventBus
	orderRepo        OrderRepo
	warehouseService WarehouseService
	antiFraudService AntiFraudService
	userService      UserService
}

func New(
	eventBus messaging.EventBus,
	orderRepo OrderRepo,
	warehouseService WarehouseService,
	antiFraudService AntiFraudService,
	userService UserService) *Worker {
	return &Worker{
		eventBus:         eventBus,
		orderRepo:        orderRepo,
		warehouseService: warehouseService,
		antiFraudService: antiFraudService,
		userService:      userService,
	}
}

func (w *Worker) Handle(ctx context.Context, event messaging.Event) error {
	log.Printf("[Validation Worker] Processing event: %v", event.Type())

	payload, ok := event.Payload().(events.OrderCreatedPayload)
	if !ok {
		return errors.New("invalid payload")
	}

	order, err := w.orderRepo.GetByID(ctx, payload.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	log.Printf("[Validation Worker] Starting validation for order %s", order.ID)
	success, err := w.validateOrder(ctx, &order)
	if success {
		return w.handleSuccess(ctx, &order)
	}
	return w.handleFailure(ctx, &order, err.Error())
}

func (w *Worker) handleSuccess(ctx context.Context, order *domain.Order) error {
	if err := order.MarkAsValidated(); err != nil {
		return fmt.Errorf("failed to mark as validated: %w", err)
	}
	if err := w.orderRepo.Save(ctx, *order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}
	validatedEvent := events.NewOrderValidated(events.OrderValidatedPayload{OrderID: order.ID})
	if err := w.eventBus.Publish(ctx, validatedEvent); err != nil {
		return fmt.Errorf("failed to publish OrderValidated: %w", err)
	}
	log.Printf("[Validation Worker] Order %s validated successfully", order.ID)
	return nil
}

func (w *Worker) handleFailure(ctx context.Context, order *domain.Order, reason string) error {
	validationFailedEvent := events.NewValidationFailed(events.ValidationFailedPayload{OrderID: order.ID, Reason: reason})
	if err := w.eventBus.Publish(ctx, validationFailedEvent); err != nil {
		return fmt.Errorf("failed to publish ValidationFailed: %w", err)
	}
	log.Printf("[Validation Worker] Order %s marked as validation failed: %s", order.ID, reason)
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
	success, err := w.warehouseService.CheckProductAvailability(ctx, order.ProductID, order.Amount)

	checks <- check{
		name:    "Warehouse Service Check",
		success: success,
		err:     err,
	}
}

func (w *Worker) checkAntiFraudCreditLimit(ctx context.Context, order *domain.Order, checks chan check) {
	success, err := w.antiFraudService.CheckUserCreditLimit(ctx, order.UserID)

	checks <- check{
		name:    "Anti Fraud Credit Limit Check",
		success: success,
		err:     err,
	}
}

func (w *Worker) checkAntiFraudOrder(ctx context.Context, order *domain.Order, checks chan check) {
	success, reason, err := w.antiFraudService.CheckOrder(ctx, order)

	checks <- check{
		name:    "Anti Fraud Order Check",
		success: success,
		reason:  reason,
		err:     err,
	}
}

func (w *Worker) checkUserService(ctx context.Context, order *domain.Order, checks chan check) {
	success, err := w.userService.CheckUserStatus(ctx, order.UserID)

	checks <- check{
		name:    "User Service Check",
		success: success,
		err:     err,
	}
}

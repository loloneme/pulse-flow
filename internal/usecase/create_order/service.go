package create_order

import (
	"context"

	"github.com/loloneme/pulse-flow/internal/domain/events"
	domain "github.com/loloneme/pulse-flow/internal/domain/order"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type Service struct {
	orderRepo OrderRepo
	eventBus  messaging.EventBus
}

func New(orderRepo OrderRepo, eventBus messaging.EventBus) *Service {
	return &Service{
		orderRepo: orderRepo,
		eventBus:  eventBus,
	}
}

func (s *Service) CreateOrder(ctx context.Context, req *CreateOrderRequest) error {
	order, err := domain.NewOrder(req.UserID, req.ProductID, req.Amount)
	if err != nil {
		return err
	}

	if err := s.orderRepo.Save(ctx, *order); err != nil {
		return err
	}

	orderCreatedEvent := events.NewOrderCreated(events.OrderCreatedPayload{
		OrderID:   order.ID,
		UserID:    order.UserID,
		ProductID: order.ProductID,
		Amount:    order.Amount,
	})

	if err := s.eventBus.Publish(ctx, orderCreatedEvent); err != nil {
		return err
	}

	return nil

}

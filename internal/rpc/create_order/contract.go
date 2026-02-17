package create_order

import (
	"context"

	"github.com/google/uuid"
	"github.com/loloneme/pulse-flow/internal/usecase/create_order"
)

type CreateOrderService interface {
	CreateOrder(ctx context.Context, req *create_order.CreateOrderRequest) error
}

type CreateOrderRequest struct {
	UserID    uuid.UUID `json:"user_id"`
	ProductID uuid.UUID `json:"product_id"`
	Amount    int       `json:"amount"`
}

type CreateOrderResponse struct {
	OrderID uuid.UUID `json:"order_id"`
}

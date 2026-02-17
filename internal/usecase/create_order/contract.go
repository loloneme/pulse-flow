package create_order

import (
	"context"

	"github.com/google/uuid"
	domain "github.com/loloneme/pulse-flow/internal/domain/order"
)

type OrderRepo interface {
	Save(ctx context.Context, order domain.Order) error
}

type CreateOrderRequest struct {
	UserID    uuid.UUID
	ProductID uuid.UUID
	Amount    int
}

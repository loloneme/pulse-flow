package payment

import (
	"context"

	"github.com/google/uuid"
	domain "github.com/loloneme/pulse-flow/internal/domain/order"
)

type OrderRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.Order, error)
	Save(ctx context.Context, model domain.Order) error
}

type PaymentService interface {
	ProcessPayment(ctx context.Context, order *domain.Order) error
}

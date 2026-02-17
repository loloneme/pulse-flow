package validation

import (
	"context"

	"github.com/google/uuid"
	domain "github.com/loloneme/pulse-flow/internal/domain/order"
)

type OrderRepo interface {
	GetByID(ctx context.Context, id uuid.UUID) (domain.Order, error)
	Save(ctx context.Context, model domain.Order) error
}

type WarehouseService interface {
	CheckProductAvailability(ctx context.Context, productID uuid.UUID, amount int) (bool, error)
}

type AntiFraudService interface {
	CheckUserCreditLimit(ctx context.Context, userID uuid.UUID) (bool, error)
	CheckOrder(ctx context.Context, order *domain.Order) (bool, string, error)
}

type UserService interface {
	CheckUserStatus(ctx context.Context, userID uuid.UUID) (bool, error)
}

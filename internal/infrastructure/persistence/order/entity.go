package order

import (
	"time"

	"github.com/google/uuid"
	domain "github.com/loloneme/pulse-flow/internal/domain/order"
)

type Order struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	ProductId uuid.UUID `db:"product_id"`
	Amount    int       `db:"amount"`
	Status    Status    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Status string

const (
	OrderStatusCreated          Status = "created"
	OrderStatusValidated        Status = "validated"
	OrderStatusValidationFailed Status = "validation failed"
	OrderStatusPaymentPending   Status = "payment pending"
	OrderStatusPaid             Status = "paid"
	OrderStatusPaymentFailed    Status = "payment failed"
	OrderStatusCancelled        Status = "cancelled"
	OrderStatusConfirmed        Status = "confirmed"
)

func (s Status) String() string { return string(s) }

func (o Order) Values() []any {
	return []any{
		o.UserID,
		o.ProductId,
		o.Amount,
		o.Status,
		o.CreatedAt,
		o.UpdatedAt,
	}
}

func (o Order) GetID() uuid.UUID {
	return o.ID
}

func (o Order) ToModel() domain.Order {
	return domain.Order{
		ID:        o.ID,
		UserID:    o.UserID,
		ProductID: o.ProductId,
		Amount:    o.Amount,
		Status:    domain.Status(o.Status),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func (o Order) FromModel(m domain.Order) any {
	return Order{
		ID:        m.ID,
		UserID:    m.UserID,
		ProductId: m.ProductID,
		Amount:    m.Amount,
		Status:    Status(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

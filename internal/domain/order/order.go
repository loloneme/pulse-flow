package order

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ProductID uuid.UUID
	Amount    int
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Status string

const (
	StatusCreated          Status = "created"
	StatusValidated        Status = "validated"
	StatusValidationFailed Status = "validation failed"
	StatusPaymentPending   Status = "payment pending"
	StatusPaid             Status = "paid"
	StatusPaymentFailed    Status = "payment failed"
	StatusCancelled        Status = "cancelled"
	StatusConfirmed        Status = "confirmed"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	switch s {
	case StatusCreated, StatusValidated, StatusValidationFailed, StatusPaymentPending,
		StatusPaid, StatusPaymentFailed, StatusCancelled, StatusConfirmed:
		return true
	default:
		return false
	}
}

var (
	ErrInvalidAmount    = errors.New("order amount must be greater than zero")
	ErrInvalidStatus    = errors.New("invalid order status")
	ErrStatusTransition = errors.New("invalid status transition")
)

func NewOrder(userID, productID uuid.UUID, amount int) (*Order, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	now := time.Now()
	return &Order{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: productID,
		Amount:    amount,
		Status:    StatusCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (o *Order) Validate() error {
	if o.Amount <= 0 {
		return ErrInvalidAmount
	}
	if !o.Status.IsValid() {
		return ErrInvalidStatus
	}
	return nil
}

func (o *Order) MarkAsValidated() error {
	if o.Status != StatusCreated {
		return ErrStatusTransition
	}
	o.Status = StatusValidated
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) MarkAsValidationFailed() error {
	if o.Status != StatusCreated {
		return ErrStatusTransition
	}
	o.Status = StatusValidationFailed
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) MarkAsPaymentPending() error {
	if o.Status != StatusValidated {
		return ErrStatusTransition
	}
	o.Status = StatusPaymentPending
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) MarkAsPaid() error {
	if o.Status != StatusPaymentPending {
		return ErrStatusTransition
	}
	o.Status = StatusPaid
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) MarkAsPaymentFailed() error {
	if o.Status != StatusPaymentPending {
		return ErrStatusTransition
	}
	o.Status = StatusPaymentFailed
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) MarkAsConfirmed() error {
	if o.Status != StatusPaid {
		return ErrStatusTransition
	}
	o.Status = StatusConfirmed
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) Cancel() error {
	switch o.Status {
	case StatusCreated, StatusValidated, StatusPaymentPending:
		o.Status = StatusCancelled
		o.UpdatedAt = time.Now()
		return nil
	default:
		return ErrStatusTransition
	}
}

func (o *Order) CanTransitionTo(newStatus Status) bool {
	switch newStatus {
	case StatusValidated:
		return o.Status == StatusCreated
	case StatusValidationFailed:
		return o.Status == StatusCreated
	case StatusPaymentPending:
		return o.Status == StatusValidated
	case StatusPaid:
		return o.Status == StatusPaymentPending
	case StatusPaymentFailed:
		return o.Status == StatusPaymentPending
	case StatusConfirmed:
		return o.Status == StatusPaid
	case StatusCancelled:
		return o.Status == StatusCreated || o.Status == StatusValidationFailed || o.Status == StatusPaymentFailed
	default:
		return false
	}
}

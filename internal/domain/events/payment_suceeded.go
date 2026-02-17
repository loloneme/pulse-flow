package events

import (
	"github.com/google/uuid"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type PaymentSucceededPayload struct {
	OrderID     uuid.UUID
	PaymentID   uuid.UUID
	Amount      int
	ProcessedAt int64
}

type PaymentSucceeded struct {
	payload PaymentSucceededPayload
	eventID uuid.UUID
}

func NewPaymentSucceeded(payload PaymentSucceededPayload) *PaymentSucceeded {
	return &PaymentSucceeded{
		payload: payload,
		eventID: uuid.New(),
	}
}

func (e *PaymentSucceeded) ID() uuid.UUID             { return e.eventID }
func (e *PaymentSucceeded) Type() messaging.EventType { return messaging.PaymentSucceeded }
func (e *PaymentSucceeded) Payload() any {
	return e.payload
}

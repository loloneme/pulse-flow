package events

import (
	"github.com/google/uuid"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type PaymentFailedPayload struct {
	OrderID uuid.UUID
	Reason  string
}

type PaymentFailed struct {
	eventID uuid.UUID
	payload PaymentFailedPayload
}

func NewPaymentFailed(payload PaymentFailedPayload) *PaymentFailed {
	return &PaymentFailed{
		payload: payload,
		eventID: uuid.New(),
	}
}

func (e *PaymentFailed) ID() uuid.UUID             { return e.eventID }
func (e *PaymentFailed) Type() messaging.EventType { return messaging.PaymentFailed }
func (e *PaymentFailed) Payload() any {
	return e.payload
}

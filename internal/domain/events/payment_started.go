package events

import (
	"github.com/google/uuid"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type PaymentStartedPayload struct {
	OrderID uuid.UUID
	Amount  int
	Price   float64
}

type PaymentStarted struct {
	payload PaymentStartedPayload
	eventID uuid.UUID
}

func NewPaymentStarted(payload PaymentStartedPayload) *PaymentStarted {
	return &PaymentStarted{
		payload: payload,
		eventID: uuid.New(),
	}
}

func (e *PaymentStarted) ID() uuid.UUID {
	return e.eventID
}

func (e *PaymentStarted) Type() messaging.EventType {
	return messaging.PaymentStarted
}

func (e *PaymentStarted) Payload() any {
	return e.payload
}

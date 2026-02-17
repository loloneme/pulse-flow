package events

import (
	"github.com/google/uuid"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type OrderConfirmedPayload struct {
	OrderID uuid.UUID
}

type OrderConfirmed struct {
	eventID uuid.UUID
	payload OrderConfirmedPayload
}

func NewOrderConfirmed(payload OrderConfirmedPayload) *OrderConfirmed {
	return &OrderConfirmed{
		payload: payload,
		eventID: uuid.New(),
	}
}

func (e *OrderConfirmed) ID() uuid.UUID {
	return e.eventID
}

func (e *OrderConfirmed) Type() messaging.EventType {
	return messaging.OrderConfirmed
}

func (e *OrderConfirmed) Payload() any {
	return e.payload
}

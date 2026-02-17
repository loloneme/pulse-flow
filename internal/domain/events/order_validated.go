package events

import (
	"github.com/google/uuid"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type OrderValidatedPayload struct {
	OrderID uuid.UUID
}

type OrderValidated struct {
	payload OrderValidatedPayload
	eventID uuid.UUID
}

func NewOrderValidated(payload OrderValidatedPayload) *OrderValidated {
	return &OrderValidated{
		payload: payload,
		eventID: uuid.New(),
	}
}

func (e *OrderValidated) ID() uuid.UUID {
	return e.eventID
}

func (e *OrderValidated) Type() messaging.EventType {
	return messaging.OrderValidated
}

func (e *OrderValidated) Payload() any {
	return e.payload
}

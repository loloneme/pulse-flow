package events

import (
	"github.com/google/uuid"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type OrderCreatedPayload struct {
	OrderID   uuid.UUID
	UserID    uuid.UUID
	ProductID uuid.UUID
	Amount    int
}

type OrderCreated struct {
	payload OrderCreatedPayload
	eventID uuid.UUID
}

func NewOrderCreated(payload OrderCreatedPayload) *OrderCreated {
	return &OrderCreated{
		payload: payload,
		eventID: uuid.New(),
	}
}

func (e *OrderCreated) ID() uuid.UUID {
	return e.eventID
}

func (e *OrderCreated) Type() messaging.EventType {
	return messaging.OrderCreated
}

func (e *OrderCreated) Payload() any {
	return e.payload
}

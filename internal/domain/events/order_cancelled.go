package events

import (
	"github.com/google/uuid"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type OrderCancelledPayload struct {
	OrderID uuid.UUID
	Reason  string
}

type OrderCancelled struct {
	eventID uuid.UUID
	payload OrderCancelledPayload
}

func NewOrderCancelled(payload OrderCancelledPayload) *OrderCancelled {
	return &OrderCancelled{payload: payload, eventID: uuid.New()}
}

func (e *OrderCancelled) ID() uuid.UUID             { return e.eventID }
func (e *OrderCancelled) Type() messaging.EventType { return messaging.OrderCancelled }
func (e *OrderCancelled) Payload() any {
	return e.payload
}

package messaging

import "github.com/google/uuid"

type EventType string

const (
	OrderCreated     EventType = "OrderCreated"
	OrderValidated   EventType = "OrderValidated"
	ValidationFailed EventType = "ValidationFailed"
	PaymentSucceeded EventType = "PaymentSucceeded"
	PaymentFailed    EventType = "PaymentFailed"
	OrderConfirmed   EventType = "OrderConfirmed"
	OrderCancelled   EventType = "OrderCancelled"
)

type Event interface {
	ID() uuid.UUID
	Type() EventType
	Payload() any
}

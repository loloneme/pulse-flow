package events

import (
	"github.com/google/uuid"
	"github.com/loloneme/pulse-flow/internal/infrastructure/messaging"
)

type ValidationFailedPayload struct {
	OrderID uuid.UUID
	Reason  string
}

type ValidationFailed struct {
	eventID uuid.UUID
	payload ValidationFailedPayload
}

func NewValidationFailed(payload ValidationFailedPayload) *ValidationFailed {
	return &ValidationFailed{
		payload: payload,
		eventID: uuid.New(),
	}
}

func (e *ValidationFailed) ID() uuid.UUID             { return e.eventID }
func (e *ValidationFailed) Type() messaging.EventType { return messaging.ValidationFailed }
func (e *ValidationFailed) Payload() any {
	return e.payload
}

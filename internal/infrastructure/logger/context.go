package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type contextKey string

const (
	correlationIDKey    contextKey = "correlation_id"
	eventHandleStateKey contextKey = "event_handle_state"
)

type EventHandleState struct {
	Log   *zap.Logger
	Start time.Time
}

func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}

func WithEventHandleState(ctx context.Context, state EventHandleState) context.Context {
	return context.WithValue(ctx, eventHandleStateKey, state)
}

func GetEventHandleState(ctx context.Context) (EventHandleState, bool) {
	state, ok := ctx.Value(eventHandleStateKey).(EventHandleState)
	return state, ok
}

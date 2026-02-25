package logger

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WorkerLogger struct {
	logger *zap.Logger
}

func NewWorkerLogger(workerName string) *WorkerLogger {
	logger := Log.With(
		zap.String("component", "worker"),
		zap.String("labels.worker_name", workerName),
	)

	return &WorkerLogger{
		logger: logger,
	}
}

func (wl *WorkerLogger) ForEvent(ctx context.Context, eventType string, eventID, orderID uuid.UUID) *zap.Logger {
	fields := []zap.Field{
		zap.String("event.action", eventType),
		zap.String("event.id", eventID.String()),
		zap.String("labels.order_id", orderID.String()),
	}

	if corrID := GetCorrelationID(ctx); corrID != "" {
		fields = append(fields, zap.String("trace.id", corrID))
	}

	return wl.logger.With(fields...)
}

func (wl *WorkerLogger) LogEventStart(
	ctx context.Context,
	eventType string,
	eventID uuid.UUID,
	orderID uuid.UUID,
) *zap.Logger {
	logger := wl.ForEvent(ctx, eventType, eventID, orderID)
	logger.Info("Processing event")
	return logger
}

func (wl *WorkerLogger) LogEventSuccess(
	logger *zap.Logger,
	duration time.Duration,
	additionalFields ...zap.Field,
) {
	fields := []zap.Field{
		zap.Int64("event.duration", duration.Nanoseconds()),
	}
	fields = append(fields, additionalFields...)

	logger.Info("Event processed successfully", fields...)
}

func (wl *WorkerLogger) LogEventError(
	logger *zap.Logger,
	err error,
	duration time.Duration,
	additionalFields ...zap.Field,
) {
	fields := []zap.Field{
		zap.Error(err),
		zap.Int64("event.duration", duration.Nanoseconds()),
	}
	fields = append(fields, additionalFields...)

	logger.Error("Event processing failed", fields...)
}

func (wl *WorkerLogger) Success(ctx context.Context, fields ...zap.Field) {
	if state, ok := GetEventHandleState(ctx); ok {
		wl.LogEventSuccess(state.Log, time.Since(state.Start), fields...)
	}
}

func (wl *WorkerLogger) Error(ctx context.Context, err error, fields ...zap.Field) {
	if state, ok := GetEventHandleState(ctx); ok {
		wl.LogEventError(state.Log, err, time.Since(state.Start), fields...)
	}
}

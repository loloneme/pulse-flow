package middleware

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/loloneme/pulse-flow/internal/infrastructure/logger"
	"go.uber.org/zap"
)

const RequestIDKey = "request_id"

func LoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			requestID := uuid.New().String()
			c.Set(RequestIDKey, requestID)

			ctx := logger.WithCorrelationID(c.Request().Context(), requestID)
			c.SetRequest(c.Request().WithContext(ctx))

			logger.Log.Info("Incoming request",
				zap.String("component", "http"),
				zap.String("request_id", requestID),
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.String("remote_addr", c.RealIP()),
				zap.String("user_agent", c.Request().UserAgent()),
			)

			err := next(c)

			latency := time.Since(start)
			fields := []zap.Field{
				zap.String("component", "http"),
				zap.String("request_id", requestID),
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.Int("status_code", c.Response().Status),
				zap.Int64("latency_ms", latency.Milliseconds()),
				zap.Int64("bytes_out", c.Response().Size),
			}

			if err != nil {
				fields = append(fields, zap.Error(err))
				logger.Log.Error("Request failed", fields...)
			} else if c.Response().Status >= 500 {
				logger.Log.Error("Request completed with server error", fields...)
			} else if c.Response().Status >= 400 {
				logger.Log.Warn("Request completed with client error", fields...)
			} else {
				logger.Log.Info("Request completed", fields...)
			}

			return err
		}
	}
}

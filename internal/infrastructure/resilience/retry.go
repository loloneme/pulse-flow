package resilience

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"time"

	"github.com/loloneme/pulse-flow/internal/config"
)

func DoWithRetry[T any](ctx context.Context, fn func(ctx context.Context) (T, error), cfg *config.RetryConfig) (T, error) {
	var lastErr error
	var zero T

	backoff := cfg.Delay

	for attempt := 0; attempt < cfg.Attempts; attempt++ {
		result, err := fn(ctx)
		lastErr = err

		if lastErr == nil {
			return result, nil
		}
		if !IsRetryableError(lastErr) {
			return zero, lastErr
		}

		if attempt == cfg.Attempts-1 {
			return zero, lastErr
		}

		backoff = min(2*backoff, cfg.MaxDelay)
		jitter := time.Duration(rand.Int63n(int64(backoff)))
		backoff += jitter

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(backoff):
		}
	}
	return zero, lastErr
}

func IsRetryableError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}

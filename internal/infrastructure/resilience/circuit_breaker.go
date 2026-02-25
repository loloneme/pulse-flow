package resilience

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/loloneme/pulse-flow/internal/config"
)

type State int

var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type Counts struct {
	Requests  int
	Successes int
	Failures  int
}

type CircuitBreaker[T any] struct {
	mu sync.Mutex

	cfg *config.CircuitBreakerConfig

	state  State
	counts Counts

	lastFailure time.Time
}

func NewCircuitBreaker[T any](cfg *config.CircuitBreakerConfig) *CircuitBreaker[T] {
	return &CircuitBreaker[T]{
		cfg:         cfg,
		state:       StateClosed,
		counts:      Counts{Requests: 0, Successes: 0, Failures: 0},
		lastFailure: time.Time{},
	}
}

func (cb *CircuitBreaker[T]) Execute(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
	var zero T

	cb.mu.Lock()

	if cb.isOpen() {
		if time.Since(cb.lastFailure) > cb.cfg.OpenTimeout {
			cb.state = StateHalfOpen
			cb.counts.Requests = 0
			cb.counts.Successes = 0
			cb.counts.Failures = 0
		} else {
			cb.mu.Unlock()
			return zero, ErrCircuitBreakerOpen
		}
	}
	cb.mu.Unlock()

	result, err := fn(ctx)

	cb.mu.Lock()
	defer cb.mu.Unlock()
	if err != nil {
		cb.onFailure()
		return zero, err
	}

	cb.onSuccess()
	return result, nil
}

func (cb *CircuitBreaker[T]) onFailure() {
	cb.counts.Failures++
	cb.counts.Requests++
	if cb.counts.Failures >= cb.cfg.MaxFailures {
		cb.state = StateOpen
		cb.lastFailure = time.Now()
	}
}

func (cb *CircuitBreaker[T]) onSuccess() {
	cb.counts.Failures = 0
	cb.counts.Successes++
	cb.counts.Requests++
	if cb.isHalfOpen() {
		cb.state = StateClosed
	}
}

func (cb *CircuitBreaker[T]) isOpen() bool {
	return cb.state == StateOpen
}

func (cb *CircuitBreaker[T]) isHalfOpen() bool {
	return cb.state == StateHalfOpen
}

func (cb *CircuitBreaker[T]) isClosed() bool {
	return cb.state == StateClosed
}

func WithResilience[T any](ctx context.Context, cb *CircuitBreaker[T], cfg *config.RetryConfig, fn func(ctx context.Context) (T, error)) (T, error) {
	return cb.Execute(ctx, func(ctx context.Context) (T, error) {
		return DoWithRetry(ctx, fn, cfg)
	})
}

func WithResilienceAs[T any](ctx context.Context, cb *CircuitBreaker[any], cfg *config.RetryConfig, fn func(ctx context.Context) (T, error)) (T, error) {
	var zero T
	result, err := cb.Execute(ctx, func(ctx context.Context) (any, error) {
		return DoWithRetry(ctx, fn, cfg)
	})
	if err != nil {
		return zero, err
	}
	v, _ := result.(T)
	return v, nil
}

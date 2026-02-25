package config

import "time"

type WorkerConfig struct {
	ExternalServiceTimeout time.Duration
	DatabaseTimeout        time.Duration

	RetryConfig          *RetryConfig
	CircuitBreakerConfig *CircuitBreakerConfig
}

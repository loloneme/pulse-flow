package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port string `env:"PORT" envDefault:"8080"`

	ExternalServiceTimeout time.Duration `env:"EXTERNAL_SERVICE_TIMEOUT" envDefault:"5s"`

	WorkerConfig         *WorkerConfig
	RetryConfig          *RetryConfig
	CircuitBreakerConfig *CircuitBreakerConfig
}

type RetryConfig struct {
	Attempts int           `env:"RETRY_ATTEMPTS" envDefault:"3"`
	Delay    time.Duration `env:"RETRY_DELAY" envDefault:"1s"`
	MaxDelay time.Duration `env:"RETRY_MAX_DELAY" envDefault:"10s"`
}

type CircuitBreakerConfig struct {
	MaxFailures int           `env:"CIRCUIT_BREAKER_MAX_FAILURES" envDefault:"3"`
	OpenTimeout time.Duration `env:"CIRCUIT_BREAKER_OPEN_TIMEOUT" envDefault:"10s"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	var retryConfig RetryConfig
	if err := env.Parse(&retryConfig); err != nil {
		return nil, err
	}

	var circuitBreakerConfig CircuitBreakerConfig
	if err := env.Parse(&circuitBreakerConfig); err != nil {
		return nil, err
	}

	cfg.WorkerConfig = &WorkerConfig{
		ExternalServiceTimeout: cfg.ExternalServiceTimeout,
		RetryConfig:            &retryConfig,
		CircuitBreakerConfig:   &circuitBreakerConfig,
	}

	cfg.RetryConfig = &retryConfig
	cfg.CircuitBreakerConfig = &circuitBreakerConfig
	return &cfg, nil
}

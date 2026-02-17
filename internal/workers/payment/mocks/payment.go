package mocks

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	domain "github.com/loloneme/pulse-flow/internal/domain/order"
)

var paymentErrors = []string{
	"insufficient funds on card",
	"card has expired",
	"card limit exceeded",
	"payment rejected due to fraud suspicion",
	"invalid card details",
	"payment declined by issuing bank",
	"3D-Secure verification failed",
}

type MockPaymentService struct {
	SuccessRate      float64
	AvgDelay         time.Duration
	NetworkErrorRate float64
}

func NewMockPaymentService() *MockPaymentService {
	return &MockPaymentService{
		SuccessRate:      0.9,
		AvgDelay:         500 * time.Millisecond,
		NetworkErrorRate: 0.05,
	}
}

func (m *MockPaymentService) ProcessPayment(ctx context.Context, order *domain.Order) error {
	delay := time.Duration(rand.Intn(int(m.AvgDelay * 2)))
	select {
	case <-ctx.Done():
		return fmt.Errorf("payment timeout: %w", ctx.Err())
	case <-time.After(delay):
	}

	if rand.Float64() < m.NetworkErrorRate {
		return errors.New("payment gateway network error: connection timeout")
	}

	if rand.Float64() < m.SuccessRate {
		return nil
	}

	randomError := paymentErrors[rand.Intn(len(paymentErrors))]
	return fmt.Errorf("%s", randomError)
}

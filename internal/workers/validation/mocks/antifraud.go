package mocks

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
	domain "github.com/loloneme/pulse-flow/internal/domain/order"
)

type MockAntiFraudService struct {
	SuccessRate      float64
	AvgDelay         time.Duration
	NetworkErrorRate float64
}

func NewMockAntiFraudService() *MockAntiFraudService {
	return &MockAntiFraudService{
		SuccessRate:      0.85,
		AvgDelay:         600 * time.Millisecond,
		NetworkErrorRate: 0.03,
	}
}

func (m *MockAntiFraudService) CheckUserCreditLimit(ctx context.Context, userID uuid.UUID) (bool, error) {
	delay := time.Duration(rand.Intn(int(m.AvgDelay * 2)))
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(delay):
	}

	if rand.Float64() < m.NetworkErrorRate {
		return false, errors.New("antifraud service is not available")
	}

	hasCreditLimit := rand.Float64() < m.SuccessRate
	if !hasCreditLimit {
		return false, errors.New("user credit limit exceeded")
	}
	return true, nil
}

func (m *MockAntiFraudService) CheckOrder(ctx context.Context, order *domain.Order) (bool, string, error) {
	delay := time.Duration(rand.Intn(int(m.AvgDelay * 2)))
	select {
	case <-ctx.Done():
		return false, "", ctx.Err()
	case <-time.After(delay):
	}

	passed := true
	reason := ""

	if order.Amount > 10000 {
		passed = rand.Float64() < 0.5
		if !passed {
			reason = "suspicious order amount"
		}
	} else {
		passed = rand.Float64() < m.SuccessRate
		if !passed {
			reasons := []string{
				"suspiciouse user behaviour",
				"unusual order frequency",
				"user has a history of suspicious activity",
			}
			reason = reasons[rand.Intn(len(reasons))]
		}
	}

	return passed, reason, nil
}

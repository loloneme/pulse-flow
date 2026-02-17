package mocks

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type MockUserService struct {
	SuccessRate      float64
	AvgDelay         time.Duration
	NetworkErrorRate float64
}

func NewMockUserService() *MockUserService {
	return &MockUserService{
		SuccessRate:      0.95,
		AvgDelay:         100 * time.Millisecond,
		NetworkErrorRate: 0.02,
	}
}

func (m *MockUserService) CheckUserStatus(ctx context.Context, userID uuid.UUID) (bool, error) {
	delay := time.Duration(rand.Intn(int(m.AvgDelay * 2)))
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(delay):
	}

	if rand.Float64() < m.NetworkErrorRate {
		return false, errors.New("user service is not available")
	}

	active := rand.Float64() < m.SuccessRate
	return active, nil
}

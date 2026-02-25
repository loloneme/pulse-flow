package mocks

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type MockWarehouseService struct {
	SuccessRate      float64
	AvgDelay         time.Duration
	NetworkErrorRate float64
}

func NewMockWarehouseService() *MockWarehouseService {
	return &MockWarehouseService{
		SuccessRate:      0.9,
		AvgDelay:         500 * time.Millisecond,
		NetworkErrorRate: 0.05,
	}
}

func (m *MockWarehouseService) CheckProductAvailability(ctx context.Context, productID uuid.UUID, amount int) (bool, error) {
	delay := time.Duration(rand.Intn(int(m.AvgDelay * 2)))

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(delay):
	}

	if rand.Float64() < m.NetworkErrorRate {
		return false, errors.New("warehouse service is not available")
	}

	hasStock := rand.Float64() < m.SuccessRate
	return hasStock, errors.New("out of stock")
}

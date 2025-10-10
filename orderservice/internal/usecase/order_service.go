package usecase

import (
	"context"
	"time"

	"github.com/gostratum/examples/orderservice/internal/domain"
	"github.com/gostratum/examples/orderservice/internal/ports"
)

// OrderService handles order business logic
type OrderService struct {
	repo ports.OrderRepository
}

// NewOrderService creates a new order service with repository injection
func NewOrderService(repo ports.OrderRepository) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(ctx context.Context, userID string, items []domain.Item) (*domain.Order, error) {
	// Apply context deadline
	ctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer cancel()

	order := domain.NewOrder(userID)
	for _, item := range items {
		if err := order.AddItem(item); err != nil {
			return nil, ErrInvalid
		}
	}

	if err := order.Validate(); err != nil {
		return nil, ErrInvalid
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return nil, ErrUnavailable
	}

	return order, nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	// Apply context deadline
	ctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer cancel()

	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err // Repository already maps to appropriate usecase errors
	}

	return order, nil
}

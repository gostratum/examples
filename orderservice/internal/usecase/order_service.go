package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/gostratum/examples/orderservice/internal/domain"
)

// OrderService handles order business logic
type OrderService struct {
	repo OrderRepository
}

// NewOrderService creates a new order service with repository injection
func NewOrderService(repo OrderRepository) *OrderService {
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
		return nil, s.translateError(err)
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
		return nil, s.translateError(err)
	}

	return order, nil
}

// translateError converts repository/domain errors to usecase errors
func (s *OrderService) translateError(err error) error {
	// Domain errors pass through
	if errors.Is(err, domain.ErrNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, domain.ErrConflict) {
		return ErrConflict
	}
	if errors.Is(err, domain.ErrInvalidInput) {
		return ErrInvalid
	}

	// All other errors are infrastructure/availability issues
	return ErrUnavailable
}

package ports

import (
	"context"

	"github.com/gostratum/examples/orderservice/internal/domain"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Save(ctx context.Context, u *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
}

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	Save(ctx context.Context, o *domain.Order) error
	FindByID(ctx context.Context, id string) (*domain.Order, error)
}

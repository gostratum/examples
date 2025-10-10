package usecase

import (
	"context"
	"time"

	"github.com/gostratum/examples/orderservice/internal/domain"
	"github.com/gostratum/examples/orderservice/internal/ports"
)

// UserService handles user business logic
type UserService struct {
	repo ports.UserRepository
}

// NewUserService creates a new user service with repository injection
func NewUserService(repo ports.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, name, email string) (*domain.User, error) {
	// Apply context deadline
	ctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer cancel()

	user := domain.NewUser(name, email)

	if err := user.Validate(); err != nil {
		return nil, ErrInvalid
	}

	if err := s.repo.Save(ctx, user); err != nil {
		return nil, ErrUnavailable
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	// Apply context deadline
	ctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer cancel()

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err // Repository already maps to appropriate usecase errors
	}

	return user, nil
}

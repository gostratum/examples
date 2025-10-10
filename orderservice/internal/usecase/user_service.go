package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/gostratum/examples/orderservice/internal/domain"
)

// UserService handles user business logic
type UserService struct {
	repo UserRepository
}

// NewUserService creates a new user service with repository injection
func NewUserService(repo UserRepository) *UserService {
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
		// Translate errors from repository layer
		return nil, s.translateError(err)
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
		return nil, s.translateError(err)
	}

	return user, nil
}

// translateError converts repository/domain errors to usecase errors
func (s *UserService) translateError(err error) error {
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

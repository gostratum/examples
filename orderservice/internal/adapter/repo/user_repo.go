package repo

import (
	"context"

	"github.com/gostratum/examples/orderservice/internal/domain"
	"github.com/gostratum/examples/orderservice/internal/ports"
	"github.com/gostratum/examples/orderservice/internal/usecase"
	"gorm.io/gorm"
)

// UserRepo implements the UserRepository interface using GORM
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo creates a new GORM-based user repository
func NewUserRepo(db *gorm.DB) ports.UserRepository {
	return &UserRepo{db: db}
}

// Save stores a user in the database
func (r *UserRepo) Save(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		return usecase.ErrInvalid
	}

	// Convert domain model to GORM entity
	var entity UserEntity
	entity.FromDomain(user)

	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		// Check for unique constraint violation (duplicate email)
		if gorm.ErrDuplicatedKey == err {
			return usecase.ErrInvalid
		}
		return usecase.ErrUnavailable
	}

	// Update domain model with generated values
	*user = *entity.ToDomain()
	return nil
}

// FindByID retrieves a user by their ID
func (r *UserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var entity UserEntity

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, usecase.ErrNotFound
		}
		return nil, usecase.ErrUnavailable
	}

	return entity.ToDomain(), nil
}

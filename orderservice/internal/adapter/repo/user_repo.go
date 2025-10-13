package repo

import (
	"context"
	"errors"

	"github.com/gostratum/examples/orderservice/internal/domain"
	"github.com/gostratum/examples/orderservice/internal/usecase"
	"gorm.io/gorm"
)

// UserRepo implements the UserRepository interface using GORM
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo creates a new GORM-based user repository
func NewUserRepo(db *gorm.DB) usecase.UserRepository {
	return &UserRepo{db: db}
}

// Save stores a user in the database
func (r *UserRepo) Save(ctx context.Context, user *domain.User) error {
	// Domain validation happens in use case layer, not here
	// Adapter just handles persistence

	// Convert domain model to GORM entity
	var entity UserEntity
	entity.FromDomain(user)

	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		// Check for unique constraint violation (duplicate email)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrConflict
		}
		// Return raw error - use case layer will translate to ErrUnavailable
		return err
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		// Return raw error - use case layer will translate to ErrUnavailable
		return nil, err
	}

	return entity.ToDomain(), nil
}

// Update modifies an existing user in the database
func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	// Convert domain model to GORM entity
	var entity UserEntity
	entity.FromDomain(user)

	result := r.db.WithContext(ctx).Where("id = ?", user.ID).Updates(&entity)
	if result.Error != nil {
		// Check for unique constraint violation (duplicate email)
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domain.ErrConflict
		}
		// Return raw error - use case layer will translate to ErrUnavailable
		return result.Error
	}

	// Check if the record was found and updated
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

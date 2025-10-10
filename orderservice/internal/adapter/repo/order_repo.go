package repo

import (
	"context"
	"errors"

	"github.com/gostratum/examples/orderservice/internal/domain"
	"github.com/gostratum/examples/orderservice/internal/usecase"
	"gorm.io/gorm"
)

// OrderRepo implements the OrderRepository interface using GORM
type OrderRepo struct {
	db *gorm.DB
}

// NewOrderRepo creates a new GORM-based order repository
func NewOrderRepo(db *gorm.DB) usecase.OrderRepository {
	return &OrderRepo{db: db}
}

// Save stores an order in the database
func (r *OrderRepo) Save(ctx context.Context, order *domain.Order) error {
	// Domain validation happens in use case layer, not here
	// Adapter just handles persistence

	// Convert domain model to GORM entity
	var entity OrderEntity
	entity.FromDomain(order)

	// Use transaction to ensure order and items are saved together
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the order
		if err := tx.Create(&entity).Error; err != nil {
			return err
		}

		// Update domain model with generated values
		*order = *entity.ToDomain()
		return nil
	})

	// Return raw error - use case layer will translate to ErrUnavailable if needed
	return err
}

// FindByID retrieves an order by its ID, including all items
func (r *OrderRepo) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	var entity OrderEntity

	err := r.db.WithContext(ctx).Preload("Items").Where("id = ?", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		// Return raw error - use case layer will translate to ErrUnavailable
		return nil, err
	}

	return entity.ToDomain(), nil
}

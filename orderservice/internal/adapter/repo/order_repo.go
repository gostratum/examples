package repo

import (
	"context"

	"github.com/gostratum/examples/orderservice/internal/domain"
	"github.com/gostratum/examples/orderservice/internal/ports"
	"github.com/gostratum/examples/orderservice/internal/usecase"
	"gorm.io/gorm"
)

// OrderRepo implements the OrderRepository interface using GORM
type OrderRepo struct {
	db *gorm.DB
}

// NewOrderRepo creates a new GORM-based order repository
func NewOrderRepo(db *gorm.DB) ports.OrderRepository {
	return &OrderRepo{db: db}
}

// Save stores an order in the database
func (r *OrderRepo) Save(ctx context.Context, order *domain.Order) error {
	if err := order.Validate(); err != nil {
		return usecase.ErrInvalid
	}

	// Convert domain model to GORM entity
	var entity OrderEntity
	entity.FromDomain(order)

	// Use transaction to ensure order and items are saved together
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the order
		if err := tx.Create(&entity).Error; err != nil {
			return usecase.ErrUnavailable
		}

		// Update domain model with generated values
		*order = *entity.ToDomain()
		return nil
	})
}

// FindByID retrieves an order by its ID, including all items
func (r *OrderRepo) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	var entity OrderEntity

	err := r.db.WithContext(ctx).Preload("Items").Where("id = ?", id).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, usecase.ErrNotFound
		}
		return nil, usecase.ErrUnavailable
	}

	return entity.ToDomain(), nil
}

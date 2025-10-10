package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Item represents an item in an order
// This is a pure domain model without infrastructure concerns
type Item struct {
	ID      uint
	OrderID string
	SKU     string
	Qty     int
	Price   float64
}

// Order represents an order in the system
// This is a pure domain model without infrastructure concerns
type Order struct {
	ID        string
	UserID    string
	Items     []Item
	Status    string
	Total     float64
	CreatedAt time.Time
}

// NewOrder creates a new order with a generated ID
func NewOrder(userID string) *Order {
	return &Order{
		ID:        uuid.New().String(),
		UserID:    userID,
		Status:    "pending",
		CreatedAt: time.Now(),
		Items:     []Item{},
	}
}

// AddItem adds an item to the order with validation
func (o *Order) AddItem(item Item) error {
	if item.SKU == "" {
		return errors.New("item SKU is required")
	}
	if item.Qty <= 0 {
		return errors.New("item quantity must be positive")
	}
	if item.Price < 0 {
		return errors.New("item price cannot be negative")
	}

	o.Items = append(o.Items, item)
	// Recalculate total
	o.Total = 0
	for _, i := range o.Items {
		o.Total += i.Price * float64(i.Qty)
	}
	return nil
}

// Validate performs basic validation on order fields
// Item-level validation is already done in AddItem(), so this only validates order-level rules
func (o *Order) Validate() error {
	if o.UserID == "" {
		return errors.New("user_id is required")
	}

	if len(o.Items) == 0 {
		return errors.New("order must have at least one item")
	}

	// Note: Individual item validation (SKU, Qty, Price) happens in AddItem()
	// No need to re-validate here - this avoids redundant validation

	return nil
}

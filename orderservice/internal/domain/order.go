package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Item represents an item in an order
type Item struct {
	ID      uint    `json:"id"`
	OrderID string  `json:"order_id"`
	SKU     string  `json:"sku"`
	Qty     int     `json:"qty"`
	Price   float64 `json:"price"`
}

// Order represents an order in the system
type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Items     []Item    `json:"items"`
	Status    string    `json:"status"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
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
func (o *Order) Validate() error {
	if o.UserID == "" {
		return errors.New("user_id is required")
	}

	if len(o.Items) == 0 {
		return errors.New("order must have at least one item")
	}

	for _, item := range o.Items {
		if item.SKU == "" {
			return errors.New("all items must have a SKU")
		}
		if item.Qty <= 0 {
			return errors.New("all items must have positive quantity")
		}
		if item.Price < 0 {
			return errors.New("all items must have non-negative price")
		}
	}

	return nil
}

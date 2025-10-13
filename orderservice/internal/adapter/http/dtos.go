package http

import (
	"time"

	"github.com/gostratum/examples/orderservice/internal/domain"
)

// UserResponse is the HTTP DTO for user data
// This struct handles JSON serialization concerns for the HTTP layer
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

// FromDomainUser converts a domain.User to UserResponse DTO
func FromDomainUser(user *domain.User) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt,
	}
}

// ItemResponse is the HTTP DTO for item data
type ItemResponse struct {
	ID      uint    `json:"id"`
	OrderID string  `json:"order_id"`
	SKU     string  `json:"sku"`
	Qty     int     `json:"qty"`
	Price   float64 `json:"price"`
}

// FromDomainItem converts a domain.Item to ItemResponse DTO
func FromDomainItem(item domain.Item) ItemResponse {
	return ItemResponse{
		ID:      item.ID,
		OrderID: item.OrderID,
		SKU:     item.SKU,
		Qty:     item.Qty,
		Price:   item.Price,
	}
}

// OrderResponse is the HTTP DTO for order data
type OrderResponse struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Items     []ItemResponse `json:"items"`
	Status    string         `json:"status"`
	Total     float64        `json:"total"`
	CreatedAt time.Time      `json:"created_at"`
}

// FromDomainOrder converts a domain.Order to OrderResponse DTO
func FromDomainOrder(order *domain.Order) *OrderResponse {
	if order == nil {
		return nil
	}

	items := make([]ItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = FromDomainItem(item)
	}

	return &OrderResponse{
		ID:        order.ID,
		UserID:    order.UserID,
		Items:     items,
		Status:    order.Status,
		Total:     order.Total,
		CreatedAt: order.CreatedAt,
	}
}

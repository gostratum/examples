package domain

import (
	"testing"
	"time"
)

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "valid user",
			user: User{
				ID:        "123",
				Name:      "John Doe",
				Email:     "john@example.com",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty name",
			user: User{
				ID:        "123",
				Name:      "",
				Email:     "john@example.com",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty email",
			user: User{
				ID:        "123",
				Name:      "John Doe",
				Email:     "",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			user: User{
				ID:        "123",
				Name:      "John Doe",
				Email:     "invalid-email",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrderValidate(t *testing.T) {
	tests := []struct {
		name    string
		order   Order
		wantErr bool
	}{
		{
			name: "valid order",
			order: Order{
				ID:     "123",
				UserID: "user123",
				Items: []Item{
					{SKU: "SKU1", Qty: 1, Price: 10.0},
				},
				Status:    "pending",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty user id",
			order: Order{
				ID: "123",
				Items: []Item{
					{SKU: "SKU1", Qty: 1, Price: 10.0},
				},
				Status:    "pending",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "no items",
			order: Order{
				ID:        "123",
				UserID:    "user123",
				Items:     []Item{},
				Status:    "pending",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		// Note: Item-level validation (negative price, empty SKU, etc.) is tested in AddItem()
		// Validate() only checks order-level rules: UserID exists and has items
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Order.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestOrderAddItem tests the AddItem method which validates individual items
func TestOrderAddItem(t *testing.T) {
	tests := []struct {
		name    string
		order   *Order
		item    Item
		wantErr bool
	}{
		{
			name:    "valid item",
			order:   NewOrder("user123"),
			item:    Item{SKU: "SKU1", Qty: 1, Price: 10.0},
			wantErr: false,
		},
		{
			name:    "empty SKU",
			order:   NewOrder("user123"),
			item:    Item{SKU: "", Qty: 1, Price: 10.0},
			wantErr: true,
		},
		{
			name:    "zero quantity",
			order:   NewOrder("user123"),
			item:    Item{SKU: "SKU1", Qty: 0, Price: 10.0},
			wantErr: true,
		},
		{
			name:    "negative quantity",
			order:   NewOrder("user123"),
			item:    Item{SKU: "SKU1", Qty: -1, Price: 10.0},
			wantErr: true,
		},
		{
			name:    "negative price",
			order:   NewOrder("user123"),
			item:    Item{SKU: "SKU1", Qty: 1, Price: -10.0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.AddItem(tt.item)
			if (err != nil) != tt.wantErr {
				t.Errorf("Order.AddItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrderTotal(t *testing.T) {
	order := Order{}

	// Add items using AddItem method which calculates total
	err := order.AddItem(Item{SKU: "SKU1", Qty: 2, Price: 10.0})
	if err != nil {
		t.Fatalf("Failed to add item: %v", err)
	}
	err = order.AddItem(Item{SKU: "SKU2", Qty: 1, Price: 15.0})
	if err != nil {
		t.Fatalf("Failed to add item: %v", err)
	}

	expected := 35.0 // (2 * 10.0) + (1 * 15.0)
	actual := order.Total

	if actual != expected {
		t.Errorf("Order.Total = %v, want %v", actual, expected)
	}
}

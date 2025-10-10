package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gostratum/examples/orderservice/internal/domain"
)

// MockOrderRepository implements ports.OrderRepository for testing
type MockOrderRepository struct {
	orders    map[string]*domain.Order
	saveError error
	findError error
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		orders: make(map[string]*domain.Order),
	}
}

func (m *MockOrderRepository) Save(ctx context.Context, o *domain.Order) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.orders[o.ID] = o
	return nil
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	if m.findError != nil {
		return nil, m.findError
	}
	order, exists := m.orders[id]
	if !exists {
		return nil, errors.New("not found")
	}
	return order, nil
}

func (m *MockOrderRepository) SetSaveError(err error) {
	m.saveError = err
}

func (m *MockOrderRepository) SetFindError(err error) {
	m.findError = err
}

func TestCreateOrder(t *testing.T) {
	validItems := []domain.Item{
		{SKU: "SKU1", Qty: 2, Price: 10.0},
		{SKU: "SKU2", Qty: 1, Price: 15.0},
	}

	tests := []struct {
		name      string
		userID    string
		items     []domain.Item
		saveError error
		wantErr   error
	}{
		{
			name:    "valid order creation",
			userID:  "user123",
			items:   validItems,
			wantErr: nil,
		},
		{
			name:    "empty user id should return invalid error",
			userID:  "",
			items:   validItems,
			wantErr: ErrInvalid,
		},
		{
			name:    "empty items should return invalid error",
			userID:  "user123",
			items:   []domain.Item{},
			wantErr: ErrInvalid,
		},
		{
			name:   "items with negative price should return invalid error",
			userID: "user123",
			items: []domain.Item{
				{SKU: "SKU1", Qty: 1, Price: -10.0},
			},
			wantErr: ErrInvalid,
		},
		{
			name:   "items with zero quantity should return invalid error",
			userID: "user123",
			items: []domain.Item{
				{SKU: "SKU1", Qty: 0, Price: 10.0},
			},
			wantErr: ErrInvalid,
		},
		{
			name:   "items with empty SKU should return invalid error",
			userID: "user123",
			items: []domain.Item{
				{SKU: "", Qty: 1, Price: 10.0},
			},
			wantErr: ErrInvalid,
		},
		{
			name:      "repository error should return unavailable error",
			userID:    "user123",
			items:     validItems,
			saveError: errors.New("database connection failed"),
			wantErr:   ErrUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockOrderRepository()
			if tt.saveError != nil {
				repo.SetSaveError(tt.saveError)
			}

			ctx := context.Background()
			service := NewOrderService(repo)
			order, err := service.CreateOrder(ctx, tt.userID, tt.items)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("CreateOrder() error = %v, wantErr %v", err, tt.wantErr)
				}
				if order != nil {
					t.Errorf("CreateOrder() should return nil order on error, got %v", order)
				}
			} else {
				if err != nil {
					t.Errorf("CreateOrder() unexpected error = %v", err)
				}
				if order == nil {
					t.Errorf("CreateOrder() should return order on success")
				} else {
					if order.UserID != tt.userID {
						t.Errorf("CreateOrder() order.UserID = %v, want %v", order.UserID, tt.userID)
					}
					if len(order.Items) != len(tt.items) {
						t.Errorf("CreateOrder() order.Items length = %v, want %v", len(order.Items), len(tt.items))
					}
					if order.Status != "pending" {
						t.Errorf("CreateOrder() order.Status = %v, want pending", order.Status)
					}
					if order.ID == "" {
						t.Errorf("CreateOrder() order.ID should not be empty")
					}
				}
			}
		})
	}
}

func TestGetOrder(t *testing.T) {
	tests := []struct {
		name       string
		orderID    string
		setupOrder *domain.Order
		findError  error
		wantErr    error
	}{
		{
			name:    "existing order should be returned",
			orderID: "test-order-id",
			setupOrder: &domain.Order{
				ID:     "test-order-id",
				UserID: "user123",
				Items: []domain.Item{
					{SKU: "SKU1", Qty: 2, Price: 10.0},
				},
				Status:    "pending",
				CreatedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name:      "non-existing order should return not found error",
			orderID:   "non-existing",
			findError: ErrNotFound,
			wantErr:   ErrNotFound,
		},
		{
			name:      "repository error should return unavailable error",
			orderID:   "test-order-id",
			findError: ErrUnavailable,
			wantErr:   ErrUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockOrderRepository()

			if tt.setupOrder != nil {
				repo.orders[tt.setupOrder.ID] = tt.setupOrder
			}

			if tt.findError != nil {
				repo.SetFindError(tt.findError)
			}

			ctx := context.Background()
			service := NewOrderService(repo)
			order, err := service.GetOrder(ctx, tt.orderID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetOrder() error = %v, wantErr %v", err, tt.wantErr)
				}
				if order != nil {
					t.Errorf("GetOrder() should return nil order on error, got %v", order)
				}
			} else {
				if err != nil {
					t.Errorf("GetOrder() unexpected error = %v", err)
				}
				if order == nil {
					t.Errorf("GetOrder() should return order on success")
				} else {
					if order.ID != tt.setupOrder.ID {
						t.Errorf("GetOrder() order.ID = %v, want %v", order.ID, tt.setupOrder.ID)
					}
					if order.UserID != tt.setupOrder.UserID {
						t.Errorf("GetOrder() order.UserID = %v, want %v", order.UserID, tt.setupOrder.UserID)
					}
					if order.Status != tt.setupOrder.Status {
						t.Errorf("GetOrder() order.Status = %v, want %v", order.Status, tt.setupOrder.Status)
					}
				}
			}
		})
	}
}

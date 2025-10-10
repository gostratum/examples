package repo

import (
	"context"
	"os"
	"testing"

	"github.com/gostratum/examples/orderservice/internal/domain"
	"github.com/gostratum/examples/orderservice/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	// Use a unique database name for each test to avoid conflicts
	dbName := t.Name() + ".db"
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	require.NoError(t, err)

	// Clean up the database file after test
	t.Cleanup(func() {
		os.Remove(dbName)
	})

	// Create tables manually for SQLite compatibility
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE orders (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			total REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		CREATE TABLE items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id TEXT NOT NULL,
			sku TEXT NOT NULL,
			qty INTEGER NOT NULL,
			price REAL NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
		);
	`).Error
	require.NoError(t, err)

	return db
}

// TestUserRepo_Save tests user repository save operations
func TestUserRepo_Save(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	ctx := context.Background()

	t.Run("save valid user", func(t *testing.T) {
		user := &domain.User{
			Name:  "John Doe",
			Email: "john@example.com",
		}

		err := repo.Save(ctx, user)
		assert.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.NotZero(t, user.CreatedAt)
	})

	t.Run("save user with invalid data", func(t *testing.T) {
		user := &domain.User{
			Name:  "", // Empty name - validation should happen in use case, not repo
			Email: "invalid@example.com",
		}

		// Repository layer doesn't validate, just persists
		// This should succeed at repo level; validation happens in use case
		err := repo.Save(ctx, user)
		assert.NoError(t, err) // Repository accepts any data
	})

	t.Run("save user with duplicate email", func(t *testing.T) {
		// First user
		user1 := &domain.User{
			Name:  "Jane Doe",
			Email: "jane@example.com",
		}
		err := repo.Save(ctx, user1)
		require.NoError(t, err)

		// Second user with same email
		user2 := &domain.User{
			Name:  "Jane Smith",
			Email: "jane@example.com", // Duplicate email
		}
		err = repo.Save(ctx, user2)
		// SQLite uses different error type, so we'll check for any error for now
		assert.Error(t, err)
		// TODO: Check specific error type once we determine SQLite's duplicate key error
	})
}

// TestUserRepo_FindByID tests user repository find operations
func TestUserRepo_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepo(db)

	ctx := context.Background()

	// Create a test user
	user := &domain.User{
		Name:  "Test User",
		Email: "test@example.com",
	}
	err := repo.Save(ctx, user)
	require.NoError(t, err)
	userID := user.ID

	t.Run("find existing user", func(t *testing.T) {
		found, err := repo.FindByID(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, userID, found.ID)
		assert.Equal(t, "Test User", found.Name)
		assert.Equal(t, "test@example.com", found.Email)
	})

	t.Run("find non-existing user", func(t *testing.T) {
		found, err := repo.FindByID(ctx, "non-existing-id")
		assert.Equal(t, usecase.ErrNotFound, err)
		assert.Nil(t, found)
	})
}

// TestOrderRepo_Save tests order repository save operations
func TestOrderRepo_Save(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	orderRepo := NewOrderRepo(db)

	ctx := context.Background()

	// Create a test user first
	user := &domain.User{
		Name:  "Order User",
		Email: "order@example.com",
	}
	err := userRepo.Save(ctx, user)
	require.NoError(t, err)

	t.Run("save valid order", func(t *testing.T) {
		order := &domain.Order{
			UserID: user.ID,
			Items: []domain.Item{
				{SKU: "ITEM001", Qty: 2, Price: 10.99},
				{SKU: "ITEM002", Qty: 1, Price: 25.50},
			},
		}

		err := orderRepo.Save(ctx, order)
		assert.NoError(t, err)
		assert.NotEmpty(t, order.ID)
		assert.Equal(t, "pending", order.Status)
		assert.NotZero(t, order.CreatedAt)
		assert.Len(t, order.Items, 2)
	})

	t.Run("save order with invalid data", func(t *testing.T) {
		order := &domain.Order{
			UserID: "", // Empty user ID - validation should happen in use case
			Items:  []domain.Item{},
		}

		// Repository layer doesn't validate, just persists
		// This should succeed at repo level; validation happens in use case
		err := orderRepo.Save(ctx, order)
		assert.NoError(t, err) // Repository accepts any data
	})

	t.Run("save order with invalid items", func(t *testing.T) {
		order := &domain.Order{
			UserID: user.ID,
			Items: []domain.Item{
				{SKU: "", Qty: 1, Price: 10.00}, // Empty SKU - validation in use case
			},
		}

		// Repository layer doesn't validate, just persists
		err := orderRepo.Save(ctx, order)
		assert.NoError(t, err) // Repository accepts any data
	})
}

// TestOrderRepo_FindByID tests order repository find operations
func TestOrderRepo_FindByID(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	orderRepo := NewOrderRepo(db)

	ctx := context.Background()

	// Create a test user
	user := &domain.User{
		Name:  "Order Test User",
		Email: "ordertest@example.com",
	}
	err := userRepo.Save(ctx, user)
	require.NoError(t, err)

	// Create a test order
	order := &domain.Order{
		UserID: user.ID,
		Items: []domain.Item{
			{SKU: "TEST001", Qty: 3, Price: 15.99},
		},
	}
	err = orderRepo.Save(ctx, order)
	require.NoError(t, err)
	orderID := order.ID

	t.Run("find existing order with items", func(t *testing.T) {
		found, err := orderRepo.FindByID(ctx, orderID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, orderID, found.ID)
		assert.Equal(t, user.ID, found.UserID)
		assert.Equal(t, "pending", found.Status)
		assert.Len(t, found.Items, 1)
		assert.Equal(t, "TEST001", found.Items[0].SKU)
		assert.Equal(t, 3, found.Items[0].Qty)
		assert.Equal(t, 15.99, found.Items[0].Price)
	})

	t.Run("find non-existing order", func(t *testing.T) {
		found, err := orderRepo.FindByID(ctx, "non-existing-id")
		assert.Equal(t, usecase.ErrNotFound, err)
		assert.Nil(t, found)
	})
}

// TestRepositoryIntegration tests the complete flow between repositories
func TestRepositoryIntegration(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepo(db)
	orderRepo := NewOrderRepo(db)

	ctx := context.Background()

	// Create user
	user := &domain.User{
		Name:  "Integration User",
		Email: "integration@example.com",
	}
	err := userRepo.Save(ctx, user)
	require.NoError(t, err)

	// Create order for that user
	order := &domain.Order{
		UserID: user.ID,
		Items: []domain.Item{
			{SKU: "INT001", Qty: 1, Price: 99.99},
			{SKU: "INT002", Qty: 2, Price: 49.50},
		},
	}
	// Calculate total manually for the test
	order.Total = 99.99 + (2 * 49.50) // 198.99

	err = orderRepo.Save(ctx, order)
	require.NoError(t, err)

	// Verify order was saved with correct total
	foundOrder, err := orderRepo.FindByID(ctx, order.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundOrder.UserID)
	assert.Len(t, foundOrder.Items, 2)

	// Calculate expected total: (1 * 99.99) + (2 * 49.50) = 99.99 + 99.00 = 198.99
	expectedTotal := 99.99 + (2 * 49.50)
	actualTotal := foundOrder.Total
	assert.Equal(t, expectedTotal, actualTotal)
}

package orderservice_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gostratum/core/logx"
	httpAdapter "github.com/gostratum/examples/orderservice/internal/adapter/http"
	"github.com/gostratum/examples/orderservice/internal/adapter/repo"
	"github.com/gostratum/examples/orderservice/internal/usecase"
)

func setupTestServer(t *testing.T) *gin.Engine {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables manually for SQLite compatibility (same as repo tests)
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			avatar_url TEXT,
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

	// Create repositories
	userRepo := repo.NewUserRepo(db)
	orderRepo := repo.NewOrderRepo(db)

	// Create services
	userService := usecase.NewUserService(userRepo)
	orderService := usecase.NewOrderService(orderRepo)

	// Create logger
	logger := logx.NewNoopLogger()

	// Create handlers
	userHandler := httpAdapter.NewUserHandler(userService, nil, logger)
	orderHandler := httpAdapter.NewOrderHandler(orderService, logger)

	// Create router
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Setup routes
	api := router.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
		}

		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("/:id", orderHandler.GetOrder)
		}
	}

	return router
}

func TestEndToEnd_UserLifecycle(t *testing.T) {
	router := setupTestServer(t)

	t.Run("create and retrieve user", func(t *testing.T) {
		// Create user request
		userReq := map[string]any{
			"name":  "John Doe",
			"email": "john.doe@example.com",
		}
		reqBody, _ := json.Marshal(userReq)

		// Create user
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var createEnvelope map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &createEnvelope)
		require.NoError(t, err)

		// Extract data from envelope
		assert.True(t, createEnvelope["ok"].(bool))
		createResp := createEnvelope["data"].(map[string]any)

		userID := createResp["id"].(string)
		assert.Equal(t, "John Doe", createResp["name"])
		assert.Equal(t, "john.doe@example.com", createResp["email"])

		// Retrieve user
		req, _ = http.NewRequest("GET", "/api/v1/users/"+userID, nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var getEnvelope map[string]any
		err = json.Unmarshal(w.Body.Bytes(), &getEnvelope)
		require.NoError(t, err)

		// Extract data from envelope
		assert.True(t, getEnvelope["ok"].(bool))
		getResp := getEnvelope["data"].(map[string]any)

		assert.Equal(t, userID, getResp["id"])
		assert.Equal(t, "John Doe", getResp["name"])
		assert.Equal(t, "john.doe@example.com", getResp["email"])
	})
}

func TestEndToEnd_OrderLifecycle(t *testing.T) {
	router := setupTestServer(t)

	t.Run("create and retrieve order", func(t *testing.T) {
		// First create a user
		userReq := map[string]any{
			"name":  "Jane Smith",
			"email": "jane.smith@example.com",
		}
		reqBody, _ := json.Marshal(userReq)

		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		// Create order request using the proper HTTP request format
		orderReq := map[string]any{
			"user_id": "1",
			"items": []map[string]any{
				{"sku": "Laptop", "qty": 1, "price": 1200.00},
				{"sku": "Mouse", "qty": 2, "price": 25.00},
			},
		}
		reqBody, _ = json.Marshal(orderReq)

		// Create order
		req, _ = http.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var createEnvelope map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &createEnvelope)
		require.NoError(t, err)

		// Extract data from envelope
		assert.True(t, createEnvelope["ok"].(bool))
		createResp := createEnvelope["data"].(map[string]any)

		orderID := createResp["id"].(string)
		assert.Equal(t, "1", createResp["user_id"])
		assert.Equal(t, 1250.00, createResp["total"].(float64))

		// Retrieve order
		req, _ = http.NewRequest("GET", "/api/v1/orders/"+orderID, nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var getEnvelope map[string]any
		err = json.Unmarshal(w.Body.Bytes(), &getEnvelope)
		require.NoError(t, err)

		// Extract data from envelope
		assert.True(t, getEnvelope["ok"].(bool))
		getResp := getEnvelope["data"].(map[string]any)

		assert.Equal(t, orderID, getResp["id"])
		assert.Equal(t, "1", getResp["user_id"])
		assert.Equal(t, 1250.00, getResp["total"].(float64))
	})
}

func TestEndToEnd_ErrorHandling(t *testing.T) {
	router := setupTestServer(t)

	t.Run("create user with invalid data", func(t *testing.T) {
		// Invalid email
		userReq := map[string]any{
			"name":  "John Doe",
			"email": "invalid-email",
		}
		reqBody, _ := json.Marshal(userReq)

		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("get non-existent user", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/users/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get non-existent order", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/orders/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

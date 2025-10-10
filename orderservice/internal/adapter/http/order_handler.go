package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gostratum/examples/orderservice/internal/domain"
	"github.com/gostratum/examples/orderservice/internal/usecase"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	service *usecase.OrderService
	log     *zap.Logger
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(service *usecase.OrderService, log *zap.Logger) *OrderHandler {
	return &OrderHandler{service: service, log: log}
}

// CreateOrderRequest represents the request payload for creating an order
type CreateOrderRequest struct {
	UserID string        `json:"user_id" binding:"required"`
	Items  []domain.Item `json:"items" binding:"required"`
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), req.UserID, req.Items)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetOrder handles GET /orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order id is required"})
		return
	}

	order, err := h.service.GetOrder(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, order)
}

// handleError maps usecase errors to HTTP responses
func (h *OrderHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
	case errors.Is(err, usecase.ErrInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case errors.Is(err, usecase.ErrUnavailable):
		c.Header("Retry-After", "2")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service temporarily unavailable"})
	default:
		h.log.Error("unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

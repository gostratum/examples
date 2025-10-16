package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gostratum/core/logx"
	"github.com/gostratum/httpx/responsex"

	"github.com/gostratum/examples/orderservice/internal/domain"
	"github.com/gostratum/examples/orderservice/internal/usecase"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	service *usecase.OrderService
	log     logx.Logger
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(service *usecase.OrderService, log logx.Logger) *OrderHandler {
	return &OrderHandler{service: service, log: log}
}

// ItemRequest represents the request payload for an order item
type ItemRequest struct {
	SKU   string  `json:"sku" binding:"required"`
	Qty   int     `json:"qty" binding:"required"`
	Price float64 `json:"price" binding:"required"`
}

// ToDomain converts ItemRequest to domain.Item
func (r *ItemRequest) ToDomain() domain.Item {
	return domain.Item{
		SKU:   r.SKU,
		Qty:   r.Qty,
		Price: r.Price,
	}
}

// CreateOrderRequest represents the request payload for creating an order
type CreateOrderRequest struct {
	UserID string        `json:"user_id" binding:"required"`
	Items  []ItemRequest `json:"items" binding:"required"`
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responsex.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request payload", nil)
		return
	}

	// Convert request items to domain items
	domainItems := make([]domain.Item, len(req.Items))
	for i, item := range req.Items {
		domainItems[i] = item.ToDomain()
	}

	order, err := h.service.CreateOrder(c.Request.Context(), req.UserID, domainItems)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert domain model to HTTP DTO
	orderResponse := FromDomainOrder(order)
	responsex.Created(c, "", orderResponse)
}

// GetOrder handles GET /orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		responsex.Error(c, http.StatusBadRequest, "MISSING_PARAMETER", "order id is required", nil)
		return
	}

	order, err := h.service.GetOrder(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Convert domain model to HTTP DTO
	orderResponse := FromDomainOrder(order)
	responsex.OK(c, orderResponse, nil)
}

// handleError maps usecase errors to HTTP responses
func (h *OrderHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrNotFound):
		responsex.Error(c, http.StatusNotFound, "ORDER_NOT_FOUND", "order not found", nil)
	case errors.Is(err, usecase.ErrInvalid):
		responsex.Error(c, http.StatusBadRequest, "INVALID_INPUT", "invalid input", nil)
	case errors.Is(err, usecase.ErrUnavailable):
		c.Header("Retry-After", "2")
		responsex.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "service temporarily unavailable", nil)
	default:
		h.log.Error("unexpected error", logx.Err(err))
		responsex.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", nil)
	}
}

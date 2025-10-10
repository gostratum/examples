package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gostratum/httpx/responsex"
	"go.uber.org/zap"

	"github.com/gostratum/examples/orderservice/internal/usecase"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	service *usecase.UserService
	log     *zap.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *usecase.UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{service: service, log: log}
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responsex.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request payload", nil)
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), req.Name, req.Email)
	if err != nil {
		h.handleError(c, err)
		return
	}

	responsex.Created(c, "", user)
}

// GetUser handles GET /users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		responsex.Error(c, http.StatusBadRequest, "MISSING_PARAMETER", "user id is required", nil)
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	responsex.OK(c, user, nil)
}

// handleError maps usecase errors to HTTP responses
func (h *UserHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrNotFound):
		responsex.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found", nil)
	case errors.Is(err, usecase.ErrInvalid):
		responsex.Error(c, http.StatusBadRequest, "INVALID_INPUT", "invalid input", nil)
	case errors.Is(err, usecase.ErrUnavailable):
		c.Header("Retry-After", "2")
		responsex.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "service temporarily unavailable", nil)
	default:
		h.log.Error("unexpected error", zap.Error(err))
		responsex.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", nil)
	}
}

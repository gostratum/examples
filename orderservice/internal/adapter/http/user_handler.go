package http

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gostratum/core/logx"
	"github.com/gostratum/httpx/responsex"
	"github.com/gostratum/storagex"

	"github.com/gostratum/examples/orderservice/internal/usecase"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	service       *usecase.UserService
	storageClient storagex.Storage
	log           logx.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *usecase.UserService, storageClient storagex.Storage, log logx.Logger) *UserHandler {
	return &UserHandler{
		service:       service,
		storageClient: storageClient,
		log:           log,
	}
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

	// Convert domain model to HTTP DTO
	userResponse := FromDomainUser(user)
	responsex.Created(c, "", userResponse)
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

	// Convert domain model to HTTP DTO
	userResponse := FromDomainUser(user)
	responsex.OK(c, userResponse, nil)
}

// UploadAvatar handles POST /users/:id/avatar
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		responsex.Error(c, http.StatusBadRequest, "MISSING_PARAMETER", "user id is required", nil)
		return
	}

	// Get the uploaded file
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		responsex.Error(c, http.StatusBadRequest, "INVALID_FILE", "avatar file is required", nil)
		return
	}
	defer file.Close()

	// Validate file type
	if !h.isValidImageType(header) {
		responsex.Error(c, http.StatusBadRequest, "INVALID_FILE_TYPE", "only image files are allowed", nil)
		return
	}

	// Validate file size (5MB max)
	if header.Size > 5*1024*1024 {
		responsex.Error(c, http.StatusBadRequest, "FILE_TOO_LARGE", "file size exceeds 5MB limit", nil)
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("avatars/%s_%d%s", userID, time.Now().Unix(), ext)

	// Upload to storage
	_, err = h.storageClient.Put(c.Request.Context(), filename, file, &storagex.PutOptions{
		ContentType: header.Header.Get("Content-Type"),
		Overwrite:   true,
	})
	if err != nil {
		h.log.Error("failed to upload avatar", logx.Err(err))
		responsex.Error(c, http.StatusInternalServerError, "UPLOAD_FAILED", "failed to upload avatar", nil)
		return
	}

	// For cloud storage, you might want to use a presigned URL or construct the full S3 URL
	// For now, we'll use the key as the URL - this should be customized based on your deployment
	url := filename

	// Update user avatar in database
	user, err := h.service.UpdateAvatar(c.Request.Context(), userID, url)
	if err != nil {
		h.handleError(c, err)
		return
	}

	userResponse := FromDomainUser(user)
	responsex.OK(c, userResponse, nil)
}

// isValidImageType checks if the uploaded file is a valid image type
func (h *UserHandler) isValidImageType(header *multipart.FileHeader) bool {
	contentType := header.Header.Get("Content-Type")
	validTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, validType := range validTypes {
		if strings.EqualFold(contentType, validType) {
			return true
		}
	}
	return false
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
		h.log.Error("unexpected error", logx.Err(err))
		responsex.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", nil)
	}
}

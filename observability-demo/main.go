package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gostratum/core/logx"
	"github.com/gostratum/dbx"
	"github.com/gostratum/httpx"
	"github.com/gostratum/metricsx"
	"github.com/gostratum/tracingx"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func main() {
	app := fx.New(
		// Core modules
		logx.Module(),

		// Observability modules
		metricsx.Module(),
		tracingx.Module(),

		// Service modules with observability
		httpx.Module(),
		dbx.Module(),

		// Application modules
		fx.Provide(
			NewUserService,
			NewUserHandler,
		),

		// Lifecycle hooks
		fx.Invoke(RegisterRoutes),
		fx.Invoke(SetupDatabase),
	)

	app.Run()
}

// User model for demonstration
type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email" gorm:"uniqueIndex"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserService handles user operations
type UserService struct {
	db     *gorm.DB
	logger logx.Logger
}

func NewUserService(conns dbx.Connections, logger logx.Logger) (*UserService, error) {
	conn, exists := conns["default"]
	if !exists {
		return nil, fmt.Errorf("default database connection not found")
	}
	return &UserService{db: conn, logger: logger}, nil
}

func (s *UserService) CreateUser(ctx context.Context, name, email string) (*User, error) {
	user := &User{
		Name:  name,
		Email: email,
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		s.logger.Error("failed to create user", logx.Err(err))
		return nil, err
	}
	s.logger.Info("user created", logx.Int("id", int(user.ID)), logx.String("email", email))
	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id uint) (*User, error) {
	var user User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		s.logger.Error("failed to get user", logx.Err(err), logx.Int("id", int(id)))
		return nil, err
	}
	return &user, nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]User, error) {
	var users []User
	if err := s.db.WithContext(ctx).Find(&users).Error; err != nil {
		s.logger.Error("failed to list users", logx.Err(err))
		return nil, err
	}
	return users, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uint, name, email string) (*User, error) {
	var user User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}

	user.Name = name
	user.Email = email

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		s.logger.Error("failed to update user", logx.Err(err), logx.Int("id", int(id)))
		return nil, err
	}
	s.logger.Info("user updated", logx.Int("id", int(id)))
	return &user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Delete(&User{}, id).Error; err != nil {
		s.logger.Error("failed to delete user", logx.Err(err), logx.Int("id", int(id)))
		return err
	}
	s.logger.Info("user deleted", logx.Int("id", int(id)))
	return nil
}

// UserHandler handles HTTP requests
type UserHandler struct {
	service *UserService
	logger  logx.Logger
}

func NewUserHandler(service *UserService, logger logx.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, map[string]string{"error": err.Error()})
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), req.Name, req.Email)
	if err != nil {
		c.JSON(500, map[string]string{"error": "failed to create user"})
		return
	}

	c.JSON(201, user)
}

func (h *UserHandler) Get(c *gin.Context) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}

	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(400, map[string]string{"error": err.Error()})
		return
	}

	user, err := h.service.GetUser(c.Request.Context(), uri.ID)
	if err != nil {
		c.JSON(404, map[string]string{"error": "user not found"})
		return
	}

	c.JSON(200, user)
}

func (h *UserHandler) List(c *gin.Context) {
	users, err := h.service.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(500, map[string]string{"error": "failed to list users"})
		return
	}

	c.JSON(200, users)
}

func (h *UserHandler) Update(c *gin.Context) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}
	var req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(400, map[string]string{"error": err.Error()})
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, map[string]string{"error": err.Error()})
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), uri.ID, req.Name, req.Email)
	if err != nil {
		c.JSON(404, map[string]string{"error": "user not found"})
		return
	}

	c.JSON(200, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}

	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(400, map[string]string{"error": err.Error()})
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), uri.ID); err != nil {
		c.JSON(404, map[string]string{"error": "user not found"})
		return
	}

	c.JSON(204, nil)
}

// RegisterRoutes registers HTTP routes
func RegisterRoutes(engine *gin.Engine, handler *UserHandler) {
	v1 := engine.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", handler.Create)
			users.GET("", handler.List)
			users.GET("/:id", handler.Get)
			users.PUT("/:id", handler.Update)
			users.DELETE("/:id", handler.Delete)
		}
	}
}

// SetupDatabase initializes the database schema
func SetupDatabase(lc fx.Lifecycle, conns dbx.Connections, logger logx.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			conn, exists := conns["default"]
			if !exists {
				return fmt.Errorf("default database connection not found")
			}

			// Auto-migrate the User model
			if err := conn.AutoMigrate(&User{}); err != nil {
				logger.Error("failed to migrate database", logx.Err(err))
				return err
			}

			logger.Info("database migration completed")
			return nil
		},
	})
}

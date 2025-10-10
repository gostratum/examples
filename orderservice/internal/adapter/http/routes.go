package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gostratum/httpx/responsex"
	"go.uber.org/zap"

	"github.com/gostratum/core"
	"github.com/gostratum/examples/orderservice/internal/usecase"
)

// RegisterRoutes registers all HTTP routes using the provided Gin engine
// This function is designed to be used with fx.Invoke to work with httpx.Module
func RegisterRoutes(
	e *gin.Engine,
	userService *usecase.UserService,
	orderService *usecase.OrderService,
	reg core.Registry,
	log *zap.Logger,
) {
	// Add responsex middleware for request tracking and metadata
	e.Use(responsex.MetaMiddleware("orderservice/v1.0.0"))

	// User handlers
	userHandler := NewUserHandler(userService, log)
	e.POST("/users", userHandler.CreateUser)
	e.GET("/users/:id", userHandler.GetUser)

	// Order handlers
	orderHandler := NewOrderHandler(orderService, log)
	e.POST("/orders", orderHandler.CreateOrder)
	e.GET("/orders/:id", orderHandler.GetOrder)

	// Health endpoints - readiness and liveness checks
	e.GET("/healthz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		res := reg.Aggregate(ctx, core.Readiness)
		if res.OK {
			c.JSON(http.StatusOK, gin.H{"ok": true, "details": res.Details})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"ok": false, "details": res.Details})
		}
	})

	e.GET("/livez", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		res := reg.Aggregate(ctx, core.Liveness)
		if res.OK {
			c.JSON(http.StatusOK, gin.H{"ok": true, "details": res.Details})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"ok": false, "details": res.Details})
		}
	})

	log.Info("HTTP routes registered")
}

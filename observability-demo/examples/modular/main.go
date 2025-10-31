package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gostratum/core"
	"github.com/gostratum/core/logx"
	"github.com/gostratum/dbx"
	"github.com/gostratum/httpx"
	"github.com/gostratum/metricsx"
	"github.com/gostratum/tracingx"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

// Modular example: demonstrates composing multiple modules using core.New()
// This is the recommended approach for multi-module workspaces
func main() {
	app := core.New(
		// Observability modules (opt-in)
		metricsx.Module(),
		tracingx.Module(),

		// Infrastructure modules
		dbx.Module(),
		httpx.Module(),

		// Application providers and invocations
		fx.Provide(
			NewUserService,
			NewUserHandler,
		),
		fx.Invoke(RegisterRoutes),
		fx.Invoke(SetupDatabase),
	)

	app.Run()
}

// Minimal copies of types so this example remains self-contained
type User struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email" gorm:"uniqueIndex"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserService struct {
	db     *gorm.DB
	logger logx.Logger
}

func NewUserService(db *gorm.DB, logger logx.Logger) (*UserService, error) {
	return &UserService{db: db, logger: logger}, nil
}

type UserHandler struct {
	service *UserService
	logger  logx.Logger
}

func NewUserHandler(service *UserService, logger logx.Logger) *UserHandler {
	return &UserHandler{service: service, logger: logger}
}

func RegisterRoutes(engine *gin.Engine, handler *UserHandler) { /* placeholder */ }

func SetupDatabase(lc fx.Lifecycle, db *gorm.DB, logger logx.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := db.AutoMigrate(&User{}); err != nil {
				logger.Error("failed to migrate database", logx.Err(err))
				return err
			}
			logger.Info("database migration completed")
			return nil
		},
	})
}

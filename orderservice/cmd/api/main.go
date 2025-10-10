package main

import (
	"go.uber.org/fx"

	"github.com/gostratum/core"
	"github.com/gostratum/dbx"
	httpAdapter "github.com/gostratum/examples/orderservice/internal/adapter/http"
	repoAdapter "github.com/gostratum/examples/orderservice/internal/adapter/repo"
	"github.com/gostratum/examples/orderservice/internal/usecase"
	"github.com/gostratum/httpx"
)

func main() {
	app := core.New(
		// Include dbx module without auto-migration (run migrations separately)
		dbx.Module(
			dbx.WithDefault("primary"),
			dbx.WithHealthChecks(),
		),

		// Include httpx module
		httpx.Module(),

		// Provide dependencies
		fx.Provide(
			// GORM repositories
			repoAdapter.NewUserRepo,
			repoAdapter.NewOrderRepo,

			// Usecase services
			usecase.NewUserService,
			usecase.NewOrderService,

			// HTTP handlers
			httpAdapter.NewUserHandler,
			httpAdapter.NewOrderHandler,
		),

		// Invoke setup functions
		fx.Invoke(
			httpAdapter.RegisterRoutes,
		),
	)

	app.Run()
}

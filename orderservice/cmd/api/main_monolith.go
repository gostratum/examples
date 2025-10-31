//go:build monolith
// +build monolith

package main

import (
	"github.com/gostratum/core"
	"github.com/gostratum/dbx"
	"github.com/gostratum/httpx"
	"github.com/gostratum/storagex"
	"go.uber.org/fx"

	httpAdapter "github.com/gostratum/examples/orderservice/internal/adapter/http"
	repoAdapter "github.com/gostratum/examples/orderservice/internal/adapter/repo"
	"github.com/gostratum/examples/orderservice/internal/usecase"
)

// Monolithic main selected with build tag `monolith`.
// Run: `go run -tags=monolith .` from the cmd/api directory
func main() {
	app := core.New(
		// Include dbx module without auto-migration (run migrations separately)
		dbx.Module(
			dbx.WithDefault("primary"),
			dbx.WithHealthChecks(),
		),

		// Include httpx module
		httpx.Module(),

		// Include storagex module (fx.Option variable)
		storagex.Module,

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

//go:build monolith
// +build monolith

package main

import (
	"github.com/gostratum/core/logx"
	"github.com/gostratum/dbx"
	"github.com/gostratum/httpx"
	"github.com/gostratum/metricsx"
	"github.com/gostratum/tracingx"
	"go.uber.org/fx"
)

// Monolithic main selected with build tag `monolith`.
// Run: `go run -tags=monolith .`
func main() {
	// Monolithic composition: keep wiring in one place (same as original but presented as a single unit)
	app := fx.New(
		logx.Module(),
		metricsx.Module(),
		tracingx.Module(),
		dbx.Module(),
		httpx.Module(),

		fx.Provide(
			NewUserService,
			NewUserHandler,
		),
		fx.Invoke(RegisterRoutes),
		fx.Invoke(SetupDatabase),
	)

	app.Run()
}

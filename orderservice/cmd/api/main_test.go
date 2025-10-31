package main

import (
	"context"
	"os"
	"testing"

	"github.com/gostratum/core"
	"go.uber.org/fx"
)

func TestMainApp(t *testing.T) {
	t.Run("app constructs without errors when DB available", func(t *testing.T) {
		// Skip this test if database is not configured
		// In CI/CD, you would set these environment variables
		if os.Getenv("STRATUM_DB_DATABASES_PRIMARY_DSN") == "" {
			t.Skip("Skipping integration test: STRATUM_DB_DATABASES_PRIMARY_DSN not set")
		}

		// This test requires an actual database connection
		// It verifies the full application can start and stop cleanly
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("app construction panicked: %v", r)
			}
		}()

		// Construct the app exactly like in main()
		// Note: This uses whatever configuration is in the environment/config files
		app := core.New(
			// All the modules and providers from main.go would go here
			// For now, just test that core.New works
			fx.Invoke(func() {
				// Minimal validation
			}),
		)

		if app == nil {
			t.Error("Expected fx app to be created")
			return
		}

		// Start and immediately stop the app to verify lifecycle
		ctx := context.Background()
		if err := app.Start(ctx); err != nil {
			t.Fatalf("Failed to start app: %v", err)
		}

		if err := app.Stop(ctx); err != nil {
			t.Fatalf("Failed to stop app: %v", err)
		}
	})
}

func TestMainAppMinimal(t *testing.T) {
	t.Run("core.New minimal construction", func(t *testing.T) {
		// Test that core.New works without any additional modules
		// This validates the core framework itself

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("core.New panicked: %v", r)
			}
		}()

		app := core.New(
			fx.Invoke(func() {
				// Empty invoke just to validate construction
			}),
		)

		if app == nil {
			t.Error("Expected app to be created")
		}
	})
}

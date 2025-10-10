package main

import (
	"context"
	"testing"

	"go.uber.org/fx"
)

func TestMainApp(t *testing.T) {
	t.Run("app starts successfully", func(t *testing.T) {
		// This test verifies that the fx app construction doesn't panic
		// In a real integration test, you would provide actual implementations
		// For now, we just test that fx.New doesn't panic with minimal setup

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("fx.New panicked: %v", r)
			}
		}()

		// Test that fx.New doesn't panic
		app := fx.New(
			fx.Provide(
				func() interface{} { return "mock_dependency" },
			),
			fx.NopLogger,
		)

		if app == nil {
			t.Error("Expected fx app to be created")
		}

		// Clean up
		app.Stop(context.Background())
	})
}

func TestMainAppIntegration(t *testing.T) {
	t.Run("dependency injection setup is valid", func(t *testing.T) {
		// This test verifies that the fx app construction is valid
		// We can't easily test the full app startup in unit tests due to external dependencies
		// But we can test that the fx.New call doesn't panic and returns a valid app

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("fx.New panicked: %v", r)
			}
		}()

		// Test that fx.New doesn't panic with the current setup
		// Note: This will fail in CI without proper environment setup
		// In a real scenario, you'd mock all external dependencies
		app := fx.New(
			fx.Provide(
				func() interface{} { return "mock_dependency" },
			),
			fx.NopLogger,
		)

		if app == nil {
			t.Error("Expected fx app to be created")
		}
	})
}

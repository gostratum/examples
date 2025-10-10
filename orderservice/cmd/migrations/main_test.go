package main

import (
	"testing"
	"time"

	"github.com/gostratum/dbx/migrate"
	"github.com/stretchr/testify/assert"
)

func TestMaskDatabaseURL(t *testing.T) {
	tests := []struct {
		name     string
		dbURL    string
		expected string
	}{
		{
			name:     "postgres URL with password",
			dbURL:    "postgres://postgres:secret123@localhost:5432/orders?sslmode=disable",
			expected: "postgres://postgres:***@localhost:5432/orders?sslmode=disable",
		},
		{
			name:     "postgres URL without password",
			dbURL:    "postgres://postgres@localhost:5432/orders",
			expected: "postgres://postgres@localhost:5432/orders",
		},
		{
			name:     "empty URL",
			dbURL:    "",
			expected: "",
		},
		{
			name:     "invalid URL format",
			dbURL:    "invalid-url",
			expected: "invalid-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskDatabaseURL(tt.dbURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigToOptions(t *testing.T) {
	tests := []struct {
		name   string
		config *migrate.Config
	}{
		{
			name: "filesystem migrations with all options",
			config: &migrate.Config{
				Dir:         "./migrations",
				Table:       "custom_migrations",
				LockTimeout: 30 * time.Second,
				Verbose:     true,
				UseEmbed:    false,
				AutoMigrate: false,
			},
		},
		{
			name: "embedded migrations",
			config: &migrate.Config{
				UseEmbed:    true,
				Table:       "schema_migrations",
				LockTimeout: 15 * time.Second,
				Verbose:     false,
				AutoMigrate: false,
			},
		},
		{
			name: "auto-migrate enabled",
			config: &migrate.Config{
				Dir:         "./migrations",
				AutoMigrate: true,
				Table:       "schema_migrations",
				Verbose:     true,
				UseEmbed:    false,
			},
		},
		{
			name: "minimal config",
			config: &migrate.Config{
				Dir: "./migrations",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := configToOptions(tt.config)

			// We can't directly test functional options, but we can ensure
			// the function doesn't panic and returns options when config has values
			assert.NotNil(t, opts)

			// Should have at least some options if config has values
			if tt.config.Dir != "" || tt.config.UseEmbed || tt.config.Table != "" ||
				tt.config.LockTimeout > 0 || tt.config.Verbose || tt.config.AutoMigrate {
				assert.True(t, len(opts) > 0)
			}
		})
	}
}

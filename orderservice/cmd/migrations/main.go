package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gostratum/core/configx"
	"github.com/gostratum/dbx/migrate"
)

func main() {
	var action string
	var steps int
	var version uint
	flag.StringVar(&action, "action", "up", "Action to perform: up, down, version, force, status")
	flag.IntVar(&steps, "steps", 0, "Number of migrations to apply (0 = all)")
	flag.UintVar(&version, "version", 0, "Version for force action")
	flag.Parse()

	fmt.Printf("🔄 Starting database migration: %s...\n", action)

	// Load configuration using configx
	loader := configx.New(
		configx.WithConfigPaths("./configs"),
	)

	// Bind environment variables for DSN
	if err := loader.BindEnv("databases.primary.dsn", "STRATUM_DATABASES_PRIMARY_DSN", "DATABASE_URL"); err != nil {
		log.Printf("Warning: Could not bind DSN env var: %v", err)
	}

	// For DSN, we'll use a simple environment variable approach since we only need one value
	// and configx.Loader.Bind requires Prefix() implementation
	dbURL := lookupEnv("DATABASE_URL", "STRATUM_DATABASES_PRIMARY_DSN")
	if dbURL == "" {
		// Fallback to default for local development
		dbURL = "postgres://postgres:postgres@localhost:5432/orders?sslmode=disable"
		log.Printf("Using default DSN (set DATABASE_URL or STRATUM_DATABASES_PRIMARY_DSN to override)")
	}

	// Create migration config using configx.Loader
	migrationConfig, err := migrate.NewConfig(loader)
	if err != nil {
		log.Fatalf("Failed to load migration config: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Execute migration action
	if err := runMigrationAction(ctx, dbURL, action, steps, version, migrationConfig); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("✅ Migration operation completed successfully")
}

// runMigrationAction executes the specified migration action using gostratum/dbx/migrate
func runMigrationAction(ctx context.Context, dbURL, action string, steps int, version uint, cfg *migrate.Config) error {
	// Convert config to options
	opts := configToOptions(cfg)

	switch action {
	case "up":
		fmt.Println("📦 Running migrations up...")
		if steps > 0 {
			if err := migrate.Steps(ctx, dbURL, steps, opts...); err != nil {
				return fmt.Errorf("failed to migrate up %d steps: %w", steps, err)
			}
		} else {
			if err := migrate.Up(ctx, dbURL, opts...); err != nil {
				return fmt.Errorf("failed to migrate up: %w", err)
			}
		}
		fmt.Println("✅ Migrations applied successfully")

	case "down":
		fmt.Println("⚠️  Rolling back migrations...")
		if steps > 0 {
			if err := migrate.Steps(ctx, dbURL, -steps, opts...); err != nil {
				return fmt.Errorf("failed to migrate down %d steps: %w", steps, err)
			}
		} else {
			if err := migrate.Down(ctx, dbURL, opts...); err != nil {
				return fmt.Errorf("failed to migrate down: %w", err)
			}
		}
		fmt.Println("✅ Rollback completed")

	case "version", "status":
		status, err := migrate.GetStatus(ctx, dbURL, opts...)
		if err != nil {
			return fmt.Errorf("failed to get migration status: %w", err)
		}

		fmt.Printf("📋 Migration Status:\n")
		fmt.Printf("  Database: %s\n", maskDatabaseURL(dbURL))
		fmt.Printf("  Current Version: %d\n", status.Current)
		fmt.Printf("  Dirty: %v\n", status.Dirty)
		fmt.Printf("  Applied: %v\n", status.Applied)
		fmt.Printf("  Pending: %v\n", status.Pending)

	case "force":
		if version == 0 && steps == 0 {
			return fmt.Errorf("force action requires a version specified via -version flag or -steps flag")
		}

		forceVersion := int(version)
		if forceVersion == 0 {
			forceVersion = steps
		}

		fmt.Printf("⚠️  Forcing version to %d...\n", forceVersion)
		if err := migrate.Force(ctx, dbURL, forceVersion, opts...); err != nil {
			return fmt.Errorf("failed to force version: %w", err)
		}
		fmt.Println("✅ Version forced successfully")

	default:
		return fmt.Errorf("unknown action: %s. Use up, down, version, status, or force", action)
	}

	return nil
}

// configToOptions converts migration config to functional options
func configToOptions(cfg *migrate.Config) []migrate.Option {
	var opts []migrate.Option

	if cfg.Dir != "" {
		opts = append(opts, migrate.WithDir(cfg.Dir))
	}

	if cfg.UseEmbed {
		opts = append(opts, migrate.WithEmbed())
	}

	if cfg.Table != "" {
		opts = append(opts, migrate.WithTable(cfg.Table))
	}

	if cfg.LockTimeout > 0 {
		opts = append(opts, migrate.WithLockTimeout(cfg.LockTimeout))
	}

	if cfg.Verbose {
		opts = append(opts, migrate.WithVerbose())
	}

	if cfg.AutoMigrate {
		opts = append(opts, migrate.WithAutoMigrate())
	}

	return opts
}

// lookupEnv checks multiple environment variable names and returns the first non-empty value
func lookupEnv(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

// maskDatabaseURL masks sensitive information in database URL for logging
func maskDatabaseURL(dbURL string) string {
	// Simple masking - replace password with ***
	// This is a basic implementation, could be enhanced
	if len(dbURL) == 0 {
		return ""
	}

	// Look for pattern like postgres://user:password@host
	start := strings.Index(dbURL, "://")
	if start == -1 {
		return dbURL
	}

	at := strings.Index(dbURL[start+3:], "@")
	if at == -1 {
		return dbURL
	}

	colon := strings.Index(dbURL[start+3:start+3+at], ":")
	if colon == -1 {
		return dbURL
	}

	// Replace password with ***
	masked := dbURL[:start+3+colon+1] + "***" + dbURL[start+3+at:]
	return masked
}

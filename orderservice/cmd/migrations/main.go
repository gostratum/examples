package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	repoAdapter "github.com/gostratum/examples/orderservice/internal/adapter/repo"
)

func main() {
	var action string
	flag.StringVar(&action, "action", "migrate", "Action to perform: migrate, rollback, status")
	flag.Parse()

	fmt.Printf("🔄 Starting database %s...\n", action)

	// Load configuration
	v := viper.New()
	v.SetConfigName("base")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")

	// Set environment variable defaults
	v.SetDefault("databases.primary.driver", "postgres")
	v.SetDefault("databases.primary.dsn", "host=localhost user=postgres password=postgres dbname=orders port=5432 sslmode=disable")

	if err := v.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	// Connect to database directly
	dsn := v.GetString("databases.primary.dsn")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("📦 Running auto-migrations...")

	// Run auto-migration directly
	if err := db.AutoMigrate(&repoAdapter.UserEntity{}, &repoAdapter.OrderEntity{}, &repoAdapter.ItemEntity{}); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}

	fmt.Println("✅ Schema migration completed")

	// Run custom migration logic based on action
	if action != "migrate" {
		runner := &MigrationRunner{db: db}
		if err := RunMigrations(runner, db, action); err != nil {
			log.Fatalf("Migration action failed: %v", err)
		}
	}

	fmt.Println("✅ Migration operation completed successfully")
}

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db *gorm.DB
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *gorm.DB) *MigrationRunner {
	return &MigrationRunner{db: db}
}

// RunMigrations executes database migrations based on action
func RunMigrations(runner *MigrationRunner, db *gorm.DB, action string) error {
	ctx := context.Background()

	switch action {
	case "migrate":
		fmt.Println("📦 Running auto-migrations...")
		// Auto-migration happens automatically through dbx.WithAutoMigrate
		fmt.Println("✅ Schema migration completed")

	case "status":
		fmt.Println("📋 Checking migration status...")
		if err := checkMigrationStatus(ctx, db); err != nil {
			return fmt.Errorf("failed to check migration status: %v", err)
		}

	case "rollback":
		fmt.Println("⚠️  Rollback not implemented - use manual SQL scripts for schema rollbacks")
		return fmt.Errorf("rollback action requires manual intervention")

	default:
		return fmt.Errorf("unknown action: %s. Use migrate, status, or rollback", action)
	}

	// Exit after migrations are done
	os.Exit(0)
	return nil
}

// checkMigrationStatus verifies that tables exist and are accessible
func checkMigrationStatus(ctx context.Context, db *gorm.DB) error {
	tables := []string{"users", "orders", "items"}

	for _, table := range tables {
		if db.WithContext(ctx).Migrator().HasTable(table) {
			fmt.Printf("✅ Table '%s' exists\n", table)
		} else {
			fmt.Printf("❌ Table '%s' does not exist\n", table)
		}
	}

	return nil
}

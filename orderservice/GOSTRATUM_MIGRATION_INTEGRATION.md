# Database Migration - gostratum/dbx Integration

## Overview

Migrated the database migration system from manual `golang-migrate` implementation to the **gostratum/dbx/migrate** package, providing a unified migration approach across all gostratum-based projects.

## Changes Made

### 1. Migration Framework

**Before:**
- Used external `github.com/golang-migrate/migrate/v4` package
- Required separate PostgreSQL driver (`github.com/lib/pq`)
- Manual setup with file source and database drivers

**After:**
- Uses **gostratum/dbx/migrate** package (built-in to the framework)
- No external migration dependencies
- Simplified configuration and usage
- Consistent with other gostratum projects

### 2. Dependencies Removed

```go
// Removed from go.mod:
github.com/golang-migrate/migrate/v4
github.com/golang-migrate/migrate/v4/database/postgres
github.com/golang-migrate/migrate/v4/source/file
github.com/lib/pq
```

### 3. Migration Runner Refactored

**File:** `cmd/migrations/main.go`

**Key Changes:**
- Replaced `golang-migrate` with `github.com/gostratum/dbx/migrate`
- Simplified database connection (uses DSN to URL conversion)
- Added context support with timeout
- Better error handling with wrapped errors
- Enhanced status command showing applied and pending migrations

**New Implementation:**
```go
import (
    "github.com/gostratum/dbx/migrate"
)

// Simple API:
migrate.Up(ctx, dbURL, migrate.WithDir("./migrations"))
migrate.Down(ctx, dbURL, migrate.WithDir("./migrations"))
migrate.GetStatus(ctx, dbURL, migrate.WithDir("./migrations"))
```

### 4. Command Line Interface

**Actions Supported:**
- `up` - Apply all pending migrations (or specific count with `-steps`)
- `down` - Rollback migrations (or specific count with `-steps`)
- `status` - Show detailed migration status (applied, pending, current version)
- `version` - Alias for status
- `force` - Force version (recovery tool, use with caution)

**New Flags:**
```bash
-action string   # Action: up, down, status, version, force (default "up")
-steps int       # Number of migrations to apply (0 = all)
-version uint    # Version for force action
```

### 5. Enhanced Status Command

The `status` action now provides comprehensive migration information:

```
📋 Migration Status:
  Database: postgres://user:pass@host:port/dbname
  Current Version: 3
  Dirty: false
  Applied: [1 2 3]
  Pending: []
```

### 6. Makefile Updates

**Updated Targets:**
```makefile
# Changed from -action=version to -action=status
migrate-version:
    go run ./cmd/migrations -action=status

# Changed from -steps=$(VERSION) to -version=$(VERSION)
migrate-force:
    go run ./cmd/migrations -action=force -version=$(VERSION)
```

### 7. Documentation Updates

- **MIGRATIONS.md**: Updated to reference gostratum/dbx/migrate
- **init.sql**: Updated deprecation notice
- **This file**: Replaced MIGRATION_REFACTORING.md

## Benefits

### Unified Framework ✅
- Uses gostratum's own migration package
- Consistent API across all gostratum projects
- No need to learn different migration tools

### Simplified Dependencies ✅
- Removed 4 external dependencies
- Everything included in gostratum/dbx
- Easier maintenance

### Better Integration ✅
- Works seamlessly with gostratum/dbx module
- Can be integrated into fx lifecycle if needed
- Consistent error handling with framework

### Enhanced Features ✅
- Context support with timeouts
- Better status reporting
- Improved error messages
- More flexible force command

## Migration File Format

**No changes** - Still uses standard SQL migration files:

```
migrations/
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_create_orders_table.up.sql
├── 000002_create_orders_table.down.sql
├── 000003_create_indexes.up.sql
└── 000003_create_indexes.down.sql
```

## Usage Examples

### Development Workflow

```bash
# Apply all pending migrations
make migrate

# Check status
make migrate-version

# Rollback last migration
make migrate-down

# Rollback specific number
make migrate-down STEPS=2

# Force to specific version (recovery)
make migrate-force VERSION=1
```

### Direct Command Usage

```bash
# Apply all migrations
go run cmd/migrations/main.go -action=up

# Apply 2 migrations
go run cmd/migrations/main.go -action=up -steps=2

# Rollback 1 migration
go run cmd/migrations/main.go -action=down -steps=1

# Check status
go run cmd/migrations/main.go -action=status

# Force version
go run cmd/migrations/main.go -action=force -version=2
```

### Production Deployment

```bash
# Build the binary
go build -o bin/migrate cmd/migrations/main.go

# Run migrations
./bin/migrate -action=up

# Verify
./bin/migrate -action=status

# Start API
./bin/api
```

## Technical Details

### DSN to URL Conversion

The migration runner automatically converts GORM DSN format to PostgreSQL URL:

**From (GORM DSN):**
```
host=localhost user=postgres password=postgres dbname=orders port=5432 sslmode=disable
```

**To (PostgreSQL URL):**
```
postgres://postgres:postgres@localhost:5432/orders?sslmode=disable
```

### Context and Timeouts

All migration operations use context with a 5-minute timeout:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

migrate.Up(ctx, dbURL, opts...)
```

### Error Handling

Errors are wrapped with context for better debugging:

```go
if err := migrate.Up(ctx, dbURL, opts...); err != nil {
    return fmt.Errorf("failed to migrate up: %w", err)
}
```

## File Changes

| File | Status | Description |
|------|--------|-------------|
| `cmd/migrations/main.go` | Modified | Refactored to use gostratum/dbx/migrate |
| `go.mod` | Modified | Removed golang-migrate dependencies |
| `go.sum` | Modified | Updated checksums |
| `Makefile` | Modified | Updated migrate-force to use -version flag |
| `MIGRATIONS.md` | Modified | Updated to reference gostratum |
| `init.sql` | Modified | Updated deprecation notice |
| `MIGRATION_REFACTORING.md` | Replaced | This document |

## Comparison: Before vs After

### Before (golang-migrate)

```go
import (
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    _ "github.com/lib/pq"
)

// Complex setup
db, _ := sql.Open("postgres", dsn)
driver, _ := postgres.WithInstance(db, &postgres.Config{})
m, _ := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)

// Run migration
m.Up()
```

### After (gostratum/dbx/migrate)

```go
import "github.com/gostratum/dbx/migrate"

// Simple, direct API
migrate.Up(ctx, dbURL, migrate.WithDir("./migrations"))
```

## gostratum Integration Path

This implementation sets up the orderservice to use gostratum's migration package consistently. Future enhancements could include:

1. **fx Integration**: Integrate migrations into the fx app lifecycle
2. **Module Pattern**: Use dbx.Module with migration options
3. **Embedded Migrations**: Support `//go:embed` for production deployments
4. **Health Checks**: Add migration status to health endpoints

For now, the standalone CLI approach provides maximum flexibility while using gostratum's migration framework.

## References

- [gostratum/dbx Documentation](https://github.com/gostratum/gostratum/tree/master/dbx)
- [gostratum/dbx/migrate Package](https://github.com/gostratum/gostratum/tree/master/dbx/migrate)
- [Migration Best Practices](../MIGRATIONS.md)
- [Migration Files Guide](./migrations/README.md)

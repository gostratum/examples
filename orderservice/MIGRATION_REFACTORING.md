# Database Migration Refactoring Summary

## Overview

Refactored the database initialization from a single `init.sql` script to a versioned migration system using `golang-migrate` for better database version control and tracking.

## Changes Made

### 1. Migration Framework Setup

**Added Dependencies:**
- `github.com/golang-migrate/migrate/v4` - Core migration library
- `github.com/golang-migrate/migrate/v4/database/postgres` - PostgreSQL driver
- `github.com/golang-migrate/migrate/v4/source/file` - File-based migrations
- `github.com/lib/pq` - PostgreSQL driver for database/sql

### 2. Migration Files Created

Created `migrations/` directory with versioned SQL migrations:

```
migrations/
├── README.md                                 (Documentation)
├── 000001_create_users_table.up.sql         (Create users table)
├── 000001_create_users_table.down.sql       (Drop users table)
├── 000002_create_orders_table.up.sql        (Create orders table)
├── 000002_create_orders_table.down.sql      (Drop orders table)
├── 000003_create_indexes.up.sql             (Create performance indexes)
└── 000003_create_indexes.down.sql           (Drop indexes)
```

**Migration Breakdown:**

| Version | Up Migration | Down Migration |
|---------|--------------|----------------|
| 000001 | Creates `users` table with id, name, email, created_at | Drops `users` table |
| 000002 | Creates `orders` table with id, user_id, items, status, created_at | Drops `orders` table |
| 000003 | Creates indexes: idx_users_email, idx_orders_user_id, idx_orders_status | Drops all indexes |

### 3. Migration Runner Refactored

**File:** `cmd/migrations/main.go`

**Before:**
- Used GORM AutoMigrate
- Limited to `migrate`, `status`, `rollback` actions
- No version tracking
- No rollback implementation

**After:**
- Uses golang-migrate framework
- Supports `up`, `down`, `version`, `force` actions
- Automatic version tracking via `schema_migrations` table
- Full rollback support with down migrations
- Supports step-by-step migrations

**New Capabilities:**
```bash
# Apply all pending migrations
go run cmd/migrations/main.go -action=up

# Apply specific number of migrations
go run cmd/migrations/main.go -action=up -steps=2

# Rollback specific number of migrations  
go run cmd/migrations/main.go -action=down -steps=1

# Check current version
go run cmd/migrations/main.go -action=version

# Force to specific version (recovery)
go run cmd/migrations/main.go -action=force -steps=2
```

### 4. Updated init.sql

**File:** `init.sql`

**Before:**
- Created all tables and indexes directly
- Single monolithic script
- No version tracking

**After:**
- Now a documentation file pointing to migration system
- Explains migration-based approach
- Lists available migrations
- Deprecated for schema changes (use migrations instead)

### 5. Makefile Updates

**New Targets:**
```makefile
make migrate          # Apply all pending migrations
make migrate-down     # Rollback migrations (STEPS=n)
make migrate-version  # Show current version
make migrate-force    # Force to version (VERSION=n)
make dev              # Migrate then start API
```

**Deprecated:**
```makefile
make status          # Now redirects to migrate-version
```

### 6. Documentation Updates

**MIGRATIONS.md:**
- Complete rewrite with golang-migrate approach
- Added migration creation guide
- Added best practices section
- Added troubleshooting guide
- Examples for all migration operations

**README.md:**
- Updated Quick Start to use migrations
- Updated available make targets
- Updated troubleshooting section
- Added reference to MIGRATIONS.md

**migrations/README.md (New):**
- Developer guide for creating migrations
- Best practices and conventions
- Testing guidelines
- Troubleshooting tips

## Benefits of New Approach

### Version Control ✅
- Each schema change is tracked with a version number
- Clear history of all database changes
- Easy to see what version is deployed

### Rollback Support ✅
- Every migration has a corresponding rollback
- Can rollback specific number of migrations
- Safe recovery from failed migrations

### Team Collaboration ✅
- Multiple developers can work on migrations
- Conflicts are visible (version numbers)
- Migrations can be code-reviewed

### Production Safety ✅
- Test migrations in dev/staging first
- Atomic migrations (all or nothing)
- Can force version for recovery
- Dirty state detection and recovery

### Audit Trail ✅
- `schema_migrations` table tracks:
  - Which migrations have been applied
  - Current version
  - Dirty state (if migration failed)

## Migration Tracking

The system automatically creates and manages a `schema_migrations` table:

```sql
CREATE TABLE schema_migrations (
    version BIGINT PRIMARY KEY,
    dirty BOOLEAN NOT NULL
);
```

This table tracks:
- **version**: The current migration version
- **dirty**: Whether a migration failed partway through

## Usage Examples

### Development Workflow

```bash
# 1. Start database
make docker-db

# 2. Run migrations
make migrate

# 3. Start API
make api

# Or combine steps 2-3
make dev
```

### Creating a New Migration

```bash
# 1. Create files
touch migrations/000004_add_user_phone.up.sql
touch migrations/000004_add_user_phone.down.sql

# 2. Write SQL in both files

# 3. Test
make migrate              # Apply
make migrate-version      # Verify
make migrate-down STEPS=1 # Test rollback
make migrate              # Re-apply
```

### Production Deployment

```bash
# 1. Build binaries
make build

# 2. Deploy and run migrations first
./bin/migrate -action=up

# 3. Verify migration succeeded
./bin/migrate -action=version

# 4. Start API
./bin/api
```

### Rollback in Production

```bash
# Rollback last migration
./bin/migrate -action=down -steps=1

# Verify
./bin/migrate -action=version

# Rollback multiple
./bin/migrate -action=down -steps=3
```

## Breaking Changes

### For Developers

- ❌ **Old:** `make db-init` to initialize schema
- ✅ **New:** `make migrate` to apply migrations

- ❌ **Old:** Edit `init.sql` for schema changes
- ✅ **New:** Create new migration files in `migrations/`

### For CI/CD

Update deployment scripts from:
```bash
psql -f init.sql  # Old
```

To:
```bash
./bin/migrate -action=up  # New
```

## File Changes Summary

| File | Status | Description |
|------|--------|-------------|
| `migrations/000001_create_users_table.up.sql` | Created | Users table creation |
| `migrations/000001_create_users_table.down.sql` | Created | Users table rollback |
| `migrations/000002_create_orders_table.up.sql` | Created | Orders table creation |
| `migrations/000002_create_orders_table.down.sql` | Created | Orders table rollback |
| `migrations/000003_create_indexes.up.sql` | Created | Index creation |
| `migrations/000003_create_indexes.down.sql` | Created | Index rollback |
| `migrations/README.md` | Created | Developer migration guide |
| `cmd/migrations/main.go` | Modified | Refactored to use golang-migrate |
| `init.sql` | Modified | Now documentation/deprecated |
| `Makefile` | Modified | Added migration targets |
| `MIGRATIONS.md` | Modified | Complete rewrite |
| `README.md` | Modified | Updated references |
| `go.mod` | Modified | Added migration dependencies |
| `go.sum` | Modified | Updated checksums |

## Next Steps

1. ✅ Test migrations in development environment
2. ✅ Update CI/CD pipelines to use new migration commands
3. ✅ Train team on new migration workflow
4. ✅ Document any environment-specific migration steps
5. ✅ Create runbook for production migration deployments

## Rollback Plan (If Needed)

If you need to revert to the old system:

```bash
# 1. Restore old cmd/migrations/main.go from git
git checkout HEAD~1 cmd/migrations/main.go

# 2. Remove migration dependencies
go mod tidy

# 3. Restore old Makefile targets
git checkout HEAD~1 Makefile

# 4. Use init.sql directly
psql -h localhost -U postgres -d orders -f init.sql
```

## Questions?

See [MIGRATIONS.md](MIGRATIONS.md) for complete documentation or the [migrations/README.md](migrations/README.md) for developer guide.

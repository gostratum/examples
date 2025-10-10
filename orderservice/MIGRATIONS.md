# OrderService - Database Migrations

This project uses versioned SQL migrations with [gostratum/dbx/migrate](https://github.com/gostratum/gostratum/tree/master/dbx) for database schema management and version control.

⚠️  **Important**: Always run migrations before starting the API service.

## Migration System

### Overview

- **Migration Files**: Located in `./migrations/` directory
- **Versioning**: Each migration has a version number (e.g., `000001_`)
- **Up/Down**: Each migration has both `.up.sql` (apply) and `.down.sql` (rollback)
- **Tracking**: gostratum/dbx/migrate automatically tracks applied migrations in `schema_migrations` table
- **Framework**: Uses gostratum's built-in migration package (part of the dbx module)

### Current Migrations

```
000001_create_users_table.up.sql     - Creates users table
000002_create_orders_table.up.sql    - Creates orders table with foreign key
000003_create_indexes.up.sql         - Creates performance indexes
```

## Migration Commands

### Using Make (Recommended)

```bash
# Apply all pending migrations
make migrate

# Rollback last migration
make migrate-down

# Rollback specific number of migrations
make migrate-down STEPS=2

# Check current migration version
make migrate-version

# Force migration to specific version (use with caution)
make migrate-force VERSION=1

# Start API (assumes migrations completed)
make api

# Development workflow (migrate then start API)
make dev
```

### Direct Commands

```bash
# Apply all pending migrations (default action)
go run cmd/migrations/main.go -action=up

# Apply specific number of migrations
go run cmd/migrations/main.go -action=up -steps=2

# Rollback all migrations
go run cmd/migrations/main.go -action=down

# Rollback specific number of migrations
go run cmd/migrations/main.go -action=down -steps=1

# Check current version
go run cmd/migrations/main.go -action=status

# Force version (use with caution!)
go run cmd/migrations/main.go -action=force -version=2
```

### API Service

```bash
# Run the API service (assumes migrations have been run)
go run cmd/api/main.go
```

## Creating New Migrations

### Naming Convention

Migrations follow the pattern: `{version}_{description}.{up|down}.sql`

Example:
```
000004_add_user_address.up.sql
000004_add_user_address.down.sql
```

### Creating a New Migration

1. **Create the files** in `migrations/` directory:
   ```bash
   # Create files manually or use a template
   touch migrations/000004_add_user_address.up.sql
   touch migrations/000004_add_user_address.down.sql
   ```

2. **Write the UP migration** (`000004_add_user_address.up.sql`):
   ```sql
   ALTER TABLE users ADD COLUMN address TEXT;
   ```

3. **Write the DOWN migration** (`000004_add_user_address.down.sql`):
   ```sql
   ALTER TABLE users DROP COLUMN address;
   ```

4. **Test the migration**:
   ```bash
   # Apply it
   make migrate
   
   # Verify it worked
   make migrate-version
   
   # Test rollback
   make migrate-down STEPS=1
   
   # Re-apply
   make migrate
   ```

## Build Commands

```bash
# Build migration binary
go build -o bin/migrate cmd/migrations/main.go

# Build API binary  
go build -o bin/api cmd/api/main.go

# Build both
make build
```

## Docker Deployment Pattern

```dockerfile
# In production, you would typically:
# 1. Build both binaries
# 2. Run migrations first: ./migrate -action=up
# 3. Then start API service: ./api
```

Example Docker entrypoint:
```bash
#!/bin/sh
# Run migrations
./migrate -action=up
# Start API
./api
```

## Environment Variables

Both commands use the same database configuration through the `configs/base.yaml` file or environment variables:

- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432) 
- `DB_NAME`: Database name (default: orders)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)

## Migration Features

- **Version Control**: Each schema change is tracked with a version number
- **Rollback Support**: Every migration has a corresponding rollback (down migration)
- **State Tracking**: `schema_migrations` table tracks which migrations are applied
- **Idempotent**: Can safely re-run migrations (already applied ones are skipped)
- **Safe Separation**: API won't run migrations, ensuring controlled deployments
- **Built-in Framework**: Uses gostratum's dbx/migrate package - no external dependencies

## Production Best Practices

1. **Test First**: Always test migrations in staging before production
2. **Backup**: Take database backups before running migrations in production
3. **Review**: Have migrations peer-reviewed before applying
4. **One-Way**: Prefer making migrations one-way (add columns, don't drop)
5. **Separate Deploys**: Run migrations separately from API deployments
6. **Monitor**: Watch migration logs and verify success before deploying API
7. **Rollback Plan**: Know how to rollback both migrations and code together
8. **Idempotent**: Write migrations that can be safely re-run if needed

## Troubleshooting

### Dirty Migration State

If a migration fails partway through, the database may be in a "dirty" state:

```bash
# Check version (will show dirty=true if there's an issue)
make migrate-version

# Fix by forcing to last known good version
make migrate-force VERSION=2
```

### Migration Conflicts

If multiple developers create the same version number:

1. Rename the newer migration to the next available version
2. Update references in the migration files if needed
3. Coordinate version numbers in your team

### Reset Database (Development Only)

```bash
# Rollback all migrations
make migrate-down

# Re-apply all migrations
make migrate
```

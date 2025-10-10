# OrderService - Database Migrations

This project separates database migrations from the API service for better control and deployment flexibility.

⚠️  **Important**: Always run migrations before starting the API service.

## Migration Commands

### Using Make (Recommended)

```bash
# Run migrations
make migrate

# Check migration status  
make status

# Start API (assumes migrations completed)
make api

# Development workflow (migrate then start API)
make dev
```

### Direct Commands

```bash
# Basic migration (default action)
go run cmd/migrations/main.go

# Explicit migration
go run cmd/migrations/main.go -action=migrate

# Check migration status
go run cmd/migrations/main.go -action=status
```

### API Service

```bash
# Run the API service (assumes migrations have been run)
go run cmd/api/main.go
```

## Build Commands

```bash
# Build migration binary
go build -o bin/migrate cmd/migrations/main.go

# Build API binary  
go build -o bin/api cmd/api/main.go
```

## Docker Deployment Pattern

```dockerfile
# In production, you would typically:
# 1. Build both binaries
# 2. Run migrations first: ./migrate
# 3. Then start API service: ./api
```

## Environment Variables

Both commands use the same database configuration through the `configs/base.yaml` file or environment variables:

- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432) 
- `DB_NAME`: Database name (default: orderservice)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)

## Migration Features

- **Auto-Migration**: Automatically creates and updates table schemas
- **Status Check**: Verifies that all required tables exist
- **Safe Separation**: API won't run migrations, ensuring controlled deployments
- **Health Checks**: Migration includes basic table existence validation

## Production Best Practices

1. Always run migrations before starting the API
2. Use versioned migration binaries in production
3. Test migrations in staging environments first
4. Keep migration and API deployments separate
5. Monitor migration logs for any schema changes
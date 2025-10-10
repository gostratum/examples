# Order Service Example

A Clean Architecture example service demonstrating production-ready patterns with `github.com/gostratum/core`, `github.com/gostratum/httpx`, and `github.com/gostratum/dbx`.

This service manages users and orders with proper health checks, graceful shutdown, and database connectivity monitoring.

## Architecture

The service follows Clean Architecture principles with these layers:

- **Domain**: Core business entities (`User`, `Order`) with validation
- **Usecase**: Business logic with typed errors and context timeouts
- **Adapter**: External interfaces (HTTP handlers, PostgreSQL repositories)
- **Infrastructure**: Database connections, HTTP server, health monitoring

## Features

- ✅ Clean Architecture with proper dependency injection
- ✅ Health checks (liveness and readiness)
- ✅ Database connection monitoring with automatic recovery
- ✅ Graceful shutdown
- ✅ Context timeouts for all operations
- ✅ Proper error mapping (HTTP status codes)
- ✅ Request ID tracing
- ✅ JSON logging with structured fields

## Prerequisites

- Go 1.25+
- PostgreSQL 16+
- Docker (for running PostgreSQL)

## Setup

### Quick Start with Docker Compose

```bash
# Start PostgreSQL with automatic schema initialization
docker-compose up -d

# Download dependencies
go mod tidy

# Run the service
make run
```

### Manual Setup

#### 1. Start PostgreSQL

```bash
# Option 1: Using Docker Compose (recommended)
docker-compose up -d postgres

# Option 2: Using Docker directly
docker run --rm \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=orders \
  -p 5432:5432 \
  postgres:16

# Option 3: Using Make
make docker-db
```

#### 2. Run Database Migrations

```bash
# Run all pending migrations
make migrate

# Or manually:
go run cmd/migrations/main.go -action=up
```

See [MIGRATIONS.md](MIGRATIONS.md) for detailed migration documentation.

#### 3. Run the Service

```bash
# Using Make (recommended)
make run

# Or manually:
export APP_ENV=dev
export CONFIG_PATHS=./configs
go run ./cmd/api
```

The service will start on port 8080 with health monitoring enabled.

## API Endpoints

### Users

#### Create User
```bash
curl -s -X POST localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

Response:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Alice",
  "email": "alice@example.com",
  "created_at": "2025-10-07T10:30:00Z"
}
```

#### Get User
```bash
curl -s localhost:8080/users/123e4567-e89b-12d3-a456-426614174000
```

### Orders

#### Create Order
```bash
curl -s -X POST localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "items": [
      {"sku": "SKU1", "qty": 2, "price": 9.50},
      {"sku": "SKU2", "qty": 1, "price": 15.00}
    ]
  }'
```

Response:
```json
{
  "id": "987fcdeb-51a2-43d1-b456-426614174000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "items": [
    {"sku": "SKU1", "qty": 2, "price": 9.50},
    {"sku": "SKU2", "qty": 1, "price": 15.00}
  ],
  "status": "pending",
  "created_at": "2025-10-07T10:35:00Z"
}
```

#### Get Order
```bash
curl -s localhost:8080/orders/987fcdeb-51a2-43d1-b456-426614174000
```

### Health Checks

#### Readiness Check
```bash
curl -s localhost:8080/healthz
```

Returns `200 OK` when all dependencies (database) are healthy:
```json
{"ok": true, "details": {"db": "OK"}}
```

Returns `503 Service Unavailable` when database is unreachable:
```json
{"ok": false, "details": {"db": "connection failed"}}
```

#### Liveness Check
```bash
curl -s localhost:8080/livez
```

Returns `200 OK` while the process is running:
```json
{"ok": true, "details": {"process.alive": "OK"}}
```

## Configuration

The service uses `configs/base.yaml` for configuration. Key settings:

```yaml
app:
  env: "dev"              # Environment (dev, staging, prod)

http:
  addr: ":8080"           # HTTP server listen address

databases:
  primary:
    driver: "postgres"          # Database driver (postgres, mysql, sqlite)
    dsn: "postgres://..."       # Database connection string
    max_open_conns: 25          # Maximum open connections
    max_idle_conns: 5           # Maximum idle connections
    conn_max_lifetime: "5m"     # Connection maximum lifetime
    conn_max_idle_time: "5m"    # Connection maximum idle time
    log_level: "warn"           # GORM log level (silent, error, warn, info)
    slow_threshold: "200ms"     # Slow query threshold
    skip_default_tx: false      # Skip default transaction
    prepare_stmt: true          # Prepare statements for better performance
```

## Health & Monitoring

### Database Health Monitoring

The service uses `gostratum/dbx` for database health monitoring:

1. **Auto-Migration**: GORM models are automatically migrated at startup
2. **Health Checks**: Database connectivity monitored via dbx health checks
3. **Connection Pooling**: Configurable connection pool settings
4. **Lifecycle Management**: Proper database connection lifecycle with graceful shutdown

### Health Check Behavior

- **Liveness** (`/livez`): Always returns `200 OK` while process runs
- **Readiness** (`/healthz`): Returns `503` when database is unreachable

### Graceful Shutdown

The service handles `SIGTERM` and `SIGINT` signals:

1. Stops accepting new requests
2. Completes in-flight requests (10s timeout)
3. Closes database connections
4. Exits cleanly

## Error Handling

The service maps internal errors to appropriate HTTP responses:

| Internal Error | HTTP Status | Response |
|---------------|-------------|----------|
| `ErrNotFound` | 404 Not Found | Resource not found |
| `ErrInvalid` | 400 Bad Request | Invalid input data |
| `ErrUnavailable` | 503 Service Unavailable | Database/network issue |

Database connectivity errors include `Retry-After: 2` header.

## Development

### Available Make Targets

```bash
make help             # Show all available targets
make run              # Run the service locally
make api              # Start API service (without migrations)
make migrate          # Run all pending database migrations
make migrate-down     # Rollback migrations (use STEPS=n)
make migrate-version  # Show current migration version
make dev              # Run migrations then start API
make build            # Build migration and API binaries
make docker-db        # Start PostgreSQL in Docker
make test             # Run tests
make fmt              # Format Go code
make vet              # Run go vet
make deps             # Download and tidy dependencies
```

For detailed migration commands, see [MIGRATIONS.md](MIGRATIONS.md).

### Project Structure

```
orderservice/
├── cmd/api/main.go              # Application entry point
├── configs/base.yaml            # Configuration file
├── internal/
│   ├── domain/                  # Business entities
│   │   ├── user.go             # User entity with validation
│   │   └── order.go            # Order entity with business logic
│   ├── usecase/                # Business logic layer
│   │   ├── interfaces.go       # Repository interfaces & errors
│   │   ├── create_user.go      # User creation logic
│   │   ├── get_user.go         # User retrieval logic
│   │   ├── create_order.go     # Order creation logic
│   │   └── get_order.go        # Order retrieval logic
│   └── adapter/                # External interfaces
│       ├── http/               # HTTP handlers
│       │   ├── routes.go       # Route registration
│       │   ├── user_handler.go # User HTTP handlers
│       │   └── order_handler.go # Order HTTP handlers
│       └── gorm/               # GORM database adapters
│           ├── user_repo.go    # User repository implementation
│           └── order_repo.go   # Order repository implementation
├── go.mod                      # Go module definition
└── README.md                   # This file
```

### Key Design Decisions

1. **Context Timeouts**: All operations have 800ms timeout
2. **Typed Errors**: Clean mapping between layers
3. **No Global State**: Everything injected via DI
4. **Health-First**: Comprehensive monitoring and recovery
5. **Production Ready**: Proper logging, error handling, graceful shutdown

## Testing the Service

### Scenario 1: Normal Operation

1. Start PostgreSQL and the service
2. Create a user → Should return 201 with user data
3. Get the user → Should return 200 with user data
4. Create an order for the user → Should return 201
5. Check health → Should return 200 OK

### Scenario 2: Database Failure

1. Start the service with PostgreSQL running
2. Stop PostgreSQL
3. Try to create/get resources → Should return 503
4. Check readiness → Should return 503
5. Restart PostgreSQL → Service should auto-recover
6. Check readiness → Should return 200 OK

### Scenario 3: Invalid Input

1. Try to create user without name → Should return 400
2. Try to create order with empty items → Should return 400
3. Try to get non-existent resource → Should return 404

## Troubleshooting

### Common Issues

#### Build Errors with go.work

If you encounter errors related to `go.work` file, use:

```bash
# For building
GOWORK=off go build ./cmd/api

# For running
GOWORK=off make run

# For testing
GOWORK=off make test
```

#### Database Connection Issues

1. **PostgreSQL not running**:
   ```bash
   # Check if PostgreSQL is running
   pg_isready -h localhost -p 5432 -U postgres -d orders
   
   # Start PostgreSQL
   docker-compose up -d postgres
   ```

2. **Database doesn't exist**:
   ```bash
   # Create database manually
   createdb -h localhost -U postgres orders
   
   # Run migrations
   make migrate
   ```

3. **Permission issues**:
   ```bash
   # Check PostgreSQL logs
   docker-compose logs postgres
   ```

#### Service Won't Start

1. **Port already in use**:
   ```bash
   # Check what's using port 8080
   lsof -i :8080
   
   # Change port in configs/base.yaml
   http:
     addr: ":8081"
   ```

2. **Missing dependencies**:
   ```bash
   # Download dependencies
   make deps
   ```

### Debugging

Enable debug logging by setting environment variable:
```bash
LOG_LEVEL=debug make run
```

View service logs in structured JSON format for better debugging.

## Dependencies

- **Core Framework**: `github.com/gostratum/core` - DI, config, logging, health
- **HTTP Framework**: `github.com/gostratum/httpx` - HTTP module with Gin integration
- **Database Framework**: `github.com/gostratum/dbx` - Database module with GORM integration
- **ORM**: `gorm.io/gorm` - Database ORM with auto-migration
- **Dependency Injection**: `go.uber.org/fx` - Application lifecycle
- **UUID Generation**: `github.com/google/uuid` - Unique identifiers

## License

This example is part of the gostratum project and follows the same license terms.
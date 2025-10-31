# Observability Demo

A comprehensive example demonstrating the full observability stack in gostratum with **modern configuration patterns**:
- **Metrics** - Prometheus metrics for HTTP requests and database queries
- **Tracing** - OpenTelemetry distributed tracing
- **Logging** - Structured logging with core/logx
- **Health Checks** - Database health monitoring with core.Registry
- **Configuration** - Using core/configx for unified configuration

## What's New in This Example

This example demonstrates the **updated gostratum patterns**:

### ✨ Modern Configuration (`core/configx`)
- Uses `core/configx.Loader` for configuration loading
- Database configuration follows the new `db.*` format
- Configuration prefix support via `Configurable` interface
- Environment variables with `STRATUM_` prefix

### ✨ Health Check Integration (`core.Registry`)
- Database health checks using `core.Registry`
- Both liveness and readiness probes
- Automatic health endpoint registration
- Kubernetes-ready health monitoring

### ✨ Dependency Injection Best Practices
- Uses `*gorm.DB` directly (not `Connections` map)
- Cleaner service constructors
- Better testability through proper DI

## Features

### Automatic HTTP Instrumentation
- Request count, duration, size metrics
- Distributed tracing with trace/span IDs
- Error tracking and status code distribution

### Automatic Database Instrumentation
- Query count, duration, errors
- Connection pool metrics
- Per-table and per-operation granularity

### RESTful API
- CRUD operations for User management
- JSON request/response
- Validation and error handling

## Architecture

```
┌─────────────────┐
│  HTTP Requests  │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────┐
│  httpx (with observability)     │
│  - MetricsMiddleware            │
│  - TracingMiddleware            │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│  UserHandler                    │
│  - Route handling               │
│  - Request validation           │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│  UserService                    │
│  - Business logic               │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│  dbx (with metrics)             │
│  - GORM MetricsPlugin           │
│  - Connection pool monitoring   │
└─────────────────────────────────┘
```

## Quick Start

### 1. Build and Run
```bash
go mod tidy
go run .
```

### Examples: modular vs monolithic

This example supports two composition styles in the same codebase:

- Modular (default): the `main.go` demonstrates composing the app via separate modules (preferred for multi-module workspaces).
- Monolithic: an alternate `main` is provided and enabled via the `monolith` build tag.

Run the default (modular) example:
```bash
go run .
```

Run the monolithic variant:
```bash
go run -tags=monolith .
```

### 2. Test the API

Create a user:
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'
```

List users:
```bash
curl http://localhost:8080/api/v1/users
```

Get a user:
```bash
curl http://localhost:8080/api/v1/users/1
```

Update a user:
```bash
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Doe", "email": "jane@example.com"}'
```

Delete a user:
```bash
curl -X DELETE http://localhost:8080/api/v1/users/1
```

### 3. View Metrics

Open http://localhost:9090/metrics in your browser to see Prometheus metrics:

**HTTP Metrics:**
```
http_requests_total{method="POST",path="/api/v1/users",status="201"} 1
http_request_duration_seconds_bucket{method="POST",path="/api/v1/users",le="0.1"} 1
http_request_size_bytes_bucket{method="POST",path="/api/v1/users",le="1024"} 1
http_response_size_bytes_bucket{method="POST",path="/api/v1/users",le="1024"} 1
```

**Database Metrics:**
```
db_queries_total{database="default",operation="create",table="users"} 1
db_query_duration_seconds_bucket{database="default",operation="select",table="users",le="0.01"} 5
db_connection_pool_open_connections{database="default"} 2
db_connection_pool_in_use_connections{database="default"} 0
```

### 4. View Traces (Optional)

To view distributed traces, run Jaeger:

```bash
docker run -d --name jaeger \
  -p 4317:4317 \
  -p 16686:16686 \
  jaegertracing/all-in-one:latest
```

Then open http://localhost:16686 to view traces in Jaeger UI.

## Configuration

This example uses the **new core/configx configuration pattern**:

### Database Configuration (New Format)
```yaml
db:
  default: primary       # Default database name
  databases:
    primary:
      driver: sqlite
      dsn: file:demo.db?cache=shared&mode=memory
      max_open_conns: 10
      max_idle_conns: 5
      conn_max_lifetime: 1h
      conn_max_idle_time: 10m
      log_level: info          # GORM log level
      slow_threshold: 200ms    # Slow query threshold
```

**Key Changes:**
- New structure: `db.databases.<name>` instead of `database.default`
- Added `log_level` and `slow_threshold` for better GORM configuration
- Implements `configx.Configurable` interface with prefix `"db"`

### Environment Variables
Override any config with environment variables using `STRATUM_` prefix:
```bash
export STRATUM_DB_DEFAULT=primary
export STRATUM_DB_DATABASES_PRIMARY_DSN="postgres://localhost/mydb"
export STRATUM_DB_DATABASES_PRIMARY_MAX_OPEN_CONNS=50
```

### Metrics
```yaml
metrics:
  enabled: true          # Toggle metrics on/off
  provider: prometheus
  prometheus:
    port: 9090          # Metrics endpoint port
    path: /metrics      # Metrics endpoint path
```

### Tracing
```yaml
tracing:
  enabled: true         # Toggle tracing on/off
  provider: otlp
  otlp:
    endpoint: localhost:4317  # OTLP collector endpoint
    insecure: true            # Use insecure connection
  service_name: observability-demo
  sample_rate: 1.0      # Sample 100% of traces
```

### Health Checks
The database module automatically registers health checks:
```bash
# Readiness check - Is the app ready to serve traffic?
curl http://localhost:8080/health/ready

# Liveness check - Is the app alive?
curl http://localhost:8080/health/live
```

**Example Response:**
```json
{
  "ok": true,
  "details": {
    "db-primary-readiness": {"ok": true, "error": ""},
    "db-primary-liveness": {"ok": true, "error": ""}
  }
}
```

## Code Patterns Demonstrated

### 1. Dependency Injection with *gorm.DB
```go
// ✅ New pattern - inject *gorm.DB directly
func NewUserService(db *gorm.DB, logger logx.Logger) (*UserService, error) {
    return &UserService{db: db, logger: logger}, nil
}

// ❌ Old pattern - inject Connections map
func NewUserService(conns dbx.Connections, logger logx.Logger) (*UserService, error) {
    conn, exists := conns["default"]
    if !exists {
        return nil, fmt.Errorf("default database connection not found")
    }
    return &UserService{db: conn, logger: logger}, nil
}
```

### 2. Health Check Integration
Health checks are automatically registered by the `dbx.Module()`:
```go
// Automatic registration - no manual code needed
app := fx.New(
    httpx.Module(),    // Provides HTTP server + health endpoints
    dbx.Module(),      // Automatically registers DB health checks
    // ...
)
```

### 3. Configuration Loading
Configuration is automatically loaded using `core/configx`:
```go
// In your config.yaml - uses new format
db:
  default: primary
  databases:
    primary:
      driver: sqlite
      # ... other settings
```

## Observability Features

### 1. Opt-In/Opt-Out

Observability is **optional** by default. To disable:

**Disable Metrics:**
```yaml
metrics:
  enabled: false
```

**Disable Tracing:**
```yaml
tracing:
  enabled: false
```

Or remove the modules from `main.go`:
```go
fx.New(
    core.Module,
    // metricsx.Module,  // ← Comment out
    // tracingx.Module,  // ← Comment out
    httpx.Module(),
    dbx.Module(),
)
```

### 2. Metrics Collected

**HTTP Metrics:**
- `http_requests_total` - Total HTTP requests (counter)
- `http_request_duration_seconds` - Request duration (histogram)
- `http_request_size_bytes` - Request body size (histogram)
- `http_response_size_bytes` - Response body size (histogram)
- `http_requests_in_flight` - Active requests (gauge)

**Database Metrics:**
- `db_queries_total` - Total queries (counter)
- `db_query_duration_seconds` - Query duration (histogram)
- `db_query_errors_total` - Query errors (counter)
- `db_queries_in_flight` - Active queries (gauge)
- `db_rows_affected` - Rows affected (histogram)
- `db_connection_pool_*` - Connection pool stats (gauge)

### 3. Trace Context Propagation

Each HTTP request includes trace headers:
```
X-Trace-ID: 80f198ee56343ba864fe8b2a57d3eff7
X-Span-ID: e457b5a2e4d86bd1
```

These can be used to correlate logs, metrics, and traces across services.

## Testing

Run tests:
```bash
go test -v ./...
```

## Next Steps

1. **Add Resilience Patterns** - Integrate circuit breakers and retry logic using `resiliencex`
2. **External API Calls** - Add `httpc` with automatic retry and circuit breaking
3. **Custom Metrics** - Add business-specific metrics using `metricsx.Metrics`
4. **Alerting** - Configure Prometheus AlertManager for metric-based alerts
5. **Dashboards** - Create Grafana dashboards for visualization

## License

MIT

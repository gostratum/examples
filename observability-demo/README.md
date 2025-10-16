# Observability Demo

A comprehensive example demonstrating the full observability stack in gostratum:
- **Metrics** - Prometheus metrics for HTTP requests and database queries
- **Tracing** - OpenTelemetry distributed tracing
- **Resilience** - Circuit breakers, retry, rate limiting (available via resiliencex)

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

Edit `config.yaml` to customize:

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

### Database
```yaml
database:
  default:
    driver: sqlite
    dsn: file:demo.db?cache=shared&mode=memory
    max_open_conns: 10
    max_idle_conns: 5
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

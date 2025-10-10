# Order Service Example - Updates for Storage & Response Modules

This document describes the updates made to integrate the new `storagex` and `responsex` modules into the order service example.

## Changes Made

### 1. Dependencies (`go.mod`)

**Added:**
- `github.com/gostratum/storagex v0.1.0` - Object storage abstraction with S3 support
- Import for S3 storage provider: `github.com/gostratum/storagex/internal/s3`

**Updated:**
- Added replace directive for storagex module pointing to local development version

### 2. HTTP Handlers

**Files Updated:**
- `internal/adapter/http/user_handler.go`
- `internal/adapter/http/order_handler.go`

**Changes:**
- Replaced `c.JSON()` calls with `responsex` functions:
  - `responsex.OK(c, data, nil)` for successful GET requests (HTTP 200)
  - `responsex.Created(c, "", data)` for successful POST requests (HTTP 201)
  - `responsex.Error(c, status, code, message, nil)` for error responses
- Added structured error codes (e.g., `USER_NOT_FOUND`, `INVALID_INPUT`, `SERVICE_UNAVAILABLE`)
- Improved error messages with consistent formatting

**Response Structure:**
All responses now follow the standardized envelope format:
```json
{
  "ok": true/false,
  "data": {...},           // on success
  "error": {               // on error
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": []
  },
  "pagination": {...},     // when applicable
  "meta": {
    "request_id": "...",
    "timestamp": "...",
    "duration_ms": 123,
    "server": "orderservice/v1.0.0"
  }
}
```

### 3. Routes & Middleware

**File Updated:**
- `internal/adapter/http/routes.go`

**Changes:**
- Added `responsex.MetaMiddleware("orderservice/v1.0.0")` to track requests
- Middleware automatically adds:
  - Request ID (X-Request-Id header)
  - Response metadata (timestamp, duration, server version)
  - Date header

### 4. Main Application

**File Updated:**
- `cmd/api/main.go`

**Changes:**
- Added `storagex.Module` to the fx dependency injection container
- Imported S3 storage provider (registers via init())
- StorageX is now available for injection into any service that needs object storage

**Note:** Due to AWS SDK dependency issues in the storagex module, the application may not compile until those are resolved. The storagex team is working on fixing the dependency versions.

### 5. Configuration

**File Updated:**
- `configs/base.yaml`

**Added StorageX Configuration:**
```yaml
storagex:
  provider: "s3"
  bucket: "orderservice-files"
  region: "us-east-1"
  # endpoint: "http://localhost:9000"  # For MinIO local testing
  # use_path_style: true                # Required for MinIO
  # disable_ssl: true                   # For local MinIO
  request_timeout: "30s"
  max_retries: 3
  backoff_initial: "200ms"
  backoff_max: "5s"
  default_part_size: 8388608   # 8MB for multipart uploads
  default_parallel: 4
  enable_logging: false
  # base_prefix: "orders/"     # Optional: prefix all keys
```

### 6. Tests

**File Updated:**
- `internal/adapter/http/user_handler_test.go`

**Changes:**
- Updated tests to expect `responsex.Envelope` structure
- Changed assertions to check `envelope.Ok`, `envelope.Data`, and `envelope.Error`
- Tests now validate the standardized response format

## Benefits

### ResponseX Module
1. **Standardized API Responses**: All endpoints return consistent envelope structure
2. **Better Error Handling**: Structured error codes and messages
3. **Request Tracking**: Automatic request ID generation and tracking
4. **Performance Metrics**: Response duration automatically tracked
5. **Client-Friendly**: Clients can always expect the same response structure

### StorageX Module  
1. **Unified Object Storage**: Single interface for S3, MinIO, and other providers
2. **Production-Ready**: Built-in retry logic, timeouts, and error handling
3. **Multipart Uploads**: Automatic handling of large file uploads
4. **Presigned URLs**: Generate temporary URLs for client-side uploads/downloads
5. **DI Integration**: Seamless integration with fx dependency injection

## Usage Examples

### Using ResponseX in Handlers

```go
// Success response
responsex.OK(c, user, nil)

// Created response with location
responsex.Created(c, "/users/123", user)

// Error response
responsex.Error(c, http.StatusBadRequest, "INVALID_INPUT", "Email is required", nil)

// Error with field details
details := []responsex.ErrDetail{
    {Field: "email", Message: "must be a valid email"},
}
responsex.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid input", details)

// With pagination
pg := &responsex.Pagination{
    Total:  ptr(int64(100)),
    Limit:  ptr(20),
    Offset: ptr(0),
}
responsex.OK(c, users, pg)
```

### Using StorageX for File Operations

```go
// Inject storage into your service
type FileService struct {
    storage storagex.Storage
}

func NewFileService(storage storagex.Storage) *FileService {
    return &FileService{storage: storage}
}

// Upload a file
func (s *FileService) UploadFile(ctx context.Context, key string, data []byte) error {
    opts := &storagex.PutOptions{
        ContentType: "application/pdf",
        Metadata: map[string]string{
            "user_id": "123",
        },
    }
    
    _, err := s.storage.PutBytes(ctx, key, data, opts)
    return err
}

// Download a file
func (s *FileService) DownloadFile(ctx context.Context, key string) ([]byte, error) {
    reader, _, err := s.storage.Get(ctx, key)
    if err != nil {
        return nil, err
    }
    defer reader.Close()
    
    return io.ReadAll(reader)
}

// Generate presigned URL (for client-side uploads)
func (s *FileService) GetUploadURL(ctx context.Context, key string) (string, error) {
    opts := &storagex.PresignOptions{
        Expiry:      15 * time.Minute,
        ContentType: "image/jpeg",
    }
    
    return s.storage.PresignPut(ctx, key, opts)
}
```

## Next Steps

1. **Resolve StorageX Dependencies**: The storagex module needs AWS SDK dependency versions fixed
2. **Add File Upload Endpoints**: Create endpoints for file uploads using storagex
3. **Update E2E Tests**: Add tests for new response envelope structure
4. **Add Pagination**: Implement pagination for list endpoints using responsex.Pagination
5. **Rate Limiting**: Use responsex.WithRateLimit for API rate limiting

## Migration Guide for Existing Code

If you have existing handlers, here's how to migrate:

### Before (Old Style):
```go
c.JSON(http.StatusOK, user)
c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
```

### After (New Style):
```go
responsex.OK(c, user, nil)
responsex.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user not found", nil)
```

### Client-Side Changes:
Clients need to extract data from the envelope:
```javascript
// Old
const user = await response.json()

// New
const envelope = await response.json()
if (envelope.ok) {
    const user = envelope.data
} else {
    const error = envelope.error
    console.error(`${error.code}: ${error.message}`)
}
```

## Known Issues

1. **StorageX Compilation**: AWS SDK dependency resolution issue prevents compilation
   - Error: `unknown revision aws/v1.30.3`
   - Status: Under investigation by storagex team
   
2. **Workaround**: Comment out storagex imports in `cmd/api/main.go` to compile without storage support:
   ```go
   // Comment these lines temporarily:
   // "github.com/gostratum/storagex/pkg/storagex"
   // _ "github.com/gostratum/storagex/internal/s3"
   // And remove storagex.Module from core.New()
   ```

## Testing

Run tests with:
```bash
go test ./internal/adapter/http/... -v
```

Note: Tests are updated to work with the new response envelope structure.

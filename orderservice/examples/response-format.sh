#!/bin/bash

# Example API Requests with New Response Format
# This demonstrates the updated responsex envelope structure

echo "=== Testing ResponseX Integration ==="
echo ""

# Note: These are example requests. To actually test, start the server with:
# make run

echo "1. Create User (Success - 201 Created)"
echo "Request:"
echo "  POST /api/v1/users"
echo '  Body: {"name": "John Doe", "email": "john@example.com"}'
echo ""
echo "Response:"
cat << 'EOF'
{
  "ok": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2025-10-10T10:17:26Z"
  },
  "meta": {
    "request_id": "req_abc123",
    "timestamp": "2025-10-10T10:17:26Z",
    "duration_ms": 45,
    "server": "orderservice/v1.0.0"
  }
}
EOF
echo ""
echo "---"
echo ""

echo "2. Get User (Success - 200 OK)"
echo "Request:"
echo "  GET /api/v1/users/550e8400-e29b-41d4-a716-446655440000"
echo ""
echo "Response:"
cat << 'EOF'
{
  "ok": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2025-10-10T10:17:26Z"
  },
  "meta": {
    "request_id": "req_def456",
    "timestamp": "2025-10-10T10:17:30Z",
    "duration_ms": 12,
    "server": "orderservice/v1.0.0"
  }
}
EOF
echo ""
echo "---"
echo ""

echo "3. Invalid Request (Error - 400 Bad Request)"
echo "Request:"
echo "  POST /api/v1/users"
echo '  Body: {"name": "", "email": "invalid"}'
echo ""
echo "Response:"
cat << 'EOF'
{
  "ok": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "invalid request payload",
    "details": []
  },
  "meta": {
    "request_id": "req_ghi789",
    "timestamp": "2025-10-10T10:17:35Z",
    "duration_ms": 5,
    "server": "orderservice/v1.0.0"
  }
}
EOF
echo ""
echo "---"
echo ""

echo "4. Not Found (Error - 404 Not Found)"
echo "Request:"
echo "  GET /api/v1/users/non-existent-id"
echo ""
echo "Response:"
cat << 'EOF'
{
  "ok": false,
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "user not found",
    "details": []
  },
  "meta": {
    "request_id": "req_jkl012",
    "timestamp": "2025-10-10T10:17:40Z",
    "duration_ms": 8,
    "server": "orderservice/v1.0.0"
  }
}
EOF
echo ""
echo "---"
echo ""

echo "✅ All responses now use the standardized envelope format!"
echo ""
echo "Key Benefits:"
echo "  - Consistent structure across all endpoints"
echo "  - Built-in request tracking (request_id)"
echo "  - Performance metrics (duration_ms)"
echo "  - Clear success/error indication (ok field)"
echo "  - Structured error codes for client handling"
echo "  - Server version tracking"
echo ""

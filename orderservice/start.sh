#!/bin/bash

# Startup script for the order service
# This script sets up the environment and starts the service

set -e

echo "🚀 Starting Order Service..."

# Set environment variables
export APP_ENV=dev
export CONFIG_PATHS=./configs
export GOWORK=off

# Check if PostgreSQL is running
echo "📡 Checking PostgreSQL connection..."
if ! pg_isready -h localhost -p 5432 -U postgres -d orders >/dev/null 2>&1; then
    echo "❌ PostgreSQL is not running or not accessible"
    echo "💡 Start PostgreSQL with: docker-compose up -d postgres"
    echo "💡 Or use: make docker-db"
    exit 1
fi

echo "✅ PostgreSQL is running"

# Download dependencies if needed
if [ ! -f "go.sum" ] || [ "go.mod" -nt "go.sum" ]; then
    echo "📦 Downloading dependencies..."
    go mod tidy
fi

# Build and run the service
echo "🔨 Building service..."
go build -o bin/orderservice ./cmd/api

echo "🎯 Starting service on :8080..."
echo "📋 Health checks available at:"
echo "   - http://localhost:8080/healthz (readiness)"
echo "   - http://localhost:8080/livez (liveness)"
echo ""

./bin/orderservice
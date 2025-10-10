.PHONY: help run build clean docker-db migrate status api dev test fmt vet

# Default target
help:
	@echo "Available targets:"
	@echo "  run        - Run the service locally (legacy)"
	@echo "  api        - Start the API service (without migrations)"
	@echo "  migrate    - Run database migrations"
	@echo "  status     - Check migration status"
	@echo "  dev        - Run migrations then start API (development)"
	@echo "  build      - Build both migration and API binaries"
	@echo "  clean      - Clean build artifacts"
	@echo "  docker-db  - Start PostgreSQL in Docker"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format Go code"
	@echo "  vet        - Run go vet"

# Run the service locally
run:
	@echo "Starting order service..."
	APP_ENV=dev CONFIG_PATHS=./configs GOWORK=off go run ./cmd/api

# Build both binaries
build:
	@echo "Building migration and API binaries..."
	@mkdir -p bin
	GOWORK=off go build -o bin/migrate ./cmd/migrations
	GOWORK=off go build -o bin/api ./cmd/api
	@echo "✅ Build completed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/

# Start PostgreSQL in Docker
docker-db:
	@echo "Starting PostgreSQL in Docker..."
	docker run --rm --name postgres-orders \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_DB=orders \
		-p 5432:5432 \
		postgres:16

# Run database migrations
migrate:
	@echo "🔄 Running database migrations..."
	APP_ENV=dev CONFIG_PATHS=./configs GOWORK=off go run ./cmd/migrations -action=migrate

# Check migration status  
status:
	@echo "📋 Checking migration status..."
	APP_ENV=dev CONFIG_PATHS=./configs GOWORK=off go run ./cmd/migrations -action=status

# Start API service (assumes migrations have been run)
api:
	@echo "🚀 Starting API service..."
	APP_ENV=dev CONFIG_PATHS=./configs GOWORK=off go run ./cmd/api

# Development workflow: migrate then start API
dev: migrate
	@echo "🏃‍♂️ Starting development server..."
	APP_ENV=dev CONFIG_PATHS=./configs GOWORK=off go run ./cmd/api

# Run tests
test:
	@echo "Running tests..."
	GOWORK=off go test -v ./...

# Format Go code
fmt:
	@echo "Formatting Go code..."
	GOWORK=off go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	GOWORK=off go vet ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	GOWORK=off go mod download
	GOWORK=off go mod tidy
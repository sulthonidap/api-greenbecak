.PHONY: help build run test clean seed migrate

# Default target
help:
	@echo "Available commands:"
	@echo "  build    - Build the application"
	@echo "  run      - Run the application in development mode"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  seed     - Seed database with initial data"
	@echo "  migrate  - Run database migrations"
	@echo "  docker   - Build and run with Docker"

# Build the application
build:
	@echo "Building GreenBecak Backend..."
	go build -o greenbecak-backend main.go

# Run the application
run:
	@echo "Running GreenBecak Backend..."
	go run main.go

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f greenbecak-backend
	rm -f greenbecak-backend.exe

# Seed database
seed:
	@echo "Seeding database..."
	go run scripts/seed.go

# Run database migrations (auto-migrate is enabled in main.go)
migrate:
	@echo "Running database migrations..."
	go run main.go

# Build for different platforms
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o greenbecak-backend-linux main.go

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o greenbecak-backend.exe main.go

build-macos:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -o greenbecak-backend-macos main.go

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run

# Development setup
dev-setup: deps seed
	@echo "Development setup completed!"

# Production build
prod-build: build-linux
	@echo "Production build completed!"

# Database backup commands
backup:
	@echo "Creating database backup..."
	go run cmd/backup/main.go -backup

backup-list:
	@echo "Listing backup files..."
	go run cmd/backup/main.go -list

backup-cleanup:
	@echo "Cleaning up old backup files..."
	go run cmd/backup/main.go -cleanup=7

backup-restore:
	@echo "Usage: make backup-restore file=backup_file.sql"
	@if [ -z "$(file)" ]; then echo "Please specify backup file"; exit 1; fi
	go run cmd/backup/main.go -restore=$(file)

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t greenbecak-backend .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env greenbecak-backend

docker-compose-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-compose-down:
	@echo "Stopping services..."
	docker-compose down

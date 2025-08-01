.PHONY: build test run clean docker-build docker-run help

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the gateway binary"
	@echo "  test         - Run tests"
	@echo "  run          - Run the gateway locally"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  deps         - Download dependencies"

# Build the gateway binary
build:
	@echo "Building gateway binary..."
	go build -o cmd/gateway/api_gateway cmd/gateway/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./tests/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./tests/...

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. ./tests/...

# Run the gateway locally
run:
	@echo "Running gateway locally..."
	go run cmd/gateway/main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f cmd/gateway/api_gateway
	rm -f backend/backend1
	rm -rf tmp/

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -f docker/Dockerfile -t api-gateway .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	cd docker && docker-compose up -d

# Stop Docker Compose
docker-stop:
	@echo "Stopping Docker Compose services..."
	cd docker && docker-compose down

# View Docker logs
docker-logs:
	@echo "Viewing Docker logs..."
	cd docker && docker-compose logs -f

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest 
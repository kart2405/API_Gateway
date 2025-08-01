#!/bin/bash

# Development setup script for API Gateway

set -e

echo "Setting up development environment for API Gateway..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.24.2 or higher."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "Go version: $GO_VERSION"

# Download dependencies
echo "Downloading Go dependencies..."
go mod download

# Create necessary directories
echo "Creating necessary directories..."
mkdir -p tmp
mkdir -p logs

# Set up environment variables
echo "Setting up environment variables..."
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=2405
export DB_NAME=apigateway
export REDIS_ADDR=localhost:6379

echo "Development environment setup complete!"
echo ""
echo "Next steps:"
echo "1. Start PostgreSQL: docker run -d --name postgres -e POSTGRES_PASSWORD=2405 -e POSTGRES_DB=apigateway -p 5432:5432 postgres:14"
echo "2. Start Redis: docker run -d --name redis -p 6379:6379 redis:7"
echo "3. Initialize database: ./scripts/init-db.sh"
echo "4. Run the gateway: make run"
echo ""
echo "Or use Docker Compose:"
echo "cd docker && docker-compose up -d" 
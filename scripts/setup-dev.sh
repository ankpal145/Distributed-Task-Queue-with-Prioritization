#!/bin/bash

# Development environment setup script

set -e

echo "Setting up development environment..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Warning: Docker is not installed. You'll need it for running the full stack."
fi

# Check if Redis is running
if ! redis-cli ping &> /dev/null; then
    echo "Warning: Redis is not running. Starting Redis with Docker..."
    docker run -d -p 6379:6379 --name taskqueue-redis redis:7-alpine
fi

# Install Go tools
echo "Installing Go tools..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Download dependencies
echo "Downloading dependencies..."
go mod download
go mod tidy

# Generate protobuf code
echo "Generating protobuf code..."
./scripts/generate-proto.sh

# Build binaries
echo "Building binaries..."
make build

echo "Development environment setup complete!"
echo ""
echo "To start the system:"
echo "  1. Start the server: make run-server"
echo "  2. Start workers: make run-worker (in separate terminals)"
echo "  3. Access dashboard: http://localhost:8080"
echo ""
echo "Or use Docker Compose:"
echo "  make docker-up"

.PHONY: help build run test clean docker-build docker-up docker-down proto

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build server and worker binaries"
	@echo "  run-server    - Run the server"
	@echo "  run-worker    - Run a worker"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-up     - Start all services with Docker Compose"
	@echo "  docker-down   - Stop all services"
	@echo "  proto         - Generate protobuf code"
	@echo "  deps          - Download dependencies"
	@echo "  lint          - Run linters"

# Build binary
build:
	@echo "Building taskqueue binary..."
	@go build -o bin/taskqueue .
	@echo "Build complete!"

# Run in server mode
run-server:
	@echo "Starting server..."
	@go run main.go server

# Run in worker mode
run-worker:
	@echo "Starting worker..."
	@go run main.go worker

# Run in both modes
run-both:
	@echo "Starting server and worker together..."
	@go run main.go both

# Quick run (default - both mode)
run:
	@echo "Starting task queue system..."
	@go run main.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t taskqueue:latest .
	@echo "Docker image built!"

# Start all services
docker-up:
	@echo "Starting services..."
	@docker-compose up -d
	@echo "Services started!"
	@echo "Dashboard: http://localhost:8080"
	@echo "Metrics: http://localhost:2112/metrics"
	@echo "Grafana: http://localhost:3000 (admin/admin)"

# Stop all services
docker-down:
	@echo "Stopping services..."
	@docker-compose down
	@echo "Services stopped!"

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/task_queue.proto
	@echo "Protobuf code generated!"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies downloaded!"

# Run linters
lint:
	@echo "Running linters..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed"; exit 1; }
	@golangci-lint run ./...
	@echo "Linting complete!"

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Development environment ready!"


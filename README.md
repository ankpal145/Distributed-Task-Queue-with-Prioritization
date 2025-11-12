# Distributed Task Queue with Prioritization

A production-grade, high-throughput, fault-tolerant distributed task queue system built with Go, Redis, gRPC, and REST APIs.

## 🏗️ Architecture

This project follows **Clean Architecture** principles with a beacon-style consumer pattern:

```
├── cmd/                          # Application entry points
│   ├── main.go                  # Unified CLI (server/worker modes)
│   ├── server/                  # Server initialization
│   └── worker/                  # Worker initialization
├── config/                      # Configuration management (Viper)
├── internal/                    # Application code
│   ├── constants/              # Centralized constants
│   ├── entity/                 # Domain entities (Task, Worker)
│   ├── handler/                # Request handlers (REST + gRPC)
│   ├── service/                # Business logic layer
│   ├── repo/                   # Data access layer
│   ├── worker/                 # Worker management
│   │   ├── worker_service.go  # Worker entity management
│   │   ├── consumer/          # Beacon-style consumer
│   │   │   └── consumer_processor.go
│   │   └── processors/        # Task processors
│   │       ├── provider.go    # Consolidated DI provider
│   │       ├── processors.go  # Processor registry
│   │       ├── default_task_processor.go
│   │       └── high_priority_processor.go
│   ├── monitoring/             # Metrics & observability
│   └── router/                 # Route registration
├── pkg/                        # Reusable packages
│   ├── logger/                # Context-aware structured logging
│   ├── middleware/            # gRPC/HTTP middleware
│   ├── redis/                 # Redis client
│   ├── errors/                # Error handling
│   └── contexts/              # Context utilities
├── proto/                      # Protocol buffer definitions
└── docs/                       # Design documentation
```

## ✨ Key Features

- **🚀 High Performance**: Built with Go for maximum throughput and low latency
- **📊 Priority-based Queuing**: Support for Critical, High, Medium, and Low priorities
- **🔄 Fault Tolerance**: Automatic retries with exponential backoff and jitter
- **💀 Dead Letter Queue**: Automatic handling of permanently failed tasks
- **👥 Worker Pool Management**: Dynamic worker registration and health monitoring
- **🔌 Dual API**: Unified port for REST and gRPC (via cmux)
- **📈 Observability**: Built-in Prometheus metrics and health checks
- **🎯 Clean Architecture**: Layered design with proper separation of concerns
- **💉 Dependency Injection**: Clean DI with uber-go/fx
- **🔐 Middleware**: Authentication, logging, recovery, request context
- **📝 Structured Logging**: Context-aware logging throughout

## 🚀 Quick Start

### Prerequisites
- **Go**: 1.21 or higher
- **Redis**: 6.0 or higher  
- **Docker**: (optional) for containerized deployment

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/Distributed-Task-Queue-with-Prioritization.git
cd Distributed-Task-Queue-with-Prioritization

# Install dependencies
go mod download

# Build the binary
go build -o bin/taskqueue ./cmd
```

### Running the Application

#### Start Server Mode
```bash
./bin/taskqueue server config.yaml
```
- Starts REST + gRPC APIs on port 9090
- Exposes metrics on `/metrics`
- Health check available at `/healthz`

#### Start Worker Mode
```bash
./bin/taskqueue worker config.yaml
```
- Starts task processor worker
- Consumes tasks from Redis priority queues
- Reports metrics to Prometheus

#### Using Docker Compose
```bash
# Start all services (Redis + Server + Worker)
docker-compose up

# Start specific service
docker-compose up server
docker-compose up worker
```

## 📋 Configuration

Edit `config.yaml`:

```yaml
server:
  grpc_port: 9090        # Unified port for HTTP + gRPC

redis:
  host: localhost
  port: 6379
  db: 0
  password: ""
  pool_size: 10

worker:
  concurrency: 10
  heartbeat_interval: 5s
  worker_timeout: 30s
  cleanup_interval: 1m

task:
  default_timeout: 5m
  max_retries: 3
  retry_base_delay: 1s
  retry_max_delay: 5m

metrics:
  enabled: true
  port: 9091

logger:
  level: info
  format: json
```

## 📊 API Examples

### Submit a Task (REST)

```bash
curl -X POST http://localhost:9090/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type": "order_processing",
    "priority": "high",
    "payload": {"order_id": "12345", "amount": 99.99},
    "max_retries": 3,
    "timeout": "5m"
  }'
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "order_processing",
  "priority": "high",
  "status": "pending",
  "created_at": "2025-01-08T10:00:00Z"
}
```

### Get Task Status

```bash
curl http://localhost:9090/api/v1/tasks/550e8400-e29b-41d4-a716-446655440000
```

### List Workers

```bash
curl http://localhost:9090/api/v1/workers
```

### Get Metrics

```bash
curl http://localhost:9090/metrics
```

See [examples/](./examples/) directory for more examples.

## 🎯 Architecture Highlights

### Beacon-Style Consumer Pattern

The worker implementation follows the beacon pattern with:
- **Consumer Processor**: Manages worker lifecycle and task consumption
- **Task Processors**: Separate processors for different priorities
- **Registry Pattern**: Centralized processor registry
- **Panic Recovery**: Graceful error handling with recovery
- **Context Management**: Proper context propagation and timeouts
- **Health Monitoring**: Continuous worker health checks

### Clean Dependency Injection

All dependencies are managed via a **single consolidated provider**:
```go
// internal/worker/processors/provider.go
var FxWorkerProcessorsProvider = fx.Options(
    // Worker service providers
    fx.Provide(worker.NewWorkerConfigFromConfig),
    fx.Provide(worker.NewWorkerService),
    
    // Task processor providers
    fx.Provide(NewTaskProcessors),
    fx.Provide(NewDefaultTaskProcessor),
    fx.Provide(NewHighPriorityTaskProcessor),
)
```

### Centralized Constants

All magic strings and numbers are defined in `internal/constants/constants.go`:
- Queue names
- Status strings
- HTTP endpoints
- Metric names
- Sleep durations
- Default values

## 📈 Monitoring & Observability

### Prometheus Metrics

The system exposes comprehensive metrics:

**Task Metrics:**
- `tasks_submitted_total` - Total tasks submitted
- `tasks_processed_total` - Total tasks processed successfully
- `tasks_failed_total` - Total failed tasks
- `tasks_completed_total` - Total completed tasks
- `task_duration_seconds` - Task processing duration histogram
- `tasks_dead_total` - Tasks moved to DLQ

**Worker Metrics:**
- `workers_active` - Number of active workers
- `worker_heartbeats_total` - Worker heartbeat count
- `workers_cleaned_up_total` - Dead workers cleaned up

**System Metrics:**
- Available at `http://localhost:9090/metrics`

### Health Checks

- **Liveness**: `http://localhost:3000/healthz` (worker)
- **Readiness**: `http://localhost:9090/health` (server)

## 🧪 Development

### Build Commands

```bash
# Build the main binary
go build -o bin/taskqueue ./cmd

# Run go vet
go vet ./...

# Format code
go fmt ./...

# Run tests
go test ./... -v
```

### Generate Protocol Buffers

```bash
./scripts/generate-proto.sh
```

### Development Setup

```bash
# Install development dependencies
./scripts/setup-dev.sh

# Run locally
./bin/taskqueue server config.yaml
./bin/taskqueue worker config.yaml
```

## 📚 Documentation

- **[Quick Reference](./QUICK_REFERENCE.md)** - Commands and structure guide
- **[API Documentation](./docs/API.md)** - Complete REST/gRPC API specs
- **[High Level Design](./docs/HLD.md)** - Architecture overview
- **[Low Level Design](./docs/LLD.md)** - Implementation details
- **[Protobuf Setup](./docs/PROTOBUF_SETUP.md)** - Protocol buffer guide

## 🏗️ Project Structure Philosophy

This project adheres to:
- ✅ **Clean Architecture**: Clear separation of concerns
- ✅ **Domain-Driven Design**: Entity-centric modeling
- ✅ **SOLID Principles**: Maintainable and extensible code
- ✅ **Dependency Injection**: Testable and decoupled components
- ✅ **Observability First**: Metrics and logging built-in
- ✅ **Production Ready**: Error handling, retries, DLQ

## 🐳 Docker Deployment

### Build Image

```bash
docker build -t task-queue:latest .
```

### Run with Docker Compose

```yaml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
  
  server:
    image: task-queue:latest
    command: ["./bin/taskqueue", "server", "config.yaml"]
    ports:
      - "9090:9090"
    depends_on:
      - redis
  
  worker:
    image: task-queue:latest
    command: ["./bin/taskqueue", "worker", "config.yaml"]
    depends_on:
      - redis
    deploy:
      replicas: 3
```

## 🚢 Kubernetes Deployment

See [deployment/kubernetes/](./deployment/kubernetes/) for complete manifests.

```bash
# Apply manifests
kubectl apply -f deployment/kubernetes/

# Scale workers
kubectl scale deployment task-queue-worker --replicas=10
```

## 🧪 Testing

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🙏 Acknowledgments

- Inspired by Celery (Python) and Sidekiq (Ruby)
- Built with best practices from the Go community
- Follows the beacon-style architecture pattern

## 📞 Support

For questions and support:
- Open an issue on GitHub
- Check the [documentation](./docs/)
- Review [examples](./examples/)

---

**Built with ❤️ using Go, Redis, and Clean Architecture principles**

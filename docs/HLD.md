# High-Level Design (HLD) - Distributed Task Queue System

## 1. System Overview

### 1.1 Purpose
A production-grade distributed task queue system designed for high throughput, fault tolerance, and horizontal scalability. Built from scratch in Go with clean architecture and beacon-style consumer pattern.

### 1.2 Key Features
- **Priority-based Task Queuing**: Support for multiple priority levels (critical, high, medium, low)
- **Worker Pool Management**: Dynamic worker registration, health monitoring, and graceful shutdown
- **Beacon-Style Consumer**: Production-ready consumer pattern with processor architecture
- **Fault Tolerance**: Automatic retries with exponential backoff and jitter
- **Dead Letter Queue (DLQ)**: Automatic handling of permanently failed tasks
- **Real-time Monitoring**: Prometheus metrics for latency, throughput, and system health
- **Dual API Support**: REST and gRPC on unified port (via cmux)
- **Persistence**: Redis-backed storage for durability and performance
- **Clean Architecture**: Layered design with proper separation of concerns
- **Structured Logging**: Context-aware logging throughout the application
- **Constants-Driven**: Centralized constants for maintainability

## 2. Architecture Overview

```
┌───────────────────────────────────────────────────────────────────┐
│                      CLIENT APPLICATIONS                          │
└──────────────────────────┬────────────────────────────────────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │   UNIFIED API PORT   │
                │   (Port 9090)        │
                │   ┌────────┬────────┐│
                │   │  REST  │  gRPC  ││
                │   │  (Gin) │(Proto) ││
                │   └────────┴────────┘│
                │      (via cmux)      │
                └──────────┬───────────┘
                           │
                ┌──────────┴──────────────────────┐
                │         SERVER MODE             │
                │  ┌──────────────────────────┐   │
                │  │   HANDLER LAYER          │   │
                │  │  - REST Handler          │   │
                │  │  - gRPC Server           │   │
                │  │  - Middleware            │   │
                │  └─────────┬────────────────┘   │
                │            │                    │
                │  ┌─────────▼────────────────┐   │
                │  │   SERVICE LAYER          │   │
                │  │  - Task Service          │   │
                │  │  - Worker Service        │   │
                │  │  - Metrics Collection    │   │
                │  └─────────┬────────────────┘   │
                │            │                    │
                │  ┌─────────▼────────────────┐   │
                │  │   REPOSITORY LAYER       │   │
                │  │  - Task Repository       │   │
                │  │  - Worker Repository     │   │
                │  └─────────┬────────────────┘   │
                └────────────┼────────────────────┘
                             │
              ┌──────────────▼──────────────┐
              │      REDIS CLUSTER          │
              │  ┌──────────────────────┐   │
              │  │  Priority Queues     │   │
              │  │  - queue:critical    │   │
              │  │  - queue:high        │   │
              │  │  - queue:medium      │   │
              │  │  - queue:low         │   │
              │  └──────────────────────┘   │
              │  ┌──────────────────────┐   │
              │  │  Dead Letter Queue   │   │
              │  │  - queue:dlq         │   │
              │  └──────────────────────┘   │
              │  ┌──────────────────────┐   │
              │  │  Processing Tasks    │   │
              │  │  - queue:processing  │   │
              │  └──────────────────────┘   │
              │  ┌──────────────────────┐   │
              │  │  Worker Registry     │   │
              │  │  - workers:*         │   │
              │  └──────────────────────┘   │
              └──────────────┬──────────────┘
                             │
                ┌────────────┴────────────┐
                │                         │
                ▼                         ▼
     ┌──────────────────┐      ┌──────────────────┐
     │  WORKER POD 1    │      │  WORKER POD N    │
     │  ┌────────────┐  │      │  ┌────────────┐  │
     │  │  Consumer  │  │      │  │  Consumer  │  │
     │  │ Processor  │  │ .... │  │ Processor  │  │
     │  └──┬─────────┘  │      │  └──┬─────────┘  │
     │     │            │      │     │            │
     │  ┌──▼─────────┐  │      │  ┌──▼─────────┐  │
     │  │ Task       │  │      │  │ Task       │  │
     │  │ Processors │  │      │  │ Processors │  │
     │  │  - Default │  │      │  │  - Default │  │
     │  │  - HighPri │  │      │  │  - HighPri │  │
     │  └────────────┘  │      │  └────────────┘  │
     └──────────────────┘      └──────────────────┘
              │                         │
              └────────────┬────────────┘
                           ▼
                ┌──────────────────────┐
                │  MONITORING          │
                │  - Prometheus        │
                │  - Metrics @ :9090   │
                │  - Health Checks     │
                └──────────────────────┘
```

### 2.1 Application Modes

The system runs in **two distinct modes**:

1. **Server Mode** (`./bin/taskqueue server`)
   - Exposes REST and gRPC APIs
   - Handles task submission and queries
   - Manages worker registration
   - Serves metrics endpoint

2. **Worker Mode** (`./bin/taskqueue worker`)
   - Consumes tasks from Redis queues
   - Executes task processing logic
   - Reports heartbeat to server
   - Handles task success/failure

## 3. Core Components

### 3.1 API Layer
**Responsibility**: External interface for clients to interact with the system

**Components**:
- **REST API Server** (Gin framework)
  - Task submission endpoints
  - Task status queries
  - Worker management
  - System metrics
  
- **gRPC Server**
  - High-performance binary protocol
  - Streaming support for real-time updates
  - Service definitions via Protocol Buffers

**Endpoints**:
```
REST:
POST   /api/v1/tasks              - Submit new task
GET    /api/v1/tasks/:id          - Get task status
DELETE /api/v1/tasks/:id          - Cancel task
GET    /api/v1/tasks              - List tasks (with filters)
POST   /api/v1/workers/register   - Register worker
GET    /api/v1/workers            - List workers
GET    /api/v1/metrics            - System metrics
GET    /api/v1/health             - Health check

gRPC:
SubmitTask(TaskRequest) -> TaskResponse
GetTaskStatus(TaskID) -> TaskStatus
StreamTasks(StreamRequest) -> stream TaskUpdate
RegisterWorker(WorkerInfo) -> WorkerID
```

### 3.2 Service Layer
**Responsibility**: Business logic and orchestration

**Components**:
- **Task Service** (`internal/service/task_service.go`)
  - Task creation and validation
  - Priority assignment
  - Retry logic with exponential backoff
  - DLQ management
  - Task status updates
  
- **Worker Service** (`internal/worker/worker_service.go`)
  - Worker entity management (register, deregister, list)
  - Worker health monitoring
  - Dead worker cleanup
  - Worker statistics
  
- **Monitoring Service** (`internal/monitoring/prometheus.go`)
  - Metrics collection and aggregation
  - Prometheus integration
  - Performance analytics

### 3.3 Repository Layer
**Responsibility**: Data persistence and retrieval

**Components**:
- **Task Repository**
  - CRUD operations for tasks
  - Priority queue management
  - Atomic operations for task state transitions
  
- **Worker Repository**
  - Worker registry management
  - Heartbeat tracking
  - Worker capacity management

### 3.4 Redis Data Store
**Responsibility**: Fast, reliable data storage

**Data Structures**:
- **Sorted Sets** for priority queues (score = priority + timestamp)
- **Hashes** for task metadata and worker information
- **Lists** for dead letter queue
- **Pub/Sub** for real-time notifications
- **TTL** for automatic cleanup of stale data

### 3.5 Worker Consumer Pattern (Beacon-Style)
**Responsibility**: Task consumption and processing

**Components**:

1. **Consumer Processor** (`internal/worker/consumer/consumer_processor.go`)
   - Worker lifecycle management (start/stop)
   - Task consumption loop
   - Heartbeat management
   - Dead worker cleanup
   - Panic recovery

2. **Task Processors** (`internal/worker/processors/`)
   - **Default Task Processor**: Handles low, medium, default priority tasks
   - **High Priority Processor**: Handles high and critical priority tasks
   - **Processor Registry**: Centralized registry of all processors
   - **Singleton Pattern**: Using `sync.Once` for processor instantiation

3. **Provider** (`internal/worker/processors/provider.go`)
   - Consolidated Fx dependency injection provider
   - Provides all worker-related services and processors

**Key Features**:
- Context-aware processing with timeouts
- Panic recovery at task level
- Request ID propagation
- Structured logging throughout
- Metrics reporting for all operations

## 4. Key Workflows

### 4.1 Task Submission Flow
```
1. Client submits task via REST/gRPC
2. API layer validates request
3. Task Service creates task with unique ID
4. Task assigned priority and added to appropriate queue
5. Task metadata stored in Redis hash
6. Response returned with task ID
7. Monitoring service records submission metric
```

### 4.2 Task Processing Flow (Beacon-Style)
```
1. Consumer Processor starts task consumption loop
2. Consumer polls Redis for tasks (priority order: critical > high > medium > low)
3. Task dequeued atomically (ZPOPMIN)
4. Task routed to appropriate Task Processor based on priority
5. Task Processor creates context with timeout and request ID
6. Task execution with panic recovery:
   a. Unmarshal task payload
   b. Execute business logic
   c. Marshal result
7. On success:
   - Task marked as completed
   - Result reported back to task service
   - Metrics updated (tasks_processed_total)
8. On failure:
   - Error logged with context
   - Task reported as failed
   - Task service handles retry logic:
     * Increment retry count
     * If retries remaining:
       - Calculate exponential backoff with jitter
       - Re-queue with delay
     * Else:
       - Move to Dead Letter Queue
       - Update metrics (tasks_dead_total)
9. Worker updates heartbeat every 5 seconds
10. Loop continues until shutdown signal
```

### 4.3 Worker Registration Flow
```
1. Worker pod starts (cmd/worker/worker.go)
2. Consumer Processor initialized with:
   - Task Service
   - Worker Service
   - Task Processors (registry)
   - Logger and Metrics
3. Worker registers with server:
   - Unique worker ID assigned
   - Worker metadata stored (queues, concurrency)
   - Status set to "active"
4. Worker starts three concurrent operations:
   a. Task consumption loop (startTaskConsumer)
   b. Heartbeat loop (updates every 5s)
   c. Cleanup worker (removes dead workers every 1m)
5. System monitors worker health via heartbeat
6. On worker failure:
   - Heartbeat timeout detected (30s default)
   - Worker marked as "dead"
   - Cleanup worker removes dead worker
   - In-progress tasks reassigned
7. On graceful shutdown:
   - Worker deregisters
   - Current tasks complete
   - Worker status updated
```

### 4.4 Retry with Exponential Backoff
```
Formula: delay = min(base_delay * (2 ^ retry_count) + jitter, max_delay)

Example:
- Attempt 1: 1s + jitter
- Attempt 2: 2s + jitter
- Attempt 3: 4s + jitter
- Attempt 4: 8s + jitter
- Attempt 5: 16s + jitter
- Max delay: 300s (5 minutes)

Jitter: Random value (0-1s) to prevent thundering herd
```

## 5. Scalability Design

### 5.1 Horizontal Scaling
- **Stateless API servers**: Scale independently behind load balancer
- **Worker pools**: Add/remove workers dynamically based on load
- **Redis cluster**: Shard data across multiple nodes
- **No single point of failure**: All components are redundant

### 5.2 Performance Optimizations
- **Connection pooling**: Reuse Redis connections
- **Batch operations**: Group multiple Redis commands
- **Async processing**: Non-blocking I/O throughout
- **Efficient serialization**: Protocol Buffers for gRPC
- **In-memory caching**: Hot path optimization

### 5.3 Load Balancing
- **Round-robin** task distribution across workers
- **Least-loaded** worker selection for heavy tasks
- **Sticky sessions** for stateful operations
- **Queue-specific** workers for specialized tasks

## 6. Fault Tolerance & Reliability

### 6.1 Failure Scenarios
| Scenario | Detection | Recovery |
|----------|-----------|----------|
| Worker crash | Heartbeat timeout | Reassign tasks, mark worker dead |
| Redis failure | Connection error | Retry with backoff, failover to replica |
| Network partition | Request timeout | Circuit breaker, queue locally |
| Task timeout | Execution timer | Kill worker, retry task |
| Memory overflow | Resource monitoring | Reject new tasks, alert operators |

### 6.2 Data Durability
- **Redis persistence**: RDB snapshots + AOF log
- **Replication**: Master-slave setup with automatic failover
- **Backup strategy**: Regular snapshots to S3/GCS
- **Transaction logs**: Audit trail for debugging

### 6.3 Monitoring & Alerting
- **Prometheus metrics**: Counters, gauges, histograms
- **Health checks**: Liveness and readiness probes
- **Alerting rules**: 
  - Task failure rate > 5%
  - Queue depth > 10,000
  - Worker availability < 50%
  - Task latency > 10s (p99)

## 7. Security Considerations

### 7.1 Authentication & Authorization
- **API Keys**: For client authentication
- **JWT Tokens**: For session management
- **RBAC**: Role-based access control
- **TLS/mTLS**: Encrypted communication

### 7.2 Rate Limiting
- **Per-client limits**: Prevent abuse
- **Global limits**: Protect system capacity
- **Adaptive throttling**: Based on system load

## 8. Deployment Architecture

### 8.1 Containerization
```
- task-queue-api:latest
- task-queue-worker:latest
- redis:7-alpine
- monitoring-dashboard:latest
```

### 8.2 Orchestration (Kubernetes)
```yaml
- Deployment: API servers (3+ replicas)
- Deployment: Workers (auto-scaling, min: 5, max: 50)
- StatefulSet: Redis cluster (3+ nodes)
- Service: Load balancer for API
- ConfigMap: Application configuration
- Secret: Credentials and keys
```

## 9. Monitoring Metrics

### 9.1 Task Metrics
- **task_submissions_total**: Counter
- **task_completions_total**: Counter (by status)
- **task_duration_seconds**: Histogram
- **task_retry_count**: Counter
- **queue_depth**: Gauge (by priority)

### 9.2 Worker Metrics
- **workers_active**: Gauge
- **worker_task_processing_duration**: Histogram
- **worker_heartbeats_total**: Counter
- **worker_failures_total**: Counter

### 9.3 System Metrics
- **redis_connection_pool_size**: Gauge
- **api_request_duration_seconds**: Histogram
- **api_requests_total**: Counter (by endpoint)
- **error_rate**: Gauge

## 10. Comparison with Existing Solutions

| Feature | Our System | Celery | Sidekiq |
|---------|-----------|--------|---------|
| Language | Go | Python | Ruby |
| Backend | Redis | Multiple | Redis |
| Performance | High (compiled) | Medium | High |
| Memory | Low | High | Low |
| Concurrency | Goroutines | Processes/Threads | Threads |
| API | REST + gRPC | Python API | Ruby API |
| Monitoring | Built-in dashboard | External | Sidekiq Web |
| DLQ | Native | Plugin | Native |
| Type Safety | Yes | No | No |

## 11. Code Organization Principles

### 11.1 Clean Architecture
- **Handler Layer**: Handles HTTP/gRPC requests, delegates to services
- **Service Layer**: Business logic, orchestration, no infrastructure concerns
- **Repository Layer**: Data access, Redis operations
- **Entity Layer**: Core domain models (Task, Worker)

### 11.2 Dependency Injection
- **Uber FX**: All dependencies managed via FX
- **Provider Files**: Consolidated providers per package
- **Interface-Based**: All services use interfaces for testability

### 11.3 Constants Management
All constants defined in `internal/constants/constants.go`:
- Queue names (`QUEUE_HIGH`, `QUEUE_CRITICAL`, etc.)
- Status strings (`STATUS_PENDING`, `STATUS_COMPLETED`, etc.)
- Metric names (`MetricTasksSubmittedTotal`, etc.)
- Default values (`DefaultWorkerConcurrency`, `DefaultTaskTimeout`, etc.)

### 11.4 Logging Strategy
- **Structured Logging**: Using custom logger package
- **Context-Aware**: `logger.InfoWithCtx`, `logger.ErrorWithCtx`
- **No fmt Package**: All output via logger
- **Request ID Propagation**: Tracked across all operations

### 11.5 Error Handling
- **Typed Errors**: Domain-specific error types
- **Error Wrapping**: Using `fmt.Errorf` with `%w`
- **Logging**: All errors logged with context
- **Metrics**: Error counters for monitoring

## 12. Future Enhancements
- Task scheduling (cron-like)
- Task dependencies and workflows (DAG)
- WebSocket support for real-time updates
- Multi-tenancy support
- Geographic distribution
- Plugin system for custom task processors
- Auto-scaling based on queue depth
- Task result caching


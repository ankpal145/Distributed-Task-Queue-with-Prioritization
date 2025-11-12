# API Documentation

Complete API reference for the Distributed Task Queue System.

## Table of Contents

- [REST API](#rest-api)
- [gRPC API](#grpc-api)
- [Authentication](#authentication)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Examples](#examples)

## REST API

Base URL: `http://localhost:8080/api/v1`

### Task Endpoints

#### Submit Task

Submit a new task to the queue.

**Endpoint:** `POST /tasks`

**Request Body:**
```json
{
  "type": "string",           // Task type identifier (required)
  "priority": integer,        // Priority level: 10, 20, 30, 40 (required)
  "payload": object,          // Task data (required)
  "max_retries": integer,     // Maximum retry attempts (optional, default: 3)
  "timeout": integer,         // Timeout in nanoseconds (optional, default: 5m)
  "scheduled_at": string,     // ISO 8601 timestamp (optional, default: now)
  "metadata": object          // Custom metadata (optional)
}
```

**Example Request:**
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "priority": 30,
    "payload": {
      "to": "user@example.com",
      "subject": "Hello",
      "body": "Welcome!"
    },
    "max_retries": 3,
    "timeout": 300000000000
  }'
```

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "email",
  "priority": 30,
  "payload": {...},
  "status": "pending",
  "retry_count": 0,
  "max_retries": 3,
  "created_at": "2024-01-15T10:30:00Z",
  "scheduled_at": "2024-01-15T10:30:00Z",
  "timeout": 300000000000,
  "metadata": {}
}
```

#### Get Task

Retrieve task details by ID.

**Endpoint:** `GET /tasks/:id`

**Example:**
```bash
curl http://localhost:8080/api/v1/tasks/550e8400-e29b-41d4-a716-446655440000
```

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "email",
  "priority": 30,
  "status": "completed",
  "result": {"status": "sent"},
  "created_at": "2024-01-15T10:30:00Z",
  "started_at": "2024-01-15T10:30:05Z",
  "completed_at": "2024-01-15T10:30:07Z",
  "worker_id": "worker-abc123"
}
```

#### Cancel Task

Cancel a pending task.

**Endpoint:** `DELETE /tasks/:id`

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/550e8400-e29b-41d4-a716-446655440000
```

**Response:** `200 OK`
```json
{
  "message": "task cancelled"
}
```

#### List Tasks

List tasks with optional filters.

**Endpoint:** `GET /tasks`

**Query Parameters:**
- `status`: Filter by status (pending, processing, completed, failed, dead)
- `priority`: Filter by priority (10, 20, 30, 40)
- `type`: Filter by task type
- `limit`: Maximum number of results (default: 100)
- `offset`: Pagination offset (default: 0)

**Example:**
```bash
curl "http://localhost:8080/api/v1/tasks?status=completed&limit=10"
```

**Response:** `200 OK`
```json
[
  {
    "id": "...",
    "type": "email",
    "status": "completed",
    ...
  },
  ...
]
```

### Dead Letter Queue Endpoints

#### Get Dead Letter Queue

Retrieve tasks from the dead letter queue.

**Endpoint:** `GET /dlq`

**Query Parameters:**
- `limit`: Maximum number of results (default: 100)

**Example:**
```bash
curl "http://localhost:8080/api/v1/dlq?limit=50"
```

**Response:** `200 OK`
```json
[
  {
    "id": "...",
    "type": "email",
    "status": "dead",
    "error": "max retries exceeded",
    "retry_count": 3,
    ...
  },
  ...
]
```

#### Retry Dead Task

Retry a task from the dead letter queue.

**Endpoint:** `POST /dlq/:id/retry`

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/dlq/550e8400-e29b-41d4-a716-446655440000/retry
```

**Response:** `200 OK`
```json
{
  "message": "task requeued"
}
```

### Worker Endpoints

#### Register Worker

Register a new worker.

**Endpoint:** `POST /workers/register`

**Request Body:**
```json
{
  "name": "string",           // Worker name (required)
  "queues": ["string"],       // Queue names to process (required)
  "concurrency": integer,     // Max concurrent tasks (optional, default: 10)
  "metadata": object          // Custom metadata (optional)
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/workers/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "worker-1",
    "queues": ["default"],
    "concurrency": 10
  }'
```

**Response:** `201 Created`
```json
{
  "id": "worker-abc123",
  "name": "worker-1",
  "status": "active",
  "queues": ["default"],
  "concurrency": 10,
  "started_at": "2024-01-15T10:00:00Z",
  ...
}
```

#### Get Worker

Get worker details by ID.

**Endpoint:** `GET /workers/:id`

**Example:**
```bash
curl http://localhost:8080/api/v1/workers/worker-abc123
```

**Response:** `200 OK`
```json
{
  "id": "worker-abc123",
  "name": "worker-1",
  "status": "active",
  "tasks_processed": 1500,
  "tasks_failed": 10,
  "last_seen": "2024-01-15T11:30:00Z",
  ...
}
```

#### List Workers

List all active workers.

**Endpoint:** `GET /workers`

**Example:**
```bash
curl http://localhost:8080/api/v1/workers
```

**Response:** `200 OK`
```json
[
  {
    "id": "worker-abc123",
    "name": "worker-1",
    "status": "active",
    ...
  },
  ...
]
```

#### Deregister Worker

Remove a worker.

**Endpoint:** `DELETE /workers/:id`

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/v1/workers/worker-abc123
```

**Response:** `200 OK`
```json
{
  "message": "worker deregistered"
}
```

#### Get Worker Statistics

Get performance statistics for a worker.

**Endpoint:** `GET /workers/:id/stats`

**Example:**
```bash
curl http://localhost:8080/api/v1/workers/worker-abc123/stats
```

**Response:** `200 OK`
```json
{
  "worker_id": "worker-abc123",
  "name": "worker-1",
  "status": "active",
  "tasks_processed": 1500,
  "tasks_failed": 10,
  "success_rate": 99.33,
  "current_load": 5,
  "max_load": 10,
  "uptime": "2h15m30s",
  "last_seen": "2024-01-15T11:30:00Z"
}
```

### System Endpoints

#### Health Check

Check if the service is healthy.

**Endpoint:** `GET /health`

**Example:**
```bash
curl http://localhost:8080/api/v1/health
```

**Response:** `200 OK`
```json
{
  "status": "healthy",
  "service": "task-queue"
}
```

#### Queue Statistics

Get statistics for all queues.

**Endpoint:** `GET /queues/stats`

**Example:**
```bash
curl http://localhost:8080/api/v1/queues/stats
```

**Response:** `200 OK`
```json
{
  "critical": {
    "name": "critical",
    "priority": 40,
    "depth": 5,
    "processed_count": 1000,
    "failed_count": 2
  },
  "high": {...},
  "normal": {...},
  "low": {...}
}
```

#### System Statistics

Get overall system statistics.

**Endpoint:** `GET /system/stats`

**Example:**
```bash
curl http://localhost:8080/api/v1/system/stats
```

**Response:** `200 OK`
```json
{
  "queues": {
    "critical": {...},
    "high": {...},
    "normal": {...},
    "low": {...}
  },
  "workers": {
    "total": 5,
    "stats": [...]
  },
  "tasks": {
    "total_in_queue": 150,
    "total_processed": 10000,
    "total_failed": 50
  }
}
```

## gRPC API

See [task_queue.proto](../internal/api/grpc/task_queue.proto) for complete definitions.

### Service Definition

```protobuf
service TaskQueueService {
  rpc SubmitTask(SubmitTaskRequest) returns (TaskResponse);
  rpc GetTaskStatus(GetTaskStatusRequest) returns (TaskResponse);
  rpc CancelTask(CancelTaskRequest) returns (CancelTaskResponse);
  rpc StreamTasks(StreamTasksRequest) returns (stream TaskUpdate);
  rpc RegisterWorker(RegisterWorkerRequest) returns (WorkerResponse);
  rpc UpdateHeartbeat(HeartbeatRequest) returns (HeartbeatResponse);
}
```

### Example Usage

```go
import (
    "context"
    grpcapi "github.com/ankurpal/distributed-task-queue/internal/api/grpc"
    "google.golang.org/grpc"
)

// Connect to server
conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
client := grpcapi.NewTaskQueueServiceClient(conn)

// Submit task
task, err := client.SubmitTask(context.Background(), &grpcapi.SubmitTaskRequest{
    Type:       "email",
    Priority:   grpcapi.Priority_HIGH,
    Payload:    []byte(`{"to":"user@example.com"}`),
    MaxRetries: 3,
    TimeoutMs:  300000,
})

// Get task status
status, err := client.GetTaskStatus(context.Background(), &grpcapi.GetTaskStatusRequest{
    TaskId: task.Id,
})

// Stream task updates
stream, err := client.StreamTasks(context.Background(), &grpcapi.StreamTasksRequest{
    TaskIds: []string{task.Id},
})
for {
    update, err := stream.Recv()
    if err == io.EOF {
        break
    }
    fmt.Printf("Task %s: %s\n", update.TaskId, update.Status)
}
```

## Authentication

Currently, the API does not require authentication. For production:

### API Key Authentication (Recommended)

Add API key header:
```bash
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/tasks
```

### JWT Token Authentication

Add bearer token:
```bash
curl -H "Authorization: Bearer your-jwt-token" http://localhost:8080/api/v1/tasks
```

## Error Handling

### Error Response Format

```json
{
  "error": "error message description"
}
```

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid request format |
| 404 | Not Found | Resource not found |
| 500 | Internal Server Error | Server error |

### Common Errors

**Invalid Task Payload:**
```json
{
  "error": "invalid request: task payload is required"
}
```

**Task Not Found:**
```json
{
  "error": "task not found"
}
```

**Worker Not Available:**
```json
{
  "error": "no task available"
}
```

## Rate Limiting

Rate limiting is not currently implemented. For production:

**Recommended limits:**
- 1000 requests/minute per client
- 10,000 task submissions/minute globally
- Implement using middleware or API gateway

## Examples

### Submit Multiple Tasks

```bash
for i in {1..100}; do
  curl -X POST http://localhost:8080/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d '{
      "type": "email",
      "priority": 20,
      "payload": {"to": "user'$i'@example.com"}
    }'
done
```

### Monitor Queue Depth

```bash
watch -n 1 'curl -s http://localhost:8080/api/v1/queues/stats | jq'
```

### Get Failed Tasks

```bash
curl "http://localhost:8080/api/v1/tasks?status=failed" | jq
```

### Retry All Dead Tasks

```bash
curl -s http://localhost:8080/api/v1/dlq | jq -r '.[].id' | while read id; do
  curl -X POST http://localhost:8080/api/v1/dlq/$id/retry
done
```

## Best Practices

1. **Use appropriate priority levels** - Reserve CRITICAL for truly urgent tasks
2. **Set reasonable timeouts** - Prevent workers from being blocked indefinitely
3. **Handle idempotency** - Tasks may be retried, design them to be idempotent
4. **Monitor metrics** - Track queue depths and processing times
5. **Implement circuit breakers** - Prevent cascading failures
6. **Use structured payloads** - JSON is recommended for complex data
7. **Set max retries appropriately** - Consider task importance and cost

## SDK Support

### Go SDK

Built-in - import the internal packages directly.

### Python SDK (Coming Soon)

```python
from taskqueue import Client

client = Client('http://localhost:8080')
task = client.submit_task(
    type='email',
    priority='high',
    payload={'to': 'user@example.com'}
)
```

### JavaScript SDK (Coming Soon)

```javascript
const TaskQueue = require('taskqueue-client');

const client = new TaskQueue('http://localhost:8080');
const task = await client.submitTask({
    type: 'email',
    priority: 'high',
    payload: {to: 'user@example.com'}
});
```


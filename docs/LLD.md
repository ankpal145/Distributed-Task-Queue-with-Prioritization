# Low-Level Design (LLD) - Distributed Task Queue System

## 1. Domain Models

### 1.1 Task Model
```go
type Task struct {
    ID            string        // UUID v4
    Type          string        // Task type identifier
    Priority      Priority      // CRITICAL, HIGH, NORMAL, LOW
    Payload       []byte        // JSON-encoded task data
    Status        TaskStatus    // PENDING, PROCESSING, COMPLETED, FAILED, DEAD
    Result        []byte        // Task execution result
    Error         string        // Error message if failed
    RetryCount    int           // Current retry attempt
    MaxRetries    int           // Maximum retry attempts (default: 3)
    CreatedAt     time.Time     // Task creation timestamp
    ScheduledAt   time.Time     // When to process (for delayed tasks)
    StartedAt     *time.Time    // When processing started
    CompletedAt   *time.Time    // When processing finished
    Timeout       time.Duration // Max execution time
    WorkerID      string        // ID of worker processing this task
    Metadata      map[string]string // Custom metadata
}

type Priority int
const (
    PriorityCritical Priority = 40
    PriorityHigh     Priority = 30
    PriorityNormal   Priority = 20
    PriorityLow      Priority = 10
)

type TaskStatus string
const (
    TaskStatusPending    TaskStatus = "pending"
    TaskStatusProcessing TaskStatus = "processing"
    TaskStatusCompleted  TaskStatus = "completed"
    TaskStatusFailed     TaskStatus = "failed"
    TaskStatusDead       TaskStatus = "dead"
)
```

### 1.2 Worker Model
```go
type Worker struct {
    ID          string            // UUID v4
    Name        string            // Human-readable name
    Status      WorkerStatus      // ACTIVE, IDLE, BUSY, DEAD
    Queues      []string          // Queue names this worker processes
    Concurrency int               // Max concurrent tasks
    LastSeen    time.Time         // Last heartbeat timestamp
    ProcessingTasks []string      // Currently processing task IDs
    Metadata    map[string]string // Custom metadata
    StartedAt   time.Time         // Worker start time
    TasksProcessed int64          // Total tasks processed
    TasksFailed    int64          // Total tasks failed
}

type WorkerStatus string
const (
    WorkerStatusActive WorkerStatus = "active"
    WorkerStatusIdle   WorkerStatus = "idle"
    WorkerStatusBusy   WorkerStatus = "busy"
    WorkerStatusDead   WorkerStatus = "dead"
)
```

### 1.3 Queue Model
```go
type Queue struct {
    Name        string    // Queue identifier
    Priority    Priority  // Queue priority level
    Depth       int64     // Current number of tasks
    ProcessedCount int64  // Total processed tasks
    FailedCount    int64  // Total failed tasks
    AvgLatency     float64 // Average task latency (seconds)
}
```

## 2. Repository Layer Implementation

### 2.1 Task Repository Interface
```go
type TaskRepository interface {
    // Create new task
    Create(ctx context.Context, task *Task) error
    
    // Get task by ID
    Get(ctx context.Context, taskID string) (*Task, error)
    
    // Update task
    Update(ctx context.Context, task *Task) error
    
    // Delete task
    Delete(ctx context.Context, taskID string) error
    
    // Enqueue task to priority queue
    Enqueue(ctx context.Context, task *Task) error
    
    // Dequeue highest priority task
    Dequeue(ctx context.Context, workerID string) (*Task, error)
    
    // Re-enqueue task with delay
    RequeueWithDelay(ctx context.Context, task *Task, delay time.Duration) error
    
    // Move task to dead letter queue
    MoveToDLQ(ctx context.Context, task *Task) error
    
    // List tasks with filters
    List(ctx context.Context, filter TaskFilter) ([]*Task, error)
    
    // Get queue statistics
    GetQueueStats(ctx context.Context) (map[string]*Queue, error)
}
```

### 2.2 Redis Implementation Details

#### Redis Key Patterns
```
tasks:{task_id}                     -> Hash (task metadata)
queue:{priority}:pending            -> Sorted Set (score = scheduled_at)
queue:processing                    -> Hash (task_id -> worker_id)
queue:dlq                           -> List (failed tasks)
workers:{worker_id}                 -> Hash (worker metadata)
workers:active                      -> Set (active worker IDs)
metrics:tasks:{date}                -> Hash (daily metrics)
```

#### Priority Queue Implementation
```go
func (r *RedisTaskRepository) Enqueue(ctx context.Context, task *Task) error {
    pipe := r.client.Pipeline()
    
    // 1. Store task metadata
    taskKey := fmt.Sprintf("tasks:%s", task.ID)
    taskData, _ := json.Marshal(task)
    pipe.HSet(ctx, taskKey, "data", taskData)
    pipe.Expire(ctx, taskKey, 24*time.Hour)
    
    // 2. Add to priority queue with score
    queueKey := fmt.Sprintf("queue:%d:pending", task.Priority)
    score := float64(task.ScheduledAt.Unix())
    pipe.ZAdd(ctx, queueKey, &redis.Z{
        Score:  score,
        Member: task.ID,
    })
    
    // 3. Execute pipeline
    _, err := pipe.Exec(ctx)
    return err
}

func (r *RedisTaskRepository) Dequeue(ctx context.Context, workerID string) (*Task, error) {
    // Try each priority queue from highest to lowest
    priorities := []Priority{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow}
    
    now := float64(time.Now().Unix())
    
    for _, priority := range priorities {
        queueKey := fmt.Sprintf("queue:%d:pending", priority)
        
        // Atomic pop from sorted set (only if scheduled time has passed)
        result, err := r.client.ZPopMin(ctx, queueKey, 1).Result()
        if err == redis.Nil {
            continue // Queue empty, try next priority
        }
        if err != nil {
            return nil, err
        }
        
        if len(result) == 0 {
            continue
        }
        
        taskID := result[0].Member.(string)
        scheduledAt := result[0].Score
        
        // Check if task is ready to process
        if scheduledAt > now {
            // Put it back, not ready yet
            r.client.ZAdd(ctx, queueKey, &redis.Z{
                Score:  scheduledAt,
                Member: taskID,
            })
            continue
        }
        
        // Get task data
        taskKey := fmt.Sprintf("tasks:%s", taskID)
        taskData, err := r.client.HGet(ctx, taskKey, "data").Result()
        if err != nil {
            continue
        }
        
        var task Task
        json.Unmarshal([]byte(taskData), &task)
        
        // Mark as processing
        task.Status = TaskStatusProcessing
        task.WorkerID = workerID
        now := time.Now()
        task.StartedAt = &now
        
        // Update in Redis
        r.client.HSet(ctx, fmt.Sprintf("queue:processing"), task.ID, workerID)
        r.Update(ctx, &task)
        
        return &task, nil
    }
    
    return nil, ErrNoTaskAvailable
}
```

### 2.3 Worker Repository Interface
```go
type WorkerRepository interface {
    // Register worker
    Register(ctx context.Context, worker *Worker) error
    
    // Get worker by ID
    Get(ctx context.Context, workerID string) (*Worker, error)
    
    // Update worker heartbeat
    UpdateHeartbeat(ctx context.Context, workerID string) error
    
    // Deregister worker
    Deregister(ctx context.Context, workerID string) error
    
    // List active workers
    ListActive(ctx context.Context) ([]*Worker, error)
    
    // Find dead workers (heartbeat timeout)
    FindDeadWorkers(ctx context.Context, timeout time.Duration) ([]*Worker, error)
}
```

## 3. Service Layer Implementation

### 3.1 Task Service
```go
type TaskService struct {
    taskRepo      TaskRepository
    workerRepo    WorkerRepository
    metrics       MetricsCollector
    config        *Config
}

func (s *TaskService) SubmitTask(ctx context.Context, req *SubmitTaskRequest) (*Task, error) {
    // 1. Validate request
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // 2. Create task
    task := &Task{
        ID:          uuid.New().String(),
        Type:        req.Type,
        Priority:    req.Priority,
        Payload:     req.Payload,
        Status:      TaskStatusPending,
        MaxRetries:  req.MaxRetries,
        CreatedAt:   time.Now(),
        ScheduledAt: req.ScheduledAt,
        Timeout:     req.Timeout,
        Metadata:    req.Metadata,
    }
    
    // 3. Store task
    if err := s.taskRepo.Create(ctx, task); err != nil {
        return nil, fmt.Errorf("failed to create task: %w", err)
    }
    
    // 4. Enqueue task
    if err := s.taskRepo.Enqueue(ctx, task); err != nil {
        return nil, fmt.Errorf("failed to enqueue task: %w", err)
    }
    
    // 5. Record metrics
    s.metrics.IncrementCounter("tasks_submitted_total", map[string]string{
        "priority": task.Priority.String(),
        "type":     task.Type,
    })
    
    return task, nil
}

func (s *TaskService) HandleTaskFailure(ctx context.Context, task *Task, err error) error {
    task.RetryCount++
    task.Error = err.Error()
    
    // Check if retries exhausted
    if task.RetryCount >= task.MaxRetries {
        task.Status = TaskStatusDead
        
        // Move to dead letter queue
        if err := s.taskRepo.MoveToDLQ(ctx, task); err != nil {
            return fmt.Errorf("failed to move task to DLQ: %w", err)
        }
        
        s.metrics.IncrementCounter("tasks_dead_total", map[string]string{
            "type": task.Type,
        })
        
        return nil
    }
    
    // Calculate exponential backoff with jitter
    delay := s.calculateBackoff(task.RetryCount)
    
    // Re-enqueue with delay
    task.Status = TaskStatusPending
    if err := s.taskRepo.RequeueWithDelay(ctx, task, delay); err != nil {
        return fmt.Errorf("failed to requeue task: %w", err)
    }
    
    s.metrics.IncrementCounter("tasks_retried_total", map[string]string{
        "type":  task.Type,
        "retry": fmt.Sprintf("%d", task.RetryCount),
    })
    
    return nil
}

func (s *TaskService) calculateBackoff(retryCount int) time.Duration {
    // Exponential backoff: min(base * 2^retry, max) + jitter
    base := s.config.RetryBaseDelay
    maxDelay := s.config.RetryMaxDelay
    
    delay := base * time.Duration(math.Pow(2, float64(retryCount)))
    if delay > maxDelay {
        delay = maxDelay
    }
    
    // Add jitter (0-1 second)
    jitter := time.Duration(rand.Int63n(int64(time.Second)))
    return delay + jitter
}
```

### 3.2 Worker Service
```go
// Worker Service manages worker entities (registration, deregistration, health)
// Moved from internal/service to internal/worker for better organization
type WorkerService struct {
    workerRepo    repo.WorkerRepository
    metrics       service.MetricsCollector
    logger        *logger.Logger
    config        *WorkerConfig
}

func (s *WorkerService) RegisterWorker(ctx context.Context, req *RegisterWorkerRequest) (*entity.Worker, error) {
    worker := &entity.Worker{
        ID:        uuid.New().String(),
        Name:      req.Name,
        Status:    constants.STATUS_ACTIVE,
        Queues:    req.Queues,
        LastSeen:  time.Now(),
        StartedAt: time.Now(),
        Metadata:  req.Metadata,
    }
    
    if err := s.workerRepo.CreateWorker(ctx, worker); err != nil {
        s.logger.ErrorWithCtx(ctx, "failed to register worker", "error", err)
        return nil, fmt.Errorf("failed to register worker: %w", err)
    }
    
    s.metrics.IncrementGauge(constants.MetricWorkersActive, 1)
    s.logger.InfoWithCtx(ctx, "worker registered successfully", "worker_id", worker.ID)
    
    return worker, nil
}

func (s *WorkerService) DeregisterWorker(ctx context.Context, workerID string) error {
    if err := s.workerRepo.DeleteWorker(ctx, workerID); err != nil {
        s.logger.ErrorWithCtx(ctx, "failed to deregister worker", "worker_id", workerID, "error", err)
        return fmt.Errorf("failed to deregister worker: %w", err)
    }
    
    s.metrics.DecrementGauge(constants.MetricWorkersActive, 1)
    s.logger.InfoWithCtx(ctx, "worker deregistered", "worker_id", workerID)
    
    return nil
}

func (s *WorkerService) CleanupDeadWorkers(ctx context.Context) error {
    workers, err := s.workerRepo.GetAllWorkers(ctx)
    if err != nil {
        return fmt.Errorf("failed to get workers: %w", err)
    }
    
    timeout := s.config.WorkerTimeout
    for _, worker := range workers {
        if time.Since(worker.LastSeen) > timeout {
            s.logger.WarnWithCtx(ctx, "cleaning up dead worker", "worker_id", worker.ID)
            if err := s.workerRepo.DeleteWorker(ctx, worker.ID); err != nil {
                s.logger.ErrorWithCtx(ctx, "failed to delete dead worker", "worker_id", worker.ID, "error", err)
                continue
            }
            s.metrics.IncrementCounter(constants.MetricWorkersCleanedUpTotal, nil)
        }
    }
    
    return nil
}
```

### 3.3 Worker Consumer Pattern (Beacon-Style)

The worker implementation follows the beacon pattern with a consumer processor and task processors.

#### Consumer Processor
```go
// internal/worker/consumer/consumer_processor.go
type Processor struct {
    workerID        string
    taskService     *service.TaskService
    workerService   *worker.WorkerService
    processorEngine func(ctx context.Context, task *entity.Task) ([]byte, error)
    logger          *logger.Logger
    config          *worker.WorkerConfig
    metrics         service.MetricsCollector
}

func NewProcessor(
    taskService *service.TaskService,
    workerService *worker.WorkerService,
    processorEngine func(ctx context.Context, task *entity.Task) ([]byte, error),
    logger *logger.Logger,
    config *worker.WorkerConfig,
    metrics service.MetricsCollector,
) *Processor {
    return &Processor{
        workerID:        uuid.New().String(),
        taskService:     taskService,
        workerService:   workerService,
        processorEngine: processorEngine,
        logger:          logger,
        config:          config,
        metrics:         metrics,
    }
}

func (p *Processor) Start(ctx context.Context) error {
    // Panic recovery
    defer func() {
        if r := recover(); r != nil {
            p.logger.ErrorWithCtx(ctx, "panic in processor start", "error", r)
        }
    }()
    
    // Register worker
    registerReq := &worker.RegisterWorkerRequest{
        Name:   fmt.Sprintf("worker-%s", p.workerID),
        Queues: []string{constants.QUEUE_CRITICAL, constants.QUEUE_HIGH, constants.QUEUE_MEDIUM, constants.QUEUE_LOW},
    }
    
    _, err := p.workerService.RegisterWorker(ctx, registerReq)
    if err != nil {
        return fmt.Errorf("failed to register worker: %w", err)
    }
    
    p.logger.InfoWithCtx(ctx, "worker started", "worker_id", p.workerID)
    
    // Start concurrent operations
    go p.startTaskConsumer(ctx)
    go p.heartbeatLoop(ctx)
    go p.startCleanupWorker(ctx)
    
    return nil
}

func (p *Processor) Stop(ctx context.Context) error {
    p.logger.InfoWithCtx(ctx, "stopping worker", "worker_id", p.workerID)
    
    // Deregister worker
    if err := p.workerService.DeregisterWorker(ctx, p.workerID); err != nil {
        p.logger.ErrorWithCtx(ctx, "failed to deregister worker", "error", err)
        return err
    }
    
    return nil
}

func (p *Processor) startTaskConsumer(ctx context.Context) {
    ctx = contexts.NewConsumerCtx(p.workerID, []string{constants.QUEUE_CRITICAL, constants.QUEUE_HIGH})
    
    for {
        select {
        case <-ctx.Done():
            return
        default:
            task, err := p.taskService.GetNextTask(ctx)
            if err != nil {
                time.Sleep(constants.SleepNoTaskAvailable)
                continue
            }
            
            // Process task using processor engine
            result, err := p.processorEngine(ctx, task)
            if err != nil {
                p.logger.ErrorWithCtx(ctx, "task processing failed", "task_id", task.ID, "error", err)
                p.taskService.ReportTaskFailure(ctx, task.ID, err)
            } else {
                p.taskService.ReportTaskSuccess(ctx, task.ID, result)
            }
        }
    }
}

func (p *Processor) heartbeatLoop(ctx context.Context) {
    ticker := time.NewTicker(constants.DefaultHeartbeatInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := p.workerService.UpdateHeartbeat(ctx, p.workerID); err != nil {
                p.logger.ErrorWithCtx(ctx, "heartbeat failed", "error", err)
            }
        }
    }
}

func (p *Processor) startCleanupWorker(ctx context.Context) {
    ticker := time.NewTicker(constants.DefaultCleanupInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := p.workerService.CleanupDeadWorkers(ctx); err != nil {
                p.logger.ErrorWithCtx(ctx, "cleanup failed", "error", err)
            }
        }
    }
}
```

#### Task Processors
```go
// internal/worker/processors/default_task_processor.go
type DefaultTaskProcessor struct {
    logger *logger.Logger
}

func (p *DefaultTaskProcessor) Process(ctx context.Context, task *entity.Task) (result []byte, processorError error) {
    // Panic recovery
    defer func() {
        if r := recover(); r != nil {
            p.logger.ErrorWithCtx(ctx, "panic in DefaultTaskProcessor", "error", r)
            processorError = fmt.Errorf("task processing panicked: %v", r)
        }
    }()
    
    // Create context with timeout and request ID
    ctx, cancel := context.WithTimeout(ctx, constants.DefaultTaskTimeout)
    defer cancel()
    
    ctx = context.WithValue(ctx, contexts.RequestIDKey, uuid.New().String())
    
    p.logger.InfoWithCtx(ctx, "processing task", "task_id", task.ID, "priority", task.Priority)
    
    // Unmarshal payload
    var payload map[string]interface{}
    if err := json.Unmarshal(task.Payload, &payload); err != nil {
        return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
    }
    
    // Business logic here
    // ... task execution ...
    
    // Return result
    resultData := map[string]interface{}{
        "status": "success",
        "task_id": task.ID,
    }
    
    return json.Marshal(resultData)
}

// internal/worker/processors/high_priority_processor.go
type HighPriorityTaskProcessor struct {
    logger *logger.Logger
}

func (p *HighPriorityTaskProcessor) Process(ctx context.Context, task *entity.Task) (result []byte, processorError error) {
    // Similar implementation with high priority specific logic
    // ...
}

// internal/worker/processors/processors.go
type TaskProcessors struct {
    DefaultTaskProcessor      *DefaultTaskProcessor
    HighPriorityTaskProcessor *HighPriorityTaskProcessor
}

var (
    taskProcessorsDoOnce sync.Once
    taskProcessorsStruct TaskProcessors
)

func NewTaskProcessors(
    defaultTaskProcessor *DefaultTaskProcessor,
    highPriorityTaskProcessor *HighPriorityTaskProcessor,
) *TaskProcessors {
    taskProcessorsDoOnce.Do(func() {
        taskProcessorsStruct = TaskProcessors{
            DefaultTaskProcessor:      defaultTaskProcessor,
            HighPriorityTaskProcessor: highPriorityTaskProcessor,
        }
    })
    return &taskProcessorsStruct
}
```

## 4. API Layer Implementation

### 4.1 REST API Handlers
```go
type TaskHandler struct {
    taskService *TaskService
}

func (h *TaskHandler) SubmitTask(c *gin.Context) {
    var req SubmitTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    task, err := h.taskService.SubmitTask(c.Request.Context(), &req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(201, task)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
    taskID := c.Param("id")
    
    task, err := h.taskService.GetTask(c.Request.Context(), taskID)
    if err != nil {
        c.JSON(404, gin.H{"error": "task not found"})
        return
    }
    
    c.JSON(200, task)
}
```

### 4.2 gRPC Service Implementation
```protobuf
syntax = "proto3";

package taskqueue;

service TaskQueueService {
  rpc SubmitTask(SubmitTaskRequest) returns (TaskResponse);
  rpc GetTaskStatus(GetTaskStatusRequest) returns (TaskResponse);
  rpc CancelTask(CancelTaskRequest) returns (CancelTaskResponse);
  rpc StreamTasks(StreamTasksRequest) returns (stream TaskUpdate);
}

message SubmitTaskRequest {
  string type = 1;
  int32 priority = 2;
  bytes payload = 3;
  int32 max_retries = 4;
  int64 timeout_ms = 5;
  map<string, string> metadata = 6;
}

message TaskResponse {
  string id = 1;
  string type = 2;
  string status = 3;
  bytes result = 4;
  string error = 5;
  int64 created_at = 6;
}
```

## 5. Monitoring Implementation

### 5.1 Metrics Collector
```go
type PrometheusCollector struct {
    taskSubmissions   *prometheus.CounterVec
    taskCompletions   *prometheus.CounterVec
    taskDuration      *prometheus.HistogramVec
    queueDepth        *prometheus.GaugeVec
    workersActive     prometheus.Gauge
}

func NewPrometheusCollector() *PrometheusCollector {
    return &PrometheusCollector{
        taskSubmissions: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "tasks_submitted_total",
                Help: "Total number of tasks submitted",
            },
            []string{"priority", "type"},
        ),
        // ... other metrics
    }
}
```

## 6. Dependency Injection (Uber FX)

### 6.1 Consolidated Provider Pattern

All worker-related dependencies are provided via a single consolidated provider:

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

### 6.2 Application Modules

```go
// cmd/main.go
func startFxApp(module string) error {
    var modules []fx.Option
    
    switch module {
    case constants.SERVER:
        modules = append(modules, grpcModules...)
        modules = append(modules, fx.Invoke(startUnifiedServer))
        
    case constants.WORKER:
        modules = append(modules, workerModules...)
        modules = append(modules, fx.Invoke(startWorker))
        
    default:
        return fmt.Errorf("unknown module: %s", module)
    }
    
    app := fx.New(modules...)
    return app.Err()
}
```

## 7. Constants Management

All constants are centralized in `internal/constants/constants.go`:

```go
package constants

// Module Types
const (
    SERVER = "server"
    WORKER = "worker"
)

// Queue Names
const (
    QUEUE_CRITICAL = "queue:critical"
    QUEUE_HIGH     = "queue:high"
    QUEUE_MEDIUM   = "queue:medium"
    QUEUE_LOW      = "queue:low"
    QUEUE_DLQ      = "queue:dlq"
)

// Status Strings
const (
    STATUS_PENDING    = "pending"
    STATUS_PROCESSING = "processing"
    STATUS_COMPLETED  = "completed"
    STATUS_FAILED     = "failed"
    STATUS_DEAD       = "dead"
    STATUS_ACTIVE     = "active"
)

// Metric Names
const (
    MetricTasksSubmittedTotal   = "tasks_submitted_total"
    MetricTasksProcessedTotal   = "tasks_processed_total"
    MetricTasksCompletedTotal   = "tasks_completed_total"
    MetricTasksFailedTotal      = "tasks_failed_total"
    MetricTasksDeadTotal        = "tasks_dead_total"
    MetricWorkersActive         = "workers_active"
    MetricWorkersCleanedUpTotal = "workers_cleaned_up_total"
)

// Sleep Durations
const (
    SleepNoTaskAvailable = 500 * time.Millisecond
)

// Default Values
const (
    DefaultWorkerConcurrency = 10
    DefaultHeartbeatInterval = 5 * time.Second
    DefaultWorkerTimeout     = 30 * time.Second
    DefaultCleanupInterval   = 1 * time.Minute
    DefaultTaskTimeout       = 5 * time.Minute
    DefaultMaxRetries        = 3
)
```

## 8. Configuration
```go
type Config struct {
    // Redis
    RedisAddr     string
    RedisPassword string
    RedisDB       int
    PoolSize      int
    
    // Server
    GRPCPort     int
    UnifiedPort  int
    
    // Worker
    WorkerConcurrency int
    HeartbeatInterval time.Duration
    WorkerTimeout     time.Duration
    CleanupInterval   time.Duration
    
    // Task
    RetryBaseDelay time.Duration
    RetryMaxDelay  time.Duration
    DefaultTimeout time.Duration
    MaxRetries     int
    
    // Monitoring
    MetricsEnabled bool
    MetricsPort    int
    
    // Logger
    LogLevel  string
    LogFormat string
}
```

## 9. Error Handling

### 9.1 Domain Errors
```go
// internal/entity/task.go
var (
    ErrTaskNotFound         = errors.New("task not found")
    ErrNoTaskAvailable      = errors.New("no task available")
    ErrTaskTimeout          = errors.New("task execution timeout")
    ErrMaxRetriesReached    = errors.New("max retries reached")
    ErrInvalidPriority      = errors.New("invalid priority")
    ErrTaskProcessingFailed = errors.New("task processing failed")
)

// internal/entity/worker.go
var (
    ErrWorkerNotFound    = errors.New("worker not found")
    ErrWorkerRegistration = errors.New("worker registration failed")
)
```

### 9.2 Error Wrapping and Logging
```go
// Example from task service
func (s *TaskService) ReportTaskFailure(ctx context.Context, taskID string, err error) error {
    s.logger.ErrorWithCtx(ctx, "task failed", "task_id", taskID, "error", err)
    
    task, getErr := s.GetTask(ctx, taskID)
    if getErr != nil {
        return fmt.Errorf("failed to get task: %w", getErr)
    }
    
    // Handle retry logic
    if err := s.handleTaskRetry(ctx, task, err); err != nil {
        return fmt.Errorf("failed to handle retry: %w", err)
    }
    
    return nil
}
```

## 10. Structured Logging

### 10.1 Logger Interface
```go
// pkg/logger/logger.go
type Logger struct {
    *logrus.Logger
}

func (l *Logger) Get(ctx context.Context) *logrus.Entry {
    entry := l.Logger.WithContext(ctx)
    
    // Extract request ID from context
    if reqID := ctx.Value(contexts.RequestIDKey); reqID != nil {
        entry = entry.WithField("request_id", reqID)
    }
    
    return entry
}

func (l *Logger) InfoWithCtx(ctx context.Context, msg string, args ...interface{}) {
    l.Get(ctx).WithFields(argsToFields(args...)).Info(msg)
}

func (l *Logger) ErrorWithCtx(ctx context.Context, msg string, args ...interface{}) {
    l.Get(ctx).WithFields(argsToFields(args...)).Error(msg)
}

func (l *Logger) WarnWithCtx(ctx context.Context, msg string, args ...interface{}) {
    l.Get(ctx).WithFields(argsToFields(args...)).Warn(msg)
}
```

### 10.2 Usage Throughout Codebase
- **No fmt Package**: All output via structured logger
- **Context Propagation**: Request ID tracked across all operations
- **Consistent Fields**: Standard fields (task_id, worker_id, error)

## 11. Testing Strategy

### 11.1 Unit Tests
- Each service method has unit tests
- Mock repositories and dependencies
- Test error paths and edge cases

### 11.2 Integration Tests
- Redis test container
- End-to-end task flow
- Worker registration and heartbeat
- Retry and DLQ logic

### 11.3 Load Tests
- Performance benchmarking
- Throughput measurement
- Latency percentiles (p50, p95, p99)

### 11.4 Chaos Testing
- Worker crash scenarios
- Redis connection failures
- Network partitions
- Timeout handling

## 12. Project Structure

```
internal/
├── constants/              # Centralized constants
│   └── constants.go
├── entity/                 # Domain entities
│   ├── task.go
│   └── worker.go
├── handler/                # Request handlers
│   ├── grpc-server/
│   └── rest-handler/
├── service/                # Business logic
│   ├── task_service.go
│   └── metrics.go
├── repo/                   # Data access
│   ├── task_repository.go
│   └── worker_repository.go
├── worker/                 # Worker management
│   ├── worker_service.go
│   ├── consumer/
│   │   └── consumer_processor.go
│   └── processors/
│       ├── provider.go
│       ├── processors.go
│       ├── default_task_processor.go
│       └── high_priority_processor.go
└── monitoring/             # Observability
    └── prometheus.go

pkg/
├── logger/                 # Structured logging
├── middleware/            # HTTP/gRPC middleware
├── redis/                 # Redis client
├── errors/                # Error handling
└── contexts/              # Context utilities
```


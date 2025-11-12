package entity

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Priority int

const (
	PriorityLow      Priority = 10
	PriorityNormal   Priority = 20
	PriorityHigh     Priority = 30
	PriorityCritical Priority = 40
)

func (p Priority) String() string {
	switch p {
	case PriorityCritical:
		return "critical"
	case PriorityHigh:
		return "high"
	case PriorityNormal:
		return "normal"
	case PriorityLow:
		return "low"
	default:
		return "unknown"
	}
}

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusDead       TaskStatus = "dead"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type Task struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Priority    Priority          `json:"priority"`
	Payload     json.RawMessage   `json:"payload"`
	Status      TaskStatus        `json:"status"`
	Result      json.RawMessage   `json:"result,omitempty"`
	Error       string            `json:"error,omitempty"`
	RetryCount  int               `json:"retry_count"`
	MaxRetries  int               `json:"max_retries"`
	CreatedAt   time.Time         `json:"created_at"`
	ScheduledAt time.Time         `json:"scheduled_at"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Timeout     time.Duration     `json:"timeout"`
	WorkerID    string            `json:"worker_id,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func NewTask(taskType string, priority Priority, payload json.RawMessage) *Task {
	now := time.Now()
	return &Task{
		ID:          uuid.New().String(),
		Type:        taskType,
		Priority:    priority,
		Payload:     payload,
		Status:      TaskStatusPending,
		RetryCount:  0,
		MaxRetries:  3,
		CreatedAt:   now,
		ScheduledAt: now,
		Timeout:     5 * time.Minute,
		Metadata:    make(map[string]string),
	}
}

func (t *Task) Validate() error {
	if t.ID == "" {
		return errors.New("task ID is required")
	}
	if t.Type == "" {
		return errors.New("task type is required")
	}
	if t.Priority < PriorityLow || t.Priority > PriorityCritical {
		return errors.New("invalid priority")
	}
	if len(t.Payload) == 0 {
		return errors.New("task payload is required")
	}
	if t.MaxRetries < 0 {
		return errors.New("max retries must be non-negative")
	}
	if t.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}
	return nil
}

func (t *Task) MarkProcessing(workerID string) {
	t.Status = TaskStatusProcessing
	t.WorkerID = workerID
	now := time.Now()
	t.StartedAt = &now
}

func (t *Task) MarkCompleted(result json.RawMessage) {
	t.Status = TaskStatusCompleted
	t.Result = result
	now := time.Now()
	t.CompletedAt = &now
}

func (t *Task) MarkFailed(err error) {
	t.Status = TaskStatusFailed
	t.Error = err.Error()
	t.RetryCount++
	now := time.Now()
	t.CompletedAt = &now
}

func (t *Task) MarkDead() {
	t.Status = TaskStatusDead
	now := time.Now()
	t.CompletedAt = &now
}

func (t *Task) MarkCancelled() {
	t.Status = TaskStatusCancelled
	now := time.Now()
	t.CompletedAt = &now
}

func (t *Task) ShouldRetry() bool {
	return t.RetryCount < t.MaxRetries
}

type Queue struct {
	Name           string    `json:"name"`
	Priority       Priority  `json:"priority"`
	Depth          int64     `json:"depth"`
	ProcessedCount int64     `json:"processed_count"`
	FailedCount    int64     `json:"failed_count"`
	AvgLatency     float64   `json:"avg_latency"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type TaskFilter struct {
	Status    []TaskStatus
	Priority  []Priority
	Type      string
	WorkerID  string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

var (
	ErrTaskNotFound         = errors.New("task not found")
	ErrInvalidTask          = errors.New("invalid task")
	ErrNoTaskAvailable      = errors.New("no task available")
	ErrTaskTimeout          = errors.New("task execution timeout")
	ErrMaxRetriesReached    = errors.New("max retries reached")
	ErrTaskProcessingFailed = errors.New("task processing failed")
)

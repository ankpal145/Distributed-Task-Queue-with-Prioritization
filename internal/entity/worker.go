package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type WorkerStatus string

const (
	WorkerStatusActive WorkerStatus = "active"
	WorkerStatusIdle   WorkerStatus = "idle"
	WorkerStatusBusy   WorkerStatus = "busy"
	WorkerStatusDead   WorkerStatus = "dead"
)

type Worker struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Status          WorkerStatus      `json:"status"`
	Queues          []string          `json:"queues"`
	Concurrency     int               `json:"concurrency"`
	LastSeen        time.Time         `json:"last_seen"`
	ProcessingTasks []string          `json:"processing_tasks"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	StartedAt       time.Time         `json:"started_at"`
	TasksProcessed  int64             `json:"tasks_processed"`
	TasksFailed     int64             `json:"tasks_failed"`
	Version         string            `json:"version"`
	Hostname        string            `json:"hostname"`
}

func NewWorker(name string, queues []string, concurrency int) *Worker {
	now := time.Now()
	return &Worker{
		ID:              uuid.New().String(),
		Name:            name,
		Status:          WorkerStatusIdle,
		Queues:          queues,
		Concurrency:     concurrency,
		LastSeen:        now,
		ProcessingTasks: make([]string, 0),
		Metadata:        make(map[string]string),
		StartedAt:       now,
		TasksProcessed:  0,
		TasksFailed:     0,
	}
}

func (w *Worker) Validate() error {
	if w.ID == "" {
		return errors.New("worker ID is required")
	}
	if w.Name == "" {
		return errors.New("worker name is required")
	}
	if w.Concurrency < 1 {
		return errors.New("concurrency must be at least 1")
	}
	if w.Concurrency > 100 {
		return errors.New("concurrency cannot exceed 100")
	}
	if len(w.Queues) == 0 {
		return errors.New("worker must handle at least one queue")
	}
	return nil
}

func (w *Worker) UpdateHeartbeat() {
	w.LastSeen = time.Now()
}

func (w *Worker) AddProcessingTask(taskID string) {
	w.ProcessingTasks = append(w.ProcessingTasks, taskID)
	w.updateStatus()
}

func (w *Worker) RemoveProcessingTask(taskID string) {
	for i, id := range w.ProcessingTasks {
		if id == taskID {
			w.ProcessingTasks = append(w.ProcessingTasks[:i], w.ProcessingTasks[i+1:]...)
			break
		}
	}
	w.updateStatus()
}

func (w *Worker) updateStatus() {
	if len(w.ProcessingTasks) == 0 {
		w.Status = WorkerStatusIdle
	} else if len(w.ProcessingTasks) >= w.Concurrency {
		w.Status = WorkerStatusBusy
	} else {
		w.Status = WorkerStatusActive
	}
}

func (w *Worker) IsAvailable() bool {
	return len(w.ProcessingTasks) < w.Concurrency && w.Status != WorkerStatusDead
}

func (w *Worker) IsDead(timeout time.Duration) bool {
	return time.Since(w.LastSeen) > timeout
}

func (w *Worker) IncrementProcessed() {
	w.TasksProcessed++
}

func (w *Worker) IncrementFailed() {
	w.TasksFailed++
}

type WorkerStats struct {
	WorkerID       string    `json:"worker_id"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	TasksProcessed int64     `json:"tasks_processed"`
	TasksFailed    int64     `json:"tasks_failed"`
	SuccessRate    float64   `json:"success_rate"`
	CurrentLoad    int       `json:"current_load"`
	MaxLoad        int       `json:"max_load"`
	Uptime         string    `json:"uptime"`
	LastSeen       time.Time `json:"last_seen"`
}

func (w *Worker) GetStats() *WorkerStats {
	successRate := 0.0
	total := w.TasksProcessed + w.TasksFailed
	if total > 0 {
		successRate = float64(w.TasksProcessed) / float64(total) * 100
	}

	return &WorkerStats{
		WorkerID:       w.ID,
		Name:           w.Name,
		Status:         string(w.Status),
		TasksProcessed: w.TasksProcessed,
		TasksFailed:    w.TasksFailed,
		SuccessRate:    successRate,
		CurrentLoad:    len(w.ProcessingTasks),
		MaxLoad:        w.Concurrency,
		Uptime:         time.Since(w.StartedAt).String(),
		LastSeen:       w.LastSeen,
	}
}

var (
	ErrWorkerNotFound    = errors.New("worker not found")
	ErrInvalidWorker     = errors.New("invalid worker")
	ErrWorkerUnavailable = errors.New("worker unavailable")
	ErrWorkerDead        = errors.New("worker is dead")
)

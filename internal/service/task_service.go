package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/ankurpal/distributed-task-queue/config"
	"github.com/ankurpal/distributed-task-queue/internal/constants"
	"github.com/ankurpal/distributed-task-queue/internal/entity"
	"github.com/ankurpal/distributed-task-queue/internal/repo"
	"github.com/ankurpal/distributed-task-queue/pkg/logger"
	"github.com/google/uuid"
)

type TaskService struct {
	taskRepo   repo.TaskRepository
	workerRepo repo.WorkerRepository
	metrics    MetricsCollector
	config     *TaskConfig
	logger     *logger.Logger
}

type TaskConfig struct {
	DefaultTimeout    time.Duration
	DefaultMaxRetries int
	RetryBaseDelay    time.Duration
	RetryMaxDelay     time.Duration
	CleanupInterval   time.Duration
}

func DefaultTaskConfig() *TaskConfig {
	return &TaskConfig{
		DefaultTimeout:    constants.DefaultTaskTimeout,
		DefaultMaxRetries: constants.DefaultMaxRetries,
		RetryBaseDelay:    constants.DefaultRetryDelay,
		RetryMaxDelay:     constants.DefaultMaxBackoffTime,
		CleanupInterval:   10 * time.Minute,
	}
}

func NewTaskConfigFromConfig(cfg *config.Configuration) *TaskConfig {
	return &TaskConfig{
		DefaultTimeout:    cfg.Task.DefaultTimeout,
		DefaultMaxRetries: cfg.Task.DefaultMaxRetries,
		RetryBaseDelay:    cfg.Task.RetryBaseDelay,
		RetryMaxDelay:     cfg.Task.RetryMaxDelay,
		CleanupInterval:   cfg.Task.CleanupInterval,
	}
}

func NewTaskService(
	taskRepo repo.TaskRepository,
	workerRepo repo.WorkerRepository,
	metrics MetricsCollector,
	config *TaskConfig,
	log *logger.Logger,
) *TaskService {
	if config == nil {
		config = DefaultTaskConfig()
	}

	return &TaskService{
		taskRepo:   taskRepo,
		workerRepo: workerRepo,
		metrics:    metrics,
		config:     config,
		logger:     log,
	}
}

type SubmitTaskRequest struct {
	Type        string            `json:"type"`
	Priority    entity.Priority   `json:"priority"`
	Payload     json.RawMessage   `json:"payload"`
	MaxRetries  int               `json:"max_retries,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func (r *SubmitTaskRequest) Validate() error {
	if r.Type == "" {
		return errors.New("task type is required")
	}
	if len(r.Payload) == 0 {
		return errors.New("task payload is required")
	}
	if r.Priority < entity.PriorityLow || r.Priority > entity.PriorityCritical {
		return errors.New("invalid priority")
	}
	return nil
}

func (s *TaskService) SubmitTask(ctx context.Context, req *SubmitTaskRequest) (*entity.Task, error) {
	if err := req.Validate(); err != nil {
		s.logger.WarnWithCtx(ctx, "Invalid task submission request: %v", err)
		return nil, err
	}

	if req.MaxRetries == 0 {
		req.MaxRetries = s.config.DefaultMaxRetries
	}
	if req.Timeout == 0 {
		req.Timeout = s.config.DefaultTimeout
	}

	now := time.Now()
	scheduledAt := now
	if req.ScheduledAt != nil {
		scheduledAt = *req.ScheduledAt
	}

	task := &entity.Task{
		ID:          uuid.New().String(),
		Type:        req.Type,
		Priority:    req.Priority,
		Payload:     req.Payload,
		Status:      entity.TaskStatusPending,
		MaxRetries:  req.MaxRetries,
		RetryCount:  0,
		CreatedAt:   now,
		ScheduledAt: scheduledAt,
		Timeout:     req.Timeout,
		Metadata:    req.Metadata,
	}

	if task.Metadata == nil {
		task.Metadata = make(map[string]string)
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		s.logger.ErrorWithCtx(ctx, "Failed to create task: %v", err)
		return nil, err
	}

	if err := s.taskRepo.Enqueue(ctx, task); err != nil {
		s.logger.ErrorWithCtx(ctx, "Failed to enqueue task: %v", err)
		return nil, err
	}

	s.metrics.IncrementCounter(constants.MetricTasksSubmittedTotal, map[string]string{
		constants.LabelPriority: task.Priority.String(),
		constants.LabelType:     task.Type,
	})

	return task, nil
}

func (s *TaskService) GetTask(ctx context.Context, taskID string) (*entity.Task, error) {
	task, err := s.taskRepo.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) CancelTask(ctx context.Context, taskID string) error {
	task, err := s.taskRepo.Get(ctx, taskID)
	if err != nil {
		return err
	}

	if task.Status != entity.TaskStatusPending {
		s.logger.WarnWithCtx(ctx, "Cannot cancel task %s: status is %s", taskID, task.Status)
		return errors.New("task cannot be cancelled: invalid status")
	}

	task.MarkCancelled()
	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.ErrorWithCtx(ctx, "Failed to cancel task %s: %v", taskID, err)
		return err
	}

	s.metrics.IncrementCounter(constants.MetricTasksCancelledTotal, map[string]string{
		constants.LabelType: task.Type,
	})

	return nil
}

func (s *TaskService) ListTasks(ctx context.Context, filter *entity.TaskFilter) ([]*entity.Task, error) {
	tasks, err := s.taskRepo.List(ctx, filter)
	if err != nil {
		s.logger.ErrorWithCtx(ctx, "Failed to list tasks: %v", err)
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) HandleTaskSuccess(ctx context.Context, taskID string, result json.RawMessage) error {
	task, err := s.taskRepo.Get(ctx, taskID)
	if err != nil {
		return err
	}

	task.MarkCompleted(result)
	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.ErrorWithCtx(ctx, "Failed to update task %s after completion: %v", taskID, err)
		return err
	}

	if task.StartedAt != nil {
		latency := task.CompletedAt.Sub(*task.StartedAt).Seconds()
		s.metrics.ObserveHistogram(constants.MetricTaskDurationSeconds, latency, map[string]string{
			constants.LabelType:     task.Type,
			constants.LabelPriority: task.Priority.String(),
		})
	}

	s.metrics.IncrementCounter(constants.MetricTasksCompletedTotal, map[string]string{
		constants.LabelType:     task.Type,
		constants.LabelPriority: task.Priority.String(),
		constants.LabelWorkerID: task.WorkerID,
	})

	return nil
}

func (s *TaskService) HandleTaskFailure(ctx context.Context, taskID string, taskErr error) error {
	task, err := s.taskRepo.Get(ctx, taskID)
	if err != nil {
		return err
	}

	task.MarkFailed(taskErr)

	if !task.ShouldRetry() {
		task.MarkDead()

		if err := s.taskRepo.MoveToDLQ(ctx, task); err != nil {
			s.logger.ErrorWithCtx(ctx, "Failed to move task %s to DLQ: %v", taskID, err)
			return err
		}

		s.metrics.IncrementCounter(constants.MetricTasksDeadTotal, map[string]string{
			constants.LabelType: task.Type,
		})

		return nil
	}

	delay := s.calculateBackoff(task.RetryCount)

	if err := s.taskRepo.RequeueWithDelay(ctx, task, delay); err != nil {
		s.logger.ErrorWithCtx(ctx, "Failed to requeue task %s with delay: %v", task.ID, err)
		return err
	}

	s.metrics.IncrementCounter(constants.MetricTasksRetriedTotal, map[string]string{
		constants.LabelType:  task.Type,
		constants.LabelRetry: fmt.Sprintf("%d", task.RetryCount),
	})

	return nil
}

func (s *TaskService) calculateBackoff(retryCount int) time.Duration {
	base := s.config.RetryBaseDelay
	maxDelay := s.config.RetryMaxDelay

	delay := base * time.Duration(math.Pow(2, float64(retryCount-1)))
	if delay > maxDelay {
		delay = maxDelay
	}

	jitter := time.Duration(rand.Int63n(int64(time.Second)))
	return delay + jitter
}

func (s *TaskService) GetQueueStats(ctx context.Context) (map[string]*entity.Queue, error) {
	return s.taskRepo.GetQueueStats(ctx)
}

func (s *TaskService) GetDeadLetterQueue(ctx context.Context, limit int) ([]*entity.Task, error) {
	return s.taskRepo.GetFromDLQ(ctx, limit)
}

func (s *TaskService) RetryDeadTask(ctx context.Context, taskID string) error {
	task, err := s.taskRepo.Get(ctx, taskID)
	if err != nil {
		return err
	}

	if task.Status != entity.TaskStatusDead {
		s.logger.WarnWithCtx(ctx, "Cannot retry task %s: not in dead letter queue (status: %s)", taskID, task.Status)
		return errors.New("task is not in dead letter queue")
	}

	task.Status = entity.TaskStatusPending
	task.RetryCount = 0
	task.Error = ""
	task.ScheduledAt = time.Now()
	task.WorkerID = ""
	task.StartedAt = nil
	task.CompletedAt = nil

	if err := s.taskRepo.Update(ctx, task); err != nil {
		s.logger.ErrorWithCtx(ctx, "Failed to update task %s: %v", taskID, err)
		return err
	}

	if err := s.taskRepo.Enqueue(ctx, task); err != nil {
		s.logger.ErrorWithCtx(ctx, "Failed to enqueue task %s: %v", taskID, err)
		return err
	}

	return nil
}

func (s *TaskService) StartCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := s.taskRepo.CleanupExpiredTasks(ctx, 7*24*time.Hour)
			if err != nil {
				s.logger.ErrorWithCtx(ctx, "Failed to cleanup expired tasks: %v", err)
				continue
			}
			if count > 0 {
				s.metrics.IncrementCounter(constants.MetricTasksCleanedUpTotal, map[string]string{
					constants.LabelCount: fmt.Sprintf("%d", count),
				})
			}

			count, err = s.taskRepo.RecoverOrphanedTasks(ctx, 10*time.Minute)
			if err != nil {
				s.logger.ErrorWithCtx(ctx, "Failed to recover orphaned tasks: %v", err)
				continue
			}
			if count > 0 {
				s.metrics.IncrementCounter(constants.MetricTasksRecoveredTotal, map[string]string{
					constants.LabelCount: fmt.Sprintf("%d", count),
				})
			}
		}
	}
}

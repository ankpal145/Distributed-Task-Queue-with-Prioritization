package repo

import (
	"context"
	"time"

	"github.com/ankurpal/distributed-task-queue/internal/entity"
)

type TaskRepository interface {
	Create(ctx context.Context, task *entity.Task) error

	Get(ctx context.Context, taskID string) (*entity.Task, error)

	Update(ctx context.Context, task *entity.Task) error

	Delete(ctx context.Context, taskID string) error

	Enqueue(ctx context.Context, task *entity.Task) error

	Dequeue(ctx context.Context, workerID string) (*entity.Task, error)

	RequeueWithDelay(ctx context.Context, task *entity.Task, delay time.Duration) error

	MoveToDLQ(ctx context.Context, task *entity.Task) error

	GetFromDLQ(ctx context.Context, limit int) ([]*entity.Task, error)

	List(ctx context.Context, filter *entity.TaskFilter) ([]*entity.Task, error)

	GetQueueStats(ctx context.Context) (map[string]*entity.Queue, error)

	GetQueueDepth(ctx context.Context, priority entity.Priority) (int64, error)

	CleanupExpiredTasks(ctx context.Context, ttl time.Duration) (int64, error)

	GetProcessingTasks(ctx context.Context) ([]*entity.Task, error)

	RecoverOrphanedTasks(ctx context.Context, timeout time.Duration) (int64, error)
}

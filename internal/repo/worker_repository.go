package repo

import (
	"context"
	"time"

	"github.com/ankurpal/distributed-task-queue/internal/entity"
)

type WorkerRepository interface {
	Register(ctx context.Context, worker *entity.Worker) error

	Get(ctx context.Context, workerID string) (*entity.Worker, error)

	Update(ctx context.Context, worker *entity.Worker) error

	Deregister(ctx context.Context, workerID string) error

	UpdateHeartbeat(ctx context.Context, workerID string) error

	ListActive(ctx context.Context) ([]*entity.Worker, error)

	FindDeadWorkers(ctx context.Context, timeout time.Duration) ([]*entity.Worker, error)

	GetStats(ctx context.Context, workerID string) (*entity.WorkerStats, error)

	GetAllStats(ctx context.Context) ([]*entity.WorkerStats, error)

	CleanupDeadWorkers(ctx context.Context, timeout time.Duration) (int64, error)
}

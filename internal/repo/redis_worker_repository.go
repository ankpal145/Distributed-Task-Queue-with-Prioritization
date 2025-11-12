package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ankurpal/distributed-task-queue/internal/entity"
	"github.com/go-redis/redis/v8"
)

type RedisWorkerRepository struct {
	client *redis.Client
}

func NewRedisWorkerRepository(client *redis.Client) *RedisWorkerRepository {
	return &RedisWorkerRepository{
		client: client,
	}
}

func AsWorkerRepository(impl *RedisWorkerRepository) WorkerRepository {
	return impl
}

const (
	workerKeyPrefix  = "worker:"
	workersActiveKey = "workers:active"
)

func workerKey(workerID string) string {
	return fmt.Sprintf("%s%s", workerKeyPrefix, workerID)
}

func (r *RedisWorkerRepository) Register(ctx context.Context, worker *entity.Worker) error {
	if err := worker.Validate(); err != nil {
		return fmt.Errorf("invalid worker: %w", err)
	}

	data, err := json.Marshal(worker)
	if err != nil {
		return fmt.Errorf("failed to marshal worker: %w", err)
	}

	pipe := r.client.Pipeline()

	key := workerKey(worker.ID)
	pipe.Set(ctx, key, data, 0)

	pipe.SAdd(ctx, workersActiveKey, worker.ID)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	return nil
}

func (r *RedisWorkerRepository) Get(ctx context.Context, workerID string) (*entity.Worker, error) {
	key := workerKey(workerID)
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, entity.ErrWorkerNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get worker: %w", err)
	}

	var worker entity.Worker
	if err := json.Unmarshal([]byte(data), &worker); err != nil {
		return nil, fmt.Errorf("failed to unmarshal worker: %w", err)
	}

	return &worker, nil
}

func (r *RedisWorkerRepository) Update(ctx context.Context, worker *entity.Worker) error {
	data, err := json.Marshal(worker)
	if err != nil {
		return fmt.Errorf("failed to marshal worker: %w", err)
	}

	key := workerKey(worker.ID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check worker existence: %w", err)
	}
	if exists == 0 {
		return entity.ErrWorkerNotFound
	}

	err = r.client.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to update worker: %w", err)
	}

	return nil
}

func (r *RedisWorkerRepository) Deregister(ctx context.Context, workerID string) error {
	pipe := r.client.Pipeline()

	key := workerKey(workerID)
	pipe.Del(ctx, key)

	pipe.SRem(ctx, workersActiveKey, workerID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to deregister worker: %w", err)
	}

	return nil
}

func (r *RedisWorkerRepository) UpdateHeartbeat(ctx context.Context, workerID string) error {
	worker, err := r.Get(ctx, workerID)
	if err != nil {
		return err
	}

	worker.UpdateHeartbeat()
	return r.Update(ctx, worker)
}

func (r *RedisWorkerRepository) ListActive(ctx context.Context) ([]*entity.Worker, error) {
	workerIDs, err := r.client.SMembers(ctx, workersActiveKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list active workers: %w", err)
	}

	workers := make([]*entity.Worker, 0, len(workerIDs))
	for _, workerID := range workerIDs {
		worker, err := r.Get(ctx, workerID)
		if err != nil {
			continue
		}
		workers = append(workers, worker)
	}

	return workers, nil
}

func (r *RedisWorkerRepository) FindDeadWorkers(ctx context.Context, timeout time.Duration) ([]*entity.Worker, error) {
	workers, err := r.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	deadWorkers := make([]*entity.Worker, 0)
	for _, worker := range workers {
		if worker.IsDead(timeout) {
			worker.Status = entity.WorkerStatusDead
			deadWorkers = append(deadWorkers, worker)
		}
	}

	return deadWorkers, nil
}

func (r *RedisWorkerRepository) GetStats(ctx context.Context, workerID string) (*entity.WorkerStats, error) {
	worker, err := r.Get(ctx, workerID)
	if err != nil {
		return nil, err
	}

	return worker.GetStats(), nil
}

func (r *RedisWorkerRepository) GetAllStats(ctx context.Context) ([]*entity.WorkerStats, error) {
	workers, err := r.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	stats := make([]*entity.WorkerStats, 0, len(workers))
	for _, worker := range workers {
		stats = append(stats, worker.GetStats())
	}

	return stats, nil
}

func (r *RedisWorkerRepository) CleanupDeadWorkers(ctx context.Context, timeout time.Duration) (int64, error) {
	deadWorkers, err := r.FindDeadWorkers(ctx, timeout)
	if err != nil {
		return 0, fmt.Errorf("failed to find dead workers: %w", err)
	}

	var count int64
	for _, worker := range deadWorkers {
		if err := r.Deregister(ctx, worker.ID); err != nil {
			continue
		}
		count++
	}

	return count, nil
}

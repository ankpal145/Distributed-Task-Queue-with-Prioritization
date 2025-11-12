package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ankurpal/distributed-task-queue/internal/entity"
	"github.com/go-redis/redis/v8"
)

type RedisTaskRepository struct {
	client *redis.Client
}

func NewRedisTaskRepository(client *redis.Client) *RedisTaskRepository {
	return &RedisTaskRepository{
		client: client,
	}
}

func AsTaskRepository(impl *RedisTaskRepository) TaskRepository {
	return impl
}

const (
	taskKeyPrefix       = "task:"
	queueKeyPrefix      = "queue:"
	processingKey       = "queue:processing"
	dlqKey              = "queue:dlq"
	queueStatsKeyPrefix = "queue:stats:"
)

func taskKey(taskID string) string {
	return fmt.Sprintf("%s%s", taskKeyPrefix, taskID)
}

func queueKey(priority entity.Priority) string {
	return fmt.Sprintf("%spending:%d", queueKeyPrefix, priority)
}

func queueStatsKey(priority entity.Priority) string {
	return fmt.Sprintf("%s%d", queueStatsKeyPrefix, priority)
}

func (r *RedisTaskRepository) Create(ctx context.Context, task *entity.Task) error {
	if err := task.Validate(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	key := taskKey(task.ID)
	err = r.client.Set(ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

func (r *RedisTaskRepository) Get(ctx context.Context, taskID string) (*entity.Task, error) {
	key := taskKey(taskID)
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, entity.ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var task entity.Task
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

func (r *RedisTaskRepository) Update(ctx context.Context, task *entity.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	key := taskKey(task.ID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check task existence: %w", err)
	}
	if exists == 0 {
		return entity.ErrTaskNotFound
	}

	err = r.client.Set(ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (r *RedisTaskRepository) Delete(ctx context.Context, taskID string) error {
	key := taskKey(taskID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (r *RedisTaskRepository) Enqueue(ctx context.Context, task *entity.Task) error {
	pipe := r.client.Pipeline()

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	key := taskKey(task.ID)
	pipe.Set(ctx, key, data, 24*time.Hour)

	qKey := queueKey(task.Priority)
	score := float64(task.ScheduledAt.Unix())
	pipe.ZAdd(ctx, qKey, &redis.Z{
		Score:  score,
		Member: task.ID,
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

func (r *RedisTaskRepository) Dequeue(ctx context.Context, workerID string) (*entity.Task, error) {
	priorities := []entity.Priority{
		entity.PriorityCritical,
		entity.PriorityHigh,
		entity.PriorityNormal,
		entity.PriorityLow,
	}

	now := float64(time.Now().Unix())

	for _, priority := range priorities {
		qKey := queueKey(priority)

		script := redis.NewScript(`
			local queueKey = KEYS[1]
			local processingKey = KEYS[2]
			local now = tonumber(ARGV[1])
			local workerID = ARGV[2]

			-- Get the first task that's ready to process
			local tasks = redis.call('ZRANGEBYSCORE', queueKey, '-inf', now, 'LIMIT', 0, 1)
			if #tasks == 0 then
				return nil
			end

			local taskID = tasks[1]
			
			-- Remove from pending queue
			redis.call('ZREM', queueKey, taskID)
			
			-- Add to processing set
			redis.call('HSET', processingKey, taskID, workerID)
			
			return taskID
		`)

		result, err := script.Run(ctx, r.client, []string{qKey, processingKey}, now, workerID).Result()
		if err == redis.Nil || result == nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to dequeue task: %w", err)
		}

		taskID, ok := result.(string)
		if !ok {
			continue
		}

		task, err := r.Get(ctx, taskID)
		if err != nil {
			r.client.HDel(ctx, processingKey, taskID)
			continue
		}

		task.MarkProcessing(workerID)
		if err := r.Update(ctx, task); err != nil {
			return nil, fmt.Errorf("failed to update task status: %w", err)
		}

		return task, nil
	}

	return nil, entity.ErrNoTaskAvailable
}

func (r *RedisTaskRepository) RequeueWithDelay(ctx context.Context, task *entity.Task, delay time.Duration) error {
	task.Status = entity.TaskStatusPending
	task.ScheduledAt = time.Now().Add(delay)
	task.WorkerID = ""
	task.StartedAt = nil

	pipe := r.client.Pipeline()

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	key := taskKey(task.ID)
	pipe.Set(ctx, key, data, 24*time.Hour)

	pipe.HDel(ctx, processingKey, task.ID)

	qKey := queueKey(task.Priority)
	score := float64(task.ScheduledAt.Unix())
	pipe.ZAdd(ctx, qKey, &redis.Z{
		Score:  score,
		Member: task.ID,
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to requeue task: %w", err)
	}

	return nil
}

func (r *RedisTaskRepository) MoveToDLQ(ctx context.Context, task *entity.Task) error {
	task.MarkDead()

	pipe := r.client.Pipeline()

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}
	key := taskKey(task.ID)
	pipe.Set(ctx, key, data, 7*24*time.Hour)

	pipe.HDel(ctx, processingKey, task.ID)

	pipe.LPush(ctx, dlqKey, task.ID)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to move task to DLQ: %w", err)
	}

	return nil
}

func (r *RedisTaskRepository) GetFromDLQ(ctx context.Context, limit int) ([]*entity.Task, error) {
	taskIDs, err := r.client.LRange(ctx, dlqKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get DLQ tasks: %w", err)
	}

	tasks := make([]*entity.Task, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		task, err := r.Get(ctx, taskID)
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *RedisTaskRepository) List(ctx context.Context, filter *entity.TaskFilter) ([]*entity.Task, error) {

	var taskIDs []string

	iter := r.client.Scan(ctx, 0, taskKeyPrefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		taskID := iter.Val()[len(taskKeyPrefix):]
		taskIDs = append(taskIDs, taskID)
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan tasks: %w", err)
	}

	tasks := make([]*entity.Task, 0)
	for _, taskID := range taskIDs {
		task, err := r.Get(ctx, taskID)
		if err != nil {
			continue
		}

		if filter != nil {
			if len(filter.Status) > 0 {
				match := false
				for _, status := range filter.Status {
					if task.Status == status {
						match = true
						break
					}
				}
				if !match {
					continue
				}
			}

			if len(filter.Priority) > 0 {
				match := false
				for _, priority := range filter.Priority {
					if task.Priority == priority {
						match = true
						break
					}
				}
				if !match {
					continue
				}
			}

			if filter.Type != "" && task.Type != filter.Type {
				continue
			}

			if filter.WorkerID != "" && task.WorkerID != filter.WorkerID {
				continue
			}
		}

		tasks = append(tasks, task)
	}

	if filter != nil {
		start := filter.Offset
		if start > len(tasks) {
			start = len(tasks)
		}
		end := start + filter.Limit
		if end > len(tasks) || filter.Limit == 0 {
			end = len(tasks)
		}
		tasks = tasks[start:end]
	}

	return tasks, nil
}

func (r *RedisTaskRepository) GetQueueStats(ctx context.Context) (map[string]*entity.Queue, error) {
	priorities := []entity.Priority{
		entity.PriorityCritical,
		entity.PriorityHigh,
		entity.PriorityNormal,
		entity.PriorityLow,
	}

	stats := make(map[string]*entity.Queue)
	for _, priority := range priorities {
		qKey := queueKey(priority)

		depth, err := r.client.ZCard(ctx, qKey).Result()
		if err != nil {
			depth = 0
		}

		stats[priority.String()] = &entity.Queue{
			Name:      priority.String(),
			Priority:  priority,
			Depth:     depth,
			UpdatedAt: time.Now(),
		}
	}

	return stats, nil
}

func (r *RedisTaskRepository) GetQueueDepth(ctx context.Context, priority entity.Priority) (int64, error) {
	qKey := queueKey(priority)
	depth, err := r.client.ZCard(ctx, qKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue depth: %w", err)
	}
	return depth, nil
}

func (r *RedisTaskRepository) CleanupExpiredTasks(ctx context.Context, ttl time.Duration) (int64, error) {

	cutoff := time.Now().Add(-ttl)
	var count int64

	iter := r.client.Scan(ctx, 0, taskKeyPrefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		taskID := key[len(taskKeyPrefix):]

		task, err := r.Get(ctx, taskID)
		if err != nil {
			continue
		}

		if task.CompletedAt != nil && task.CompletedAt.Before(cutoff) {
			r.Delete(ctx, taskID)
			count++
		}
	}

	return count, iter.Err()
}

func (r *RedisTaskRepository) GetProcessingTasks(ctx context.Context) ([]*entity.Task, error) {
	taskIDs, err := r.client.HKeys(ctx, processingKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get processing tasks: %w", err)
	}

	tasks := make([]*entity.Task, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		task, err := r.Get(ctx, taskID)
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *RedisTaskRepository) RecoverOrphanedTasks(ctx context.Context, timeout time.Duration) (int64, error) {
	processingTasks, err := r.GetProcessingTasks(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get processing tasks: %w", err)
	}

	var count int64
	cutoff := time.Now().Add(-timeout)

	for _, task := range processingTasks {
		if task.StartedAt != nil && task.StartedAt.Before(cutoff) {
			if err := r.RequeueWithDelay(ctx, task, 0); err != nil {
				continue
			}
			count++
		}
	}

	return count, nil
}

package processors

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/ankurpal/distributed-task-queue/internal/entity"
	"github.com/ankurpal/distributed-task-queue/pkg/contexts"
	"github.com/ankurpal/distributed-task-queue/pkg/logger"
	"github.com/google/uuid"
)

type DefaultTaskProcessor struct {
	logger *logger.Logger
}

var (
	defaultTaskProcessorDoOnce sync.Once
	defaultTaskProcessorStruct DefaultTaskProcessor
)

func NewDefaultTaskProcessor(logger *logger.Logger) *DefaultTaskProcessor {
	defaultTaskProcessorDoOnce.Do(func() {
		defaultTaskProcessorStruct = DefaultTaskProcessor{
			logger: logger,
		}
	})
	return &defaultTaskProcessorStruct
}

func (d *DefaultTaskProcessor) Process(ctx context.Context, task *entity.Task) (processorError error) {
	ctx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, contexts.RequestIDKey, requestID)

	d.logger.InfoWithCtx(ctx, "Message processing started by DefaultTaskProcessor for task %s (type: %s)", task.ID, task.Type)

	defer func() {
		if er := recover(); er != nil {
			d.logger.ErrorWithCtx(ctx, "Recovered from panic in DefaultTaskProcessor: %v", er)
			processorError = entity.ErrTaskProcessingFailed
		}
	}()

	var payload map[string]interface{}
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		d.logger.ErrorWithCtx(ctx, "Error while unmarshalling task payload for task %s: %v", task.ID, err)
		return err
	}

	d.logger.InfoWithCtx(ctx, "Processing task %s with payload: %v", task.ID, payload)

	time.Sleep(1 * time.Second)

	result := map[string]interface{}{
		"task_id":      task.ID,
		"task_type":    task.Type,
		"status":       "completed",
		"processed_at": time.Now().Format(time.RFC3339),
		"request_id":   requestID,
		"payload":      payload,
	}

	if _, marshalErr := json.Marshal(result); marshalErr != nil {
		d.logger.ErrorWithCtx(ctx, "Error while marshalling result for task %s: %v", task.ID, marshalErr)
		return marshalErr
	}

	d.logger.InfoWithCtx(ctx, "Message processing finished by DefaultTaskProcessor for task %s", task.ID)
	return nil
}

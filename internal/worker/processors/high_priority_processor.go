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

type HighPriorityTaskProcessor struct {
	logger *logger.Logger
}

var (
	highPriorityTaskProcessorDoOnce sync.Once
	highPriorityTaskProcessorStruct HighPriorityTaskProcessor
)

func NewHighPriorityTaskProcessor(logger *logger.Logger) *HighPriorityTaskProcessor {
	highPriorityTaskProcessorDoOnce.Do(func() {
		highPriorityTaskProcessorStruct = HighPriorityTaskProcessor{
			logger: logger,
		}
	})
	return &highPriorityTaskProcessorStruct
}

func (h *HighPriorityTaskProcessor) Process(ctx context.Context, task *entity.Task) (processorError error) {
	ctx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, contexts.RequestIDKey, requestID)

	h.logger.InfoWithCtx(ctx, "Message processing started by HighPriorityTaskProcessor for task %s", task.ID)

	defer func() {
		if er := recover(); er != nil {
			h.logger.ErrorWithCtx(ctx, "Recovered from panic in HighPriorityTaskProcessor: %v", er)
			processorError = entity.ErrTaskProcessingFailed
		}
	}()

	var payload map[string]interface{}
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		h.logger.ErrorWithCtx(ctx, "Error while unmarshalling task payload for task %s: %v", task.ID, err)
		return err
	}

	h.logger.InfoWithCtx(ctx, "Processing HIGH PRIORITY task %s with payload: %v", task.ID, payload)

	time.Sleep(500 * time.Millisecond)

	result := map[string]interface{}{
		"task_id":      task.ID,
		"task_type":    task.Type,
		"priority":     "HIGH",
		"status":       "completed",
		"processed_at": time.Now().Format(time.RFC3339),
		"request_id":   requestID,
		"payload":      payload,
	}

	if _, marshalErr := json.Marshal(result); marshalErr != nil {
		h.logger.ErrorWithCtx(ctx, "Error while marshalling result for task %s: %v", task.ID, marshalErr)
		return marshalErr
	}

	h.logger.InfoWithCtx(ctx, "Message processing finished by HighPriorityTaskProcessor for task %s", task.ID)
	return nil
}

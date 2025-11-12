package processors

import (
	"sync"
)

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

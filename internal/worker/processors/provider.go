package processors

import (
	"github.com/ankurpal/distributed-task-queue/internal/worker"
	"go.uber.org/fx"
)

var FxWorkerProcessorsProvider = fx.Options(
	fx.Provide(worker.NewWorkerConfigFromConfig),
	fx.Provide(worker.NewWorkerService),
	fx.Provide(NewTaskProcessors),
	fx.Provide(NewDefaultTaskProcessor),
	fx.Provide(NewHighPriorityTaskProcessor),
)

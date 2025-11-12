package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ankurpal/distributed-task-queue/config"
	"github.com/ankurpal/distributed-task-queue/internal/constants"
	"github.com/ankurpal/distributed-task-queue/internal/entity"
	"github.com/ankurpal/distributed-task-queue/internal/repo"
	"github.com/ankurpal/distributed-task-queue/internal/service"
	workerPkg "github.com/ankurpal/distributed-task-queue/internal/worker"
	"github.com/ankurpal/distributed-task-queue/internal/worker/consumer"
	"github.com/ankurpal/distributed-task-queue/internal/worker/processors"
	"github.com/ankurpal/distributed-task-queue/pkg/logger"
	"go.uber.org/fx"
)

type Worker struct {
	config         *config.Configuration
	logger         *logger.Logger
	ctx            context.Context
	cancelFunc     context.CancelFunc
	processor      *consumer.Processor
	taskProcessors *processors.TaskProcessors
}

func NewWorker(
	lifecycle fx.Lifecycle,
	config *config.Configuration,
	logger *logger.Logger,
	workerService *workerPkg.WorkerService,
	workerRepo repo.WorkerRepository,
	taskRepo repo.TaskRepository,
	taskService *service.TaskService,
	metrics service.MetricsCollector,
	taskProcessors *processors.TaskProcessors,
) *Worker {
	ctx, cancelFunc := context.WithCancel(context.Background())

	hostname, _ := os.Hostname()
	workerID := fmt.Sprintf("worker-%s-%d", hostname, time.Now().Unix())

	processorConfig := &consumer.ProcessorConfig{
		HeartbeatInterval: config.Worker.HeartbeatInterval,
		WorkerTimeout:     config.Worker.WorkerTimeout,
		Concurrency:       config.Worker.Concurrency,
		CleanupInterval:   config.Worker.CleanupInterval,
	}

	processor := consumer.NewProcessor(
		func(ctx context.Context, task *entity.Task) error {
			return processTask(ctx, task, taskProcessors, taskService, metrics, logger)
		},
		workerService,
		taskService,
		workerRepo,
		taskRepo,
		metrics,
		logger,
		workerID,
		[]string{constants.QueueDefault, constants.QueueHigh, constants.QueueMedium, constants.QueueLow},
		processorConfig,
	)

	worker := &Worker{
		config:         config,
		logger:         logger,
		ctx:            ctx,
		cancelFunc:     cancelFunc,
		processor:      processor,
		taskProcessors: taskProcessors,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			worker.logger.InfoWithCtx(ctx, "Starting worker...")
			go worker.Start(ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			worker.logger.InfoWithCtx(ctx, "Stopping worker...")
			worker.Stop(ctx)
			return nil
		},
	})

	return worker
}

func (w *Worker) Start(ctx context.Context) {
	go func() {
		http.HandleFunc(constants.HealthEndpoint, func(writer http.ResponseWriter, req *http.Request) {
			_, err := fmt.Fprintf(writer, "{\"status\":\"%s\"}", constants.StatusServing)
			if err != nil {
				w.logger.ErrorWithCtx(ctx, "Failed to write health response: %v", err)
			}
		})
		w.logger.InfoWithCtx(ctx, "Starting health monitoring at %s", constants.HealthCheckAddress)
		if err := http.ListenAndServe(constants.HealthCheckAddress, nil); err != nil {
			w.logger.ErrorWithCtx(ctx, "Failed to start health monitoring: %v", err)
		}
	}()

	w.processor.Start(w.ctx)

	w.waitForShutDown(ctx)
}

func (w *Worker) Stop(ctx context.Context) {
	w.logger.InfoWithCtx(ctx, "Shutting down worker...")
	w.processor.Stop(ctx)
	w.cancelFunc()
}

func (w *Worker) waitForShutDown(ctx context.Context) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	w.logger.InfoWithCtx(ctx, "Shutdown signal received")
}

func processTask(
	ctx context.Context,
	task *entity.Task,
	taskProcessors *processors.TaskProcessors,
	taskService *service.TaskService,
	metrics service.MetricsCollector,
	logger *logger.Logger,
) error {
	startTime := time.Now()

	logger.InfoWithCtx(ctx, "Routing task %s (type: %s, priority: %s) to appropriate processor", task.ID, task.Type, task.Priority)

	var processorErr error
	switch task.Priority {
	case entity.PriorityCritical, entity.PriorityHigh:
		processorErr = taskProcessors.HighPriorityTaskProcessor.Process(ctx, task)
	default:
		processorErr = taskProcessors.DefaultTaskProcessor.Process(ctx, task)
	}

	duration := time.Since(startTime)

	if processorErr != nil {
		logger.ErrorWithCtx(ctx, "Task %s processing failed: %v", task.ID, processorErr)
		taskService.HandleTaskFailure(ctx, task.ID, processorErr)

		metrics.IncrementCounter(constants.MetricTasksFailedTotal, map[string]string{
			constants.LabelType:     task.Type,
			constants.LabelPriority: task.Priority.String(),
		})
		return processorErr
	}

	result := map[string]interface{}{
		constants.MetadataTaskID: task.ID,
		constants.LabelType:      task.Type,
		constants.LabelPriority:  task.Priority.String(),
		constants.LabelStatus:    constants.StatusCompleted,
		"processed_at":           time.Now().Format(time.RFC3339),
	}

	resultJSON, _ := json.Marshal(result)
	taskService.HandleTaskSuccess(ctx, task.ID, resultJSON)

	metrics.IncrementCounter(constants.MetricTasksProcessedTotal, map[string]string{
		constants.LabelType:     task.Type,
		constants.LabelPriority: task.Priority.String(),
	})

	metrics.ObserveHistogram(constants.MetricTaskDurationSeconds, duration.Seconds(), map[string]string{
		constants.LabelType:     task.Type,
		constants.LabelPriority: task.Priority.String(),
	})

	logger.InfoWithCtx(ctx, "Task %s completed successfully in %v", task.ID, duration)
	return nil
}

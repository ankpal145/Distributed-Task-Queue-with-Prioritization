package consumer

import (
	"context"
	"os"
	"time"

	"github.com/ankurpal/distributed-task-queue/internal/constants"
	"github.com/ankurpal/distributed-task-queue/internal/entity"
	"github.com/ankurpal/distributed-task-queue/internal/repo"
	"github.com/ankurpal/distributed-task-queue/internal/service"
	"github.com/ankurpal/distributed-task-queue/internal/worker"
	"github.com/ankurpal/distributed-task-queue/pkg/contexts"
	"github.com/ankurpal/distributed-task-queue/pkg/logger"
)

type Processor struct {
	processorEngine func(ctx context.Context, task *entity.Task) error
	workerService   *worker.WorkerService
	taskService     *service.TaskService
	workerRepo      repo.WorkerRepository
	taskRepo        repo.TaskRepository
	metrics         service.MetricsCollector
	workerID        string
	queues          []string
	config          *ProcessorConfig
	logger          *logger.Logger
}

type ProcessorConfig struct {
	HeartbeatInterval time.Duration
	WorkerTimeout     time.Duration
	Concurrency       int
	CleanupInterval   time.Duration
}

func NewProcessor(
	pe func(ctx context.Context, task *entity.Task) error,
	workerService *worker.WorkerService,
	taskService *service.TaskService,
	workerRepo repo.WorkerRepository,
	taskRepo repo.TaskRepository,
	metrics service.MetricsCollector,
	logger *logger.Logger,
	workerID string,
	queues []string,
	config *ProcessorConfig,
) *Processor {
	return &Processor{
		processorEngine: pe,
		workerService:   workerService,
		taskService:     taskService,
		workerRepo:      workerRepo,
		taskRepo:        taskRepo,
		metrics:         metrics,
		workerID:        workerID,
		queues:          queues,
		config:          config,
		logger:          logger,
	}
}

func (p *Processor) Start(ctx context.Context) {
	defer func() {
		if er := recover(); er != nil {
			p.logger.ErrorWithCtx(ctx, "Recovered from panic in processor: %v", er)
		}
	}()

	hostname, _ := os.Hostname()
	req := &worker.RegisterWorkerRequest{
		Name:        p.workerID,
		Queues:      p.queues,
		Concurrency: p.config.Concurrency,
		Metadata:    map[string]string{constants.MetadataHostname: hostname},
	}

	registeredWorker, err := p.workerService.RegisterWorker(ctx, req)
	if err != nil {
		p.logger.ErrorWithCtx(ctx, "Failed to register worker: %v", err)
		return
	}

	p.workerID = registeredWorker.ID
	p.logger.InfoWithCtx(ctx, "Worker %s registered successfully", p.workerID)

	go p.heartbeatLoop(ctx, p.workerID)
	go p.workerService.StartCleanupWorker(ctx)

	p.startTaskConsumer(ctx)
}

func (p *Processor) Stop(ctx context.Context) {
	p.logger.InfoWithCtx(ctx, "Shutting down message processor...")

	if err := p.workerService.DeregisterWorker(ctx, p.workerID); err != nil {
		p.logger.ErrorWithCtx(ctx, "Failed to deregister worker %s: %v", p.workerID, err)
	} else {
		p.logger.InfoWithCtx(ctx, "Worker %s deregistered successfully", p.workerID)
	}
}

func (p *Processor) startTaskConsumer(ctx context.Context) {
	p.logger.InfoWithCtx(ctx, "Consuming tasks for worker %s on queues: %v", p.workerID, p.queues)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			taskCtx := contexts.NewConsumerCtx(p.workerID, p.queues)

			worker, fetchErr := p.workerRepo.Get(taskCtx, p.workerID)
			if fetchErr != nil {
				p.logger.ErrorWithCtx(taskCtx, "Failed to get worker %s: %v", p.workerID, fetchErr)
				time.Sleep(constants.SleepOnError)
				continue
			}

			if !worker.IsAvailable() {
				time.Sleep(constants.SleepOnWorkerBusy)
				continue
			}

			task, fetchErr := p.taskRepo.Dequeue(taskCtx, p.workerID)
			if fetchErr == entity.ErrNoTaskAvailable {
				time.Sleep(constants.SleepOnNoTask)
				continue
			}
			if fetchErr != nil {
				p.logger.ErrorWithCtx(taskCtx, "Error while fetching tasks: %v", fetchErr.Error())
				time.Sleep(constants.SleepOnError)
				continue
			}

			worker.AddProcessingTask(task.ID)
			p.workerRepo.Update(taskCtx, worker)

			processorError := p.processorEngine(taskCtx, task)

			if processorError != nil {
				p.logger.ErrorWithCtx(taskCtx, "Failed to process task %s: %v", task.ID, processorError)
			}

			worker.RemoveProcessingTask(task.ID)
			p.workerRepo.Update(taskCtx, worker)
		}
	}
}

func (p *Processor) heartbeatLoop(ctx context.Context, workerID string) {
	ticker := time.NewTicker(p.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.workerRepo.UpdateHeartbeat(ctx, workerID); err != nil {
				p.logger.ErrorWithCtx(ctx, "Failed to update heartbeat for worker %s: %v", workerID, err)
				continue
			}

			p.metrics.IncrementCounter(constants.MetricWorkerHeartbeatsTotal, map[string]string{
				constants.LabelWorkerID: workerID,
			})
		}
	}
}

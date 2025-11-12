package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ankurpal/distributed-task-queue/config"
	"github.com/ankurpal/distributed-task-queue/internal/constants"
	"github.com/ankurpal/distributed-task-queue/internal/entity"
	"github.com/ankurpal/distributed-task-queue/internal/repo"
	"github.com/ankurpal/distributed-task-queue/internal/service"
	"github.com/ankurpal/distributed-task-queue/pkg/logger"
)

type WorkerService struct {
	workerRepo repo.WorkerRepository
	metrics    service.MetricsCollector
	config     *WorkerConfig
	logger     *logger.Logger
}

type WorkerConfig struct {
	HeartbeatInterval  time.Duration
	WorkerTimeout      time.Duration
	DefaultConcurrency int
	CleanupInterval    time.Duration
}

func DefaultWorkerConfig() *WorkerConfig {
	return &WorkerConfig{
		HeartbeatInterval:  constants.DefaultHeartbeatInterval,
		WorkerTimeout:      constants.DefaultWorkerTimeout,
		DefaultConcurrency: constants.DefaultConcurrency,
		CleanupInterval:    constants.DefaultCleanupInterval,
	}
}

func NewWorkerConfigFromConfig(cfg *config.Configuration) *WorkerConfig {
	return &WorkerConfig{
		HeartbeatInterval:  cfg.Worker.HeartbeatInterval,
		WorkerTimeout:      cfg.Worker.WorkerTimeout,
		DefaultConcurrency: cfg.Worker.Concurrency,
		CleanupInterval:    cfg.Worker.CleanupInterval,
	}
}

func NewWorkerService(
	workerRepo repo.WorkerRepository,
	metrics service.MetricsCollector,
	config *WorkerConfig,
	log *logger.Logger,
) *WorkerService {
	if config == nil {
		config = DefaultWorkerConfig()
	}

	return &WorkerService{
		workerRepo: workerRepo,
		metrics:    metrics,
		config:     config,
		logger:     log,
	}
}

type RegisterWorkerRequest struct {
	Name        string            `json:"name"`
	Queues      []string          `json:"queues"`
	Concurrency int               `json:"concurrency,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func (r *RegisterWorkerRequest) Validate() error {
	if r.Name == "" {
		return errors.New("worker name is required")
	}
	if len(r.Queues) == 0 {
		return errors.New("at least one queue is required")
	}
	return nil
}

func (s *WorkerService) RegisterWorker(ctx context.Context, req *RegisterWorkerRequest) (*entity.Worker, error) {
	if err := req.Validate(); err != nil {
		s.logger.WarnWithCtx(ctx, "Invalid worker registration request: %v", err)
		return nil, err
	}

	concurrency := req.Concurrency
	if concurrency == 0 {
		concurrency = s.config.DefaultConcurrency
	}

	worker := entity.NewWorker(req.Name, req.Queues, concurrency)
	worker.Metadata = req.Metadata

	if err := s.workerRepo.Register(ctx, worker); err != nil {
		s.logger.ErrorWithCtx(ctx, "Failed to register worker %s: %v", worker.Name, err)
		return nil, err
	}

	s.metrics.IncrementGauge(constants.MetricWorkersActive, 1)

	return worker, nil
}

func (s *WorkerService) DeregisterWorker(ctx context.Context, workerID string) error {
	if err := s.workerRepo.Deregister(ctx, workerID); err != nil {
		s.logger.ErrorWithCtx(ctx, "Failed to deregister worker %s: %v", workerID, err)
		return err
	}

	s.metrics.DecrementGauge(constants.MetricWorkersActive, 1)

	return nil
}

func (s *WorkerService) GetWorker(ctx context.Context, workerID string) (*entity.Worker, error) {
	return s.workerRepo.Get(ctx, workerID)
}

func (s *WorkerService) ListWorkers(ctx context.Context) ([]*entity.Worker, error) {
	return s.workerRepo.ListActive(ctx)
}

func (s *WorkerService) GetWorkerStats(ctx context.Context, workerID string) (*entity.WorkerStats, error) {
	return s.workerRepo.GetStats(ctx, workerID)
}

func (s *WorkerService) GetAllWorkerStats(ctx context.Context) ([]*entity.WorkerStats, error) {
	return s.workerRepo.GetAllStats(ctx)
}

func (s *WorkerService) StartCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := s.workerRepo.CleanupDeadWorkers(ctx, s.config.WorkerTimeout)
			if err != nil {
				s.logger.ErrorWithCtx(ctx, "Failed to cleanup dead workers: %v", err)
				continue
			}
			if count > 0 {
				s.metrics.DecrementGauge(constants.MetricWorkersActive, float64(count))
				s.metrics.IncrementCounter(constants.MetricWorkersCleanedUpTotal, map[string]string{
					constants.LabelCount: fmt.Sprintf("%d", count),
				})
			}
		}
	}
}

package constants

import "time"

const (
	// Module types
	SERVER = "server"
	WORKER = "worker"

	// Queue names
	QueueDefault = "default"
	QueueHigh    = "high"
	QueueMedium  = "medium"
	QueueLow     = "low"

	// Status strings
	StatusCompleted = "completed"
	StatusServing   = "SERVING"
	StatusFailed    = "failed"

	// HTTP endpoints
	HealthEndpoint     = "/healthz"
	APIVersion         = "/api/v1"
	MetricsEndpoint    = "/metrics"
	HealthCheckAddress = "0.0.0.0:3000"
	DefaultGRPCPort    = ":9090"
	DefaultMetricsPort = 9091

	// Context keys
	RequestIDKey     = "request_id"
	CorrelationIDKey = "correlation_id"

	// Worker defaults
	DefaultConcurrency       = 10
	DefaultHeartbeatInterval = 5 * time.Second
	DefaultWorkerTimeout     = 30 * time.Second
	DefaultCleanupInterval   = 1 * time.Minute

	// Task defaults
	DefaultTaskTimeout    = 5 * time.Minute
	DefaultMaxRetries     = 3
	DefaultRetryDelay     = 1 * time.Second
	DefaultBackoffFactor  = 2.0
	DefaultMaxBackoffTime = 1 * time.Hour

	// Sleep durations
	SleepOnError       = 1 * time.Second
	SleepOnNoTask      = 500 * time.Millisecond
	SleepOnWorkerBusy  = 100 * time.Millisecond
	SleepAfterShutdown = 1 * time.Second

	// Metadata keys
	MetadataHostname = "hostname"
	MetadataTaskID   = "task_id"
	MetadataWorkerID = "worker_id"

	// Metric names
	MetricWorkersActive         = "workers_active"
	MetricWorkersCleanedUpTotal = "workers_cleaned_up_total"
	MetricWorkerHeartbeatsTotal = "worker_heartbeats_total"
	MetricTasksProcessedTotal   = "tasks_processed_total"
	MetricTasksFailedTotal      = "tasks_failed_total"
	MetricTaskDurationSeconds   = "task_duration_seconds"
	MetricTasksEnqueuedTotal    = "tasks_enqueued_total"
	MetricTasksSubmittedTotal   = "tasks_submitted_total"
	MetricTasksCancelledTotal   = "tasks_cancelled_total"
	MetricTasksCompletedTotal   = "tasks_completed_total"
	MetricTasksDeadTotal        = "tasks_dead_total"
	MetricTasksRetriedTotal     = "tasks_retried_total"
	MetricTasksCleanedUpTotal   = "tasks_cleaned_up_total"
	MetricTasksRecoveredTotal   = "tasks_recovered_total"

	// Metric label keys
	LabelWorkerID = "worker_id"
	LabelType     = "type"
	LabelPriority = "priority"
	LabelCount    = "count"
	LabelStatus   = "status"
	LabelRetry    = "retry"
)

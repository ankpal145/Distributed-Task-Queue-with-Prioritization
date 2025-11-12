package monitoring

import (
	"github.com/ankurpal/distributed-task-queue/internal/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusCollector struct {
	tasksSubmitted   *prometheus.CounterVec
	tasksCompleted   *prometheus.CounterVec
	tasksFailed      *prometheus.CounterVec
	tasksRetried     *prometheus.CounterVec
	tasksDead        *prometheus.CounterVec
	tasksCancelled   *prometheus.CounterVec
	taskDuration     *prometheus.HistogramVec
	queueDepth       *prometheus.GaugeVec
	workersActive    prometheus.Gauge
	workerHeartbeats *prometheus.CounterVec
}

func NewPrometheusCollector() *PrometheusCollector {
	return &PrometheusCollector{
		tasksSubmitted: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "task_queue_tasks_submitted_total",
				Help: "Total number of tasks submitted to the queue",
			},
			[]string{"priority", "type"},
		),
		tasksCompleted: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "task_queue_tasks_completed_total",
				Help: "Total number of tasks completed successfully",
			},
			[]string{"type", "priority", "worker_id"},
		),
		tasksFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "task_queue_tasks_failed_total",
				Help: "Total number of tasks that failed",
			},
			[]string{"type", "worker_id"},
		),
		tasksRetried: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "task_queue_tasks_retried_total",
				Help: "Total number of task retries",
			},
			[]string{"type", "retry"},
		),
		tasksDead: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "task_queue_tasks_dead_total",
				Help: "Total number of tasks moved to dead letter queue",
			},
			[]string{"type"},
		),
		tasksCancelled: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "task_queue_tasks_cancelled_total",
				Help: "Total number of tasks cancelled",
			},
			[]string{"type"},
		),
		taskDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "task_queue_task_duration_seconds",
				Help:    "Task execution duration in seconds",
				Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60, 120, 300},
			},
			[]string{"type", "priority"},
		),
		queueDepth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "task_queue_queue_depth",
				Help: "Current number of tasks in each priority queue",
			},
			[]string{"priority"},
		),
		workersActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "task_queue_workers_active",
				Help: "Current number of active workers",
			},
		),
		workerHeartbeats: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "task_queue_worker_heartbeats_total",
				Help: "Total number of worker heartbeats",
			},
			[]string{"worker_id"},
		),
	}
}

func AsMetricsCollector(impl *PrometheusCollector) service.MetricsCollector {
	return impl
}

func (p *PrometheusCollector) IncrementCounter(name string, labels map[string]string) {
	switch name {
	case "tasks_submitted_total":
		p.tasksSubmitted.With(prometheus.Labels(labels)).Inc()
	case "tasks_completed_total":
		p.tasksCompleted.With(prometheus.Labels(labels)).Inc()
	case "tasks_failed_total":
		p.tasksFailed.With(prometheus.Labels(labels)).Inc()
	case "tasks_retried_total":
		p.tasksRetried.With(prometheus.Labels(labels)).Inc()
	case "tasks_dead_total":
		p.tasksDead.With(prometheus.Labels(labels)).Inc()
	case "tasks_cancelled_total":
		p.tasksCancelled.With(prometheus.Labels(labels)).Inc()
	case "worker_heartbeats_total":
		p.workerHeartbeats.With(prometheus.Labels(labels)).Inc()
	}
}

func (p *PrometheusCollector) IncrementGauge(name string, value float64) {
	switch name {
	case "workers_active":
		p.workersActive.Add(value)
	}
}

func (p *PrometheusCollector) DecrementGauge(name string, value float64) {
	switch name {
	case "workers_active":
		p.workersActive.Sub(value)
	}
}

func (p *PrometheusCollector) SetGauge(name string, value float64) {
	switch name {
	case "workers_active":
		p.workersActive.Set(value)
	}
}

func (p *PrometheusCollector) ObserveHistogram(name string, value float64, labels map[string]string) {
	switch name {
	case "task_duration_seconds":
		p.taskDuration.With(prometheus.Labels(labels)).Observe(value)
	}
}

func (p *PrometheusCollector) UpdateQueueDepth(priority string, depth float64) {
	p.queueDepth.With(prometheus.Labels{"priority": priority}).Set(depth)
}

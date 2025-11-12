package service

type MetricsCollector interface {
	IncrementCounter(name string, labels map[string]string)

	IncrementGauge(name string, value float64)

	DecrementGauge(name string, value float64)

	SetGauge(name string, value float64)

	ObserveHistogram(name string, value float64, labels map[string]string)
}

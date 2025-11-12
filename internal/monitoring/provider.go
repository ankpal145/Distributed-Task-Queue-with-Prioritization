package monitoring

import (
	"go.uber.org/fx"
)

var FxMonitoringProvider = fx.Options(
	fx.Provide(NewPrometheusCollector),
	fx.Provide(AsMetricsCollector),
)

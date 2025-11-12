package service

import (
	"go.uber.org/fx"
)

var FxServiceProvider = fx.Options(
	fx.Provide(NewTaskConfigFromConfig),
	fx.Provide(NewTaskService),
)

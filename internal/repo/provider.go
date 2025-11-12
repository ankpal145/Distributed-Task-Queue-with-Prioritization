package repo

import (
	"go.uber.org/fx"
)

var FxRepoProvider = fx.Options(
	fx.Provide(NewRedisTaskRepository),
	fx.Provide(AsTaskRepository),
	fx.Provide(NewRedisWorkerRepository),
	fx.Provide(AsWorkerRepository),
)

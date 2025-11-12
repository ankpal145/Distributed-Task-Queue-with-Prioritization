package pkg

import (
	"github.com/ankurpal/distributed-task-queue/pkg/logger"
	"github.com/ankurpal/distributed-task-queue/pkg/middleware"
	"github.com/ankurpal/distributed-task-queue/pkg/redis"
	"go.uber.org/fx"
)

var FxPkgProvider = fx.Options(
	fx.Provide(logger.SetupLogger),
	fx.Provide(middleware.NewAuthMiddleware),
	fx.Provide(middleware.NewRequestContextMiddleware),
	fx.Provide(redis.NewRedisClient),
	fx.Provide(redis.ProvideRawClient),
)

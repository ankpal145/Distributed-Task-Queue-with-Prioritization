package grpcserver

import (
	"go.uber.org/fx"
)

var FxGrpcServerProvider = fx.Options(
	fx.Provide(NewServer),
)

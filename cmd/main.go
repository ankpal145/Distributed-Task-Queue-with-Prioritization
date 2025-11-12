package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ankurpal/distributed-task-queue/cmd/server"
	"github.com/ankurpal/distributed-task-queue/cmd/worker"
	"github.com/ankurpal/distributed-task-queue/config"
	"github.com/ankurpal/distributed-task-queue/internal/constants"
	"github.com/ankurpal/distributed-task-queue/internal/dao"
	grpcserver "github.com/ankurpal/distributed-task-queue/internal/handler/grpc-server"
	"github.com/ankurpal/distributed-task-queue/internal/monitoring"
	"github.com/ankurpal/distributed-task-queue/internal/repo"
	"github.com/ankurpal/distributed-task-queue/internal/service"
	"github.com/ankurpal/distributed-task-queue/internal/worker/processors"
	"github.com/ankurpal/distributed-task-queue/pkg"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

// @title           Go Template Swagger
// @version         1.0
// @description     Use this to write some description about the API.

// @contact.name   Developers
// @contact.email  ankurpal1073@gmail.com

// @BasePath  /api/v1.
func main() {
	commandArgs := os.Args
	config.Load(commandArgs)

	app := &cli.App{
		Name:  "taskqueue",
		Usage: "Distributed Task Queue System",
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "Run API server (REST + gRPC) with metrics",
				Action: func(c *cli.Context) error {
					return startFxApp(constants.SERVER)
				},
			},
			{
				Name:  "worker",
				Usage: "Run task processor worker",
				Action: func(c *cli.Context) error {
					return startFxApp(constants.WORKER)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func NewContext() (context.Context, error) {
	return context.Background(), nil
}

func startFxApp(module string) error {
	provideConfig := fx.Provide(func() *config.Configuration {
		return config.NewConfig(module)
	})

	commonModules := fx.Options(
		provideConfig,
		fx.Provide(NewContext),
		pkg.FxPkgProvider,
		monitoring.FxMonitoringProvider,
		repo.FxRepoProvider,
		dao.FxDaoProvider,
		service.FxServiceProvider,
	)

	grpcModules := fx.Options(
		processors.FxWorkerProcessorsProvider,
		grpcserver.FxGrpcServerProvider,
		fx.Invoke(server.NewGRPCServer),
	)

	workerModules := fx.Options(
		processors.FxWorkerProcessorsProvider,
		fx.Invoke(worker.NewWorker),
	)

	var app *fx.App

	switch module {
	case constants.SERVER:
		app = fx.New(commonModules, grpcModules)
	case constants.WORKER:
		app = fx.New(commonModules, workerModules)
	default:
		return fmt.Errorf("unknown module: %s. Use 'server' or 'worker'", module)
	}

	err := app.Start(context.Background())
	if err != nil {
		return err
	}

	<-app.Done()

	if err = app.Stop(context.Background()); err != nil {
		return err
	}

	return nil
}

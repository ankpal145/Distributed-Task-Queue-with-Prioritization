package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/ankurpal/distributed-task-queue/config"
	grpcserver "github.com/ankurpal/distributed-task-queue/internal/handler/grpc-server"
	"github.com/ankurpal/distributed-task-queue/internal/service"
	"github.com/ankurpal/distributed-task-queue/internal/worker"
	"github.com/ankurpal/distributed-task-queue/pkg/logger"
	"github.com/ankurpal/distributed-task-queue/pkg/middleware"
	pb "github.com/ankurpal/distributed-task-queue/proto"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	cmux       cmux.CMux
	listener   net.Listener
}

func NewGRPCServer(
	lc fx.Lifecycle,
	ctx context.Context,
	cfg *config.Configuration,
	log *logger.Logger,
	authMw *middleware.AuthMiddleware,
	reqCtxMw *middleware.RequestContextMiddleware,
	taskService *service.TaskService,
	workerService *worker.WorkerService,
) *Server {
	server := &Server{}

	lc.Append(fx.Hook{
		OnStart: func(startCtx context.Context) error {
			return startUnifiedServer(ctx, cfg, log, authMw, reqCtxMw, taskService, workerService, server)
		},
		OnStop: func(stopCtx context.Context) error {
			return stopUnifiedServer(ctx, log, server)
		},
	})

	return server
}

func NewRESTServer(lc fx.Lifecycle, cfg *config.Configuration, log *logger.Logger) {
	log.InfoWithCtx(context.Background(), "REST server will be started via unified cmux server")
}

func NewMetricsServer(lc fx.Lifecycle, cfg *config.Configuration, log *logger.Logger) {
	if !cfg.Metrics.Enabled {
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go startMetricsServer(cfg, log)
			log.InfoWithCtx(ctx, "📈 Metrics server starting on port %d", cfg.Metrics.Port)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.InfoWithCtx(ctx, "Stopping metrics server...")
			return nil
		},
	})
}

func startUnifiedServer(
	ctx context.Context,
	cfg *config.Configuration,
	log *logger.Logger,
	authMw *middleware.AuthMiddleware,
	reqCtxMw *middleware.RequestContextMiddleware,
	taskService *service.TaskService,
	workerService *worker.WorkerService,
	server *Server,
) error {
	portInfo := fmt.Sprintf(":%d", cfg.Server.GRPCPort)

	logrusEntry := logrus.NewEntry(log.GetLogger())
	grpc_logrus.ReplaceGrpcLogger(logrusEntry)

	logOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
		grpc_logrus.WithDurationField(grpc_logrus.DefaultDurationToField),
		grpc_logrus.WithDecider(func(fullMethodName string, err error) bool {
			return true
		}),
	}

	recoveryOpt := grpc_recovery.WithRecoveryHandlerContext(
		func(ctx context.Context, p interface{}) error {
			log.ErrorWithCtx(ctx, "[GRPC PANIC RECOVERY] panic: %v, stack: %s", p, string(debug.Stack()))
			return status.Errorf(codes.Internal, "%s", p)
		},
	)

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.StreamServerInterceptor(logrusEntry, logOpts...),
			grpc_prometheus.StreamServerInterceptor,
			grpc_recovery.StreamServerInterceptor(recoveryOpt),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			reqCtxMw.RequestContextUnaryInterceptor,
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.UnaryServerInterceptor(logrusEntry, logOpts...),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_recovery.UnaryServerInterceptor(recoveryOpt),
		)),
	)

	grpcServerImpl := grpcserver.NewServer(taskService, workerService)
	pb.RegisterTaskQueueServiceServer(grpcServer, grpcServerImpl)

	gatewayMux := runtime.NewServeMux()

	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", promhttp.Handler())
	httpMux.Handle("/api/", gatewayMux)
	httpMux.Handle("/", gatewayMux)

	httpServer := &http.Server{
		Handler: corsHandler(httpMux),
	}

	handleHealthCheck(ctx, gatewayMux, log)

	listener, err := net.Listen("tcp", portInfo)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", portInfo, err)
	}

	m := cmux.New(listener)
	httpListener := m.Match(cmux.HTTP1Fast("PATCH"))
	grpcListener := m.Match(cmux.HTTP2())

	server.grpcServer = grpcServer
	server.httpServer = httpServer
	server.cmux = m
	server.listener = listener

	go func() {
		log.InfoWithCtx(ctx, "🌐 HTTP server starting on %s", portInfo)
		if err := httpServer.Serve(httpListener); err != nil && err != http.ErrServerClosed {
			log.ErrorWithCtx(ctx, "HTTP server error: %v", err)
		}
	}()

	go func() {
		log.InfoWithCtx(ctx, "⚡ gRPC server starting on %s", portInfo)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.ErrorWithCtx(ctx, "gRPC server error: %v", err)
		}
	}()

	go func() {
		log.InfoWithCtx(ctx, "🚀 Unified server (cmux) starting on %s", portInfo)
		if err := m.Serve(); err != nil {
			log.ErrorWithCtx(ctx, "cmux server error: %v", err)
		}
	}()

	go waitForShutdown(ctx, server, log)

	log.InfoWithCtx(ctx, "✅ Unified server started successfully on %s", portInfo)
	return nil
}

func stopUnifiedServer(ctx context.Context, log *logger.Logger, server *Server) error {
	log.InfoWithCtx(ctx, "Initiating graceful shutdown...")

	if server.cmux != nil {
		server.cmux.Close()
	}

	if server.grpcServer != nil {
		server.grpcServer.GracefulStop()
	}

	if server.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := server.httpServer.Shutdown(shutdownCtx); err != nil {
			log.ErrorWithCtx(ctx, "HTTP server shutdown error: %v", err)
			return err
		}
	}

	log.InfoWithCtx(ctx, "Server shutdown completed")
	return nil
}

func waitForShutdown(ctx context.Context, server *Server, log *logger.Logger) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
	log.InfoWithCtx(ctx, "Shutdown signal received")

	stopUnifiedServer(ctx, log, server)
}

func corsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType, x-request-id, x-correlation-id")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	})
}

type PingResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func handleHealthCheck(ctx context.Context, mux *runtime.ServeMux, log *logger.Logger) {
	err := mux.HandlePath(http.MethodGet, "/health", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		payload, _ := json.Marshal(PingResponse{
			Status:  "healthy",
			Service: "taskqueue",
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(payload); err != nil {
			log.ErrorWithCtx(ctx, "Failed to write health check response: %v", err)
		}
	})
	if err != nil {
		log.ErrorWithCtx(ctx, "Failed to register health check handler: %v", err)
	}
}

type healthCheckClient struct {
	status grpc_health_v1.HealthCheckResponse_ServingStatus
	code   codes.Code
}

func (h healthCheckClient) Check(ctx context.Context, r *grpc_health_v1.HealthCheckRequest, opts ...grpc.CallOption) (*grpc_health_v1.HealthCheckResponse, error) {
	if h.code != codes.OK {
		return nil, status.Error(h.code, r.GetService())
	}
	return &grpc_health_v1.HealthCheckResponse{Status: h.status}, nil
}

func (h healthCheckClient) Watch(ctx context.Context, in *grpc_health_v1.HealthCheckRequest, opts ...grpc.CallOption) (grpc_health_v1.Health_WatchClient, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func startMetricsServer(cfg *config.Configuration, log *logger.Logger) {
	http.Handle(cfg.Metrics.Path, promhttp.Handler())

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Metrics.Port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("Metrics server error", err)
	}
}

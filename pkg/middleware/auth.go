package middleware

import (
	"context"

	"github.com/ankurpal/distributed-task-queue/config"
	"github.com/ankurpal/distributed-task-queue/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	XApiKey = "x-api-key"
)

type AuthMiddleware struct {
	logger *logger.Logger
	config *config.Configuration
}

func NewAuthMiddleware(logger *logger.Logger, config *config.Configuration) *AuthMiddleware {
	return &AuthMiddleware{
		logger: logger,
		config: config,
	}
}

func (a *AuthMiddleware) AuthUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod == "/grpc.health.v1.Health/Check" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		a.logger.WarnWithCtx(ctx, "missing metadata in request")
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	apiKeys := md.Get(XApiKey)
	if len(apiKeys) == 0 {
		a.logger.WarnWithCtx(ctx, "x-api-key is not provided")
		return nil, status.Errorf(codes.Unauthenticated, "x-api-key is not provided")
	}

	return handler(ctx, req)
}

func (a *AuthMiddleware) AuthStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		a.logger.WarnWithCtx(ctx, "missing metadata in stream request")
		return status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	apiKeys := md.Get(XApiKey)
	if len(apiKeys) == 0 {
		a.logger.WarnWithCtx(ctx, "x-api-key is not provided in stream")
		return status.Errorf(codes.Unauthenticated, "x-api-key is not provided")
	}

	return handler(srv, ss)
}

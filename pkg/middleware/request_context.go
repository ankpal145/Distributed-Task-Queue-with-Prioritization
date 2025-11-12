package middleware

import (
	"context"

	"github.com/ankurpal/distributed-task-queue/pkg/contexts"
	"github.com/ankurpal/distributed-task-queue/pkg/logger"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	XRequestIDHeader     = "x-request-id"
	XCorrelationIDHeader = "x-correlation-id"
	XUserIDHeader        = "x-user-id"
	XTaskIDHeader        = "x-task-id"
	XWorkerIDHeader      = "x-worker-id"
)

type RequestContextMiddleware struct {
	logger *logger.Logger
}

func NewRequestContextMiddleware(logger *logger.Logger) *RequestContextMiddleware {
	return &RequestContextMiddleware{
		logger: logger,
	}
}

func (r *RequestContextMiddleware) RequestContextUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	reqContext := contexts.NewRequestContext()

	if requestIDs := md.Get(XRequestIDHeader); len(requestIDs) > 0 {
		reqContext.RequestID = requestIDs[0]
	} else {
		reqContext.RequestID = uuid.New().String()
	}

	if correlationIDs := md.Get(XCorrelationIDHeader); len(correlationIDs) > 0 {
		reqContext.CorrelationID = correlationIDs[0]
	} else {
		reqContext.CorrelationID = uuid.New().String()
	}

	if userIDs := md.Get(XUserIDHeader); len(userIDs) > 0 {
		reqContext.UserID = userIDs[0]
	}

	if taskIDs := md.Get(XTaskIDHeader); len(taskIDs) > 0 {
		reqContext.TaskID = taskIDs[0]
	}

	if workerIDs := md.Get(XWorkerIDHeader); len(workerIDs) > 0 {
		reqContext.WorkerID = workerIDs[0]
	}

	reqContext.Method = info.FullMethod

	ctx = contexts.SetRequestContext(ctx, reqContext)

	r.logger.DebugWithCtx(ctx, "Request context set for %s", info.FullMethod)

	return handler(ctx, req)
}

func (r *RequestContextMiddleware) RequestContextStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	reqContext := contexts.NewRequestContext()

	if requestIDs := md.Get(XRequestIDHeader); len(requestIDs) > 0 {
		reqContext.RequestID = requestIDs[0]
	} else {
		reqContext.RequestID = uuid.New().String()
	}

	if correlationIDs := md.Get(XCorrelationIDHeader); len(correlationIDs) > 0 {
		reqContext.CorrelationID = correlationIDs[0]
	} else {
		reqContext.CorrelationID = uuid.New().String()
	}

	if userIDs := md.Get(XUserIDHeader); len(userIDs) > 0 {
		reqContext.UserID = userIDs[0]
	}

	if taskIDs := md.Get(XTaskIDHeader); len(taskIDs) > 0 {
		reqContext.TaskID = taskIDs[0]
	}

	if workerIDs := md.Get(XWorkerIDHeader); len(workerIDs) > 0 {
		reqContext.WorkerID = workerIDs[0]
	}

	reqContext.Method = info.FullMethod

	ctx = contexts.SetRequestContext(ctx, reqContext)

	r.logger.DebugWithCtx(ctx, "Request context set for stream %s", info.FullMethod)

	wrappedStream := &wrappedServerStream{
		ServerStream: ss,
		ctx:          ctx,
	}

	return handler(srv, wrappedStream)
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

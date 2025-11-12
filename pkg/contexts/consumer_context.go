package contexts

import (
	"context"

	"github.com/google/uuid"
)

func NewConsumerCtx(workerID string, queues []string) context.Context {
	ctx := context.Background()

	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, RequestIDKey, requestID)

	reqCtx := NewRequestContext().
		WithRequestID(requestID).
		WithWorkerID(workerID)

	return SetRequestContext(ctx, reqCtx)
}

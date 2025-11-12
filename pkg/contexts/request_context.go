package contexts

import (
	"context"
)

type contextKey string

const (
	requestContextKey contextKey = "request_context"
	RequestIDKey      contextKey = "request_id"
	CorrelationIDKey  contextKey = "correlation_id"
)

type RequestContext struct {
	RequestID     string
	CorrelationID string
	UserID        string
	URL           string
	Method        string
	TaskID        string
	WorkerID      string
}

func SetRequestContext(ctx context.Context, reqCtx *RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey, reqCtx)
}

func GetRequestContext(ctx context.Context) *RequestContext {
	if reqCtx, ok := ctx.Value(requestContextKey).(*RequestContext); ok {
		return reqCtx
	}
	return nil
}

func NewRequestContext() *RequestContext {
	return &RequestContext{}
}

func (rc *RequestContext) WithRequestID(requestID string) *RequestContext {
	rc.RequestID = requestID
	return rc
}

func (rc *RequestContext) WithCorrelationID(correlationID string) *RequestContext {
	rc.CorrelationID = correlationID
	return rc
}

func (rc *RequestContext) WithUserID(userID string) *RequestContext {
	rc.UserID = userID
	return rc
}

func (rc *RequestContext) WithURL(url string) *RequestContext {
	rc.URL = url
	return rc
}

func (rc *RequestContext) WithMethod(method string) *RequestContext {
	rc.Method = method
	return rc
}

func (rc *RequestContext) WithTaskID(taskID string) *RequestContext {
	rc.TaskID = taskID
	return rc
}

func (rc *RequestContext) WithWorkerID(workerID string) *RequestContext {
	rc.WorkerID = workerID
	return rc
}

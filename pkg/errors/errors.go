package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServiceError struct {
	Code    codes.Code
	Message string
	Details string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

func (e *ServiceError) ToGRPCError() error {
	return status.Error(e.Code, e.Error())
}

func NewServiceError(code codes.Code, message string, err error) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func NewInternalError(message string, err error) *ServiceError {
	return NewServiceError(codes.Internal, message, err)
}

func NewNotFoundError(message string, err error) *ServiceError {
	return NewServiceError(codes.NotFound, message, err)
}

func NewInvalidArgumentError(message string, err error) *ServiceError {
	return NewServiceError(codes.InvalidArgument, message, err)
}

func NewAlreadyExistsError(message string, err error) *ServiceError {
	return NewServiceError(codes.AlreadyExists, message, err)
}

func NewUnauthenticatedError(message string, err error) *ServiceError {
	return NewServiceError(codes.Unauthenticated, message, err)
}

func NewPermissionDeniedError(message string, err error) *ServiceError {
	return NewServiceError(codes.PermissionDenied, message, err)
}

func NewUnavailableError(message string, err error) *ServiceError {
	return NewServiceError(codes.Unavailable, message, err)
}

func NewDeadlineExceededError(message string, err error) *ServiceError {
	return NewServiceError(codes.DeadlineExceeded, message, err)
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func NewErrorResponse(code, message, details string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func FromServiceError(err *ServiceError) *ErrorResponse {
	return &ErrorResponse{
		Code:    err.Code.String(),
		Message: err.Message,
		Details: err.Details,
	}
}

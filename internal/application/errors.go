package application

import "fmt"

type ErrorCode string

const (
	CodeInvalidArgument ErrorCode = "INVALID_ARGUMENT"
	CodeNotFound        ErrorCode = "NOT_FOUND"
	CodeUnauthenticated ErrorCode = "UNAUTHENTICATED"
	CodeRateLimited     ErrorCode = "RATE_LIMITED"
	CodeInternal        ErrorCode = "INTERNAL"
)

type AppError struct {
	Code    ErrorCode      `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	Err     error          `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
}

func (e *AppError) Unwrap() error { return e.Err }

func NewInvalidArgument(message string, details map[string]any) *AppError {
	return &AppError{Code: CodeInvalidArgument, Message: message, Details: details}
}

func NewNotFound(message string, details map[string]any) *AppError {
	return &AppError{Code: CodeNotFound, Message: message, Details: details}
}

func NewUnauthenticated(message string) *AppError {
	return &AppError{Code: CodeUnauthenticated, Message: message}
}

func NewRateLimited(message string, details map[string]any) *AppError {
	return &AppError{Code: CodeRateLimited, Message: message, Details: details}
}

func NewInternal(message string, err error) *AppError {
	return &AppError{Code: CodeInternal, Message: message, Err: err}
}

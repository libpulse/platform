package errors

import (
	"net/http"
)

// ErrorCode represents standardized error codes for LibPulse API
type ErrorCode string

// Error codes - only what you need
const (
	ErrUnauthorized  ErrorCode = "unauthorized"
	ErrInvalidToken  ErrorCode = "invalid_token"
	ErrInternalError ErrorCode = "internal_error"
)

// APIError represents a standardized API error response
type APIError struct {
	Error  string    `json:"error"`
	Code   ErrorCode `json:"code"`
	Status int       `json:"-"`
}

// ErrorMapping maps error codes to HTTP status codes and messages
var ErrorMapping = map[ErrorCode]*APIError{
	ErrUnauthorized: {
		Error:  "Missing or invalid Authorization header",
		Code:   ErrUnauthorized,
		Status: http.StatusUnauthorized,
	},
	ErrInvalidToken: {
		Error:  "invalid token",
		Code:   ErrInvalidToken,
		Status: http.StatusUnauthorized,
	},
	ErrInternalError: {
		Error:  "failed to fetch user",
		Code:   ErrInternalError,
		Status: http.StatusInternalServerError,
	},
}

// NewAPIError creates a new API error with the given code
func NewAPIError(code ErrorCode) *APIError {
	baseError, exists := ErrorMapping[code]
	if !exists {
		baseError = ErrorMapping[ErrInternalError]
	}

	return &APIError{
		Error:  baseError.Error,
		Code:   baseError.Code,
		Status: baseError.Status,
	}
}

// StatusCode returns the HTTP status code
func (e *APIError) StatusCode() int {
	return e.Status
}

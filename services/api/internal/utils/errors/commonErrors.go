package errors

// Common error codes used across different handlers
const (
	ErrBadRequest    ErrorCode = "bad_request"
	ErrForbidden     ErrorCode = "forbidden"
	ErrNotFound      ErrorCode = "not_found"
	ErrConflict      ErrorCode = "conflict"
	ErrTooManyRequests ErrorCode = "too_many_requests"
	ErrInternalError ErrorCode = "internal_error"
)

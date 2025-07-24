package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode represents different types of application errors
type ErrorCode string

const (
	CodeValidation   ErrorCode = "VALIDATION_ERROR"
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeConflict     ErrorCode = "CONFLICT"
	CodeInternal     ErrorCode = "INTERNAL_ERROR"
	CodeTimeout      ErrorCode = "TIMEOUT"
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeRateLimit    ErrorCode = "RATE_LIMIT"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	StatusCode int       `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
}

// New creates a new AppError
func New(code ErrorCode, message string, details string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: getStatusCodeForErrorCode(code),
	}
}

// Validation creates a validation error
func Validation(message, details string) *AppError {
	return New(CodeValidation, message, details)
}

// NotFound creates a not found error
func NotFound(resource string) *AppError {
	return New(CodeNotFound, fmt.Sprintf("%s not found", resource), "")
}

// Conflict creates a conflict error
func Conflict(message, details string) *AppError {
	return New(CodeConflict, message, details)
}

// Internal creates an internal server error
func Internal(message, details string) *AppError {
	return New(CodeInternal, message, details)
}

// WriteErrorResponse writes an error response to the HTTP response writer
func WriteErrorResponse(w http.ResponseWriter, err error, requestID string) {
	var appErr *AppError

	// Convert to AppError if it's not already one
	if e, ok := err.(*AppError); ok {
		appErr = e
	} else {
		appErr = Internal("An unexpected error occurred", err.Error())
	}

	// Set request ID if provided
	if requestID != "" {
		appErr.RequestID = requestID
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)

	response := map[string]interface{}{
		"error": appErr,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback to plain text if JSON encoding fails
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}
}

func getStatusCodeForErrorCode(code ErrorCode) int {
	switch code {
	case CodeValidation:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeRateLimit:
		return http.StatusTooManyRequests
	case CodeTimeout:
		return http.StatusRequestTimeout
	case CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

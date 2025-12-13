package models

import "net/http"

// AppError represents an application error
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// Predefined errors
var (
	ErrInvalidCredentials    = &AppError{Code: http.StatusUnauthorized, Message: "Invalid credentials"}
	ErrUserNotFound          = &AppError{Code: http.StatusNotFound, Message: "User not found"}
	ErrUserAlreadyExists     = &AppError{Code: http.StatusConflict, Message: "User already exists"}
	ErrEmailAlreadyExists    = &AppError{Code: http.StatusConflict, Message: "Email already exists"}
	ErrPhoneAlreadyExists    = &AppError{Code: http.StatusConflict, Message: "Phone already exists"}
	ErrUsernameAlreadyExists = &AppError{Code: http.StatusConflict, Message: "Username already exists"}
	ErrInvalidToken          = &AppError{Code: http.StatusUnauthorized, Message: "Invalid token"}
	ErrTokenExpired          = &AppError{Code: http.StatusUnauthorized, Message: "Token expired"}
	ErrTokenRevoked          = &AppError{Code: http.StatusUnauthorized, Message: "Token revoked"}
	ErrUnauthorized          = &AppError{Code: http.StatusUnauthorized, Message: "Unauthorized"}
	ErrForbidden             = &AppError{Code: http.StatusForbidden, Message: "Forbidden"}
	ErrBadRequest            = &AppError{Code: http.StatusBadRequest, Message: "Bad request"}
	ErrInternalServer        = &AppError{Code: http.StatusInternalServerError, Message: "Internal server error"}
	ErrRateLimitExceeded     = &AppError{Code: http.StatusTooManyRequests, Message: "Rate limit exceeded"}
	ErrInvalidProvider       = &AppError{Code: http.StatusBadRequest, Message: "Invalid OAuth provider"}

	// API Key errors
	ErrAPIKeyNotFound = &AppError{Code: http.StatusNotFound, Message: "API key not found"}
	ErrInvalidAPIKey  = &AppError{Code: http.StatusUnauthorized, Message: "Invalid API key"}
	ErrAPIKeyExpired  = &AppError{Code: http.StatusUnauthorized, Message: "API key expired"}
	ErrAPIKeyRevoked  = &AppError{Code: http.StatusUnauthorized, Message: "API key revoked"}

	// Generic errors
	ErrNotFound            = &AppError{Code: http.StatusNotFound, Message: "Resource not found"}
	ErrAlreadyExists       = &AppError{Code: http.StatusConflict, Message: "Resource already exists"}
	ErrForeignKeyViolation = &AppError{Code: http.StatusBadRequest, Message: "Foreign key constraint violation"}
	ErrRequiredField       = &AppError{Code: http.StatusBadRequest, Message: "Required field is missing or null"}
)

// NewAppError creates a new application error
func NewAppError(code int, message string, details ...string) *AppError {
	err := &AppError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err error) *ErrorResponse {
	if appErr, ok := err.(*AppError); ok {
		return &ErrorResponse{
			Error:   http.StatusText(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
		}
	}
	return &ErrorResponse{
		Error:   "Internal Server Error",
		Message: err.Error(),
	}
}

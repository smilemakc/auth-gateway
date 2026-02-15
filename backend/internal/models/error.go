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
	ErrTokenCompromised      = &AppError{Code: http.StatusUnauthorized, Message: "Token may be compromised: device mismatch"}
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
	// HTTP error status text
	Error string `json:"error" example:"Bad Request"`
	// Human-readable error message
	Message string `json:"message" example:"Invalid request parameters"`
	// Additional error details
	Details string `json:"details,omitempty" example:"Email field is required"`
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

// MessageResponse represents a simple message response
type MessageResponse struct {
	// Response message
	Message string `json:"message" example:"Operation completed successfully"`
}

// PasswordlessLoginRequest represents a passwordless login request
type PasswordlessLoginRequest struct {
	// Email address to receive the OTP code
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// PasswordlessLoginVerifyRequest represents passwordless login verification
type PasswordlessLoginVerifyRequest struct {
	// Email address that received the OTP
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	// 6-digit OTP code
	Code string `json:"code" binding:"required,len=6" example:"123456"`
}

// RegenerateBackupCodesRequest represents a request to regenerate 2FA backup codes
type RegenerateBackupCodesRequest struct {
	// User's current password for verification
	Password string `json:"password" binding:"required" example:"SecurePass123!"`
}

// BackupCodesResponse represents the response containing backup codes
type BackupCodesResponse struct {
	// List of new backup codes
	BackupCodes []string `json:"backup_codes" example:"ABC123,DEF456,GHI789,JKL012,MNO345"`
	// Response message
	Message string `json:"message" example:"Backup codes regenerated successfully. Save them in a secure location."`
}

// TelegramAuthData represents Telegram widget authentication data
type TelegramAuthData struct {
	// Telegram user ID
	ID int64 `json:"id" example:"123456789"`
	// User's first name
	FirstName string `json:"first_name" example:"John"`
	// User's last name (optional)
	LastName string `json:"last_name,omitempty" example:"Doe"`
	// Username without @ (optional)
	Username string `json:"username,omitempty" example:"johndoe"`
	// Profile photo URL (optional)
	PhotoURL string `json:"photo_url,omitempty" example:"https://t.me/i/userpic/123.jpg"`
	// Authentication timestamp
	AuthDate int64 `json:"auth_date" example:"1234567890"`
	// Hash for verification
	Hash string `json:"hash" example:"abc123def456..."`
}

// ResendVerificationRequest represents request to resend verification email
type ResendVerificationRequest struct {
	// Email address to resend verification to
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	// Email address to verify
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	// 6-digit verification code
	Code string `json:"code" binding:"required,len=6" example:"123456"`
}

// VerifyEmailResponse represents email verification response
type VerifyEmailResponse struct {
	// Whether verification was successful
	Valid bool `json:"valid" example:"true"`
	// Human-readable message
	Message string `json:"message" example:"Email verified successfully"`
}

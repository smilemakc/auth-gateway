package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Context keys
const (
	UserIDKey   = "user_id"
	UserEmailKey = "user_email"
	UserRoleKey  = "user_role"
)

// GetUserIDFromContext retrieves the user ID from the Gin context
func GetUserIDFromContext(c *gin.Context) (*uuid.UUID, bool) {
	value, exists := c.Get(UserIDKey)
	if !exists {
		return nil, false
	}

	userID, ok := value.(uuid.UUID)
	if !ok {
		return nil, false
	}

	return &userID, true
}

// GetUserEmailFromContext retrieves the user email from the Gin context
func GetUserEmailFromContext(c *gin.Context) (string, bool) {
	value, exists := c.Get(UserEmailKey)
	if !exists {
		return "", false
	}

	email, ok := value.(string)
	if !ok {
		return "", false
	}

	return email, true
}

// GetUserRoleFromContext retrieves the user role from the Gin context
func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	value, exists := c.Get(UserRoleKey)
	if !exists {
		return "", false
	}

	role, ok := value.(string)
	if !ok {
		return "", false
	}

	return role, true
}

// GetClientIP gets the real client IP address
func GetClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return c.ClientIP()
}

// GetUserAgent gets the user agent string
func GetUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}

package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// Context keys
const (
	UserIDKey    = "user_id"
	UserEmailKey = "user_email"
	UserRoleKey  = "user_role"
	UserRolesKey = "user_roles"
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

// GetUserRolesFromContext retrieves user roles from Gin context
func GetUserRolesFromContext(c *gin.Context) ([]string, bool) {
	roles, exists := c.Get(UserRolesKey)
	if !exists {
		return nil, false
	}
	roleSlice, ok := roles.([]string)
	return roleSlice, ok
}

// HasRole checks if user has a specific role
func HasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the required roles
func HasAnyRole(userRoles, requiredRoles []string) bool {
	for _, required := range requiredRoles {
		for _, userRole := range userRoles {
			if userRole == required {
				return true
			}
		}
	}
	return false
}

// GetDeviceInfoFromContext retrieves User-Agent header from Gin context and parses it
func GetDeviceInfoFromContext(c *gin.Context) models.DeviceInfo {
	userAgent := GetUserAgent(c)
	return ParseUserAgent(userAgent)
}

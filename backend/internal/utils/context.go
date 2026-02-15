package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// Context keys
const (
	UserIDKey        = "user_id"
	UserEmailKey     = "user_email"
	UserRoleKey      = "user_role"
	UserRolesKey     = "user_roles"
	TokenKey         = "access_token"
	ApplicationIDKey = "application_id"
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

// MustGetUserID extracts the user ID from context and responds with 401 if not found.
// Returns the user ID and true on success. On failure, it responds and aborts — caller should return.
func MustGetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		c.Abort()
		return uuid.Nil, false
	}
	return *userID, true
}

// GetApplicationIDFromContext retrieves the application ID from the Gin context
func GetApplicationIDFromContext(c *gin.Context) (*uuid.UUID, bool) {
	value, exists := c.Get(ApplicationIDKey)
	if !exists {
		return nil, false
	}

	// Can be a string or uuid.UUID
	switch v := value.(type) {
	case uuid.UUID:
		return &v, true
	case *uuid.UUID:
		if v == nil {
			return nil, false
		}
		return v, true
	case string:
		if v == "" {
			return nil, false
		}
		parsed, err := uuid.Parse(v)
		if err != nil {
			return nil, false
		}
		return &parsed, true
	default:
		return nil, false
	}
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

// GetTokenFromContext retrieves raw access token set by auth middleware
func GetTokenFromContext(c *gin.Context) (string, bool) {
	value, exists := c.Get(TokenKey)
	if !exists {
		return "", false
	}

	token, ok := value.(string)
	return token, ok && token != ""
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

// SetApplicationIDInContext sets the application ID in the Gin context
func SetApplicationIDInContext(c *gin.Context, applicationID uuid.UUID) {
	c.Set(ApplicationIDKey, applicationID)
}

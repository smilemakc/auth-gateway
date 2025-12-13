package middleware

import (
	"net/http"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"

	"github.com/gin-gonic/gin"
)

// IPFilterMiddleware provides IP-based access control
type IPFilterMiddleware struct {
	ipFilterService *service.IPFilterService
}

// NewIPFilterMiddleware creates a new IP filter middleware
func NewIPFilterMiddleware(ipFilterService *service.IPFilterService) *IPFilterMiddleware {
	return &IPFilterMiddleware{
		ipFilterService: ipFilterService,
	}
}

// CheckIPFilter validates if the client IP is allowed
func (m *IPFilterMiddleware) CheckIPFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := utils.GetClientIP(c)

		// Check if IP is allowed
		result, err := m.ipFilterService.CheckIPAllowed(c.Request.Context(), clientIP)
		if err != nil {
			// On error, log and allow (fail-open for availability)
			c.Next()
			return
		}

		if !result.Allowed {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Access denied: " + result.Reason,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

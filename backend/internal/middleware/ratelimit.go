package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/config"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

// RateLimitMiddleware provides rate limiting functionality
type RateLimitMiddleware struct {
	redis  *service.RedisService
	config *config.RateLimitConfig
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(redis *service.RedisService, cfg *config.RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redis:  redis,
		config: cfg,
	}
}

// LimitByIP limits requests by IP address
func (m *RateLimitMiddleware) LimitByIP(endpoint string, max int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := utils.GetClientIP(c)
		key := fmt.Sprintf("ratelimit:%s:%s", ip, endpoint)

		count, err := m.redis.IncrementRateLimit(c.Request.Context(), key, window)
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Rate limit error: %v\n", err)
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", max))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max-int(count)))

		if count > int64(max) {
			c.JSON(http.StatusTooManyRequests, models.NewErrorResponse(models.ErrRateLimitExceeded))
			c.Abort()
			return
		}

		c.Next()
	}
}

// LimitSignup limits signup requests
func (m *RateLimitMiddleware) LimitSignup() gin.HandlerFunc {
	return m.LimitByIP("signup", m.config.SignupMax, m.config.SignupWindow)
}

// LimitSignin limits signin requests
func (m *RateLimitMiddleware) LimitSignin() gin.HandlerFunc {
	return m.LimitByIP("signin", m.config.SigninMax, m.config.SigninWindow)
}

// LimitAPI limits general API requests
func (m *RateLimitMiddleware) LimitAPI() gin.HandlerFunc {
	return m.LimitByIP("api", m.config.APIMax, m.config.APIWindow)
}

// LimitByUserID limits requests by user ID (for authenticated endpoints)
func (m *RateLimitMiddleware) LimitByUserID(endpoint string, max int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := utils.GetUserIDFromContext(c)
		if !exists {
			// Fall back to IP-based limiting
			m.LimitByIP(endpoint, max, window)(c)
			return
		}

		key := fmt.Sprintf("ratelimit:%s:%s", userID.String(), endpoint)

		count, err := m.redis.IncrementRateLimit(c.Request.Context(), key, window)
		if err != nil {
			fmt.Printf("Rate limit error: %v\n", err)
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", max))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max-int(count)))

		if count > int64(max) {
			c.JSON(http.StatusTooManyRequests, models.NewErrorResponse(models.ErrRateLimitExceeded))
			c.Abort()
			return
		}

		c.Next()
	}
}

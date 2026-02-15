package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
)

const csrfTokenLength = 32
const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"

// CSRFProtection implements the Double Submit Cookie CSRF protection pattern.
// On safe methods (GET/HEAD/OPTIONS), a csrf_token cookie is set.
// On state-changing methods (POST/PUT/DELETE/PATCH), the X-CSRF-Token header
// must match the csrf_token cookie value.
func CSRFProtection(enabled bool, secureCookie bool) gin.HandlerFunc {
	if !enabled {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			// Safe methods: only set CSRF token cookie if one does not already exist
			if _, err := c.Cookie(csrfCookieName); err != nil {
				token, err := generateCSRFToken()
				if err != nil {
					c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
						models.NewAppError(http.StatusInternalServerError, "CSRF_GENERATION_FAILED", "Failed to generate CSRF token"),
					))
					c.Abort()
					return
				}
				c.SetCookie(csrfCookieName, token, 3600, "/", "", secureCookie, false)
			}
			c.Next()
		default:
			// State-changing methods: verify CSRF token
			cookie, err := c.Cookie(csrfCookieName)
			if err != nil || cookie == "" {
				c.JSON(http.StatusForbidden, models.NewErrorResponse(
					models.NewAppError(http.StatusForbidden, "CSRF_MISSING", "CSRF token cookie required"),
				))
				c.Abort()
				return
			}
			header := c.GetHeader(csrfHeaderName)
			if !hmac.Equal([]byte(cookie), []byte(header)) {
				c.JSON(http.StatusForbidden, models.NewErrorResponse(
					models.NewAppError(http.StatusForbidden, "CSRF_INVALID", "CSRF token mismatch"),
				))
				c.Abort()
				return
			}
			c.Next()
		}
	}
}

func generateCSRFToken() (string, error) {
	b := make([]byte, csrfTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

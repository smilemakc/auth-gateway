package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// OAuthHandler handles OAuth-related requests
type OAuthHandler struct {
	oauthService *service.OAuthService
	logger       *logger.Logger
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	oauthService *service.OAuthService,
	logger *logger.Logger,
) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		logger:       logger,
	}
}

// Login initiates OAuth login flow
// @Summary Initiate OAuth login
// @Description Redirect to OAuth provider for authentication
// @Tags OAuth
// @Param provider path string true "OAuth provider" Enums(google, yandex, github, instagram, telegram)
// @Success 302
// @Router /auth/{provider} [get]
func (h *OAuthHandler) Login(c *gin.Context) {
	provider := c.Param("provider")

	if !models.IsValidProvider(provider) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrInvalidProvider))
		return
	}

	// Generate state for CSRF protection
	state, err := h.oauthService.GenerateState()
	if err != nil {
		h.logger.Error("Failed to generate OAuth state", map[string]interface{}{
			"error":    err.Error(),
			"provider": provider,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	// Store state in session/cookie for validation
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // 10 minutes

	// Get authorization URL
	authURL, err := h.oauthService.GetAuthURL(models.OAuthProvider(provider), state)
	if err != nil {
		if err == models.ErrInvalidProvider {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
			return
		}
		h.logger.Error("Failed to get OAuth URL", map[string]interface{}{
			"error":    err.Error(),
			"provider": provider,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	// Redirect to OAuth provider
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// Callback handles OAuth callback
// @Summary Handle OAuth callback
// @Description Process OAuth callback and create/login user
// @Tags OAuth
// @Param provider path string true "OAuth provider"
// @Param code query string true "Authorization code"
// @Param state query string true "CSRF state"
// @Success 302
// @Router /auth/{provider}/callback [get]
func (h *OAuthHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	if !models.IsValidProvider(provider) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrInvalidProvider))
		return
	}

	// Validate state for CSRF protection
	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState != state {
		h.logger.Warn("Invalid OAuth state", map[string]interface{}{
			"provider": provider,
			"error":    "state_mismatch",
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid state parameter",
		})
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization code is required",
		})
		return
	}

	// Handle OAuth callback
	response, err := h.oauthService.HandleCallback(
		c.Request.Context(),
		models.OAuthProvider(provider),
		code,
	)

	if err != nil {
		h.logger.Error("OAuth callback failed", map[string]interface{}{
			"error":    err.Error(),
			"provider": provider,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	// For web app, redirect to frontend with tokens
	// For API, return JSON response
	if c.Query("response_type") == "json" {
		c.JSON(http.StatusOK, response)
		return
	}

	// Redirect to frontend with tokens in URL (not recommended for production)
	// Better approach: set httpOnly cookies or use a redirect with a one-time code
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3001")
	redirectURL := fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s&is_new_user=%v",
		frontendURL,
		response.AccessToken,
		response.RefreshToken,
		response.IsNewUser,
	)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// TelegramCallback handles Telegram widget callback
// @Summary Handle Telegram auth callback
// @Description Process Telegram widget authentication
// @Tags OAuth
// @Accept json
// @Produce json
// @Param data body map[string]interface{} true "Telegram auth data"
// @Success 200 {object} models.OAuthLoginResponse
// @Router /auth/telegram/callback [post]
func (h *OAuthHandler) TelegramCallback(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	// Verify Telegram data hash
	if !h.verifyTelegramAuth(data) {
		h.logger.Warn("Invalid Telegram auth data", nil)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid authentication data",
		})
		return
	}

	// Create OAuth user info from Telegram data
	userInfo := &models.OAuthUserInfo{
		Provider:       string(models.ProviderTelegram),
		ProviderUserID: fmt.Sprintf("%v", data["id"]),
		Name:           getString(data, "first_name"),
		Username:       getString(data, "username"),
		ProfilePicture: getString(data, "photo_url"),
	}

	if lastName := getString(data, "last_name"); lastName != "" {
		userInfo.Name += " " + lastName
	}

	// Create a dummy code for HandleCallback
	// In real implementation, you might want to refactor HandleCallback
	// to accept userInfo directly for Telegram
	response, err := h.oauthService.HandleCallback(
		c.Request.Context(),
		models.ProviderTelegram,
		"telegram_auth", // Telegram doesn't use OAuth code flow
	)

	if err != nil {
		h.logger.Error("Telegram auth failed", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, response)
}

// verifyTelegramAuth verifies Telegram widget authentication data
func (h *OAuthHandler) verifyTelegramAuth(data map[string]interface{}) bool {
	// TODO: Implement proper Telegram auth verification
	// https://core.telegram.org/widgets/login#checking-authorization
	// This requires crypto/sha256 and crypto/hmac
	// For now, accept all (NOT SECURE for production!)
	return true
}

// GetProviders returns available OAuth providers
// @Summary Get available OAuth providers
// @Description List all configured OAuth providers
// @Tags OAuth
// @Produce json
// @Success 200 {array} models.OAuthProviderInfo
// @Router /auth/providers [get]
func (h *OAuthHandler) GetProviders(c *gin.Context) {
	providers := []models.OAuthProviderInfo{
		{
			Name:        "google",
			DisplayName: "Google",
			IconURL:     "/icons/google.svg",
			Enabled:     getEnv("GOOGLE_CLIENT_ID", "") != "",
		},
		{
			Name:        "yandex",
			DisplayName: "Yandex",
			IconURL:     "/icons/yandex.svg",
			Enabled:     getEnv("YANDEX_CLIENT_ID", "") != "",
		},
		{
			Name:        "github",
			DisplayName: "GitHub",
			IconURL:     "/icons/github.svg",
			Enabled:     getEnv("GITHUB_CLIENT_ID", "") != "",
		},
		{
			Name:        "instagram",
			DisplayName: "Instagram",
			IconURL:     "/icons/instagram.svg",
			Enabled:     getEnv("INSTAGRAM_CLIENT_ID", "") != "",
		},
		{
			Name:        "telegram",
			DisplayName: "Telegram",
			IconURL:     "/icons/telegram.svg",
			Enabled:     getEnv("TELEGRAM_BOT_TOKEN", "") != "",
		},
	}

	c.JSON(http.StatusOK, providers)
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

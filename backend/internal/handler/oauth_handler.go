package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// OAuthHandler handles OAuth-related requests
type OAuthHandler struct {
	oauthService     *service.OAuthService
	logger           *logger.Logger
	telegramBotToken string
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	oauthService *service.OAuthService,
	logger *logger.Logger,
	telegramBotToken string,
) *OAuthHandler {
	return &OAuthHandler{
		oauthService:     oauthService,
		logger:           logger,
		telegramBotToken: telegramBotToken,
	}
}

// Login initiates OAuth login flow
// @Summary Initiate OAuth login
// @Description Redirect to OAuth provider for authentication (Google, Yandex, GitHub, Instagram, Telegram, 1C)
// @Tags OAuth
// @Param provider path string true "OAuth provider" Enums(google, yandex, github, instagram, telegram, onec)
// @Success 302 {string} string "Redirect to OAuth provider"
// @Failure 400 {object} models.ErrorResponse "Invalid provider"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /api/auth/{provider} [get]
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

	// Get authorization URL (with optional app context)
	appID, _ := utils.GetApplicationIDFromContext(c)
	authURL, err := h.oauthService.GetAuthURL(c.Request.Context(), models.OAuthProvider(provider), state, appID)
	if err != nil {
		if errors.Is(err, models.ErrInvalidProvider) {
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
// @Description Process OAuth callback from provider, create or login user, and redirect with tokens
// @Tags OAuth
// @Produce json
// @Param provider path string true "OAuth provider" Enums(google, yandex, github, instagram, onec)
// @Param code query string true "Authorization code from OAuth provider"
// @Param state query string true "CSRF state parameter"
// @Param response_type query string false "Response type: 'json' for JSON response, otherwise redirect" Enums(json)
// @Success 200 {object} models.OAuthLoginResponse "JSON response when response_type=json"
// @Success 302 {string} string "Redirect to frontend with tokens"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /api/auth/{provider}/callback [get]
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

	// Handle OAuth callback (with optional app context)
	appID, _ := utils.GetApplicationIDFromContext(c)
	response, err := h.oauthService.HandleCallback(
		c.Request.Context(),
		models.OAuthProvider(provider),
		code,
		utils.GetClientIP(c),
		c.Request.UserAgent(),
		appID,
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
// @Description Process Telegram widget authentication data and return tokens
// @Tags OAuth
// @Accept json
// @Produce json
// @Param data body models.TelegramAuthData true "Telegram widget auth data"
// @Success 200 {object} models.OAuthLoginResponse
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 401 {object} models.ErrorResponse "Invalid authentication data"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /api/auth/telegram/callback [post]
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
	appID, _ := utils.GetApplicationIDFromContext(c)
	response, err := h.oauthService.HandleCallback(
		c.Request.Context(),
		models.ProviderTelegram,
		"telegram_auth", // Telegram doesn't use OAuth code flow
		utils.GetClientIP(c),
		c.Request.UserAgent(),
		appID,
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

const telegramAuthMaxAge = 86400

func (h *OAuthHandler) verifyTelegramAuth(data map[string]interface{}) bool {
	if h.telegramBotToken == "" {
		return false
	}

	receivedHash, ok := data["hash"].(string)
	if !ok || receivedHash == "" {
		return false
	}

	if !h.isTelegramAuthDateValid(data) {
		return false
	}

	expectedHash := h.computeTelegramHash(data)
	return hmac.Equal([]byte(receivedHash), []byte(expectedHash))
}

func (h *OAuthHandler) isTelegramAuthDateValid(data map[string]interface{}) bool {
	authDateVal, exists := data["auth_date"]
	if !exists {
		return false
	}

	authDateFloat, ok := authDateVal.(float64)
	if !ok {
		return false
	}

	authDate := time.Unix(int64(authDateFloat), 0)
	elapsed := time.Since(authDate)
	return elapsed >= 0 && elapsed.Seconds() <= telegramAuthMaxAge
}

func (h *OAuthHandler) computeTelegramHash(data map[string]interface{}) string {
	dataCheckString := buildTelegramDataCheckString(data)
	secretKey := sha256.Sum256([]byte(h.telegramBotToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	return hex.EncodeToString(mac.Sum(nil))
}

func buildTelegramDataCheckString(data map[string]interface{}) string {
	var parts []string
	for k, v := range data {
		if k == "hash" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", k, formatTelegramValue(v)))
	}
	sort.Strings(parts)
	return strings.Join(parts, "\n")
}

func formatTelegramValue(v interface{}) string {
	if floatVal, ok := v.(float64); ok && floatVal == float64(int64(floatVal)) {
		return fmt.Sprintf("%d", int64(floatVal))
	}
	return fmt.Sprintf("%v", v)
}

// GetProviders returns available OAuth providers
// @Summary Get available OAuth providers
// @Description List all configured OAuth providers with their enabled status
// @Tags OAuth
// @Produce json
// @Success 200 {array} models.OAuthProviderInfo
// @Router /api/auth/providers [get]
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
		{
			Name:        "onec",
			DisplayName: "1C",
			IconURL:     "/icons/onec.svg",
			Enabled:     getEnv("OAUTH_ONEC_ENABLED", "") == "true" && getEnv("OAUTH_ONEC_CLIENT_ID", "") != "",
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

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

type TokenHandler struct {
	jwtService    *jwt.Service
	apiKeyService *service.APIKeyService
	redis         *service.RedisService
	logger        *logger.Logger
}

func NewTokenHandler(
	jwtService *jwt.Service,
	apiKeyService *service.APIKeyService,
	redis *service.RedisService,
	log *logger.Logger,
) *TokenHandler {
	return &TokenHandler{
		jwtService:    jwtService,
		apiKeyService: apiKeyService,
		redis:         redis,
		logger:        log,
	}
}

type ValidateTokenRequest struct {
	AccessToken string `json:"access_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type ValidateTokenResponse struct {
	Valid     bool     `json:"valid" example:"true"`
	UserID    string   `json:"user_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email     string   `json:"email,omitempty" example:"user@example.com"`
	Username  string   `json:"username,omitempty" example:"johndoe"`
	Roles     []string `json:"roles,omitempty" example:"user,admin"`
	ExpiresAt int64    `json:"expires_at,omitempty" example:"1234567890"`
	IsActive  bool     `json:"is_active,omitempty" example:"true"`
}

type ValidateTokenErrorResponse struct {
	Valid        bool   `json:"valid" example:"false"`
	ErrorMessage string `json:"error_message" example:"Token expired"`
}

// ValidateToken validates a JWT access token or API key
// @Summary Validate token
// @Description Validates a JWT access token or API key and returns user information. Supports both JWT tokens and API keys with 'agw_' prefix.
// @Tags Token
// @Accept json
// @Produce json
// @Param request body ValidateTokenRequest true "Token validation request"
// @Success 200 {object} ValidateTokenResponse "Token is valid"
// @Success 401 {object} ValidateTokenErrorResponse "Token is invalid"
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/token/validate [post]
func (h *TokenHandler) ValidateToken(c *gin.Context) {
	var req ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	if req.AccessToken == "" {
		c.JSON(http.StatusUnauthorized, ValidateTokenErrorResponse{
			Valid:        false,
			ErrorMessage: "access_token is required",
		})
		return
	}

	// Check if it's an API key (starts with "agw_")
	if len(req.AccessToken) > 4 && req.AccessToken[:4] == "agw_" {
		h.validateAPIKey(c, req.AccessToken)
		return
	}

	h.validateJWT(c, req.AccessToken)
}

func (h *TokenHandler) validateAPIKey(c *gin.Context, apiKey string) {
	_, user, err := h.apiKeyService.ValidateAPIKey(c.Request.Context(), apiKey)
	if err != nil {
		h.logger.Debug("API key validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusUnauthorized, ValidateTokenErrorResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		})
		return
	}

	roleNames := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roleNames[i] = role.Name
	}

	c.JSON(http.StatusOK, ValidateTokenResponse{
		Valid:     user.IsActive,
		UserID:    user.ID.String(),
		Email:     user.Email,
		Username:  user.Username,
		Roles:     roleNames,
		ExpiresAt: 0,
		IsActive:  user.IsActive,
	})
}

func (h *TokenHandler) validateJWT(c *gin.Context, token string) {
	claims, err := h.jwtService.ValidateAccessToken(token)
	if err != nil {
		h.logger.Debug("Token validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusUnauthorized, ValidateTokenErrorResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		})
		return
	}

	tokenHash := utils.HashToken(token)
	blacklisted, err := h.redis.IsBlacklisted(c.Request.Context(), tokenHash)
	if err != nil {
		h.logger.Warn("Redis blacklist check failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	if blacklisted {
		c.JSON(http.StatusUnauthorized, ValidateTokenErrorResponse{
			Valid:        false,
			ErrorMessage: "token is blacklisted",
		})
		return
	}

	c.JSON(http.StatusOK, ValidateTokenResponse{
		Valid:     claims.IsActive,
		UserID:    claims.UserID.String(),
		Email:     claims.Email,
		Username:  claims.Username,
		Roles:     claims.Roles,
		ExpiresAt: claims.ExpiresAt.Unix(),
		IsActive:  claims.IsActive,
	})
}

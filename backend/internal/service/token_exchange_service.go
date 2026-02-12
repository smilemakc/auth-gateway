package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

const (
	tokenExchangePrefix     = "token_exchange:"
	tokenExchangeRatePrefix = "token_exchange_rate:"
	tokenExchangeTTL        = 30 * time.Second
	tokenExchangeRateWindow = 60 * time.Second
	tokenExchangeRateLimit  = 10
)

type TokenExchangeService struct {
	redis        *RedisService
	jwtService   TokenService
	appRepo      ApplicationStore
	userRepo     UserStore
	auditService AuditLogger
}

func NewTokenExchangeService(
	redis *RedisService,
	jwtService TokenService,
	appRepo ApplicationStore,
	userRepo UserStore,
	auditService AuditLogger,
) *TokenExchangeService {
	return &TokenExchangeService{
		redis:        redis,
		jwtService:   jwtService,
		appRepo:      appRepo,
		userRepo:     userRepo,
		auditService: auditService,
	}
}

func (s *TokenExchangeService) CreateExchange(ctx context.Context, req *models.CreateTokenExchangeRequest, sourceAppID *uuid.UUID) (*models.CreateTokenExchangeResponse, error) {
	claims, err := s.jwtService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return nil, models.NewAppError(401, "Invalid or expired access token")
	}

	if err := s.checkRateLimit(ctx, claims.UserID); err != nil {
		return nil, err
	}

	targetAppID, err := uuid.Parse(req.TargetAppID)
	if err != nil {
		return nil, models.NewAppError(400, "Invalid target application ID")
	}

	targetApp, err := s.appRepo.GetApplicationByID(ctx, targetAppID)
	if err != nil {
		return nil, models.NewAppError(404, "Target application not found")
	}

	if !targetApp.IsActive {
		return nil, models.NewAppError(403, "Target application is not active")
	}

	code, err := generateExchangeCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate exchange code: %w", err)
	}

	var resolvedSourceAppID uuid.UUID
	if sourceAppID != nil {
		resolvedSourceAppID = *sourceAppID
	}

	exchangeData := &models.TokenExchangeCode{
		Code:        code,
		UserID:      claims.UserID,
		SourceAppID: resolvedSourceAppID,
		TargetAppID: targetAppID,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(tokenExchangeTTL),
	}

	if err := s.storeExchangeCode(ctx, code, exchangeData); err != nil {
		return nil, fmt.Errorf("failed to store exchange code: %w", err)
	}

	s.auditService.Log(AuditLogParams{
		UserID: &claims.UserID,
		Action: models.ActionTokenExchangeCreate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"target_app_id": targetAppID.String(),
		},
	})

	var redirectURL string
	if targetApp.HomepageURL != "" {
		redirectURL = targetApp.HomepageURL
	}

	return &models.CreateTokenExchangeResponse{
		ExchangeCode: code,
		ExpiresAt:    exchangeData.ExpiresAt,
		RedirectURL:  redirectURL,
	}, nil
}

func (s *TokenExchangeService) RedeemExchange(ctx context.Context, req *models.RedeemTokenExchangeRequest, redeemingAppID *uuid.UUID) (*models.RedeemTokenExchangeResponse, error) {
	exchangeData, err := s.getAndDeleteExchangeCode(ctx, req.ExchangeCode)
	if err != nil {
		return nil, models.NewAppError(400, "Invalid or expired exchange code")
	}

	if redeemingAppID != nil && *redeemingAppID != exchangeData.TargetAppID {
		s.auditService.Log(AuditLogParams{
			UserID: &exchangeData.UserID,
			Action: models.ActionTokenExchangeRedeem,
			Status: models.StatusFailed,
			Details: map[string]interface{}{
				"reason":           "app_mismatch",
				"expected_app_id":  exchangeData.TargetAppID.String(),
				"redeeming_app_id": redeemingAppID.String(),
			},
		})
		return nil, models.NewAppError(403, "Exchange code was not issued for this application")
	}

	user, err := s.userRepo.GetByID(ctx, exchangeData.UserID, utils.Ptr(true), UserGetWithRoles())
	if err != nil {
		return nil, models.NewAppError(404, "User not found")
	}

	accessToken, err := s.jwtService.GenerateAccessToken(user, &exchangeData.TargetAppID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user, &exchangeData.TargetAppID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	s.auditService.Log(AuditLogParams{
		UserID: &exchangeData.UserID,
		Action: models.ActionTokenExchangeRedeem,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"target_app_id": exchangeData.TargetAppID.String(),
			"source_app_id": exchangeData.SourceAppID.String(),
		},
	})

	return &models.RedeemTokenExchangeResponse{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		ExpiresIn:     int64(s.jwtService.GetAccessTokenExpiration().Seconds()),
		User:          user.PublicUser(),
		ApplicationID: exchangeData.TargetAppID.String(),
	}, nil
}

func (s *TokenExchangeService) checkRateLimit(ctx context.Context, userID uuid.UUID) error {
	key := tokenExchangeRatePrefix + userID.String()
	count, err := s.redis.IncrementRateLimit(ctx, key, tokenExchangeRateWindow)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if count > tokenExchangeRateLimit {
		return models.ErrRateLimitExceeded
	}
	return nil
}

func (s *TokenExchangeService) storeExchangeCode(ctx context.Context, code string, data *models.TokenExchangeCode) error {
	key := tokenExchangePrefix + code
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal exchange data: %w", err)
	}
	return s.redis.Set(ctx, key, jsonData, tokenExchangeTTL)
}

func (s *TokenExchangeService) getAndDeleteExchangeCode(ctx context.Context, code string) (*models.TokenExchangeCode, error) {
	key := tokenExchangePrefix + code

	data, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if err := s.redis.Delete(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to delete exchange code: %w", err)
	}

	var exchangeData models.TokenExchangeCode
	if err := json.Unmarshal([]byte(data), &exchangeData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal exchange data: %w", err)
	}

	return &exchangeData, nil
}

func generateExchangeCode() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}

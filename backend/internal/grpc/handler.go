package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// AuthHandlerV2 implements the gRPC AuthService with API key support
type AuthHandlerV2 struct {
	UnimplementedAuthServiceServer
	jwtService    *jwt.Service
	userRepo      *repository.UserRepository
	tokenRepo     *repository.TokenRepository
	apiKeyService *service.APIKeyService
	redis         *service.RedisService
	logger        *logger.Logger
}

// NewAuthHandlerV2 creates a new auth handler with API key support
func NewAuthHandlerV2(
	jwtService *jwt.Service,
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	apiKeyService *service.APIKeyService,
	redis *service.RedisService,
	log *logger.Logger,
) *AuthHandlerV2 {
	return &AuthHandlerV2{
		jwtService:    jwtService,
		userRepo:      userRepo,
		tokenRepo:     tokenRepo,
		apiKeyService: apiKeyService,
		redis:         redis,
		logger:        log,
	}
}

// ValidateToken validates a JWT access token or API key and returns user information
func (h *AuthHandlerV2) ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return &ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "access_token is required",
		}, nil
	}

	// Check if it's an API key (starts with "agw_")
	if len(req.AccessToken) > 4 && req.AccessToken[:4] == "agw_" {
		// Validate API key
		_, user, err := h.apiKeyService.ValidateAPIKey(ctx, req.AccessToken)
		if err != nil {
			h.logger.Debug("API key validation failed", map[string]interface{}{
				"error": err.Error(),
			})
			return &ValidateTokenResponse{
				Valid:        false,
				ErrorMessage: err.Error(),
			}, nil
		}

		// Return successful validation
		return &ValidateTokenResponse{
			Valid:     true,
			UserId:    user.ID.String(),
			Email:     user.Email,
			Username:  user.Username,
			Role:      user.Role,
			ExpiresAt: 0, // API keys don't have JWT expiration
		}, nil
	}

	// Validate JWT token
	claims, err := h.jwtService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		h.logger.Debug("Token validation failed", map[string]interface{}{
			"error": err.Error(),
		})

		return &ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Check if token is blacklisted
	tokenHash := utils.HashToken(req.AccessToken)
	blacklisted, err := h.redis.IsBlacklisted(ctx, tokenHash)
	if err != nil {
		// Log error but check database as fallback
		h.logger.Warn("Redis blacklist check failed", map[string]interface{}{
			"error": err.Error(),
		})
		blacklisted, _ = h.tokenRepo.IsBlacklisted(ctx, tokenHash)
	}

	if blacklisted {
		return &ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "token is blacklisted",
		}, nil
	}

	// Return successful validation
	return &ValidateTokenResponse{
		Valid:     true,
		UserId:    claims.UserID.String(),
		Email:     claims.Email,
		Username:  claims.Username,
		Role:      claims.Role,
		ExpiresAt: claims.ExpiresAt.Unix(),
	}, nil
}

// GetUser retrieves user information by ID
func (h *AuthHandlerV2) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Get user from repository
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == models.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		h.logger.Error("Failed to get user", map[string]interface{}{
			"user_id": req.UserId,
			"error":   err.Error(),
		})
		return nil, status.Error(codes.Internal, "internal error")
	}

	// Convert to proto user
	return &GetUserResponse{
		User: &User{
			Id:                user.ID.String(),
			Email:             user.Email,
			Username:          user.Username,
			FullName:          user.FullName,
			ProfilePictureUrl: user.ProfilePictureURL,
			Role:              user.Role,
			EmailVerified:     user.EmailVerified,
			IsActive:          user.IsActive,
			CreatedAt:         user.CreatedAt.Unix(),
			UpdatedAt:         user.UpdatedAt.Unix(),
		},
	}, nil
}

// CheckPermission checks if a user has specific permission
func (h *AuthHandlerV2) CheckPermission(ctx context.Context, req *CheckPermissionRequest) (*CheckPermissionResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Get user to check role
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == models.ErrUserNotFound {
			return &CheckPermissionResponse{
				Allowed:      false,
				ErrorMessage: "user not found",
			}, nil
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	// Simple RBAC: admin has all permissions
	if user.Role == string(models.RoleAdmin) {
		return &CheckPermissionResponse{
			Allowed: true,
			Role:    user.Role,
		}, nil
	}

	// Moderator has read permissions on all resources
	if user.Role == string(models.RoleModerator) && req.Action == "read" {
		return &CheckPermissionResponse{
			Allowed: true,
			Role:    user.Role,
		}, nil
	}

	// Regular users can only read their own resources
	if user.Role == string(models.RoleUser) && req.Action == "read" {
		return &CheckPermissionResponse{
			Allowed: true,
			Role:    user.Role,
		}, nil
	}

	// Permission denied
	return &CheckPermissionResponse{
		Allowed: false,
		Role:    user.Role,
	}, nil
}

// IntrospectToken provides detailed information about a token
func (h *AuthHandlerV2) IntrospectToken(ctx context.Context, req *IntrospectTokenRequest) (*IntrospectTokenResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}

	// Validate token
	claims, err := h.jwtService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return &IntrospectTokenResponse{
			Active:       false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Check if token is blacklisted
	tokenHash := utils.HashToken(req.AccessToken)
	blacklisted, err := h.redis.IsBlacklisted(ctx, tokenHash)
	if err != nil {
		blacklisted, _ = h.tokenRepo.IsBlacklisted(ctx, tokenHash)
	}

	// Return detailed token information
	return &IntrospectTokenResponse{
		Active:      !blacklisted,
		UserId:      claims.UserID.String(),
		Email:       claims.Email,
		Username:    claims.Username,
		Role:        claims.Role,
		IssuedAt:    claims.IssuedAt.Unix(),
		ExpiresAt:   claims.ExpiresAt.Unix(),
		NotBefore:   claims.NotBefore.Unix(),
		Subject:     claims.Subject,
		Blacklisted: blacklisted,
	}, nil
}

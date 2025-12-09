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

// AuthHandler implements the gRPC AuthService
type AuthHandler struct {
	UnimplementedAuthServiceServer
	jwtService *jwt.Service
	userRepo   *repository.UserRepository
	tokenRepo  *repository.TokenRepository
	redis      *service.RedisService
	logger     *logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	jwtService *jwt.Service,
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	redis *service.RedisService,
	log *logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		jwtService: jwtService,
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		redis:      redis,
		logger:     log,
	}
}

// ValidateToken validates a JWT access token and returns user information
func (h *AuthHandler) ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return &ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "access_token is required",
		}, nil
	}

	// Validate token
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
		blacklisted, _ = h.tokenRepo.IsBlacklisted(tokenHash)
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
func (h *AuthHandler) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Get user from repository
	user, err := h.userRepo.GetByID(userID)
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
func (h *AuthHandler) CheckPermission(ctx context.Context, req *CheckPermissionRequest) (*CheckPermissionResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Get user to check role
	user, err := h.userRepo.GetByID(userID)
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
	// This is a simple example - in production, implement proper RBAC/ABAC
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
func (h *AuthHandler) IntrospectToken(ctx context.Context, req *IntrospectTokenRequest) (*IntrospectTokenResponse, error) {
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
		blacklisted, _ = h.tokenRepo.IsBlacklisted(tokenHash)
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

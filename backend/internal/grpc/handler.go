package grpc

import (
	"context"
	"errors"

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
	rbacRepo      *repository.RBACRepository
	apiKeyService *service.APIKeyService
	authService   *service.AuthService
	redis         *service.RedisService
	logger        *logger.Logger
}

// NewAuthHandlerV2 creates a new auth handler with API key support
func NewAuthHandlerV2(
	jwtService *jwt.Service,
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	rbacRepo *repository.RBACRepository,
	apiKeyService *service.APIKeyService,
	authService *service.AuthService,
	redis *service.RedisService,
	log *logger.Logger,
) *AuthHandlerV2 {
	return &AuthHandlerV2{
		jwtService:    jwtService,
		userRepo:      userRepo,
		tokenRepo:     tokenRepo,
		rbacRepo:      rbacRepo,
		apiKeyService: apiKeyService,
		authService:   authService,
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

		// Load user with roles
		userWithRoles, err := h.userRepo.GetByIDWithRoles(ctx, user.ID, nil)
		if err != nil {
			h.logger.Error("Failed to load user roles", map[string]interface{}{
				"user_id": user.ID.String(),
				"error":   err.Error(),
			})
			return &ValidateTokenResponse{
				Valid:        false,
				ErrorMessage: "failed to load user roles",
			}, nil
		}

		// Return successful validation
		return &ValidateTokenResponse{
			Valid:     user.IsActive,
			UserId:    userWithRoles.ID.String(),
			Email:     userWithRoles.Email,
			Username:  userWithRoles.Username,
			Roles:     extractRoleNames(userWithRoles.Roles),
			ExpiresAt: 0, // API keys don't have JWT expiration
			IsActive:  user.IsActive,
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
		Valid:     claims.IsActive,
		UserId:    claims.UserID.String(),
		Email:     claims.Email,
		Username:  claims.Username,
		Roles:     claims.Roles,
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

	// Get user from repository with roles
	user, err := h.userRepo.GetByIDWithRoles(ctx, userID, nil)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
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
			Roles:             mapRolesToProto(user.Roles),
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

	// Get user roles with permissions
	roles, err := h.rbacRepo.GetUserRoles(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user roles", map[string]interface{}{
			"user_id": req.UserId,
			"error":   err.Error(),
		})
		return nil, status.Error(codes.Internal, "internal error")
	}

	if len(roles) == 0 {
		return &CheckPermissionResponse{
			Allowed:      false,
			Roles:        []string{},
			ErrorMessage: "user has no roles",
		}, nil
	}

	// Check if any role has the required permission
	for _, role := range roles {
		for _, permission := range role.Permissions {
			if permission.Resource == req.Resource && permission.Action == req.Action {
				return &CheckPermissionResponse{
					Allowed: true,
					Roles:   extractRoleNames(roles),
				}, nil
			}
		}
	}

	// Permission denied
	return &CheckPermissionResponse{
		Allowed: false,
		Roles:   extractRoleNames(roles),
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
		Roles:       claims.Roles,
		IssuedAt:    claims.IssuedAt.Unix(),
		ExpiresAt:   claims.ExpiresAt.Unix(),
		NotBefore:   claims.NotBefore.Unix(),
		Subject:     claims.Subject,
		Blacklisted: blacklisted,
	}, nil
}

// mapRolesToProto converts model Roles to proto RoleInfo array
func mapRolesToProto(roles []models.Role) []RoleInfo {
	protoRoles := make([]RoleInfo, len(roles))
	for i, role := range roles {
		protoRoles[i] = RoleInfo{
			Id:          role.ID.String(),
			Name:        role.Name,
			DisplayName: role.DisplayName,
		}
	}
	return protoRoles
}

// extractRoleNames extracts role names from Role array
func extractRoleNames(roles []models.Role) []string {
	names := make([]string, len(roles))
	for i, role := range roles {
		names[i] = role.Name
	}
	return names
}

// CreateUser creates a new user account
func (h *AuthHandlerV2) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	// Validate required fields
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if req.Email == "" && req.Phone == "" {
		return nil, status.Error(codes.InvalidArgument, "either email or phone is required")
	}

	// Convert phone string to pointer if provided
	var phone *string
	if req.Phone != "" {
		phone = &req.Phone
	}

	// Create user request for AuthService
	createReq := &models.CreateUserRequest{
		Email:       req.Email,
		Phone:       phone,
		Username:    req.Username,
		Password:    req.Password,
		FullName:    req.FullName,
		AccountType: req.AccountType,
	}

	// Create device info for token generation (gRPC doesn't have browser/device info)
	deviceInfo := models.DeviceInfo{
		DeviceType: "grpc_client",
		OS:         "unknown",
		Browser:    "grpc",
	}

	// Call AuthService.SignUp
	authResp, err := h.authService.SignUp(ctx, createReq, "", "", deviceInfo)
	if err != nil {
		h.logger.Error("Failed to create user via gRPC", map[string]interface{}{
			"error":    err.Error(),
			"username": req.Username,
		})

		// Convert error to appropriate gRPC status
		if appErr, ok := err.(*models.AppError); ok {
			switch appErr.Code {
			case 400:
				return nil, status.Error(codes.InvalidArgument, appErr.Message)
			case 409:
				return nil, status.Error(codes.AlreadyExists, appErr.Message)
			default:
				return nil, status.Error(codes.Internal, appErr.Message)
			}
		}

		// Check for known error types
		switch err {
		case models.ErrEmailAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		case models.ErrUsernameAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, "username already exists")
		case models.ErrPhoneAlreadyExists:
			return nil, status.Error(codes.AlreadyExists, "phone already exists")
		default:
			return nil, status.Error(codes.Internal, "failed to create user")
		}
	}

	// Convert response
	user := authResp.User
	return &CreateUserResponse{
		User: &User{
			Id:                user.ID.String(),
			Email:             user.Email,
			Username:          user.Username,
			FullName:          user.FullName,
			ProfilePictureUrl: user.ProfilePictureURL,
			Roles:             mapRolesToProto(user.Roles),
			EmailVerified:     user.EmailVerified,
			IsActive:          user.IsActive,
			CreatedAt:         user.CreatedAt.Unix(),
			UpdatedAt:         user.UpdatedAt.Unix(),
		},
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresIn:    authResp.ExpiresIn,
	}, nil
}

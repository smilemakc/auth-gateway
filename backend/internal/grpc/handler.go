package grpc

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	pb "github.com/smilemakc/auth-gateway/proto"
)

// AuthHandlerV2 implements the gRPC AuthService with API key support
type AuthHandlerV2 struct {
	pb.UnimplementedAuthServiceServer
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
func (h *AuthHandlerV2) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return &pb.ValidateTokenResponse{
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
			return &pb.ValidateTokenResponse{
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
			return &pb.ValidateTokenResponse{
				Valid:        false,
				ErrorMessage: "failed to load user roles",
			}, nil
		}

		// Return successful validation
		return &pb.ValidateTokenResponse{
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

		return &pb.ValidateTokenResponse{
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
		return &pb.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: "token is blacklisted",
		}, nil
	}

	// Return successful validation
	return &pb.ValidateTokenResponse{
		Valid:     claims.IsActive,
		UserId:    claims.UserID.String(),
		Email:     claims.Email,
		Username:  claims.Username,
		Roles:     claims.Roles,
		ExpiresAt: claims.ExpiresAt.Unix(),
	}, nil
}

// GetUser retrieves user information by ID
func (h *AuthHandlerV2) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
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
	return &pb.GetUserResponse{
		User: &pb.User{
			Id:                user.ID.String(),
			Email:             user.Email,
			Username:          user.Username,
			FullName:          user.FullName,
			ProfilePictureUrl: user.ProfilePictureURL,
			Roles:             extractRoleNames(user.Roles),
			EmailVerified:     user.EmailVerified,
			IsActive:          user.IsActive,
			CreatedAt:         user.CreatedAt.Unix(),
			UpdatedAt:         user.UpdatedAt.Unix(),
		},
	}, nil
}

// CheckPermission checks if a user has specific permission
func (h *AuthHandlerV2) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
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
		return &pb.CheckPermissionResponse{
			Allowed:      false,
			ErrorMessage: "user has no roles",
		}, nil
	}

	// Check if any role has the required permission
	for _, role := range roles {
		for _, permission := range role.Permissions {
			if permission.Resource == req.Resource && permission.Action == req.Action {
				return &pb.CheckPermissionResponse{
					Allowed: true,
					Role:    role.Name,
				}, nil
			}
		}
	}

	// Permission denied
	return &pb.CheckPermissionResponse{
		Allowed: false,
	}, nil
}

// IntrospectToken provides detailed information about a token
func (h *AuthHandlerV2) IntrospectToken(ctx context.Context, req *pb.IntrospectTokenRequest) (*pb.IntrospectTokenResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}

	// Validate token
	claims, err := h.jwtService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return &pb.IntrospectTokenResponse{
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

	// Get role (use first role if multiple)
	role := ""
	if len(claims.Roles) > 0 {
		role = claims.Roles[0]
	}

	// Return detailed token information
	return &pb.IntrospectTokenResponse{
		Active:      !blacklisted,
		UserId:      claims.UserID.String(),
		Email:       claims.Email,
		Username:    claims.Username,
		Role:        role,
		IssuedAt:    claims.IssuedAt.Unix(),
		ExpiresAt:   claims.ExpiresAt.Unix(),
		NotBefore:   claims.NotBefore.Unix(),
		Subject:     claims.Subject,
		Blacklisted: blacklisted,
	}, nil
}

// InitPasswordlessRegistration initiates passwordless two-step registration
func (h *AuthHandlerV2) InitPasswordlessRegistration(ctx context.Context, req *pb.InitPasswordlessRegistrationRequest) (*pb.InitPasswordlessRegistrationResponse, error) {
	// Validate that either email or phone is provided
	if req.Email == "" && req.Phone == "" {
		return nil, status.Error(codes.InvalidArgument, "either email or phone is required")
	}

	// Convert to internal model
	var email, phone *string
	if req.Email != "" {
		email = &req.Email
	}
	if req.Phone != "" {
		phone = &req.Phone
	}

	internalReq := &models.InitPasswordlessRegistrationRequest{
		Email:    email,
		Phone:    phone,
		Username: req.Username,
		FullName: req.FullName,
	}

	// Call AuthService (gRPC doesn't have IP/UserAgent)
	err := h.authService.InitPasswordlessRegistration(ctx, internalReq, "", "grpc")
	if err != nil {
		h.logger.Error("Failed to init passwordless registration via gRPC", map[string]interface{}{
			"error": err.Error(),
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
		if errors.Is(err, models.ErrEmailAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		}
		if errors.Is(err, models.ErrPhoneAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "phone already exists")
		}
		return nil, status.Error(codes.Internal, "failed to initiate registration")
	}

	return &pb.InitPasswordlessRegistrationResponse{
		Success: true,
		Message: "OTP sent successfully. Please verify to complete registration.",
	}, nil
}

// CompletePasswordlessRegistration completes registration after OTP verification
func (h *AuthHandlerV2) CompletePasswordlessRegistration(ctx context.Context, req *pb.CompletePasswordlessRegistrationRequest) (*pb.CompletePasswordlessRegistrationResponse, error) {
	// Validate required fields
	if req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}
	if req.Email == "" && req.Phone == "" {
		return nil, status.Error(codes.InvalidArgument, "either email or phone is required")
	}

	// Convert to internal model
	var email, phone *string
	if req.Email != "" {
		email = &req.Email
	}
	if req.Phone != "" {
		phone = &req.Phone
	}

	internalReq := &models.CompletePasswordlessRegistrationRequest{
		Email: email,
		Phone: phone,
		Code:  req.Code,
	}

	// Create device info for token generation (gRPC doesn't have browser/device info)
	deviceInfo := models.DeviceInfo{
		DeviceType: "grpc_client",
		OS:         "unknown",
		Browser:    "grpc",
	}

	// Call AuthService
	resp, err := h.authService.CompletePasswordlessRegistration(ctx, internalReq, "", "", deviceInfo)
	if err != nil {
		h.logger.Error("Failed to complete passwordless registration via gRPC", map[string]interface{}{
			"error": err.Error(),
		})

		// Convert error to appropriate gRPC status
		if appErr, ok := err.(*models.AppError); ok {
			switch appErr.Code {
			case 400:
				return nil, status.Error(codes.InvalidArgument, appErr.Message)
			case 404:
				return nil, status.Error(codes.NotFound, appErr.Message)
			default:
				return nil, status.Error(codes.Internal, appErr.Message)
			}
		}

		return nil, status.Error(codes.Internal, "failed to complete registration")
	}

	// Convert user to proto
	user := resp.User
	return &pb.CompletePasswordlessRegistrationResponse{
		User: &pb.User{
			Id:                user.ID.String(),
			Email:             user.Email,
			Username:          user.Username,
			FullName:          user.FullName,
			ProfilePictureUrl: user.ProfilePictureURL,
			Roles:             extractRoleNames(user.Roles),
			EmailVerified:     user.EmailVerified,
			IsActive:          user.IsActive,
			CreatedAt:         user.CreatedAt.Unix(),
			UpdatedAt:         user.UpdatedAt.Unix(),
		},
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
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
func (h *AuthHandlerV2) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Validate required fields
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if req.Email == "" && req.Phone == "" {
		return nil, status.Error(codes.InvalidArgument, "either email or phone is required")
	}

	email := utils.NormalizeEmail(req.Email)
	phone := utils.NormalizePhone(req.Phone)
	username := req.Username
	if username == "" {
		username = utils.Default(email, strings.ReplaceAll(phone, "+", ""))
	}

	// Create user request for AuthService
	createReq := &models.CreateUserRequest{
		Email:       email,
		Phone:       utils.Ptr(phone),
		Username:    username,
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
			"username": username,
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
		if errors.Is(err, models.ErrEmailAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		}
		if errors.Is(err, models.ErrUsernameAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "username already exists")
		}
		if errors.Is(err, models.ErrPhoneAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "phone already exists")
		}
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	// Convert response
	user := authResp.User
	return &pb.CreateUserResponse{
		User: &pb.User{
			Id:                user.ID.String(),
			Email:             user.Email,
			Username:          user.Username,
			FullName:          user.FullName,
			ProfilePictureUrl: user.ProfilePictureURL,
			Roles:             extractRoleNames(user.Roles),
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

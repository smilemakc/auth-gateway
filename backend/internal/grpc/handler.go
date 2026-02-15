package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	pb "github.com/smilemakc/auth-gateway/proto"
)

// ClientInfo holds client identification from gRPC metadata
type ClientInfo struct {
	IP            string
	UserAgent     string
	ApplicationID string
	ClientName    string
	ClientVersion string
	Platform      string
	Environment   string
}

// parseOSFromUserAgent attempts to extract OS info from user agent string
func parseOSFromUserAgent(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "mac") || strings.Contains(ua, "darwin"):
		return "macOS"
	case strings.Contains(ua, "linux"):
		return "Linux"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "ios") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		return "iOS"
	default:
		return ""
	}
}

// parseRuntimeFromUserAgent extracts runtime/language info from user agent
func parseRuntimeFromUserAgent(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "go/") || strings.Contains(ua, "golang") || strings.Contains(ua, "grpc-go"):
		return "Go"
	case strings.Contains(ua, "python") || strings.Contains(ua, "grpc-python"):
		return "Python"
	case strings.Contains(ua, "node") || strings.Contains(ua, "javascript") || strings.Contains(ua, "grpc-node"):
		return "Node.js"
	case strings.Contains(ua, "java") || strings.Contains(ua, "grpc-java"):
		return "Java"
	case strings.Contains(ua, "csharp") || strings.Contains(ua, "dotnet") || strings.Contains(ua, "grpc-dotnet"):
		return ".NET"
	case strings.Contains(ua, "ruby"):
		return "Ruby"
	case strings.Contains(ua, "php"):
		return "PHP"
	case strings.Contains(ua, "rust"):
		return "Rust"
	default:
		return ""
	}
}

// extractClientInfo extracts client IP, User-Agent, and Application ID from gRPC context
// Clients should send metadata:
//   - "x-application-id": application identifier (UUID)
//   - "x-client-name": service/client name (e.g., "payment-service", "mobile-app")
//   - "x-client-version": client version (e.g., "2.1.0")
//   - "x-platform": platform info (e.g., "kubernetes", "docker", "aws-lambda")
//   - "x-environment": environment (e.g., "production", "staging", "development")
//   - "user-agent": standard gRPC user agent
//   - "x-forwarded-for": original client IP (if behind proxy)
func extractClientInfo(ctx context.Context) ClientInfo {
	info := ClientInfo{
		IP:        "",
		UserAgent: "gRPC Client",
	}

	// Extract IP from peer
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		addr := p.Addr.String()
		// Remove port if present
		if idx := strings.LastIndex(addr, ":"); idx != -1 {
			info.IP = addr[:idx]
		} else {
			info.IP = addr
		}
	}

	// Extract metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// Check for forwarded IP (proxy)
		if vals := md.Get("x-forwarded-for"); len(vals) > 0 && vals[0] != "" {
			ips := strings.Split(vals[0], ",")
			info.IP = strings.TrimSpace(ips[0])
		}

		// Application ID
		if vals := md.Get("x-application-id"); len(vals) > 0 && vals[0] != "" {
			info.ApplicationID = vals[0]
		}

		// Client identification
		if vals := md.Get("x-client-name"); len(vals) > 0 && vals[0] != "" {
			info.ClientName = vals[0]
		}
		if vals := md.Get("x-client-version"); len(vals) > 0 && vals[0] != "" {
			info.ClientVersion = vals[0]
		}
		if vals := md.Get("x-platform"); len(vals) > 0 && vals[0] != "" {
			info.Platform = vals[0]
		}
		if vals := md.Get("x-environment"); len(vals) > 0 && vals[0] != "" {
			info.Environment = vals[0]
		}

		// Get base user-agent
		baseUA := ""
		if vals := md.Get("user-agent"); len(vals) > 0 && vals[0] != "" {
			baseUA = vals[0]
		}

		// Build informative user agent string
		info.UserAgent = buildUserAgent(info, baseUA)
	}

	return info
}

// buildUserAgent creates an informative user agent string from client info
func buildUserAgent(info ClientInfo, baseUA string) string {
	var parts []string

	// Add client name and version
	if info.ClientName != "" {
		if info.ClientVersion != "" {
			parts = append(parts, info.ClientName+"/"+info.ClientVersion)
		} else {
			parts = append(parts, info.ClientName)
		}
	}

	// Add runtime info from base UA
	runtime := parseRuntimeFromUserAgent(baseUA)
	if runtime != "" {
		parts = append(parts, runtime)
	}

	// Add OS info
	os := parseOSFromUserAgent(baseUA)
	if os != "" {
		parts = append(parts, os)
	}

	// Add platform
	if info.Platform != "" {
		parts = append(parts, info.Platform)
	}

	// Add environment
	if info.Environment != "" {
		parts = append(parts, "["+info.Environment+"]")
	}

	// If we have custom parts, use them
	if len(parts) > 0 {
		return strings.Join(parts, " | ") + " (gRPC)"
	}

	// Fallback: use base UA or default
	if baseUA != "" {
		// Clean up standard grpc-go user agent to be more readable
		if strings.HasPrefix(baseUA, "grpc-go/") {
			return "gRPC Go Client " + strings.TrimPrefix(baseUA, "grpc-go/")
		}
		return baseUA
	}

	return "gRPC Client"
}

// buildDeviceInfo creates DeviceInfo from ClientInfo for session tracking
func buildDeviceInfo(clientInfo ClientInfo) models.DeviceInfo {
	os := parseOSFromUserAgent(clientInfo.UserAgent)
	if os == "" {
		// Try to extract runtime info for non-browser clients
		runtime := parseRuntimeFromUserAgent(clientInfo.UserAgent)
		if runtime != "" {
			os = runtime + " Runtime"
		} else if clientInfo.Platform != "" {
			os = clientInfo.Platform
		} else {
			os = "gRPC Client"
		}
	}

	deviceType := "grpc_client"
	if clientInfo.ClientName != "" {
		deviceType = clientInfo.ClientName
	}

	return models.DeviceInfo{
		DeviceType: deviceType,
		OS:         os,
		Browser:    clientInfo.UserAgent,
	}
}

// AuthHandlerV2 implements the gRPC AuthService with API key support
type AuthHandlerV2 struct {
	pb.UnimplementedAuthServiceServer
	jwtService           *jwt.Service
	userRepo             *repository.UserRepository
	tokenRepo            *repository.TokenRepository
	rbacRepo             *repository.RBACRepository
	apiKeyService        *service.APIKeyService
	authService          *service.AuthService
	oauthProviderService *service.OAuthProviderService
	otpService           *service.OTPService
	emailProfileService  *service.EmailProfileService
	adminService         *service.AdminService
	appService           *service.ApplicationService
	redis                *service.RedisService
	tokenExchangeService *service.TokenExchangeService
	logger               *logger.Logger
}

// NewAuthHandlerV2 creates a new auth handler with API key support
func NewAuthHandlerV2(
	jwtService *jwt.Service,
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	rbacRepo *repository.RBACRepository,
	apiKeyService *service.APIKeyService,
	authService *service.AuthService,
	oauthProviderService *service.OAuthProviderService,
	otpService *service.OTPService,
	emailProfileService *service.EmailProfileService,
	adminService *service.AdminService,
	appService *service.ApplicationService,
	redis *service.RedisService,
	tokenExchangeService *service.TokenExchangeService,
	log *logger.Logger,
) *AuthHandlerV2 {
	return &AuthHandlerV2{
		jwtService:           jwtService,
		userRepo:             userRepo,
		tokenRepo:            tokenRepo,
		rbacRepo:             rbacRepo,
		apiKeyService:        apiKeyService,
		authService:          authService,
		oauthProviderService: oauthProviderService,
		otpService:           otpService,
		emailProfileService:  emailProfileService,
		adminService:         adminService,
		appService:           appService,
		redis:                redis,
		tokenExchangeService: tokenExchangeService,
		logger:               log,
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
		userWithRoles, err := h.userRepo.GetByID(ctx, user.ID, nil, service.UserGetWithRoles())
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

		response := &pb.ValidateTokenResponse{
			Valid:     user.IsActive,
			UserId:    userWithRoles.ID.String(),
			Email:     userWithRoles.Email,
			Username:  userWithRoles.Username,
			Roles:     extractRoleNames(userWithRoles.Roles),
			ExpiresAt: 0,
			IsActive:  user.IsActive,
		}

		if resolvedAppID := ResolveApplicationID(ctx, req.ApplicationId); resolvedAppID != "" {
			appID, parseErr := uuid.Parse(resolvedAppID)
			if parseErr == nil {
				response.ApplicationId = resolvedAppID
				appRoles, roleErr := h.rbacRepo.GetUserRolesInApp(ctx, user.ID, &appID)
				if roleErr == nil {
					response.AppRoles = extractRoleNames(appRoles)
				}
			}
		}

		return response, nil
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

	response := &pb.ValidateTokenResponse{
		Valid:     claims.IsActive,
		UserId:    claims.UserID.String(),
		Email:     claims.Email,
		Username:  claims.Username,
		Roles:     claims.Roles,
		ExpiresAt: claims.ExpiresAt.Unix(),
		IsActive:  claims.IsActive,
	}

	if claims.ApplicationID != nil {
		response.ApplicationId = claims.ApplicationID.String()

		appRoles, roleErr := h.rbacRepo.GetUserRolesInApp(ctx, claims.UserID, claims.ApplicationID)
		if roleErr == nil {
			response.AppRoles = extractRoleNames(appRoles)
		}
	}

	if resolvedAppID := ResolveApplicationID(ctx, req.ApplicationId); resolvedAppID != "" {
		reqAppID, parseErr := uuid.Parse(resolvedAppID)
		if parseErr != nil {
			return &pb.ValidateTokenResponse{
				Valid:        false,
				ErrorMessage: "invalid application_id format",
			}, nil
		}

		if claims.ApplicationID != nil && *claims.ApplicationID != reqAppID {
			return &pb.ValidateTokenResponse{
				Valid:        false,
				ErrorMessage: "token application_id does not match requested application_id",
			}, nil
		}

		if claims.ApplicationID == nil {
			response.ApplicationId = resolvedAppID
			appRoles, roleErr := h.rbacRepo.GetUserRolesInApp(ctx, claims.UserID, &reqAppID)
			if roleErr == nil {
				response.AppRoles = extractRoleNames(appRoles)
			}
		}
	}

	return response, nil
}

// GetUser retrieves user information by ID
func (h *AuthHandlerV2) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	user, err := h.userRepo.GetByID(ctx, userID, nil, service.UserGetWithRoles())
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

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	var roles []models.Role
	if resolvedAppID := ResolveApplicationID(ctx, req.ApplicationId); resolvedAppID != "" {
		appID, parseErr := uuid.Parse(resolvedAppID)
		if parseErr != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid application_id format")
		}
		roles, err = h.rbacRepo.GetUserRolesInApp(ctx, userID, &appID)
	} else {
		roles, err = h.rbacRepo.GetUserRoles(ctx, userID)
	}

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

	// Extract client info from gRPC metadata
	clientInfo := extractClientInfo(ctx)
	err := h.authService.InitPasswordlessRegistration(ctx, internalReq, clientInfo.IP, clientInfo.UserAgent)
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

	// Extract client info from gRPC metadata
	clientInfo := extractClientInfo(ctx)

	// Create device info for token generation
	deviceInfo := buildDeviceInfo(clientInfo)

	// Call AuthService
	resp, err := h.authService.CompletePasswordlessRegistration(ctx, internalReq, clientInfo.IP, clientInfo.UserAgent, deviceInfo)
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

// ========== OAuth Provider Methods ==========

// IntrospectOAuthToken validates OAuth access token (RFC 7662)
func (h *AuthHandlerV2) IntrospectOAuthToken(ctx context.Context, req *pb.IntrospectOAuthTokenRequest) (*pb.IntrospectOAuthTokenResponse, error) {
	if req.Token == "" {
		return &pb.IntrospectOAuthTokenResponse{
			Active:       false,
			ErrorMessage: "token is required",
		}, nil
	}

	if h.oauthProviderService == nil {
		return &pb.IntrospectOAuthTokenResponse{
			Active:       false,
			ErrorMessage: "OAuth provider service not configured",
		}, nil
	}

	result, err := h.oauthProviderService.IntrospectToken(ctx, req.Token, req.TokenTypeHint, nil)
	if err != nil {
		h.logger.Error("OAuth token introspection failed", map[string]interface{}{
			"error": err.Error(),
		})
		return &pb.IntrospectOAuthTokenResponse{
			Active:       false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.IntrospectOAuthTokenResponse{
		Active:    result.Active,
		Scope:     result.Scope,
		ClientId:  result.ClientID,
		Username:  result.Username,
		TokenType: result.TokenType,
		Exp:       result.ExpiresAt,
		Iat:       result.IssuedAt,
		Nbf:       result.NotBefore,
		Sub:       result.Subject,
		Aud:       result.Audience,
		Iss:       result.Issuer,
		Jti:       result.JWTID,
	}, nil
}

// ValidateOAuthClient validates OAuth client credentials
func (h *AuthHandlerV2) ValidateOAuthClient(ctx context.Context, req *pb.ValidateOAuthClientRequest) (*pb.ValidateOAuthClientResponse, error) {
	if req.ClientId == "" {
		return nil, status.Error(codes.InvalidArgument, "client_id is required")
	}

	if h.oauthProviderService == nil {
		return &pb.ValidateOAuthClientResponse{
			Valid:        false,
			ErrorMessage: "OAuth provider service not configured",
		}, nil
	}

	client, err := h.oauthProviderService.ValidateClientCredentials(ctx, req.ClientId, req.ClientSecret)
	if err != nil {
		h.logger.Debug("OAuth client validation failed", map[string]interface{}{
			"client_id": req.ClientId,
			"error":     err.Error(),
		})
		return &pb.ValidateOAuthClientResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.ValidateOAuthClientResponse{
		Valid:        true,
		ClientId:     client.ClientID,
		ClientName:   client.Name,
		ClientType:   client.ClientType,
		Scopes:       client.AllowedScopes,
		RedirectUris: client.RedirectURIs,
	}, nil
}

// GetOAuthClient retrieves OAuth client information by client_id
func (h *AuthHandlerV2) GetOAuthClient(ctx context.Context, req *pb.GetOAuthClientRequest) (*pb.GetOAuthClientResponse, error) {
	if req.ClientId == "" {
		return nil, status.Error(codes.InvalidArgument, "client_id is required")
	}

	if h.oauthProviderService == nil {
		return nil, status.Error(codes.Unavailable, "OAuth provider service not configured")
	}

	client, err := h.oauthProviderService.GetClientByClientID(ctx, req.ClientId)
	if err != nil {
		h.logger.Debug("OAuth client not found", map[string]interface{}{
			"client_id": req.ClientId,
			"error":     err.Error(),
		})
		return nil, status.Error(codes.NotFound, "OAuth client not found")
	}

	return &pb.GetOAuthClientResponse{
		Client: &pb.OAuthClient{
			Id:                      client.ID.String(),
			ClientId:                client.ClientID,
			ClientName:              client.Name,
			ClientType:              client.ClientType,
			RedirectUris:            client.RedirectURIs,
			Scopes:                  client.AllowedScopes,
			GrantTypes:              client.AllowedGrantTypes,
			TokenEndpointAuthMethod: "client_secret_basic",
			IsActive:                client.IsActive,
			CreatedAt:               client.CreatedAt.Unix(),
			UpdatedAt:               client.UpdatedAt.Unix(),
		},
	}, nil
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

	// Extract client info from gRPC metadata
	clientInfo := extractClientInfo(ctx)

	// Create device info for token generation
	deviceInfo := buildDeviceInfo(clientInfo)

	// Call AuthService.SignUp
	authResp, err := h.authService.SignUp(ctx, createReq, clientInfo.IP, clientInfo.UserAgent, deviceInfo, GetApplicationUUIDFromGRPCContext(ctx))
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

// Login authenticates a user with email/phone and password
func (h *AuthHandlerV2) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Validate required fields
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if req.Email == "" && req.Phone == "" {
		return nil, status.Error(codes.InvalidArgument, "either email or phone is required")
	}

	// Create sign-in request for AuthService
	var phone *string
	if req.Phone != "" {
		phone = &req.Phone
	}

	signInReq := &models.SignInRequest{
		Email:    utils.NormalizeEmail(req.Email),
		Phone:    phone,
		Password: req.Password,
	}

	// Extract client info from gRPC metadata
	clientInfo := extractClientInfo(ctx)

	// Create device info for token generation
	deviceInfo := buildDeviceInfo(clientInfo)

	// Call AuthService.SignIn
	authResp, err := h.authService.SignIn(ctx, signInReq, clientInfo.IP, clientInfo.UserAgent, deviceInfo, GetApplicationUUIDFromGRPCContext(ctx))
	if err != nil {
		h.logger.Debug("Login via gRPC failed", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})

		// Convert error to appropriate gRPC status
		if appErr, ok := err.(*models.AppError); ok {
			switch appErr.Code {
			case 400:
				return nil, status.Error(codes.InvalidArgument, appErr.Message)
			case 401:
				return nil, status.Error(codes.Unauthenticated, appErr.Message)
			default:
				return nil, status.Error(codes.Internal, appErr.Message)
			}
		}

		// Check for known error types
		if errors.Is(err, models.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "login failed")
	}

	// Handle 2FA requirement
	if authResp.Requires2FA {
		return &pb.LoginResponse{
			ErrorMessage: "2FA required - use TwoFactorToken to verify",
		}, nil
	}

	// Convert response
	user := authResp.User
	return &pb.LoginResponse{
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

// ========== OTP Methods ==========

// convertOTPType converts proto OTPType to models.OTPType
func convertOTPType(otpType pb.OTPType) models.OTPType {
	switch otpType {
	case pb.OTPType_OTP_TYPE_VERIFICATION:
		return models.OTPTypeVerification
	case pb.OTPType_OTP_TYPE_PASSWORD_RESET:
		return models.OTPTypePasswordReset
	case pb.OTPType_OTP_TYPE_TWO_FA:
		return models.OTPType2FA
	case pb.OTPType_OTP_TYPE_LOGIN:
		return models.OTPTypeLogin
	case pb.OTPType_OTP_TYPE_REGISTRATION:
		return models.OTPTypeRegistration
	default:
		return models.OTPTypeVerification
	}
}

// SendOTP sends a one-time password to email
func (h *AuthHandlerV2) SendOTP(ctx context.Context, req *pb.SendOTPRequest) (*pb.SendOTPResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if h.otpService == nil {
		return nil, status.Error(codes.Unavailable, "OTP service not configured")
	}

	// Parse profile ID if provided
	var profileID *uuid.UUID
	if req.ProfileId != "" {
		id, err := uuid.Parse(req.ProfileId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid profile_id format")
		}
		profileID = &id
	}

	// Create OTP request
	otpReq := &models.SendOTPRequest{
		Email:     &req.Email,
		Type:      convertOTPType(req.OtpType),
		ProfileID: profileID,
	}

	// Send OTP
	err := h.otpService.SendOTP(ctx, otpReq)
	if err != nil {
		h.logger.Error("Failed to send OTP via gRPC", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})

		if appErr, ok := err.(*models.AppError); ok {
			return &pb.SendOTPResponse{
				Success:      false,
				ErrorMessage: appErr.Message,
			}, nil
		}
		return &pb.SendOTPResponse{
			Success:      false,
			ErrorMessage: "failed to send OTP",
		}, nil
	}

	return &pb.SendOTPResponse{
		Success:   true,
		Message:   "OTP sent successfully",
		ExpiresIn: 600, // 10 minutes
	}, nil
}

// VerifyOTP verifies a one-time password
func (h *AuthHandlerV2) VerifyOTP(ctx context.Context, req *pb.VerifyOTPRequest) (*pb.VerifyOTPResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}

	if h.otpService == nil {
		return nil, status.Error(codes.Unavailable, "OTP service not configured")
	}

	// Create verify request
	verifyReq := &models.VerifyOTPRequest{
		Email: &req.Email,
		Code:  req.Code,
		Type:  convertOTPType(req.OtpType),
	}

	// Verify OTP
	resp, err := h.otpService.VerifyOTP(ctx, verifyReq)
	if err != nil {
		h.logger.Debug("OTP verification failed via gRPC", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})

		if appErr, ok := err.(*models.AppError); ok {
			return &pb.VerifyOTPResponse{
				Valid:        false,
				ErrorMessage: appErr.Message,
			}, nil
		}
		return &pb.VerifyOTPResponse{
			Valid:        false,
			ErrorMessage: "verification failed",
		}, nil
	}

	return &pb.VerifyOTPResponse{
		Valid: resp.Valid,
	}, nil
}

// LoginWithOTP initiates passwordless login by sending OTP to email
func (h *AuthHandlerV2) LoginWithOTP(ctx context.Context, req *pb.LoginWithOTPRequest) (*pb.LoginWithOTPResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if h.otpService == nil {
		return nil, status.Error(codes.Unavailable, "OTP service not configured")
	}

	// Check if user exists
	user, err := h.userRepo.GetByEmail(ctx, req.Email, nil)
	if err != nil || user == nil {
		return &pb.LoginWithOTPResponse{
			Success:      false,
			ErrorMessage: "user not found",
		}, nil
	}

	// Parse profile ID if provided
	var profileID *uuid.UUID
	if req.ProfileId != "" {
		id, err := uuid.Parse(req.ProfileId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid profile_id format")
		}
		profileID = &id
	}

	// Send login OTP
	otpReq := &models.SendOTPRequest{
		Email:     &req.Email,
		Type:      models.OTPTypeLogin,
		ProfileID: profileID,
	}

	err = h.otpService.SendOTP(ctx, otpReq)
	if err != nil {
		h.logger.Error("Failed to send login OTP via gRPC", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})

		if appErr, ok := err.(*models.AppError); ok {
			return &pb.LoginWithOTPResponse{
				Success:      false,
				ErrorMessage: appErr.Message,
			}, nil
		}
		return &pb.LoginWithOTPResponse{
			Success:      false,
			ErrorMessage: "failed to send OTP",
		}, nil
	}

	return &pb.LoginWithOTPResponse{
		Success:   true,
		Message:   "OTP sent to your email",
		ExpiresIn: 600,
	}, nil
}

// VerifyLoginOTP completes passwordless login by verifying OTP
func (h *AuthHandlerV2) VerifyLoginOTP(ctx context.Context, req *pb.VerifyLoginOTPRequest) (*pb.VerifyLoginOTPResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}

	if h.otpService == nil {
		return nil, status.Error(codes.Unavailable, "OTP service not configured")
	}

	// Verify OTP
	verifyReq := &models.VerifyOTPRequest{
		Email: &req.Email,
		Code:  req.Code,
		Type:  models.OTPTypeLogin,
	}

	verifyResp, err := h.otpService.VerifyOTP(ctx, verifyReq)
	if err != nil || !verifyResp.Valid {
		errMsg := "invalid or expired OTP"
		if appErr, ok := err.(*models.AppError); ok {
			errMsg = appErr.Message
		}
		return &pb.VerifyLoginOTPResponse{
			ErrorMessage: errMsg,
		}, nil
	}

	// Get user with roles and generate tokens
	userWithRoles, err := h.userRepo.GetByEmail(ctx, req.Email, nil, service.UserGetWithRoles())
	if err != nil || userWithRoles == nil {
		return &pb.VerifyLoginOTPResponse{
			ErrorMessage: "user not found",
		}, nil
	}

	// Generate tokens
	accessToken, err := h.jwtService.GenerateAccessToken(userWithRoles)
	if err != nil {
		h.logger.Error("Failed to generate access token", map[string]interface{}{
			"user_id": userWithRoles.ID.String(),
			"error":   err.Error(),
		})
		return &pb.VerifyLoginOTPResponse{
			ErrorMessage: "failed to generate tokens",
		}, nil
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(userWithRoles)
	if err != nil {
		h.logger.Error("Failed to generate refresh token", map[string]interface{}{
			"user_id": userWithRoles.ID.String(),
			"error":   err.Error(),
		})
		return &pb.VerifyLoginOTPResponse{
			ErrorMessage: "failed to generate tokens",
		}, nil
	}

	return &pb.VerifyLoginOTPResponse{
		User: &pb.User{
			Id:                userWithRoles.ID.String(),
			Email:             userWithRoles.Email,
			Username:          userWithRoles.Username,
			FullName:          userWithRoles.FullName,
			ProfilePictureUrl: userWithRoles.ProfilePictureURL,
			Roles:             extractRoleNames(userWithRoles.Roles),
			EmailVerified:     userWithRoles.EmailVerified,
			IsActive:          userWithRoles.IsActive,
			CreatedAt:         userWithRoles.CreatedAt.Unix(),
			UpdatedAt:         userWithRoles.UpdatedAt.Unix(),
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
	}, nil
}

// RegisterWithOTP initiates OTP-based registration by sending verification code
func (h *AuthHandlerV2) RegisterWithOTP(ctx context.Context, req *pb.RegisterWithOTPRequest) (*pb.RegisterWithOTPResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if h.otpService == nil {
		return nil, status.Error(codes.Unavailable, "OTP service not configured")
	}

	// Check if user already exists
	existingUser, _ := h.userRepo.GetByEmail(ctx, req.Email, nil)
	if existingUser != nil {
		return &pb.RegisterWithOTPResponse{
			Success:      false,
			ErrorMessage: "email already registered",
		}, nil
	}

	// Parse profile ID if provided
	var profileID *uuid.UUID
	if req.ProfileId != "" {
		id, err := uuid.Parse(req.ProfileId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid profile_id format")
		}
		profileID = &id
	}

	// Store registration data in Redis for later use
	if h.redis != nil {
		regData := &models.PendingRegistration{
			Email:    req.Email,
			Username: req.Username,
			FullName: req.FullName,
		}
		_ = h.redis.StorePendingRegistration(ctx, "otp_reg:"+req.Email, regData, 600*time.Second)
	}

	// Send registration OTP
	otpReq := &models.SendOTPRequest{
		Email:     &req.Email,
		Type:      models.OTPTypeRegistration,
		ProfileID: profileID,
	}

	err := h.otpService.SendOTP(ctx, otpReq)
	if err != nil {
		h.logger.Error("Failed to send registration OTP via gRPC", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})

		if appErr, ok := err.(*models.AppError); ok {
			return &pb.RegisterWithOTPResponse{
				Success:      false,
				ErrorMessage: appErr.Message,
			}, nil
		}
		return &pb.RegisterWithOTPResponse{
			Success:      false,
			ErrorMessage: "failed to send OTP",
		}, nil
	}

	return &pb.RegisterWithOTPResponse{
		Success:   true,
		Message:   "Verification code sent to your email",
		ExpiresIn: 600,
	}, nil
}

// VerifyRegistrationOTP completes OTP-based registration
func (h *AuthHandlerV2) VerifyRegistrationOTP(ctx context.Context, req *pb.VerifyRegistrationOTPRequest) (*pb.VerifyRegistrationOTPResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}

	if h.otpService == nil {
		return nil, status.Error(codes.Unavailable, "OTP service not configured")
	}

	// Verify OTP
	verifyReq := &models.VerifyOTPRequest{
		Email: &req.Email,
		Code:  req.Code,
		Type:  models.OTPTypeRegistration,
	}

	verifyResp, err := h.otpService.VerifyOTP(ctx, verifyReq)
	if err != nil || !verifyResp.Valid {
		errMsg := "invalid or expired OTP"
		if appErr, ok := err.(*models.AppError); ok {
			errMsg = appErr.Message
		}
		return &pb.VerifyRegistrationOTPResponse{
			ErrorMessage: errMsg,
		}, nil
	}

	// Get stored registration data
	username := req.Username
	fullName := req.FullName
	if h.redis != nil {
		regData, err := h.redis.GetPendingRegistration(ctx, "otp_reg:"+req.Email)
		if err == nil && regData != nil {
			if username == "" && regData.Username != "" {
				username = regData.Username
			}
			if fullName == "" && regData.FullName != "" {
				fullName = regData.FullName
			}
		}
		// Cleanup registration data
		_ = h.redis.DeletePendingRegistration(ctx, "otp_reg:"+req.Email)
	}

	// Generate username if not provided
	if username == "" {
		username = strings.Split(req.Email, "@")[0]
	}

	// Create user via AuthService
	createReq := &models.CreateUserRequest{
		Email:       req.Email,
		Username:    username,
		Password:    req.Password, // Can be empty for passwordless
		FullName:    fullName,
		AccountType: "human",
	}

	// Extract client info from gRPC metadata
	clientInfo := extractClientInfo(ctx)

	deviceInfo := models.DeviceInfo{
		DeviceType: "grpc_client",
		OS:         parseOSFromUserAgent(clientInfo.UserAgent),
		Browser:    clientInfo.UserAgent,
	}

	authResp, err := h.authService.SignUp(ctx, createReq, clientInfo.IP, clientInfo.UserAgent, deviceInfo, GetApplicationUUIDFromGRPCContext(ctx))
	if err != nil {
		h.logger.Error("Failed to create user via gRPC OTP registration", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})

		if appErr, ok := err.(*models.AppError); ok {
			return &pb.VerifyRegistrationOTPResponse{
				ErrorMessage: appErr.Message,
			}, nil
		}
		return &pb.VerifyRegistrationOTPResponse{
			ErrorMessage: "failed to create user",
		}, nil
	}

	// Mark email as verified since OTP was verified
	_ = h.userRepo.MarkEmailVerified(ctx, authResp.User.ID)

	user := authResp.User
	return &pb.VerifyRegistrationOTPResponse{
		User: &pb.User{
			Id:                user.ID.String(),
			Email:             user.Email,
			Username:          user.Username,
			FullName:          user.FullName,
			ProfilePictureUrl: user.ProfilePictureURL,
			Roles:             extractRoleNames(user.Roles),
			EmailVerified:     true,
			IsActive:          user.IsActive,
			CreatedAt:         user.CreatedAt.Unix(),
			UpdatedAt:         user.UpdatedAt.Unix(),
		},
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresIn:    authResp.ExpiresIn,
	}, nil
}

func (h *AuthHandlerV2) GetUserApplicationProfile(ctx context.Context, req *pb.GetUserAppProfileRequest) (*pb.UserAppProfileResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	resolvedAppID := ResolveApplicationID(ctx, req.ApplicationId)
	if resolvedAppID == "" {
		return nil, status.Error(codes.InvalidArgument, "application_id is required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}
	appID, err := uuid.Parse(resolvedAppID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid application_id format")
	}
	profile, err := h.appService.GetOrCreateUserProfile(ctx, userID, appID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user profile: %v", err)
	}
	return toUserAppProfileResponse(profile), nil
}

// GetUserTelegramBots returns user's Telegram bot access for an application
func (h *AuthHandlerV2) GetUserTelegramBots(ctx context.Context, req *pb.GetUserTelegramBotsRequest) (*pb.UserTelegramBotsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	resolvedAppID := ResolveApplicationID(ctx, req.ApplicationId)
	if resolvedAppID == "" {
		return nil, status.Error(codes.InvalidArgument, "application_id is required")
	}
	_ = resolvedAppID // TODO: use when implemented

	return nil, status.Errorf(codes.Unimplemented, "GetUserTelegramBots not implemented - handler requires UserTelegramRepository dependency")
}

func (h *AuthHandlerV2) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UserAppProfileResponse, error) {
	if req.UserId == "" || req.ApplicationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and application_id are required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}
	appID, err := uuid.Parse(req.ApplicationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid application_id format")
	}
	updateReq := &models.UpdateUserAppProfileRequest{
		AppRoles: req.AppRoles,
	}
	if req.DisplayName != "" {
		updateReq.DisplayName = &req.DisplayName
	}
	if req.AvatarUrl != "" {
		updateReq.AvatarURL = &req.AvatarUrl
	}
	if req.Nickname != "" {
		updateReq.Nickname = &req.Nickname
	}
	profile, err := h.appService.UpdateUserProfile(ctx, userID, appID, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user profile: %v", err)
	}
	return toUserAppProfileResponse(profile), nil
}

func (h *AuthHandlerV2) CreateUserProfile(ctx context.Context, req *pb.CreateUserProfileRequest) (*pb.UserAppProfileResponse, error) {
	if req.UserId == "" || req.ApplicationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and application_id are required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}
	appID, err := uuid.Parse(req.ApplicationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid application_id format")
	}
	profile, err := h.appService.GetOrCreateUserProfile(ctx, userID, appID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user profile: %v", err)
	}
	return toUserAppProfileResponse(profile), nil
}

func (h *AuthHandlerV2) DeleteUserProfile(ctx context.Context, req *pb.DeleteUserProfileRequest) (*pb.GenericResponse, error) {
	if req.UserId == "" || req.ApplicationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and application_id are required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}
	appID, err := uuid.Parse(req.ApplicationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid application_id format")
	}
	if err := h.appService.DeleteUserProfile(ctx, userID, appID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user profile: %v", err)
	}
	return &pb.GenericResponse{Success: true, Message: "User profile deleted"}, nil
}

func (h *AuthHandlerV2) BanUser(ctx context.Context, req *pb.BanUserRequest) (*pb.GenericResponse, error) {
	if req.UserId == "" || req.ApplicationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and application_id are required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}
	appID, err := uuid.Parse(req.ApplicationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid application_id format")
	}
	var bannedBy uuid.UUID
	if req.BannedBy != "" {
		bannedBy, err = uuid.Parse(req.BannedBy)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid banned_by format")
		}
	}
	if err := h.appService.BanUser(ctx, userID, appID, bannedBy, req.Reason); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to ban user: %v", err)
	}
	return &pb.GenericResponse{Success: true, Message: "User banned"}, nil
}

func (h *AuthHandlerV2) UnbanUser(ctx context.Context, req *pb.UnbanUserRequest) (*pb.GenericResponse, error) {
	if req.UserId == "" || req.ApplicationId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and application_id are required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}
	appID, err := uuid.Parse(req.ApplicationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid application_id format")
	}
	if err := h.appService.UnbanUser(ctx, userID, appID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unban user: %v", err)
	}
	return &pb.GenericResponse{Success: true, Message: "User unbanned"}, nil
}

func (h *AuthHandlerV2) ListApplicationUsers(ctx context.Context, req *pb.ListApplicationUsersRequest) (*pb.ListApplicationUsersResponse, error) {
	if req.ApplicationId == "" {
		return nil, status.Error(codes.InvalidArgument, "application_id is required")
	}
	appID, err := uuid.Parse(req.ApplicationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid application_id format")
	}
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	result, err := h.appService.ListApplicationUsers(ctx, appID, page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list application users: %v", err)
	}
	var profiles []*pb.UserAppProfileResponse
	for _, p := range result.Profiles {
		profiles = append(profiles, toUserAppProfileResponse(&p))
	}
	return &pb.ListApplicationUsersResponse{
		Profiles:   profiles,
		Total:      int32(result.Total),
		Page:       int32(result.Page),
		PageSize:   int32(result.PageSize),
		TotalPages: int32(result.TotalPages),
	}, nil
}

func toUserAppProfileResponse(p *models.UserApplicationProfile) *pb.UserAppProfileResponse {
	resp := &pb.UserAppProfileResponse{
		UserId:        p.UserID.String(),
		ApplicationId: p.ApplicationID.String(),
		AppRoles:      p.AppRoles,
		IsActive:      p.IsActive,
		IsBanned:      p.IsBanned,
		CreatedAt:     p.CreatedAt.Unix(),
		UpdatedAt:     p.UpdatedAt.Unix(),
	}
	if p.DisplayName != nil {
		resp.DisplayName = *p.DisplayName
	}
	if p.AvatarURL != nil {
		resp.AvatarUrl = *p.AvatarURL
	}
	if p.Nickname != nil {
		resp.Nickname = *p.Nickname
	}
	if p.BanReason != nil {
		resp.BanReason = *p.BanReason
	}
	if p.LastAccessAt != nil {
		resp.LastAccessAt = p.LastAccessAt.Unix()
	}
	if len(p.Metadata) > 0 {
		metadataMap, err := p.GetMetadataMap()
		if err == nil {
			resp.Metadata = make(map[string]string)
			for k, v := range metadataMap {
				resp.Metadata[k] = fmt.Sprintf("%v", v)
			}
		}
	}
	return resp
}

// ========== Sync & Config Methods ==========

// SyncUsers returns users updated after a given timestamp for shadow table sync
func (h *AuthHandlerV2) SyncUsers(ctx context.Context, req *pb.SyncUsersRequest) (*pb.SyncUsersResponse, error) {
	if req.UpdatedAfter == "" {
		return &pb.SyncUsersResponse{
			ErrorMessage: "updated_after is required",
		}, nil
	}

	updatedAfter, err := time.Parse(time.RFC3339, req.UpdatedAfter)
	if err != nil {
		return &pb.SyncUsersResponse{
			ErrorMessage: "invalid updated_after format, expected RFC3339",
		}, nil
	}

	var appID *uuid.UUID
	if resolvedAppID := ResolveApplicationID(ctx, req.ApplicationId); resolvedAppID != "" {
		parsed, err := uuid.Parse(resolvedAppID)
		if err != nil {
			return &pb.SyncUsersResponse{
				ErrorMessage: "invalid application_id format",
			}, nil
		}
		appID = &parsed
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 100
	}
	offset := int(req.Offset)

	result, err := h.adminService.SyncUsers(ctx, updatedAfter, appID, limit, offset)
	if err != nil {
		return &pb.SyncUsersResponse{
			ErrorMessage: err.Error(),
		}, nil
	}

	users := make([]*pb.SyncUser, len(result.Users))
	for i, u := range result.Users {
		syncUser := &pb.SyncUser{
			Id:            u.ID.String(),
			Email:         u.Email,
			Username:      u.Username,
			FullName:      u.FullName,
			IsActive:      u.IsActive,
			EmailVerified: u.EmailVerified,
			UpdatedAt:     u.UpdatedAt.Format(time.RFC3339),
		}
		if u.AppProfile != nil {
			syncUser.AppProfile = &pb.SyncUserAppProfile{
				DisplayName: u.AppProfile.DisplayName,
				AvatarUrl:   u.AppProfile.AvatarURL,
				AppRoles:    u.AppProfile.AppRoles,
				IsActive:    u.AppProfile.IsActive,
				IsBanned:    u.AppProfile.IsBanned,
			}
		}
		users[i] = syncUser
	}

	return &pb.SyncUsersResponse{
		Users:         users,
		Total:         int32(result.Total),
		HasMore:       result.HasMore,
		SyncTimestamp: result.SyncTimestamp,
	}, nil
}

// GetApplicationAuthConfig returns auth configuration for a specific application
func (h *AuthHandlerV2) GetApplicationAuthConfig(ctx context.Context, req *pb.GetApplicationAuthConfigRequest) (*pb.GetApplicationAuthConfigResponse, error) {
	resolvedAppID := ResolveApplicationID(ctx, req.ApplicationId)
	if resolvedAppID == "" {
		return &pb.GetApplicationAuthConfigResponse{
			ErrorMessage: "application_id is required",
		}, nil
	}

	appID, err := uuid.Parse(resolvedAppID)
	if err != nil {
		return &pb.GetApplicationAuthConfigResponse{
			ErrorMessage: "invalid application_id format",
		}, nil
	}

	app, err := h.appService.GetByID(ctx, appID)
	if err != nil {
		return &pb.GetApplicationAuthConfigResponse{
			ErrorMessage: "application not found",
		}, nil
	}

	authConfig, err := h.appService.GetAuthConfig(ctx, app)
	if err != nil {
		return &pb.GetApplicationAuthConfigResponse{
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.GetApplicationAuthConfigResponse{
		ApplicationId:      authConfig.ApplicationID.String(),
		Name:               authConfig.Name,
		DisplayName:        authConfig.DisplayName,
		AllowedAuthMethods: authConfig.AllowedAuthMethods,
		OauthProviders:     authConfig.OAuthProviders,
	}, nil
}

// SendEmail sends an email using a specified template
func (h *AuthHandlerV2) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	if req.TemplateType == "" {
		return nil, status.Error(codes.InvalidArgument, "template_type is required")
	}
	if req.ToEmail == "" {
		return nil, status.Error(codes.InvalidArgument, "to_email is required")
	}

	if h.emailProfileService == nil {
		return &pb.SendEmailResponse{
			Success:      false,
			ErrorMessage: "email profile service not configured",
		}, nil
	}

	// Parse optional profile ID
	var profileID *uuid.UUID
	if req.ProfileId != "" {
		id, err := uuid.Parse(req.ProfileId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid profile_id format")
		}
		profileID = &id
	}

	// Parse optional application ID (falls back to app secret context)
	var applicationID *uuid.UUID
	if resolvedAppID := ResolveApplicationID(ctx, req.ApplicationId); resolvedAppID != "" {
		id, err := uuid.Parse(resolvedAppID)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid application_id format")
		}
		applicationID = &id
	}

	// Convert map[string]string to map[string]interface{}
	variables := make(map[string]interface{}, len(req.Variables))
	for k, v := range req.Variables {
		variables[k] = v
	}

	err := h.emailProfileService.SendEmail(ctx, profileID, applicationID, req.ToEmail, req.TemplateType, variables)
	if err != nil {
		h.logger.Error("Failed to send email via gRPC", map[string]interface{}{
			"error":         err.Error(),
			"template_type": req.TemplateType,
			"to_email":      req.ToEmail,
		})
		return &pb.SendEmailResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.SendEmailResponse{
		Success: true,
		Message: "Email sent successfully",
	}, nil
}

// ========== SSO Token Exchange Methods ==========

func (h *AuthHandlerV2) CreateTokenExchange(ctx context.Context, req *pb.CreateTokenExchangeGrpcRequest) (*pb.CreateTokenExchangeGrpcResponse, error) {
	if req.AccessToken == "" {
		return &pb.CreateTokenExchangeGrpcResponse{
			ErrorMessage: "access_token is required",
		}, nil
	}
	if req.TargetApplicationId == "" {
		return &pb.CreateTokenExchangeGrpcResponse{
			ErrorMessage: "target_application_id is required",
		}, nil
	}

	if h.tokenExchangeService == nil {
		return &pb.CreateTokenExchangeGrpcResponse{
			ErrorMessage: "token exchange service not configured",
		}, nil
	}

	clientInfo := extractClientInfo(ctx)
	var sourceAppID *uuid.UUID
	if clientInfo.ApplicationID != "" {
		parsed, err := uuid.Parse(clientInfo.ApplicationID)
		if err == nil {
			sourceAppID = &parsed
		}
	}

	exchangeReq := &models.CreateTokenExchangeRequest{
		AccessToken: req.AccessToken,
		TargetAppID: req.TargetApplicationId,
	}

	resp, err := h.tokenExchangeService.CreateExchange(ctx, exchangeReq, sourceAppID)
	if err != nil {
		h.logger.Error("Failed to create token exchange via gRPC", map[string]interface{}{
			"error": err.Error(),
		})
		if appErr, ok := err.(*models.AppError); ok {
			return &pb.CreateTokenExchangeGrpcResponse{
				ErrorMessage: appErr.Message,
			}, nil
		}
		return &pb.CreateTokenExchangeGrpcResponse{
			ErrorMessage: "failed to create token exchange",
		}, nil
	}

	return &pb.CreateTokenExchangeGrpcResponse{
		ExchangeCode: resp.ExchangeCode,
		ExpiresAt:    resp.ExpiresAt.Format(time.RFC3339),
		RedirectUrl:  resp.RedirectURL,
	}, nil
}

func (h *AuthHandlerV2) RedeemTokenExchange(ctx context.Context, req *pb.RedeemTokenExchangeGrpcRequest) (*pb.RedeemTokenExchangeGrpcResponse, error) {
	if req.ExchangeCode == "" {
		return &pb.RedeemTokenExchangeGrpcResponse{
			ErrorMessage: "exchange_code is required",
		}, nil
	}

	if h.tokenExchangeService == nil {
		return &pb.RedeemTokenExchangeGrpcResponse{
			ErrorMessage: "token exchange service not configured",
		}, nil
	}

	clientInfo := extractClientInfo(ctx)
	var redeemingAppID *uuid.UUID
	if clientInfo.ApplicationID != "" {
		parsed, err := uuid.Parse(clientInfo.ApplicationID)
		if err == nil {
			redeemingAppID = &parsed
		}
	}

	redeemReq := &models.RedeemTokenExchangeRequest{
		ExchangeCode: req.ExchangeCode,
	}

	resp, err := h.tokenExchangeService.RedeemExchange(ctx, redeemReq, redeemingAppID)
	if err != nil {
		h.logger.Error("Failed to redeem token exchange via gRPC", map[string]interface{}{
			"error": err.Error(),
		})
		if appErr, ok := err.(*models.AppError); ok {
			return &pb.RedeemTokenExchangeGrpcResponse{
				ErrorMessage: appErr.Message,
			}, nil
		}
		return &pb.RedeemTokenExchangeGrpcResponse{
			ErrorMessage: "failed to redeem token exchange",
		}, nil
	}

	userEmail := ""
	if resp.User != nil {
		userEmail = resp.User.Email
	}
	userID := ""
	if resp.User != nil {
		userID = resp.User.ID.String()
	}

	return &pb.RedeemTokenExchangeGrpcResponse{
		AccessToken:   resp.AccessToken,
		RefreshToken:  resp.RefreshToken,
		UserId:        userID,
		Email:         userEmail,
		ApplicationId: resp.ApplicationID,
	}, nil
}

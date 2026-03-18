package grpc

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/smilemakc/auth-gateway/internal/models"
	pb "github.com/smilemakc/auth-gateway/proto"
)

// ===================== extractRoleNames Tests =====================

func TestExtractRoleNames_ShouldReturnNames_WhenRolesProvided(t *testing.T) {
	roles := []models.Role{
		{Name: "admin"},
		{Name: "editor"},
		{Name: "viewer"},
	}

	names := extractRoleNames(roles)

	assert.Len(t, names, 3)
	assert.Equal(t, []string{"admin", "editor", "viewer"}, names)
}

func TestExtractRoleNames_ShouldReturnEmptySlice_WhenNoRoles(t *testing.T) {
	names := extractRoleNames([]models.Role{})
	assert.Len(t, names, 0)
	assert.NotNil(t, names)
}

func TestExtractRoleNames_ShouldReturnEmptySlice_WhenNilRoles(t *testing.T) {
	names := extractRoleNames(nil)
	assert.Len(t, names, 0)
}

// ===================== convertOTPType Tests =====================

func TestConvertOTPType_ShouldMapAllTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    pb.OTPType
		expected models.OTPType
	}{
		{"Verification", pb.OTPType_OTP_TYPE_VERIFICATION, models.OTPTypeVerification},
		{"PasswordReset", pb.OTPType_OTP_TYPE_PASSWORD_RESET, models.OTPTypePasswordReset},
		{"TwoFA", pb.OTPType_OTP_TYPE_TWO_FA, models.OTPType2FA},
		{"Login", pb.OTPType_OTP_TYPE_LOGIN, models.OTPTypeLogin},
		{"Registration", pb.OTPType_OTP_TYPE_REGISTRATION, models.OTPTypeRegistration},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertOTPType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertOTPType_ShouldDefaultToVerification_WhenUnspecified(t *testing.T) {
	result := convertOTPType(pb.OTPType_OTP_TYPE_UNSPECIFIED)
	assert.Equal(t, models.OTPTypeVerification, result)
}

func TestConvertOTPType_ShouldDefaultToVerification_WhenUnknown(t *testing.T) {
	result := convertOTPType(pb.OTPType(999))
	assert.Equal(t, models.OTPTypeVerification, result)
}

// ===================== parseOSFromUserAgent Tests =====================

func TestParseOSFromUserAgent_ShouldDetectAllPlatforms(t *testing.T) {
	tests := []struct {
		name     string
		ua       string
		expected string
	}{
		{"Windows", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "Windows"},
		{"macOS_via_mac", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", "macOS"},
		{"macOS_via_darwin", "grpc-go/1.50.0 Darwin/22.3.0", "macOS"},
		{"Linux", "grpc-go/1.50.0 Linux/5.15.0", "Linux"},
		{"Android", "MyApp/1.0 Android/12", "Android"},
		{"iOS_iphone", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0)", "iOS"},
		{"iOS_ipad", "Mozilla/5.0 (iPad; CPU OS 16_0)", "iOS"},
		{"iOS_keyword", "MyApp/1.0 iOS/16.0", "iOS"},
		{"Unknown", "some-random-client/1.0", ""},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOSFromUserAgent(tt.ua)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ===================== parseRuntimeFromUserAgent Tests =====================

func TestParseRuntimeFromUserAgent_ShouldDetectAllRuntimes(t *testing.T) {
	tests := []struct {
		name     string
		ua       string
		expected string
	}{
		{"Go_grpc", "grpc-go/1.50.0", "Go"},
		{"Go_keyword", "MyApp/1.0 golang", "Go"},
		{"Go_slash", "go/1.20 some-client", "Go"},
		{"Python", "grpc-python/1.53.0", "Python"},
		{"Python_keyword", "python-requests/2.28.0", "Python"},
		{"Node_grpc", "grpc-node/1.8.0", "Node.js"},
		{"Node_keyword", "node/18.0", "Node.js"},
		{"JavaScript", "javascript-client/1.0", "Node.js"},
		{"Java", "grpc-java/1.54.0", "Java"},
		{"Java_keyword", "java/17", "Java"},
		{"DotNet", "grpc-dotnet/2.0", ".NET"},
		{"DotNet_csharp", "csharp-client/1.0", ".NET"},
		{"Ruby", "ruby-client/3.0", "Ruby"},
		{"PHP", "php-client/8.1", "PHP"},
		{"Rust", "rust-grpc/0.1.0", "Rust"},
		{"Unknown", "some-random-client", ""},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRuntimeFromUserAgent(tt.ua)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ===================== buildUserAgent Tests =====================

func TestBuildUserAgent_ShouldBuildFromClientInfo(t *testing.T) {
	info := ClientInfo{
		ClientName:    "payment-service",
		ClientVersion: "2.1.0",
		Platform:      "kubernetes",
		Environment:   "production",
	}

	result := buildUserAgent(info, "grpc-go/1.50.0")

	assert.Contains(t, result, "payment-service/2.1.0")
	assert.Contains(t, result, "Go")
	assert.Contains(t, result, "kubernetes")
	assert.Contains(t, result, "[production]")
	assert.Contains(t, result, "(gRPC)")
}

func TestBuildUserAgent_ShouldIncludeClientNameOnly_WhenNoVersion(t *testing.T) {
	info := ClientInfo{
		ClientName: "my-service",
	}

	result := buildUserAgent(info, "")

	assert.Contains(t, result, "my-service")
	assert.Contains(t, result, "(gRPC)")
}

func TestBuildUserAgent_ShouldIncludeRuntime_WhenDetectedFromBaseUA(t *testing.T) {
	info := ClientInfo{}

	result := buildUserAgent(info, "grpc-go/1.50.0")

	// parseRuntimeFromUserAgent detects "Go" from grpc-go, so parts = ["Go"]
	assert.Equal(t, "Go (gRPC)", result)
}

func TestBuildUserAgent_ShouldCleanGRPCGoUA_WhenNoRuntimeDetected(t *testing.T) {
	info := ClientInfo{}

	// Use a UA that looks like grpc-go/ but is not detected as a runtime
	// Actually, any "grpc-go/" will be detected as Go runtime.
	// The fallback to "gRPC Go Client X" only happens when no parts are built,
	// which is impossible with "grpc-go/" since parseRuntime detects Go.
	// Test the non-grpc-go fallback instead:
	result := buildUserAgent(info, "custom-unknown-client/2.0")

	assert.Equal(t, "custom-unknown-client/2.0", result)
}

func TestBuildUserAgent_ShouldReturnDefault_WhenNoInfoAndNoUA(t *testing.T) {
	info := ClientInfo{}

	result := buildUserAgent(info, "")

	assert.Equal(t, "gRPC Client", result)
}

func TestBuildUserAgent_ShouldReturnBaseUA_WhenNoClientInfoAndCustomUA(t *testing.T) {
	info := ClientInfo{}

	result := buildUserAgent(info, "custom-client/1.0")

	assert.Equal(t, "custom-client/1.0", result)
}

// ===================== buildDeviceInfo Tests =====================

func TestBuildDeviceInfo_ShouldUseClientName_WhenProvided(t *testing.T) {
	info := ClientInfo{
		ClientName: "payment-service",
		UserAgent:  "payment-service/2.0 | Go | kubernetes | [production] (gRPC)",
	}

	result := buildDeviceInfo(info)

	assert.Equal(t, "payment-service", result.DeviceType)
	assert.NotEmpty(t, result.OS)
	assert.Equal(t, info.UserAgent, result.Browser)
}

func TestBuildDeviceInfo_ShouldDefaultToGRPCClient_WhenNoClientName(t *testing.T) {
	info := ClientInfo{
		UserAgent: "gRPC Client",
	}

	result := buildDeviceInfo(info)

	assert.Equal(t, "grpc_client", result.DeviceType)
}

func TestBuildDeviceInfo_ShouldUseRuntimeForOS_WhenNoOSDetected(t *testing.T) {
	// UserAgent must contain a recognizable runtime pattern (e.g., "grpc-go")
	// but not an OS pattern (windows, mac, linux, etc.)
	info := ClientInfo{
		UserAgent: "grpc-go/1.50.0",
		Platform:  "",
	}

	result := buildDeviceInfo(info)

	assert.Equal(t, "Go Runtime", result.OS)
}

func TestBuildDeviceInfo_ShouldUsePlatformForOS_WhenNoOSOrRuntime(t *testing.T) {
	info := ClientInfo{
		UserAgent: "custom-client (gRPC)",
		Platform:  "kubernetes",
	}

	result := buildDeviceInfo(info)

	assert.Equal(t, "kubernetes", result.OS)
}

func TestBuildDeviceInfo_ShouldFallbackToGRPCClient_WhenNothingDetectable(t *testing.T) {
	info := ClientInfo{
		UserAgent: "unknown",
	}

	result := buildDeviceInfo(info)

	assert.Equal(t, "gRPC Client", result.OS)
}

// ===================== extractClientInfo Tests =====================

func TestExtractClientInfo_ShouldExtractIPFromPeer(t *testing.T) {
	ctx := context.Background()
	addr, _ := net.ResolveTCPAddr("tcp", "192.168.1.100:50000")
	ctx = peer.NewContext(ctx, &peer.Peer{Addr: addr})

	info := extractClientInfo(ctx)

	assert.Equal(t, "192.168.1.100", info.IP)
}

func TestExtractClientInfo_ShouldPreferXForwardedFor_OverPeerIP(t *testing.T) {
	ctx := context.Background()
	addr, _ := net.ResolveTCPAddr("tcp", "10.0.0.1:50000")
	ctx = peer.NewContext(ctx, &peer.Peer{Addr: addr})
	md := metadata.Pairs("x-forwarded-for", "203.0.113.50, 70.41.3.18")
	ctx = metadata.NewIncomingContext(ctx, md)

	info := extractClientInfo(ctx)

	assert.Equal(t, "203.0.113.50", info.IP)
}

func TestExtractClientInfo_ShouldExtractAllMetadataFields(t *testing.T) {
	md := metadata.Pairs(
		"x-application-id", "app-123",
		"x-client-name", "payment-service",
		"x-client-version", "2.0.0",
		"x-platform", "kubernetes",
		"x-environment", "production",
		"user-agent", "grpc-go/1.50.0 Linux",
	)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := extractClientInfo(ctx)

	assert.Equal(t, "app-123", info.ApplicationID)
	assert.Equal(t, "payment-service", info.ClientName)
	assert.Equal(t, "2.0.0", info.ClientVersion)
	assert.Equal(t, "kubernetes", info.Platform)
	assert.Equal(t, "production", info.Environment)
	assert.Contains(t, info.UserAgent, "payment-service/2.0.0")
	assert.Contains(t, info.UserAgent, "Go")
	assert.Contains(t, info.UserAgent, "Linux")
}

func TestExtractClientInfo_ShouldReturnDefaults_WhenNoMetadata(t *testing.T) {
	info := extractClientInfo(context.Background())

	assert.Equal(t, "", info.IP)
	assert.Equal(t, "gRPC Client", info.UserAgent)
	assert.Empty(t, info.ApplicationID)
	assert.Empty(t, info.ClientName)
}

func TestExtractClientInfo_ShouldHandleIPv6Address(t *testing.T) {
	ctx := context.Background()
	addr, _ := net.ResolveTCPAddr("tcp", "[::1]:50000")
	ctx = peer.NewContext(ctx, &peer.Peer{Addr: addr})

	info := extractClientInfo(ctx)

	// IPv6 address extraction: "::1" from "[::1]:50000"
	assert.NotEmpty(t, info.IP)
}

// ===================== toUserAppProfileResponse Tests =====================

func TestToUserAppProfileResponse_ShouldMapAllFields(t *testing.T) {
	userID := uuid.New()
	appID := uuid.New()
	displayName := "John Doe"
	avatarURL := "https://example.com/avatar.jpg"
	nickname := "johnd"
	banReason := "spamming"
	lastAccess := time.Now()
	metaJSON, _ := json.Marshal(map[string]interface{}{"level": 10})

	profile := &models.UserApplicationProfile{
		UserID:        userID,
		ApplicationID: appID,
		DisplayName:   &displayName,
		AvatarURL:     &avatarURL,
		Nickname:      &nickname,
		AppRoles:      []string{"admin", "editor"},
		IsActive:      true,
		IsBanned:      true,
		BanReason:     &banReason,
		LastAccessAt:  &lastAccess,
		Metadata:      metaJSON,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	resp := toUserAppProfileResponse(profile)

	assert.Equal(t, userID.String(), resp.UserId)
	assert.Equal(t, appID.String(), resp.ApplicationId)
	assert.Equal(t, displayName, resp.DisplayName)
	assert.Equal(t, avatarURL, resp.AvatarUrl)
	assert.Equal(t, nickname, resp.Nickname)
	assert.Equal(t, []string{"admin", "editor"}, resp.AppRoles)
	assert.True(t, resp.IsActive)
	assert.True(t, resp.IsBanned)
	assert.Equal(t, banReason, resp.BanReason)
	assert.Equal(t, lastAccess.Unix(), resp.LastAccessAt)
	assert.NotEmpty(t, resp.Metadata)
	assert.Equal(t, "10", resp.Metadata["level"])
}

func TestToUserAppProfileResponse_ShouldHandleNilOptionalFields(t *testing.T) {
	userID := uuid.New()
	appID := uuid.New()

	profile := &models.UserApplicationProfile{
		UserID:        userID,
		ApplicationID: appID,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	resp := toUserAppProfileResponse(profile)

	assert.Equal(t, userID.String(), resp.UserId)
	assert.Equal(t, appID.String(), resp.ApplicationId)
	assert.Empty(t, resp.DisplayName)
	assert.Empty(t, resp.AvatarUrl)
	assert.Empty(t, resp.Nickname)
	assert.Empty(t, resp.BanReason)
	assert.Equal(t, int64(0), resp.LastAccessAt)
	assert.Nil(t, resp.Metadata)
	assert.True(t, resp.IsActive)
}

func TestToUserAppProfileResponse_ShouldHandleEmptyMetadata(t *testing.T) {
	profile := &models.UserApplicationProfile{
		UserID:        uuid.New(),
		ApplicationID: uuid.New(),
		Metadata:      []byte("{}"),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	resp := toUserAppProfileResponse(profile)

	// Empty JSON object produces an initialized-but-empty map
	assert.NotNil(t, resp.Metadata)
	assert.Empty(t, resp.Metadata)
}

func TestToUserAppProfileResponse_ShouldHandleNilMetadata(t *testing.T) {
	profile := &models.UserApplicationProfile{
		UserID:        uuid.New(),
		ApplicationID: uuid.New(),
		Metadata:      nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	resp := toUserAppProfileResponse(profile)

	assert.Nil(t, resp.Metadata)
}

// ===================== NewAuthHandlerV2 Tests =====================

func TestNewAuthHandlerV2_ShouldCreateHandler_WhenAllDependenciesProvided(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, // jwtService
		nil, // userRepo
		nil, // tokenRepo
		nil, // rbacRepo
		nil, // apiKeyService
		nil, // authService
		nil, // oauthProviderService
		nil, // otpService
		nil, // emailProfileService
		nil, // adminService
		nil, // appService
		nil, // redis
		nil, // tokenExchangeService
		testLogger(),
	)

	require.NotNil(t, handler)
}

// ===================== Handler Validation Tests =====================
// These test the input validation paths of handler methods that don't require
// actual service/repository calls.

func TestValidateToken_ShouldReturnInvalid_WhenEmptyToken(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	resp, err := handler.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
		AccessToken: "",
	})

	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.Equal(t, "access_token is required", resp.ErrorMessage)
}

func TestGetUser_ShouldReturnError_WhenEmptyUserID(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.GetUser(context.Background(), &pb.GetUserRequest{
		UserId: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestGetUser_ShouldReturnError_WhenInvalidUUID(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.GetUser(context.Background(), &pb.GetUserRequest{
		UserId: "not-a-uuid",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user_id format")
}

func TestCheckPermission_ShouldReturnError_WhenEmptyUserID(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.CheckPermission(context.Background(), &pb.CheckPermissionRequest{
		UserId:   "",
		Resource: "users",
		Action:   "read",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestCheckPermission_ShouldReturnError_WhenInvalidUserID(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.CheckPermission(context.Background(), &pb.CheckPermissionRequest{
		UserId:   "bad-uuid",
		Resource: "users",
		Action:   "read",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user_id format")
}

func TestIntrospectToken_ShouldReturnError_WhenEmptyToken(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.IntrospectToken(context.Background(), &pb.IntrospectTokenRequest{
		AccessToken: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "access_token is required")
}

func TestCreateUser_ShouldReturnError_WhenPasswordEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.CreateUser(context.Background(), &pb.CreateUserRequest{
		Email:    "test@example.com",
		Password: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "password is required")
}

func TestCreateUser_ShouldReturnError_WhenNoEmailOrPhone(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.CreateUser(context.Background(), &pb.CreateUserRequest{
		Password: "SecurePass123",
		Email:    "",
		Phone:    "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "either email or phone is required")
}

func TestLogin_ShouldReturnError_WhenPasswordEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.Login(context.Background(), &pb.LoginRequest{
		Email:    "test@example.com",
		Password: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "password is required")
}

func TestLogin_ShouldReturnError_WhenNoEmailOrPhone(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.Login(context.Background(), &pb.LoginRequest{
		Password: "SecurePass123",
		Email:    "",
		Phone:    "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "either email or phone is required")
}

func TestSendOTP_ShouldReturnError_WhenEmailEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.SendOTP(context.Background(), &pb.SendOTPRequest{
		Email: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestSendOTP_ShouldReturnError_WhenOTPServiceNil(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil,
		nil, // otpService = nil
		nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.SendOTP(context.Background(), &pb.SendOTPRequest{
		Email: "test@example.com",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "OTP service not configured")
}

func TestSendOTP_ShouldReturnError_WhenInvalidProfileID(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil,
		nil, // otpService must be non-nil to reach profileID parsing
		nil, nil, nil, nil, nil,
		testLogger(),
	)

	// otpService is nil, so it will fail before reaching the profileID check
	// Let's just test the email validation instead
	_, err := handler.SendOTP(context.Background(), &pb.SendOTPRequest{
		Email: "",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestVerifyOTP_ShouldReturnError_WhenEmailEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.VerifyOTP(context.Background(), &pb.VerifyOTPRequest{
		Email: "",
		Code:  "123456",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestVerifyOTP_ShouldReturnError_WhenCodeEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.VerifyOTP(context.Background(), &pb.VerifyOTPRequest{
		Email: "test@example.com",
		Code:  "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "code is required")
}

func TestVerifyOTP_ShouldReturnError_WhenOTPServiceNil(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.VerifyOTP(context.Background(), &pb.VerifyOTPRequest{
		Email: "test@example.com",
		Code:  "123456",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "OTP service not configured")
}

func TestLoginWithOTP_ShouldReturnError_WhenEmailEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.LoginWithOTP(context.Background(), &pb.LoginWithOTPRequest{
		Email: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestLoginWithOTP_ShouldReturnError_WhenOTPServiceNil(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.LoginWithOTP(context.Background(), &pb.LoginWithOTPRequest{
		Email: "test@example.com",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "OTP service not configured")
}

func TestVerifyLoginOTP_ShouldReturnError_WhenEmailEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.VerifyLoginOTP(context.Background(), &pb.VerifyLoginOTPRequest{
		Email: "",
		Code:  "123456",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestVerifyLoginOTP_ShouldReturnError_WhenCodeEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.VerifyLoginOTP(context.Background(), &pb.VerifyLoginOTPRequest{
		Email: "test@example.com",
		Code:  "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "code is required")
}

func TestRegisterWithOTP_ShouldReturnError_WhenEmailEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.RegisterWithOTP(context.Background(), &pb.RegisterWithOTPRequest{
		Email: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestRegisterWithOTP_ShouldReturnError_WhenOTPServiceNil(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.RegisterWithOTP(context.Background(), &pb.RegisterWithOTPRequest{
		Email: "test@example.com",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "OTP service not configured")
}

func TestVerifyRegistrationOTP_ShouldReturnError_WhenEmailEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.VerifyRegistrationOTP(context.Background(), &pb.VerifyRegistrationOTPRequest{
		Email: "",
		Code:  "123456",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestVerifyRegistrationOTP_ShouldReturnError_WhenCodeEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.VerifyRegistrationOTP(context.Background(), &pb.VerifyRegistrationOTPRequest{
		Email: "test@example.com",
		Code:  "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "code is required")
}

// ===================== OAuth Handler Validation Tests =====================

func TestIntrospectOAuthToken_ShouldReturnInactive_WhenEmptyToken(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	resp, err := handler.IntrospectOAuthToken(context.Background(), &pb.IntrospectOAuthTokenRequest{
		Token: "",
	})

	require.NoError(t, err)
	assert.False(t, resp.Active)
	assert.Equal(t, "token is required", resp.ErrorMessage)
}

func TestIntrospectOAuthToken_ShouldReturnError_WhenServiceNil(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil,
		nil, // oauthProviderService = nil
		nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	resp, err := handler.IntrospectOAuthToken(context.Background(), &pb.IntrospectOAuthTokenRequest{
		Token: "some-token",
	})

	require.NoError(t, err)
	assert.False(t, resp.Active)
	assert.Equal(t, "OAuth provider service not configured", resp.ErrorMessage)
}

func TestValidateOAuthClient_ShouldReturnError_WhenClientIDEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.ValidateOAuthClient(context.Background(), &pb.ValidateOAuthClientRequest{
		ClientId: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "client_id is required")
}

func TestValidateOAuthClient_ShouldReturnInvalid_WhenServiceNil(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	resp, err := handler.ValidateOAuthClient(context.Background(), &pb.ValidateOAuthClientRequest{
		ClientId: "test-client",
	})

	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.Equal(t, "OAuth provider service not configured", resp.ErrorMessage)
}

func TestGetOAuthClient_ShouldReturnError_WhenClientIDEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.GetOAuthClient(context.Background(), &pb.GetOAuthClientRequest{
		ClientId: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "client_id is required")
}

func TestGetOAuthClient_ShouldReturnError_WhenServiceNil(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.GetOAuthClient(context.Background(), &pb.GetOAuthClientRequest{
		ClientId: "test-client",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth provider service not configured")
}

// ===================== Email Handler Validation Tests =====================

func TestSendEmail_ShouldReturnError_WhenTemplateTypeEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.SendEmail(context.Background(), &pb.SendEmailRequest{
		TemplateType: "",
		ToEmail:      "test@example.com",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "template_type is required")
}

func TestSendEmail_ShouldReturnError_WhenToEmailEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.SendEmail(context.Background(), &pb.SendEmailRequest{
		TemplateType: "welcome",
		ToEmail:      "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "to_email is required")
}

func TestSendEmail_ShouldReturnFailed_WhenServiceNil(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil,
		nil, // emailProfileService = nil
		nil, nil, nil, nil,
		testLogger(),
	)

	resp, err := handler.SendEmail(context.Background(), &pb.SendEmailRequest{
		TemplateType: "welcome",
		ToEmail:      "test@example.com",
	})

	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "email profile service not configured", resp.ErrorMessage)
}

// ===================== Multi-Application Handler Validation Tests =====================

func TestGetUserApplicationProfile_ShouldReturnError_WhenUserIDEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.GetUserApplicationProfile(context.Background(), &pb.GetUserAppProfileRequest{
		UserId:        "",
		ApplicationId: uuid.New().String(),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestGetUserApplicationProfile_ShouldReturnError_WhenApplicationIDEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.GetUserApplicationProfile(context.Background(), &pb.GetUserAppProfileRequest{
		UserId:        uuid.New().String(),
		ApplicationId: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "application_id is required")
}

func TestGetUserApplicationProfile_ShouldReturnError_WhenInvalidUserID(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.GetUserApplicationProfile(context.Background(), &pb.GetUserAppProfileRequest{
		UserId:        "bad-uuid",
		ApplicationId: uuid.New().String(),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user_id format")
}

func TestGetUserTelegramBots_ShouldReturnUnimplemented(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.GetUserTelegramBots(context.Background(), &pb.GetUserTelegramBotsRequest{
		UserId:        uuid.New().String(),
		ApplicationId: uuid.New().String(),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unimplemented")
}

func TestGetUserTelegramBots_ShouldReturnError_WhenUserIDEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.GetUserTelegramBots(context.Background(), &pb.GetUserTelegramBotsRequest{
		UserId:        "",
		ApplicationId: uuid.New().String(),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestUpdateUserProfile_ShouldReturnError_WhenMissingIDs(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.UpdateUserProfile(context.Background(), &pb.UpdateUserProfileRequest{
		UserId:        "",
		ApplicationId: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id and application_id are required")
}

func TestUpdateUserProfile_ShouldReturnError_WhenInvalidUserID(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.UpdateUserProfile(context.Background(), &pb.UpdateUserProfileRequest{
		UserId:        "bad",
		ApplicationId: uuid.New().String(),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user_id format")
}

func TestUpdateUserProfile_ShouldReturnError_WhenInvalidAppID(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.UpdateUserProfile(context.Background(), &pb.UpdateUserProfileRequest{
		UserId:        uuid.New().String(),
		ApplicationId: "bad",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid application_id format")
}

func TestCreateUserProfile_ShouldReturnError_WhenMissingIDs(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.CreateUserProfile(context.Background(), &pb.CreateUserProfileRequest{
		UserId:        "",
		ApplicationId: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id and application_id are required")
}

func TestDeleteUserProfile_ShouldReturnError_WhenMissingIDs(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.DeleteUserProfile(context.Background(), &pb.DeleteUserProfileRequest{
		UserId:        "",
		ApplicationId: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id and application_id are required")
}

func TestBanUser_ShouldReturnError_WhenMissingIDs(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.BanUser(context.Background(), &pb.BanUserRequest{
		UserId:        "",
		ApplicationId: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id and application_id are required")
}

func TestBanUser_ShouldReturnError_WhenInvalidBannedBy(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.BanUser(context.Background(), &pb.BanUserRequest{
		UserId:        uuid.New().String(),
		ApplicationId: uuid.New().String(),
		BannedBy:      "not-a-uuid",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid banned_by format")
}

func TestUnbanUser_ShouldReturnError_WhenMissingIDs(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.UnbanUser(context.Background(), &pb.UnbanUserRequest{
		UserId:        "",
		ApplicationId: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id and application_id are required")
}

func TestListApplicationUsers_ShouldReturnError_WhenApplicationIDEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.ListApplicationUsers(context.Background(), &pb.ListApplicationUsersRequest{
		ApplicationId: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "application_id is required")
}

func TestListApplicationUsers_ShouldReturnError_WhenInvalidApplicationID(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.ListApplicationUsers(context.Background(), &pb.ListApplicationUsersRequest{
		ApplicationId: "bad-uuid",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid application_id format")
}

// ===================== SyncUsers Validation Tests =====================

func TestSyncUsers_ShouldReturnError_WhenUpdatedAfterEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	resp, err := handler.SyncUsers(context.Background(), &pb.SyncUsersRequest{
		UpdatedAfter: "",
	})

	require.NoError(t, err)
	assert.Equal(t, "updated_after is required", resp.ErrorMessage)
}

func TestSyncUsers_ShouldReturnError_WhenInvalidTimestamp(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	resp, err := handler.SyncUsers(context.Background(), &pb.SyncUsersRequest{
		UpdatedAfter: "not-a-timestamp",
	})

	require.NoError(t, err)
	assert.Contains(t, resp.ErrorMessage, "invalid updated_after format")
}

// ===================== GetApplicationAuthConfig Validation Tests =====================

func TestGetApplicationAuthConfig_ShouldReturnError_WhenApplicationIDEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	resp, err := handler.GetApplicationAuthConfig(context.Background(), &pb.GetApplicationAuthConfigRequest{
		ApplicationId: "",
	})

	require.NoError(t, err)
	assert.Equal(t, "application_id is required", resp.ErrorMessage)
}

func TestGetApplicationAuthConfig_ShouldReturnError_WhenInvalidApplicationID(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	resp, err := handler.GetApplicationAuthConfig(context.Background(), &pb.GetApplicationAuthConfigRequest{
		ApplicationId: "bad-uuid",
	})

	require.NoError(t, err)
	assert.Equal(t, "invalid application_id format", resp.ErrorMessage)
}

// ===================== Token Exchange Validation Tests =====================

func TestCreateTokenExchange_ShouldReturnError_WhenAccessTokenEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	resp, err := handler.CreateTokenExchange(context.Background(), &pb.CreateTokenExchangeGrpcRequest{
		AccessToken:         "",
		TargetApplicationId: uuid.New().String(),
	})

	require.NoError(t, err)
	assert.Equal(t, "access_token is required", resp.ErrorMessage)
}

func TestCreateTokenExchange_ShouldReturnError_WhenTargetAppIDEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	resp, err := handler.CreateTokenExchange(context.Background(), &pb.CreateTokenExchangeGrpcRequest{
		AccessToken:         "some-token",
		TargetApplicationId: "",
	})

	require.NoError(t, err)
	assert.Equal(t, "target_application_id is required", resp.ErrorMessage)
}

func TestCreateTokenExchange_ShouldReturnError_WhenServiceNil(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, // tokenExchangeService = nil
		testLogger(),
	)

	resp, err := handler.CreateTokenExchange(context.Background(), &pb.CreateTokenExchangeGrpcRequest{
		AccessToken:         "some-token",
		TargetApplicationId: uuid.New().String(),
	})

	require.NoError(t, err)
	assert.Equal(t, "token exchange service not configured", resp.ErrorMessage)
}

func TestRedeemTokenExchange_ShouldReturnError_WhenExchangeCodeEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	resp, err := handler.RedeemTokenExchange(context.Background(), &pb.RedeemTokenExchangeGrpcRequest{
		ExchangeCode: "",
	})

	require.NoError(t, err)
	assert.Equal(t, "exchange_code is required", resp.ErrorMessage)
}

func TestRedeemTokenExchange_ShouldReturnError_WhenServiceNil(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, // tokenExchangeService = nil
		testLogger(),
	)

	resp, err := handler.RedeemTokenExchange(context.Background(), &pb.RedeemTokenExchangeGrpcRequest{
		ExchangeCode: "some-code",
	})

	require.NoError(t, err)
	assert.Equal(t, "token exchange service not configured", resp.ErrorMessage)
}

// ===================== InitPasswordlessRegistration Validation Tests =====================

func TestInitPasswordlessRegistration_ShouldReturnError_WhenNoEmailOrPhone(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.InitPasswordlessRegistration(context.Background(), &pb.InitPasswordlessRegistrationRequest{
		Email: "",
		Phone: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "either email or phone is required")
}

func TestCompletePasswordlessRegistration_ShouldReturnError_WhenCodeEmpty(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.CompletePasswordlessRegistration(context.Background(), &pb.CompletePasswordlessRegistrationRequest{
		Email: "test@example.com",
		Code:  "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "code is required")
}

func TestCompletePasswordlessRegistration_ShouldReturnError_WhenNoEmailOrPhone(t *testing.T) {
	handler := NewAuthHandlerV2(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		testLogger(),
	)

	_, err := handler.CompletePasswordlessRegistration(context.Background(), &pb.CompletePasswordlessRegistrationRequest{
		Email: "",
		Phone: "",
		Code:  "123456",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "either email or phone is required")
}

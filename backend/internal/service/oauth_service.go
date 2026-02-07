package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

// OAuthService provides OAuth operations
type OAuthService struct {
	userRepo             UserStore
	oauthRepo            OAuthStore
	tokenRepo            TokenStore
	auditRepo            AuditStore
	rbacRepo             RBACStore
	jwtService           TokenService
	sessionService       *SessionService
	loginAlertService    *LoginAlertService
	httpClient           HTTPClient
	providers            map[models.OAuthProvider]*OAuthProviderConfig
	jitProvisioning      bool // Enable Just-In-Time user provisioning
	appOAuthProviderRepo AppOAuthProviderStore
	appRepo              ApplicationStore
}

// OAuthProviderConfig holds OAuth provider configuration
type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(
	userRepo UserStore,
	oauthRepo OAuthStore,
	tokenRepo TokenStore,
	auditRepo AuditStore,
	rbacRepo RBACStore,
	jwtService TokenService,
	sessionService *SessionService,
	httpClient HTTPClient,
	appOAuthProviderRepo AppOAuthProviderStore,
	appRepo ApplicationStore,
	jitProvisioning bool,
	loginAlertService *LoginAlertService,
) *OAuthService {
	// Use default HTTP client if not provided
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	service := &OAuthService{
		userRepo:             userRepo,
		oauthRepo:            oauthRepo,
		tokenRepo:            tokenRepo,
		auditRepo:            auditRepo,
		rbacRepo:             rbacRepo,
		jwtService:           jwtService,
		sessionService:       sessionService,
		loginAlertService:    loginAlertService,
		httpClient:           httpClient,
		providers:            make(map[models.OAuthProvider]*OAuthProviderConfig),
		jitProvisioning:      jitProvisioning,
		appOAuthProviderRepo: appOAuthProviderRepo,
		appRepo:              appRepo,
	}

	// Initialize providers
	service.initializeProviders()

	return service
}

func (s *OAuthService) initializeProviders() {
	// Google OAuth
	s.providers[models.ProviderGoogle] = &OAuthProviderConfig{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		CallbackURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes:       []string{"openid", "profile", "email"},
	}

	// Yandex OAuth
	s.providers[models.ProviderYandex] = &OAuthProviderConfig{
		ClientID:     os.Getenv("YANDEX_CLIENT_ID"),
		ClientSecret: os.Getenv("YANDEX_CLIENT_SECRET"),
		CallbackURL:  os.Getenv("YANDEX_CALLBACK_URL"),
		AuthURL:      "https://oauth.yandex.ru/authorize",
		TokenURL:     "https://oauth.yandex.ru/token",
		UserInfoURL:  "https://login.yandex.ru/info",
		Scopes:       []string{"login:email", "login:info"},
	}

	// GitHub OAuth
	s.providers[models.ProviderGitHub] = &OAuthProviderConfig{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		CallbackURL:  os.Getenv("GITHUB_CALLBACK_URL"),
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
		Scopes:       []string{"user:email"},
	}

	// Instagram Basic Display API
	s.providers[models.ProviderInstagram] = &OAuthProviderConfig{
		ClientID:     os.Getenv("INSTAGRAM_CLIENT_ID"),
		ClientSecret: os.Getenv("INSTAGRAM_CLIENT_SECRET"),
		CallbackURL:  os.Getenv("INSTAGRAM_CALLBACK_URL"),
		AuthURL:      "https://api.instagram.com/oauth/authorize",
		TokenURL:     "https://api.instagram.com/oauth/access_token",
		UserInfoURL:  "https://graph.instagram.com/me",
		Scopes:       []string{"user_profile", "user_media"},
	}

	// Telegram (using bot API)
	s.providers[models.ProviderTelegram] = &OAuthProviderConfig{
		ClientID:     os.Getenv("TELEGRAM_BOT_TOKEN"),
		ClientSecret: "",
		CallbackURL:  os.Getenv("TELEGRAM_CALLBACK_URL"),
		// Telegram uses widget authentication, not traditional OAuth
	}

	// 1C OAuth (custom OAuth provider with configurable URLs)
	if os.Getenv("OAUTH_ONEC_ENABLED") == "true" {
		scopes := os.Getenv("OAUTH_ONEC_SCOPES")
		if scopes == "" {
			scopes = "openid profile email"
		}
		s.providers[models.ProviderOneC] = &OAuthProviderConfig{
			ClientID:     os.Getenv("OAUTH_ONEC_CLIENT_ID"),
			ClientSecret: os.Getenv("OAUTH_ONEC_CLIENT_SECRET"),
			CallbackURL:  os.Getenv("OAUTH_ONEC_REDIRECT_URI"),
			AuthURL:      os.Getenv("OAUTH_ONEC_AUTH_URL"),
			TokenURL:     os.Getenv("OAUTH_ONEC_TOKEN_URL"),
			UserInfoURL:  os.Getenv("OAUTH_ONEC_USERINFO_URL"),
			Scopes:       splitScopes(scopes),
		}
	}
}

// GenerateState generates a random state for OAuth flow
func (s *OAuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// getProviderConfigForApp returns OAuth provider config for a specific application.
// Falls back to env-based config if no app-specific config is found.
func (s *OAuthService) getProviderConfigForApp(ctx context.Context, provider models.OAuthProvider, appID *uuid.UUID) (*OAuthProviderConfig, error) {
	if appID != nil && s.appOAuthProviderRepo != nil {
		appProvider, err := s.appOAuthProviderRepo.GetByAppAndProvider(ctx, *appID, string(provider))
		if err == nil && appProvider != nil && appProvider.IsActive {
			scopes := appProvider.Scopes
			if len(scopes) == 0 {
				// Use default scopes from env-based config if app doesn't specify
				if envConfig, ok := s.providers[provider]; ok {
					scopes = envConfig.Scopes
				}
			}
			return &OAuthProviderConfig{
				ClientID:     appProvider.ClientID,
				ClientSecret: appProvider.ClientSecret,
				CallbackURL:  appProvider.CallbackURL,
				AuthURL:      appProvider.AuthURL,
				TokenURL:     appProvider.TokenURL,
				UserInfoURL:  appProvider.UserInfoURL,
				Scopes:       scopes,
			}, nil
		}
	}

	// Fallback to env-based config
	config, exists := s.providers[provider]
	if !exists || config.ClientID == "" {
		return nil, models.ErrInvalidProvider
	}
	return config, nil
}

// GetAuthURL returns the OAuth authorization URL for a provider
func (s *OAuthService) GetAuthURL(ctx context.Context, provider models.OAuthProvider, state string, appID *uuid.UUID) (string, error) {
	config, err := s.getProviderConfigForApp(ctx, provider, appID)
	if err != nil {
		return "", err
	}

	if provider == models.ProviderTelegram {
		// Telegram uses widget authentication
		return fmt.Sprintf("%s?bot_id=%s", config.CallbackURL, config.ClientID), nil
	}

	params := url.Values{}
	params.Add("client_id", config.ClientID)
	params.Add("redirect_uri", config.CallbackURL)
	params.Add("response_type", "code")
	params.Add("state", state)
	params.Add("scope", joinScopes(config.Scopes))

	// Provider-specific parameters
	if provider == models.ProviderGoogle {
		params.Add("access_type", "offline")
		params.Add("prompt", "consent")
	}

	return fmt.Sprintf("%s?%s", config.AuthURL, params.Encode()), nil
}

// ExchangeCode exchanges authorization code for access token
func (s *OAuthService) ExchangeCode(ctx context.Context, provider models.OAuthProvider, code string, appID *uuid.UUID) (*OAuthTokenResponse, error) {
	config, err := s.getProviderConfigForApp(ctx, provider, appID)
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", config.CallbackURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", config.TokenURL, nil)
	if err != nil {
		return nil, err
	}

	// Instagram uses form data
	if provider == models.ProviderInstagram {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Body = io.NopCloser(nil)
	} else {
		req.Header.Set("Accept", "application/json")
		req.URL.RawQuery = data.Encode()
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var tokenResp OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo fetches user information from OAuth provider
func (s *OAuthService) GetUserInfo(ctx context.Context, provider models.OAuthProvider, accessToken string, appID *uuid.UUID) (*models.OAuthUserInfo, error) {
	config, err := s.getProviderConfigForApp(ctx, provider, appID)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", config.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var rawUserInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawUserInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return s.parseUserInfo(provider, rawUserInfo)
}

// parseUserInfo parses provider-specific user info into standard format
func (s *OAuthService) parseUserInfo(provider models.OAuthProvider, data map[string]interface{}) (*models.OAuthUserInfo, error) {
	userInfo := &models.OAuthUserInfo{
		Provider: string(provider),
	}

	switch provider {
	case models.ProviderGoogle:
		userInfo.ProviderUserID = getString(data, "id")
		userInfo.Email = getString(data, "email")
		userInfo.Name = getString(data, "name")
		userInfo.ProfilePicture = getString(data, "picture")

	case models.ProviderYandex:
		userInfo.ProviderUserID = getString(data, "id")
		userInfo.Email = getString(data, "default_email")
		userInfo.Name = getString(data, "real_name")
		userInfo.Username = getString(data, "login")

	case models.ProviderGitHub:
		userInfo.ProviderUserID = fmt.Sprintf("%v", data["id"])
		userInfo.Email = getString(data, "email")
		userInfo.Name = getString(data, "name")
		userInfo.Username = getString(data, "login")
		userInfo.ProfilePicture = getString(data, "avatar_url")

	case models.ProviderInstagram:
		userInfo.ProviderUserID = getString(data, "id")
		userInfo.Username = getString(data, "username")
		userInfo.Name = getString(data, "username") // Instagram doesn't provide full name

	case models.ProviderTelegram:
		userInfo.ProviderUserID = fmt.Sprintf("%v", data["id"])
		userInfo.Name = getString(data, "first_name")
		if lastName := getString(data, "last_name"); lastName != "" {
			userInfo.Name += " " + lastName
		}
		userInfo.Username = getString(data, "username")
		userInfo.ProfilePicture = getString(data, "photo_url")

	case models.ProviderOneC:
		// 1C OAuth userinfo response parsing
		// Fields may vary depending on 1C system configuration
		userInfo.ProviderUserID = getString(data, "sub")
		if userInfo.ProviderUserID == "" {
			userInfo.ProviderUserID = getString(data, "id")
		}
		if userInfo.ProviderUserID == "" {
			userInfo.ProviderUserID = getString(data, "user_id")
		}
		userInfo.Email = getString(data, "email")
		userInfo.Name = getString(data, "name")
		if userInfo.Name == "" {
			// Try to compose name from parts
			firstName := getString(data, "given_name")
			lastName := getString(data, "family_name")
			if firstName != "" || lastName != "" {
				userInfo.Name = firstName
				if lastName != "" {
					if userInfo.Name != "" {
						userInfo.Name += " "
					}
					userInfo.Name += lastName
				}
			}
		}
		userInfo.Username = getString(data, "preferred_username")
		if userInfo.Username == "" {
			userInfo.Username = getString(data, "username")
		}
		userInfo.ProfilePicture = getString(data, "picture")
	}

	return userInfo, nil
}

// HandleCallback handles OAuth callback and creates/updates user
func (s *OAuthService) HandleCallback(ctx context.Context, provider models.OAuthProvider, code, ipAddress, userAgent string, appID *uuid.UUID) (*models.OAuthLoginResponse, error) {
	// Exchange code for token
	tokenResp, err := s.ExchangeCode(ctx, provider, code, appID)
	if err != nil {
		return nil, err
	}

	// Get user info
	userInfo, err := s.GetUserInfo(ctx, provider, tokenResp.AccessToken, appID)
	if err != nil {
		return nil, err
	}

	// Find or create OAuth account
	oauthAccount, err := s.oauthRepo.GetOAuthAccount(ctx, string(provider), userInfo.ProviderUserID)
	isNewUser := false

	if oauthAccount == nil {
		// Check if JIT provisioning is enabled
		if !s.jitProvisioning {
			return nil, models.NewAppError(403, "User not found. Automatic user creation is disabled.")
		}

		// OAuth account doesn't exist, create new user (JIT provisioning)
		user, err := s.createUserFromOAuth(ctx, userInfo)
		if err != nil {
			return nil, err
		}

		// Create OAuth account link
		oauthAccount = &models.OAuthAccount{
			ID:             uuid.New(),
			UserID:         user.ID,
			Provider:       string(provider),
			ProviderUserID: userInfo.ProviderUserID,
			AccessToken:    tokenResp.AccessToken,
			RefreshToken:   tokenResp.RefreshToken,
		}

		if tokenResp.ExpiresIn > 0 {
			expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
			oauthAccount.TokenExpiresAt = &expiresAt
		}

		// Store profile data as JSON
		profileData, _ := json.Marshal(userInfo)
		oauthAccount.ProfileData = profileData

		if err := s.oauthRepo.CreateOAuthAccount(ctx, oauthAccount); err != nil {
			return nil, err
		}

		isNewUser = true
	} else {
		// Update existing OAuth account
		oauthAccount.AccessToken = tokenResp.AccessToken
		if tokenResp.RefreshToken != "" {
			oauthAccount.RefreshToken = tokenResp.RefreshToken
		}
		if tokenResp.ExpiresIn > 0 {
			expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
			oauthAccount.TokenExpiresAt = &expiresAt
		}

		profileData, _ := json.Marshal(userInfo)
		oauthAccount.ProfileData = profileData

		if err := s.oauthRepo.UpdateOAuthAccount(ctx, oauthAccount); err != nil {
			return nil, err
		}
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, oauthAccount.UserID, utils.Ptr(true))
	if err != nil {
		return nil, err
	}

	// Generate JWT tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user, appID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user, appID)
	if err != nil {
		return nil, err
	}

	// Parse device info for RefreshToken and Session
	deviceInfo := utils.ParseUserAgent(userAgent)
	sessionName := utils.GenerateSessionName(deviceInfo)

	// Save refresh token with device tracking
	tokenHash := utils.HashToken(refreshToken)
	refreshExpiration := s.jwtService.GetRefreshTokenExpiration()
	dbToken := &models.RefreshToken{
		ID:          uuid.New(),
		UserID:      user.ID,
		TokenHash:   tokenHash,
		ExpiresAt:   time.Now().Add(refreshExpiration),
		DeviceType:  deviceInfo.DeviceType,
		OS:          deviceInfo.OS,
		Browser:     deviceInfo.Browser,
		SessionName: sessionName,
		IPAddress:   ipAddress,
	}

	if err := s.tokenRepo.CreateRefreshToken(ctx, dbToken); err != nil {
		return nil, err
	}

	// Create session using SessionService (non-fatal to not block auth)
	if s.sessionService != nil {
		s.sessionService.CreateSessionNonFatal(ctx, SessionCreationParams{
			UserID:          user.ID,
			TokenHash:       tokenHash,
			AccessTokenHash: utils.HashToken(accessToken),
			IPAddress:       ipAddress,
			UserAgent:       userAgent,
			ExpiresAt:       time.Now().Add(refreshExpiration),
			SessionName:     sessionName,
		})
	}

	// Check for new device and send login alert email (async, non-blocking)
	if s.loginAlertService != nil {
		go func() {
			alertCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			s.loginAlertService.CheckAndAlert(alertCtx, LoginAlertParams{
				UserID:    user.ID,
				Username:  user.Username,
				Email:     user.Email,
				IP:        ipAddress,
				UserAgent: userAgent,
				Device:    deviceInfo,
				AppID:     appID,
				IsNewUser: isNewUser,
			})
		}()
	}

	return &models.OAuthLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		IsNewUser:    isNewUser,
	}, nil
}

// createUserFromOAuth creates a new user from OAuth data
func (s *OAuthService) createUserFromOAuth(ctx context.Context, userInfo *models.OAuthUserInfo) (*models.User, error) {
	email := userInfo.Email
	if email == "" {
		// Generate a placeholder email if provider doesn't provide one
		email = fmt.Sprintf("%s_%s@oauth.local", userInfo.Provider, userInfo.ProviderUserID)
	}

	username := userInfo.Username
	if username == "" {
		username = userInfo.Name
	}
	if username == "" {
		username = fmt.Sprintf("%s_%s", userInfo.Provider, userInfo.ProviderUserID[:8])
	}

	// Ensure unique username
	username = utils.NormalizeUsername(username)
	originalUsername := username
	counter := 1
	for {
		exists, err := s.userRepo.UsernameExists(ctx, username)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		username = fmt.Sprintf("%s%d", originalUsername, counter)
		counter++
	}

	// Get default "user" role
	defaultRole, err := s.rbacRepo.GetRoleByName(ctx, "user")
	if err != nil {
		return nil, fmt.Errorf("failed to get default role: %w", err)
	}

	user := &models.User{
		ID:            uuid.New(),
		Email:         email,
		Username:      username,
		FullName:      userInfo.Name,
		PasswordHash:  "", // OAuth users don't have passwords
		IsActive:      true,
		EmailVerified: userInfo.Email != "", // Mark as verified if email provided by OAuth
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Assign default "user" role to the new user
	if err := s.rbacRepo.AssignRoleToUser(ctx, user.ID, defaultRole.ID, user.ID); err != nil {
		return nil, fmt.Errorf("failed to assign default role: %w", err)
	}

	return user, nil
}

// OAuthTokenResponse represents OAuth token response
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func joinScopes(scopes []string) string {
	result := ""
	for i, scope := range scopes {
		if i > 0 {
			result += " "
		}
		result += scope
	}
	return result
}

func splitScopes(scopes string) []string {
	var result []string
	current := ""
	for _, char := range scopes {
		if char == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

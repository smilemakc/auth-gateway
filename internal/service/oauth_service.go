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
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
)

// OAuthService provides OAuth operations
type OAuthService struct {
	userRepo    *repository.UserRepository
	oauthRepo   *repository.OAuthRepository
	tokenRepo   *repository.TokenRepository
	auditRepo   *repository.AuditRepository
	jwtService  *jwt.Service
	providers   map[models.OAuthProvider]*OAuthProviderConfig
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
	userRepo *repository.UserRepository,
	oauthRepo *repository.OAuthRepository,
	tokenRepo *repository.TokenRepository,
	auditRepo *repository.AuditRepository,
	jwtService *jwt.Service,
) *OAuthService {
	service := &OAuthService{
		userRepo:   userRepo,
		oauthRepo:  oauthRepo,
		tokenRepo:  tokenRepo,
		auditRepo:  auditRepo,
		jwtService: jwtService,
		providers:  make(map[models.OAuthProvider]*OAuthProviderConfig),
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
}

// GenerateState generates a random state for OAuth flow
func (s *OAuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthURL returns the OAuth authorization URL for a provider
func (s *OAuthService) GetAuthURL(provider models.OAuthProvider, state string) (string, error) {
	config, exists := s.providers[provider]
	if !exists || config.ClientID == "" {
		return "", models.ErrInvalidProvider
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
func (s *OAuthService) ExchangeCode(ctx context.Context, provider models.OAuthProvider, code string) (*OAuthTokenResponse, error) {
	config, exists := s.providers[provider]
	if !exists {
		return nil, models.ErrInvalidProvider
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

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
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
func (s *OAuthService) GetUserInfo(ctx context.Context, provider models.OAuthProvider, accessToken string) (*models.OAuthUserInfo, error) {
	config, exists := s.providers[provider]
	if !exists {
		return nil, models.ErrInvalidProvider
	}

	req, err := http.NewRequestWithContext(ctx, "GET", config.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
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
	}

	return userInfo, nil
}

// HandleCallback handles OAuth callback and creates/updates user
func (s *OAuthService) HandleCallback(ctx context.Context, provider models.OAuthProvider, code string) (*models.OAuthLoginResponse, error) {
	// Exchange code for token
	tokenResp, err := s.ExchangeCode(ctx, provider, code)
	if err != nil {
		return nil, err
	}

	// Get user info
	userInfo, err := s.GetUserInfo(ctx, provider, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	// Find or create OAuth account
	oauthAccount, err := s.oauthRepo.GetOAuthAccount(string(provider), userInfo.ProviderUserID)
	isNewUser := false

	if oauthAccount == nil {
		// OAuth account doesn't exist, create new user
		user, err := s.createUserFromOAuth(userInfo)
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

		if err := s.oauthRepo.CreateOAuthAccount(oauthAccount); err != nil {
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

		if err := s.oauthRepo.UpdateOAuthAccount(oauthAccount); err != nil {
			return nil, err
		}
	}

	// Get user
	user, err := s.userRepo.GetByID(oauthAccount.UserID)
	if err != nil {
		return nil, err
	}

	// Generate JWT tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	// Save refresh token
	tokenHash := utils.HashToken(refreshToken)
	dbToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTokenExpiration()),
	}

	if err := s.tokenRepo.CreateRefreshToken(dbToken); err != nil {
		return nil, err
	}

	return &models.OAuthLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		IsNewUser:    isNewUser,
	}, nil
}

// createUserFromOAuth creates a new user from OAuth data
func (s *OAuthService) createUserFromOAuth(userInfo *models.OAuthUserInfo) (*models.User, error) {
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
		exists, err := s.userRepo.UsernameExists(username)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		username = fmt.Sprintf("%s%d", originalUsername, counter)
		counter++
	}

	user := &models.User{
		ID:             uuid.New(),
		Email:          email,
		Username:       username,
		FullName:       userInfo.Name,
		PasswordHash:   "", // OAuth users don't have passwords
		Role:           string(models.RoleUser),
		IsActive:       true,
		EmailVerified:  userInfo.Email != "", // Mark as verified if email provided by OAuth
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
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

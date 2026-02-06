package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/keys"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidClient           = errors.New("invalid_client")
	ErrInvalidGrant            = errors.New("invalid_grant")
	ErrInvalidScope            = errors.New("invalid_scope")
	ErrInvalidRequest          = errors.New("invalid_request")
	ErrUnauthorizedClient      = errors.New("unauthorized_client")
	ErrAccessDenied            = errors.New("access_denied")
	ErrUnsupportedGrantType    = errors.New("unsupported_grant_type")
	ErrUnsupportedResponseType = errors.New("unsupported_response_type")
	ErrServerError             = errors.New("server_error")
	ErrConsentRequired         = errors.New("consent_required")
	ErrLoginRequired           = errors.New("login_required")
	ErrAuthorizationPending    = errors.New("authorization_pending")
	ErrSlowDown                = errors.New("slow_down")
	ErrExpiredToken            = errors.New("expired_token")
)

const (
	authorizationCodeTTL      = 10 * time.Minute
	deviceCodeTTL             = 15 * time.Minute
	deviceCodePollingInterval = 5

	clientIDPrefix     = "agw_"
	clientSecretPrefix = "agws_"

	defaultAccessTokenTTL  = 900
	defaultRefreshTokenTTL = 604800
	defaultIDTokenTTL      = 3600

	bcryptCostForClientSecret = 10
)

type ConsentInfo struct {
	Client          *models.OAuthClient `json:"client"`
	RequestedScopes []ScopeInfo         `json:"requested_scopes"`
	AlreadyGranted  []string            `json:"already_granted,omitempty"`
}

type ScopeInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

type OAuthProviderService struct {
	repo           OAuthProviderStore
	userRepo       UserStore
	auditRepo      AuditStore
	sessionService *SessionService
	oidcJWT        *jwt.OIDCService
	keyManager     *keys.Manager
	logger         *logger.Logger
	issuer         string
	baseURL        string
}

func NewOAuthProviderService(
	repo OAuthProviderStore,
	userRepo UserStore,
	auditRepo AuditStore,
	sessionService *SessionService,
	oidcJWT *jwt.OIDCService,
	keyManager *keys.Manager,
	log *logger.Logger,
	issuer string,
	baseURL string,
) *OAuthProviderService {
	return &OAuthProviderService{
		repo:           repo,
		userRepo:       userRepo,
		auditRepo:      auditRepo,
		sessionService: sessionService,
		oidcJWT:        oidcJWT,
		keyManager:     keyManager,
		logger:         log,
		issuer:         issuer,
		baseURL:        baseURL,
	}
}

// NewOAuthProviderServiceMinimal creates a minimal service for OAuth client management
// when OIDC is not fully enabled. This allows managing OAuth clients without
// requiring the full OIDC infrastructure (signing keys, etc.)
func NewOAuthProviderServiceMinimal(
	repo OAuthProviderStore,
	auditRepo AuditStore,
	log *logger.Logger,
) *OAuthProviderService {
	return &OAuthProviderService{
		repo:      repo,
		auditRepo: auditRepo,
		logger:    log,
	}
}

func (s *OAuthProviderService) CreateClient(ctx context.Context, req *models.CreateOAuthClientRequest, ownerID *uuid.UUID) (*models.CreateOAuthClientResponse, error) {
	clientID := s.generateClientID()

	var clientSecretPlain string
	var clientSecretHash *string

	if req.ClientType == string(models.ClientTypeConfidential) {
		plain, hash, err := s.generateClientSecret()
		if err != nil {
			s.logger.Error("failed to generate client secret", map[string]interface{}{"error": err.Error()})
			return nil, fmt.Errorf("failed to generate client secret: %w", err)
		}
		clientSecretPlain = plain
		clientSecretHash = &hash
	}

	accessTokenTTL := defaultAccessTokenTTL
	if req.AccessTokenTTL != nil {
		accessTokenTTL = *req.AccessTokenTTL
	}

	refreshTokenTTL := defaultRefreshTokenTTL
	if req.RefreshTokenTTL != nil {
		refreshTokenTTL = *req.RefreshTokenTTL
	}

	idTokenTTL := defaultIDTokenTTL
	if req.IDTokenTTL != nil {
		idTokenTTL = *req.IDTokenTTL
	}

	requirePKCE := req.ClientType == string(models.ClientTypePublic)
	if req.RequirePKCE != nil {
		requirePKCE = *req.RequirePKCE
	}

	requireConsent := true
	if req.RequireConsent != nil {
		requireConsent = *req.RequireConsent
	}

	firstParty := false
	if req.FirstParty != nil {
		firstParty = *req.FirstParty
	}

	client := &models.OAuthClient{
		ID:                uuid.New(),
		ClientID:          clientID,
		ClientSecretHash:  clientSecretHash,
		Name:              req.Name,
		Description:       req.Description,
		LogoURL:           req.LogoURL,
		ClientType:        req.ClientType,
		RedirectURIs:      req.RedirectURIs,
		AllowedGrantTypes: req.AllowedGrantTypes,
		AllowedScopes:     req.AllowedScopes,
		DefaultScopes:     req.DefaultScopes,
		AccessTokenTTL:    accessTokenTTL,
		RefreshTokenTTL:   refreshTokenTTL,
		IDTokenTTL:        idTokenTTL,
		RequirePKCE:       requirePKCE,
		RequireConsent:    requireConsent,
		FirstParty:        firstParty,
		OwnerID:           ownerID,
		IsActive:          true,
	}

	if err := s.repo.CreateClient(ctx, client); err != nil {
		s.logger.Error("failed to create oauth client", map[string]interface{}{
			"error":     err.Error(),
			"client_id": clientID,
		})
		return nil, fmt.Errorf("failed to create oauth client: %w", err)
	}

	s.logger.Info("oauth client created", map[string]interface{}{
		"client_id":   clientID,
		"client_type": req.ClientType,
		"name":        req.Name,
	})

	return &models.CreateOAuthClientResponse{
		Client:       client,
		ClientSecret: clientSecretPlain,
	}, nil
}

func (s *OAuthProviderService) GetClient(ctx context.Context, id uuid.UUID) (*models.OAuthClient, error) {
	client, err := s.repo.GetClientByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth client: %w", err)
	}
	return client, nil
}

func (s *OAuthProviderService) GetClientByClientID(ctx context.Context, clientID string) (*models.OAuthClient, error) {
	client, err := s.repo.GetClientByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth client: %w", err)
	}
	return client, nil
}

func (s *OAuthProviderService) UpdateClient(ctx context.Context, id uuid.UUID, req *models.UpdateOAuthClientRequest) (*models.OAuthClient, error) {
	client, err := s.repo.GetClientByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth client: %w", err)
	}

	if req.Name != "" {
		client.Name = req.Name
	}
	if req.Description != "" {
		client.Description = req.Description
	}
	if req.LogoURL != "" {
		client.LogoURL = req.LogoURL
	}
	if len(req.RedirectURIs) > 0 {
		client.RedirectURIs = req.RedirectURIs
	}
	if len(req.AllowedGrantTypes) > 0 {
		client.AllowedGrantTypes = req.AllowedGrantTypes
	}
	if len(req.AllowedScopes) > 0 {
		client.AllowedScopes = req.AllowedScopes
	}
	if len(req.DefaultScopes) > 0 {
		client.DefaultScopes = req.DefaultScopes
	}
	if req.AccessTokenTTL != nil {
		client.AccessTokenTTL = *req.AccessTokenTTL
	}
	if req.RefreshTokenTTL != nil {
		client.RefreshTokenTTL = *req.RefreshTokenTTL
	}
	if req.IDTokenTTL != nil {
		client.IDTokenTTL = *req.IDTokenTTL
	}
	if req.RequirePKCE != nil {
		client.RequirePKCE = *req.RequirePKCE
	}
	if req.RequireConsent != nil {
		client.RequireConsent = *req.RequireConsent
	}
	if req.IsActive != nil {
		client.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateClient(ctx, client); err != nil {
		s.logger.Error("failed to update oauth client", map[string]interface{}{
			"error":     err.Error(),
			"client_id": client.ClientID,
		})
		return nil, fmt.Errorf("failed to update oauth client: %w", err)
	}

	s.logger.Info("oauth client updated", map[string]interface{}{
		"client_id": client.ClientID,
	})

	return client, nil
}

func (s *OAuthProviderService) DeleteClient(ctx context.Context, id uuid.UUID) error {
	client, err := s.repo.GetClientByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get oauth client: %w", err)
	}

	if err := s.repo.DeleteClient(ctx, id); err != nil {
		s.logger.Error("failed to delete oauth client", map[string]interface{}{
			"error":     err.Error(),
			"client_id": client.ClientID,
		})
		return fmt.Errorf("failed to delete oauth client: %w", err)
	}

	s.logger.Info("oauth client deleted", map[string]interface{}{
		"client_id": client.ClientID,
	})

	return nil
}

func (s *OAuthProviderService) ListClients(ctx context.Context, ownerID *uuid.UUID, page, perPage int) ([]*models.OAuthClient, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	clients, total, err := s.repo.ListClients(ctx, ownerID, page, perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list oauth clients: %w", err)
	}

	return clients, total, nil
}

func (s *OAuthProviderService) RotateClientSecret(ctx context.Context, id uuid.UUID) (string, error) {
	client, err := s.repo.GetClientByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get oauth client: %w", err)
	}

	if client.ClientType != string(models.ClientTypeConfidential) {
		return "", fmt.Errorf("cannot rotate secret for public client")
	}

	plain, hash, err := s.generateClientSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate client secret: %w", err)
	}

	client.ClientSecretHash = &hash

	if err := s.repo.UpdateClient(ctx, client); err != nil {
		s.logger.Error("failed to rotate client secret", map[string]interface{}{
			"error":     err.Error(),
			"client_id": client.ClientID,
		})
		return "", fmt.Errorf("failed to rotate client secret: %w", err)
	}

	s.logger.Info("client secret rotated", map[string]interface{}{
		"client_id": client.ClientID,
	})

	return plain, nil
}

func (s *OAuthProviderService) ValidateClientCredentials(ctx context.Context, clientID, clientSecret string) (*models.OAuthClient, error) {
	client, err := s.repo.GetClientByClientID(ctx, clientID)
	if err != nil {
		return nil, ErrInvalidClient
	}

	if !client.IsActive {
		return nil, ErrInvalidClient
	}

	if client.ClientType == string(models.ClientTypeConfidential) {
		if client.ClientSecretHash == nil {
			return nil, ErrInvalidClient
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*client.ClientSecretHash), []byte(clientSecret)); err != nil {
			return nil, ErrInvalidClient
		}
	}

	return client, nil
}

func (s *OAuthProviderService) Authorize(ctx context.Context, req *models.AuthorizeRequest, userID uuid.UUID) (*models.AuthorizeResponse, error) {
	client, err := s.repo.GetClientByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, ErrInvalidClient
	}

	if !client.IsActive {
		return nil, ErrInvalidClient
	}

	if !s.validateRedirectURI(req.RedirectURI, client.RedirectURIs) {
		return nil, ErrInvalidRequest
	}

	if req.ResponseType != "code" {
		return nil, ErrUnsupportedResponseType
	}

	requestedScopes := s.parseScopes(req.Scope)
	if err := s.validateScopes(requestedScopes, client.AllowedScopes); err != nil {
		return nil, err
	}

	requirePKCE := client.RequirePKCE || client.ClientType == string(models.ClientTypePublic)
	if requirePKCE {
		if req.CodeChallenge == nil || *req.CodeChallenge == "" {
			return nil, fmt.Errorf("%w: code_challenge is required", ErrInvalidRequest)
		}
	}

	if req.CodeChallenge != nil && *req.CodeChallenge != "" {
		method := CodeChallengeMethodPlain
		if req.CodeChallengeMethod != nil && *req.CodeChallengeMethod != "" {
			method = *req.CodeChallengeMethod
		}
		if !IsValidCodeChallengeMethod(method) {
			return nil, ErrInvalidRequest
		}
	}

	needsConsent := client.RequireConsent && !client.FirstParty
	if needsConsent {
		consent, err := s.repo.GetUserConsent(ctx, userID, client.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check user consent: %w", err)
		}

		if consent == nil || consent.IsRevoked() || !s.hasAllScopes(consent.Scopes, requestedScopes) {
			return nil, ErrConsentRequired
		}
	}

	codePlain, codeHash, err := s.generateAuthorizationCode()
	if err != nil {
		s.logger.Error("failed to generate authorization code", map[string]interface{}{"error": err.Error()})
		return nil, ErrServerError
	}

	authCode := &models.AuthorizationCode{
		ID:          uuid.New(),
		CodeHash:    codeHash,
		ClientID:    client.ID,
		UserID:      userID,
		RedirectURI: req.RedirectURI,
		Scope:       req.Scope,
		ExpiresAt:   time.Now().Add(authorizationCodeTTL),
	}

	if req.CodeChallenge != nil && *req.CodeChallenge != "" {
		authCode.CodeChallenge = req.CodeChallenge
		method := CodeChallengeMethodPlain
		if req.CodeChallengeMethod != nil {
			method = *req.CodeChallengeMethod
		}
		authCode.CodeChallengeMethod = &method
	}

	if req.Nonce != nil {
		authCode.Nonce = req.Nonce
	}

	if err := s.repo.CreateAuthorizationCode(ctx, authCode); err != nil {
		s.logger.Error("failed to create authorization code", map[string]interface{}{
			"error":     err.Error(),
			"client_id": client.ClientID,
		})
		return nil, ErrServerError
	}

	s.logAudit(ctx, &userID, "oauth_authorize", "success", map[string]interface{}{
		"client_id": client.ClientID,
		"scope":     req.Scope,
	})

	return &models.AuthorizeResponse{
		Code:  codePlain,
		State: req.State,
	}, nil
}

func (s *OAuthProviderService) ExchangeCode(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error) {
	if req.Code == nil || *req.Code == "" {
		return nil, fmt.Errorf("%w: code is required", ErrInvalidRequest)
	}

	if req.RedirectURI == nil || *req.RedirectURI == "" {
		return nil, fmt.Errorf("%w: redirect_uri is required", ErrInvalidRequest)
	}

	codeHash := s.hashToken(*req.Code)
	authCode, err := s.repo.GetAuthorizationCode(ctx, codeHash)
	if err != nil {
		return nil, ErrInvalidGrant
	}

	if authCode.Used {
		s.logger.Warn("authorization code reuse attempted", map[string]interface{}{
			"code_id": authCode.ID.String(),
		})
		return nil, ErrInvalidGrant
	}

	if authCode.IsExpired() {
		return nil, ErrInvalidGrant
	}

	if authCode.RedirectURI != *req.RedirectURI {
		return nil, ErrInvalidGrant
	}

	client := authCode.Client
	if client == nil {
		client, err = s.repo.GetClientByID(ctx, authCode.ClientID)
		if err != nil {
			return nil, ErrInvalidClient
		}
	}

	if client.ClientID != req.ClientID {
		return nil, ErrInvalidClient
	}

	if client.ClientType == string(models.ClientTypeConfidential) {
		if req.ClientSecret == nil || *req.ClientSecret == "" {
			return nil, ErrInvalidClient
		}
		_, err := s.ValidateClientCredentials(ctx, req.ClientID, *req.ClientSecret)
		if err != nil {
			return nil, err
		}
	}

	if authCode.CodeChallenge != nil && *authCode.CodeChallenge != "" {
		if req.CodeVerifier == nil || *req.CodeVerifier == "" {
			return nil, fmt.Errorf("%w: code_verifier is required", ErrInvalidRequest)
		}

		method := CodeChallengeMethodPlain
		if authCode.CodeChallengeMethod != nil {
			method = *authCode.CodeChallengeMethod
		}

		if err := ValidateCodeChallenge(*req.CodeVerifier, *authCode.CodeChallenge, method); err != nil {
			return nil, ErrInvalidGrant
		}
	}

	if err := s.repo.MarkAuthorizationCodeUsed(ctx, authCode.ID); err != nil {
		s.logger.Error("failed to mark authorization code as used", map[string]interface{}{
			"error":   err.Error(),
			"code_id": authCode.ID.String(),
		})
		return nil, ErrServerError
	}

	user := authCode.User
	if user == nil {
		user, err = s.userRepo.GetByID(ctx, authCode.UserID, nil, UserGetWithRoles())
		if err != nil {
			return nil, ErrServerError
		}
	}

	scopes := s.parseScopes(authCode.Scope)

	response, err := s.generateTokens(ctx, client, &authCode.UserID, user, scopes, authCode.Nonce)
	if err != nil {
		return nil, err
	}

	// Create session for the user using SessionService
	if s.sessionService != nil && response.RefreshToken != "" {
		s.sessionService.CreateSessionNonFatal(ctx, SessionCreationParams{
			UserID:          authCode.UserID,
			TokenHash:       utils.HashToken(response.RefreshToken),
			AccessTokenHash: utils.HashToken(response.AccessToken),
			IPAddress:       req.IPAddress,
			UserAgent:       req.UserAgent,
			ExpiresAt:       time.Now().Add(time.Duration(client.RefreshTokenTTL) * time.Second),
		})
	}

	s.logger.Info("authorization code exchanged", map[string]interface{}{
		"user_id":   authCode.UserID.String(),
		"client_id": client.ClientID,
		"scopes":    scopes,
	})
	s.logAudit(ctx, &authCode.UserID, "oauth_code_exchanged", "success", map[string]interface{}{
		"client_id": client.ClientID,
		"scope":     req.Scope,
	})
	return response, nil
}

func (s *OAuthProviderService) ClientCredentialsGrant(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error) {
	if req.ClientSecret == nil || *req.ClientSecret == "" {
		return nil, ErrInvalidClient
	}

	client, err := s.ValidateClientCredentials(ctx, req.ClientID, *req.ClientSecret)
	if err != nil {
		return nil, err
	}

	if !s.hasGrantType(client.AllowedGrantTypes, string(models.GrantTypeClientCredentials)) {
		return nil, ErrUnauthorizedClient
	}

	requestedScopes := []string{}
	if req.Scope != nil && *req.Scope != "" {
		requestedScopes = s.parseScopes(*req.Scope)
	}

	if err := s.validateScopes(requestedScopes, client.AllowedScopes); err != nil {
		return nil, err
	}

	if len(requestedScopes) == 0 {
		requestedScopes = client.DefaultScopes
	}

	return s.generateTokens(ctx, client, nil, nil, requestedScopes, nil)
}

func (s *OAuthProviderService) RefreshToken(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error) {
	if req.RefreshToken == nil || *req.RefreshToken == "" {
		return nil, fmt.Errorf("%w: refresh_token is required", ErrInvalidRequest)
	}

	tokenHash := s.hashToken(*req.RefreshToken)
	refreshToken, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidGrant
	}

	if !refreshToken.IsValid() {
		return nil, ErrInvalidGrant
	}

	client := refreshToken.Client
	if client == nil {
		client, err = s.repo.GetClientByID(ctx, refreshToken.ClientID)
		if err != nil {
			return nil, ErrInvalidClient
		}
	}

	if client.ClientID != req.ClientID {
		return nil, ErrInvalidClient
	}

	if client.ClientType == string(models.ClientTypeConfidential) {
		if req.ClientSecret == nil || *req.ClientSecret == "" {
			return nil, ErrInvalidClient
		}
		_, err := s.ValidateClientCredentials(ctx, req.ClientID, *req.ClientSecret)
		if err != nil {
			return nil, err
		}
	}

	user := refreshToken.User
	if user == nil {
		user, err = s.userRepo.GetByID(ctx, refreshToken.UserID, nil, UserGetWithRoles())
		if err != nil {
			return nil, ErrServerError
		}
	}

	scopes := s.parseScopes(refreshToken.Scope)

	if err := s.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		s.logger.Error("failed to revoke old refresh token", map[string]interface{}{"error": err.Error()})
	}

	response, err := s.generateTokens(ctx, client, &refreshToken.UserID, user, scopes, nil)
	if err != nil {
		return nil, err
	}

	// Update existing session with new token hashes instead of creating a new one
	if s.sessionService != nil && response.RefreshToken != "" {
		s.sessionService.RefreshSessionNonFatal(ctx, SessionRefreshParams{
			OldRefreshTokenHash: tokenHash,
			NewRefreshTokenHash: utils.HashToken(response.RefreshToken),
			NewAccessTokenHash:  utils.HashToken(response.AccessToken),
			NewExpiresAt:        time.Now().Add(time.Duration(client.RefreshTokenTTL) * time.Second),
		})
	}

	s.logger.Info("token refreshed", map[string]interface{}{
		"user_id":   refreshToken.UserID.String(),
		"client_id": client.ClientID,
	})

	return response, nil
}

func (s *OAuthProviderService) DeviceAuthorization(ctx context.Context, req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error) {
	client, err := s.repo.GetClientByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, ErrInvalidClient
	}

	if !client.IsActive {
		return nil, ErrInvalidClient
	}

	if !s.hasGrantType(client.AllowedGrantTypes, string(models.GrantTypeDeviceCode)) {
		return nil, ErrUnauthorizedClient
	}

	requestedScopes := []string{}
	if req.Scope != nil && *req.Scope != "" {
		requestedScopes = s.parseScopes(*req.Scope)
	}

	if err := s.validateScopes(requestedScopes, client.AllowedScopes); err != nil {
		return nil, err
	}

	if len(requestedScopes) == 0 {
		requestedScopes = client.DefaultScopes
	}

	scope := strings.Join(requestedScopes, " ")

	deviceCodePlain, deviceCodeHash, err := s.generateToken()
	if err != nil {
		s.logger.Error("failed to generate device code", map[string]interface{}{"error": err.Error()})
		return nil, ErrServerError
	}

	userCode := s.generateUserCode()

	verificationURI := fmt.Sprintf("%s/device", s.baseURL)
	verificationURIComplete := fmt.Sprintf("%s?user_code=%s", verificationURI, userCode)

	deviceCode := &models.DeviceCode{
		ID:                      uuid.New(),
		DeviceCodeHash:          deviceCodeHash,
		UserCode:                userCode,
		ClientID:                client.ID,
		Scope:                   scope,
		Status:                  models.DeviceCodeStatusPending,
		VerificationURI:         verificationURI,
		VerificationURIComplete: verificationURIComplete,
		ExpiresAt:               time.Now().Add(deviceCodeTTL),
		Interval:                deviceCodePollingInterval,
	}

	if err := s.repo.CreateDeviceCode(ctx, deviceCode); err != nil {
		s.logger.Error("failed to create device code", map[string]interface{}{"error": err.Error()})
		return nil, ErrServerError
	}

	s.logger.Info("device authorization initiated", map[string]interface{}{
		"client_id": client.ClientID,
		"user_code": userCode,
	})

	return &models.DeviceAuthResponse{
		DeviceCode:              deviceCodePlain,
		UserCode:                userCode,
		VerificationURI:         verificationURI,
		VerificationURIComplete: verificationURIComplete,
		ExpiresIn:               int(deviceCodeTTL.Seconds()),
		Interval:                deviceCodePollingInterval,
	}, nil
}

func (s *OAuthProviderService) PollDeviceToken(ctx context.Context, req *models.TokenRequest) (*models.TokenResponse, error) {
	if req.DeviceCode == nil || *req.DeviceCode == "" {
		return nil, fmt.Errorf("%w: device_code is required", ErrInvalidRequest)
	}

	deviceCodeHash := s.hashToken(*req.DeviceCode)
	deviceCode, err := s.repo.GetDeviceCode(ctx, deviceCodeHash)
	if err != nil {
		return nil, ErrInvalidGrant
	}

	client := deviceCode.Client
	if client == nil {
		client, err = s.repo.GetClientByClientID(ctx, req.ClientID)
		if err != nil {
			return nil, ErrInvalidClient
		}
	}

	if client.ClientID != req.ClientID {
		return nil, ErrInvalidClient
	}

	if deviceCode.IsExpired() {
		return nil, ErrExpiredToken
	}

	switch deviceCode.Status {
	case models.DeviceCodeStatusPending:
		return nil, ErrAuthorizationPending
	case models.DeviceCodeStatusDenied:
		return nil, ErrAccessDenied
	case models.DeviceCodeStatusAuthorized:
		if deviceCode.UserID == nil {
			return nil, ErrServerError
		}

		user, err := s.userRepo.GetByID(ctx, *deviceCode.UserID, nil, UserGetWithRoles())
		if err != nil {
			return nil, ErrServerError
		}

		scopes := s.parseScopes(deviceCode.Scope)

		s.logAudit(ctx, deviceCode.UserID, "oauth_device_authorized", "success", map[string]interface{}{
			"client_id": client.ClientID,
		})

		response, err := s.generateTokens(ctx, client, deviceCode.UserID, user, scopes, nil)
		if err != nil {
			return nil, err
		}

		// Create session for the device authorization using SessionService
		if s.sessionService != nil && response.RefreshToken != "" {
			s.sessionService.CreateSessionNonFatal(ctx, SessionCreationParams{
				UserID:          *deviceCode.UserID,
				TokenHash:       utils.HashToken(response.RefreshToken),
				AccessTokenHash: utils.HashToken(response.AccessToken),
				IPAddress:       req.IPAddress,
				UserAgent:       req.UserAgent,
				ExpiresAt:       time.Now().Add(time.Duration(client.RefreshTokenTTL) * time.Second),
			})
		}

		s.logger.Info("device token issued", map[string]interface{}{
			"user_id":   deviceCode.UserID.String(),
			"client_id": client.ClientID,
		})

		return response, nil
	default:
		return nil, ErrServerError
	}
}

func (s *OAuthProviderService) ApproveDeviceCode(ctx context.Context, userID uuid.UUID, userCode string, approve bool) error {
	deviceCode, err := s.repo.GetDeviceCodeByUserCode(ctx, userCode)
	if err != nil {
		return fmt.Errorf("device code not found")
	}

	if deviceCode.IsExpired() {
		return fmt.Errorf("device code expired")
	}

	if deviceCode.Status != models.DeviceCodeStatusPending {
		return fmt.Errorf("device code already processed")
	}

	var status models.DeviceCodeStatus
	var userIDPtr *uuid.UUID

	if approve {
		status = models.DeviceCodeStatusAuthorized
		userIDPtr = &userID
	} else {
		status = models.DeviceCodeStatusDenied
	}

	if err := s.repo.UpdateDeviceCodeStatus(ctx, deviceCode.ID, status, userIDPtr); err != nil {
		s.logger.Error("failed to update device code status", map[string]interface{}{
			"error":     err.Error(),
			"user_code": userCode,
		})
		return fmt.Errorf("failed to update device code: %w", err)
	}

	action := "oauth_device_denied"
	if approve {
		action = "oauth_device_approved"
	}

	s.logAudit(ctx, &userID, action, "success", map[string]interface{}{
		"user_code": userCode,
		"client_id": deviceCode.ClientID.String(),
	})

	return nil
}

func (s *OAuthProviderService) IntrospectToken(ctx context.Context, token, tokenTypeHint string, clientID *string) (*models.IntrospectionResponse, error) {
	tokenHash := s.hashToken(token)

	if tokenTypeHint == "" || tokenTypeHint == "access_token" {
		accessToken, err := s.repo.GetAccessToken(ctx, tokenHash)
		if err == nil {
			return s.buildIntrospectionResponse(accessToken, nil), nil
		}
	}

	if tokenTypeHint == "" || tokenTypeHint == "refresh_token" {
		refreshToken, err := s.repo.GetRefreshToken(ctx, tokenHash)
		if err == nil {
			return s.buildIntrospectionResponse(nil, refreshToken), nil
		}
	}

	return &models.IntrospectionResponse{Active: false}, nil
}

func (s *OAuthProviderService) RevokeToken(ctx context.Context, token, tokenTypeHint string, clientID *string) error {
	tokenHash := s.hashToken(token)

	if tokenTypeHint == "" || tokenTypeHint == "access_token" {
		err := s.repo.RevokeAccessToken(ctx, tokenHash)
		if err == nil {
			// Also revoke associated session if exists using SessionService
			if s.sessionService != nil {
				s.sessionService.RevokeSessionByTokenHash(ctx, tokenHash)
			}
			s.logger.Info("access token revoked", map[string]interface{}{
				"token_type": "access_token",
			})
			return nil
		}
	}

	if tokenTypeHint == "" || tokenTypeHint == "refresh_token" {
		err := s.repo.RevokeRefreshToken(ctx, tokenHash)
		if err == nil {
			// Also revoke associated session if exists using SessionService
			if s.sessionService != nil {
				s.sessionService.RevokeSessionByTokenHash(ctx, tokenHash)
			}
			s.logger.Info("refresh token revoked", map[string]interface{}{
				"token_type": "refresh_token",
			})
			return nil
		}
	}

	return nil
}

func (s *OAuthProviderService) GetUserInfo(ctx context.Context, accessToken string) (*models.UserInfoResponse, error) {
	tokenHash := s.hashToken(accessToken)
	tokenRecord, err := s.repo.GetAccessToken(ctx, tokenHash)
	if err != nil {
		claims, err := s.oidcJWT.ValidateOAuthAccessToken(accessToken)
		if err != nil {
			return nil, ErrInvalidGrant
		}

		if claims.Subject == "" {
			return nil, fmt.Errorf("token has no subject")
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return nil, ErrInvalidGrant
		}

		user, err := s.userRepo.GetByID(ctx, userID, nil, UserGetWithRoles())
		if err != nil {
			return nil, ErrServerError
		}

		scopes := jwt.SplitScopes(claims.Scope)
		return s.buildUserInfoResponse(user, scopes), nil
	}

	if !tokenRecord.IsValid() {
		return nil, ErrInvalidGrant
	}

	if tokenRecord.UserID == nil {
		return nil, fmt.Errorf("token has no associated user")
	}

	user := tokenRecord.User
	if user == nil {
		user, err = s.userRepo.GetByID(ctx, *tokenRecord.UserID, nil, UserGetWithRoles())
		if err != nil {
			return nil, ErrServerError
		}
	}

	scopes := s.parseScopes(tokenRecord.Scope)
	return s.buildUserInfoResponse(user, scopes), nil
}

func (s *OAuthProviderService) GetDiscoveryDocument() *models.OIDCDiscoveryDocument {
	return &models.OIDCDiscoveryDocument{
		Issuer:                      s.issuer,
		AuthorizationEndpoint:       fmt.Sprintf("%s/oauth2/authorize", s.baseURL),
		TokenEndpoint:               fmt.Sprintf("%s/oauth2/token", s.baseURL),
		UserInfoEndpoint:            fmt.Sprintf("%s/oauth2/userinfo", s.baseURL),
		JwksURI:                     fmt.Sprintf("%s/.well-known/jwks.json", s.baseURL),
		RevocationEndpoint:          fmt.Sprintf("%s/oauth2/revoke", s.baseURL),
		IntrospectionEndpoint:       fmt.Sprintf("%s/oauth2/introspect", s.baseURL),
		DeviceAuthorizationEndpoint: fmt.Sprintf("%s/oauth2/device/code", s.baseURL),
		EndSessionEndpoint:          fmt.Sprintf("%s/oauth2/logout", s.baseURL),
		ScopesSupported: []string{
			models.ScopeOpenID,
			models.ScopeProfile,
			models.ScopeEmail,
			models.ScopePhone,
			models.ScopeAddress,
			models.ScopeOfflineAccess,
		},
		ResponseTypesSupported:            []string{"code"},
		ResponseModesSupported:            []string{"query", "fragment"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token", "client_credentials", "urn:ietf:params:oauth:grant-type:device_code"},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_basic", "client_secret_post", "none"},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256", "ES256"},
		CodeChallengeMethodsSupported:     []string{"plain", "S256"},
		ClaimsSupported:                   []string{"sub", "iss", "aud", "exp", "iat", "name", "email", "email_verified", "phone_number", "phone_number_verified", "picture", "preferred_username"},
	}
}

func (s *OAuthProviderService) GetJWKS() *models.JWKSDocument {
	jwks := s.keyManager.GetJWKS()

	result := &models.JWKSDocument{
		Keys: make([]models.JWK, 0, len(jwks.Keys)),
	}

	for _, key := range jwks.Keys {
		jwk := models.JWK{
			KeyType:   key.KTY,
			Use:       key.Use,
			Algorithm: key.Alg,
			KeyID:     key.KID,
			N:         key.N,
			E:         key.E,
			CRV:       key.Crv,
			X:         key.X,
			Y:         key.Y,
		}
		result.Keys = append(result.Keys, jwk)
	}

	return result
}

func (s *OAuthProviderService) GetConsentInfo(ctx context.Context, clientID string, scopes []string) (*ConsentInfo, error) {
	client, err := s.repo.GetClientByClientID(ctx, clientID)
	if err != nil {
		return nil, ErrInvalidClient
	}

	scopeInfos := make([]ScopeInfo, 0, len(scopes))
	for _, scopeName := range scopes {
		scope, err := s.repo.GetScopeByName(ctx, scopeName)
		if err != nil {
			scopeInfos = append(scopeInfos, ScopeInfo{
				Name:        scopeName,
				DisplayName: scopeName,
				Description: "",
			})
			continue
		}
		scopeInfos = append(scopeInfos, ScopeInfo{
			Name:        scope.Name,
			DisplayName: scope.DisplayName,
			Description: scope.Description,
		})
	}

	return &ConsentInfo{
		Client:          client,
		RequestedScopes: scopeInfos,
	}, nil
}

func (s *OAuthProviderService) GrantConsent(ctx context.Context, userID uuid.UUID, clientID string, scopes []string) error {
	client, err := s.repo.GetClientByClientID(ctx, clientID)
	if err != nil {
		return ErrInvalidClient
	}

	consent := &models.UserConsent{
		ID:       uuid.New(),
		UserID:   userID,
		ClientID: client.ID,
		Scopes:   scopes,
	}

	if err := s.repo.CreateOrUpdateConsent(ctx, consent); err != nil {
		s.logger.Error("failed to grant consent", map[string]interface{}{
			"error":     err.Error(),
			"user_id":   userID.String(),
			"client_id": clientID,
		})
		return fmt.Errorf("failed to grant consent: %w", err)
	}

	s.logAudit(ctx, &userID, "oauth_consent_granted", "success", map[string]interface{}{
		"client_id": clientID,
		"scopes":    scopes,
	})

	return nil
}

func (s *OAuthProviderService) RevokeConsent(ctx context.Context, userID, clientID uuid.UUID) error {
	if err := s.repo.RevokeAllUserAccessTokens(ctx, userID, clientID); err != nil {
		s.logger.Warn("failed to revoke access tokens during consent revocation", map[string]interface{}{
			"error": err.Error(),
		})
	}

	if err := s.repo.RevokeAllUserRefreshTokens(ctx, userID, clientID); err != nil {
		s.logger.Warn("failed to revoke refresh tokens during consent revocation", map[string]interface{}{
			"error": err.Error(),
		})
	}

	if err := s.repo.RevokeConsent(ctx, userID, clientID); err != nil {
		s.logger.Error("failed to revoke consent", map[string]interface{}{
			"error":     err.Error(),
			"user_id":   userID.String(),
			"client_id": clientID.String(),
		})
		return fmt.Errorf("failed to revoke consent: %w", err)
	}

	s.logAudit(ctx, &userID, "oauth_consent_revoked", "success", map[string]interface{}{
		"client_id": clientID.String(),
	})

	return nil
}

func (s *OAuthProviderService) ListUserConsents(ctx context.Context, userID uuid.UUID) ([]*models.UserConsent, error) {
	consents, err := s.repo.ListUserConsents(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user consents: %w", err)
	}
	return consents, nil
}

func (s *OAuthProviderService) generateClientID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return clientIDPrefix + base64.RawURLEncoding.EncodeToString(bytes)
}

func (s *OAuthProviderService) generateClientSecret() (plain string, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	plain = clientSecretPrefix + base64.RawURLEncoding.EncodeToString(bytes)

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCostForClientSecret)
	if err != nil {
		return "", "", err
	}

	return plain, string(hashBytes), nil
}

func (s *OAuthProviderService) generateAuthorizationCode() (plain string, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	plain = base64.RawURLEncoding.EncodeToString(bytes)
	hash = s.hashToken(plain)

	return plain, hash, nil
}

func (s *OAuthProviderService) generateToken() (plain string, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	plain = base64.RawURLEncoding.EncodeToString(bytes)
	hash = s.hashToken(plain)

	return plain, hash, nil
}

func (s *OAuthProviderService) generateUserCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	bytes := make([]byte, 8)
	rand.Read(bytes)

	code := make([]byte, 8)
	for i := range code {
		code[i] = charset[int(bytes[i])%len(charset)]
	}

	return string(code[:4]) + "-" + string(code[4:])
}

func (s *OAuthProviderService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *OAuthProviderService) validateRedirectURI(uri string, allowedURIs []string) bool {
	for _, allowed := range allowedURIs {
		if uri == allowed {
			return true
		}
	}
	return false
}

func (s *OAuthProviderService) validateScopes(requested []string, allowed []string) error {
	if len(requested) == 0 {
		return nil
	}

	allowedSet := make(map[string]bool)
	for _, scope := range allowed {
		allowedSet[scope] = true
	}

	for _, scope := range requested {
		if !allowedSet[scope] {
			return fmt.Errorf("%w: %s", ErrInvalidScope, scope)
		}
	}

	return nil
}

func (s *OAuthProviderService) parseScopes(scope string) []string {
	if scope == "" {
		return []string{}
	}
	scopes := strings.Split(scope, " ")
	result := make([]string, 0, len(scopes))
	for _, sc := range scopes {
		sc = strings.TrimSpace(sc)
		if sc != "" {
			result = append(result, sc)
		}
	}
	return result
}

func (s *OAuthProviderService) hasGrantType(grantTypes []string, grantType string) bool {
	for _, gt := range grantTypes {
		if gt == grantType {
			return true
		}
	}
	return false
}

func (s *OAuthProviderService) hasAllScopes(granted, requested []string) bool {
	grantedSet := make(map[string]bool)
	for _, scope := range granted {
		grantedSet[scope] = true
	}

	for _, scope := range requested {
		if !grantedSet[scope] {
			return false
		}
	}
	return true
}

func (s *OAuthProviderService) buildErrorRedirect(redirectURI, errorCode, errorDesc, state string) string {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return redirectURI
	}

	q := u.Query()
	q.Set("error", errorCode)
	if errorDesc != "" {
		q.Set("error_description", errorDesc)
	}
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

func (s *OAuthProviderService) generateTokens(ctx context.Context, client *models.OAuthClient, userID *uuid.UUID, user *models.User, scopes []string, nonce *string) (*models.TokenResponse, error) {
	scope := strings.Join(scopes, " ")

	var roles []string
	if user != nil && len(user.Roles) > 0 {
		for _, role := range user.Roles {
			roles = append(roles, role.Name)
		}
	}

	accessToken, err := s.oidcJWT.GenerateOAuthAccessToken(userID, client.ClientID, scope, roles, time.Duration(client.AccessTokenTTL)*time.Second)
	if err != nil {
		s.logger.Error("failed to generate access token", map[string]interface{}{"error": err.Error()})
		return nil, ErrServerError
	}

	accessTokenHash := s.hashToken(accessToken)
	accessTokenRecord := &models.OAuthAccessToken{
		ID:        uuid.New(),
		TokenHash: accessTokenHash,
		ClientID:  client.ID,
		UserID:    userID,
		Scope:     scope,
		IsActive:  true,
		ExpiresAt: time.Now().Add(time.Duration(client.AccessTokenTTL) * time.Second),
	}

	if err := s.repo.CreateAccessToken(ctx, accessTokenRecord); err != nil {
		s.logger.Error("failed to store access token", map[string]interface{}{"error": err.Error()})
		return nil, ErrServerError
	}

	response := &models.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   client.AccessTokenTTL,
		Scope:       scope,
	}

	if userID != nil && s.hasGrantType(client.AllowedGrantTypes, string(models.GrantTypeRefreshToken)) {
		refreshTokenPlain, refreshTokenHash, err := s.generateToken()
		if err != nil {
			s.logger.Error("failed to generate refresh token", map[string]interface{}{"error": err.Error()})
			return nil, ErrServerError
		}

		refreshTokenRecord := &models.OAuthRefreshToken{
			ID:            uuid.New(),
			TokenHash:     refreshTokenHash,
			AccessTokenID: &accessTokenRecord.ID,
			ClientID:      client.ID,
			UserID:        *userID,
			Scope:         scope,
			IsActive:      true,
			ExpiresAt:     time.Now().Add(time.Duration(client.RefreshTokenTTL) * time.Second),
		}

		if err := s.repo.CreateRefreshToken(ctx, refreshTokenRecord); err != nil {
			s.logger.Error("failed to store refresh token", map[string]interface{}{"error": err.Error()})
			return nil, ErrServerError
		}

		response.RefreshToken = refreshTokenPlain
	}

	if userID != nil && user != nil && s.containsScope(scopes, models.ScopeOpenID) {
		nonceStr := ""
		if nonce != nil {
			nonceStr = *nonce
		}

		idToken, err := s.oidcJWT.GenerateIDToken(*userID, client.ClientID, nonceStr, scopes, user, time.Duration(client.IDTokenTTL)*time.Second)
		if err != nil {
			s.logger.Error("failed to generate ID token", map[string]interface{}{"error": err.Error()})
			return nil, ErrServerError
		}

		response.IDToken = idToken
	}
	s.logAudit(ctx, nil, "oauth_client_credentials", "success", map[string]interface{}{
		"client_id": client.ClientID,
		"scope":     scope,
	})
	return response, nil
}

func (s *OAuthProviderService) containsScope(scopes []string, target string) bool {
	for _, scope := range scopes {
		if scope == target {
			return true
		}
	}
	return false
}

func (s *OAuthProviderService) buildIntrospectionResponse(accessToken *models.OAuthAccessToken, refreshToken *models.OAuthRefreshToken) *models.IntrospectionResponse {
	if accessToken != nil {
		if !accessToken.IsValid() {
			return &models.IntrospectionResponse{Active: false}
		}

		response := &models.IntrospectionResponse{
			Active:    true,
			Scope:     accessToken.Scope,
			TokenType: "Bearer",
			ExpiresAt: accessToken.ExpiresAt.Unix(),
			IssuedAt:  accessToken.CreatedAt.Unix(),
			NotBefore: accessToken.CreatedAt.Unix(),
			Issuer:    s.issuer,
		}

		if accessToken.Client != nil {
			response.ClientID = accessToken.Client.ClientID
		}

		if accessToken.UserID != nil {
			response.Subject = accessToken.UserID.String()
			if accessToken.User != nil {
				response.Username = accessToken.User.Username
			}
		}

		return response
	}

	if refreshToken != nil {
		if !refreshToken.IsValid() {
			return &models.IntrospectionResponse{Active: false}
		}

		response := &models.IntrospectionResponse{
			Active:    true,
			Scope:     refreshToken.Scope,
			TokenType: "refresh_token",
			ExpiresAt: refreshToken.ExpiresAt.Unix(),
			IssuedAt:  refreshToken.CreatedAt.Unix(),
			NotBefore: refreshToken.CreatedAt.Unix(),
			Subject:   refreshToken.UserID.String(),
			Issuer:    s.issuer,
		}

		if refreshToken.Client != nil {
			response.ClientID = refreshToken.Client.ClientID
		}

		if refreshToken.User != nil {
			response.Username = refreshToken.User.Username
		}

		return response
	}

	return &models.IntrospectionResponse{Active: false}
}

func (s *OAuthProviderService) buildUserInfoResponse(user *models.User, scopes []string) *models.UserInfoResponse {
	response := &models.UserInfoResponse{
		Subject: user.ID.String(),
	}

	scopeSet := make(map[string]bool)
	for _, scope := range scopes {
		scopeSet[scope] = true
	}

	if scopeSet[models.ScopeProfile] {
		if user.FullName != "" {
			response.Name = &user.FullName
		}
		if user.Username != "" {
			response.PreferredUsername = &user.Username
		}
		if user.ProfilePictureURL != "" {
			response.Picture = &user.ProfilePictureURL
		}
		if !user.UpdatedAt.IsZero() {
			updatedAt := user.UpdatedAt.Unix()
			response.UpdatedAt = &updatedAt
		}
	}

	if scopeSet[models.ScopeEmail] {
		if user.Email != "" {
			response.Email = &user.Email
		}
		response.EmailVerified = &user.EmailVerified
	}

	if scopeSet[models.ScopePhone] {
		if user.Phone != nil && *user.Phone != "" {
			response.PhoneNumber = user.Phone
		}
		response.PhoneNumberVerified = &user.PhoneVerified
	}

	return response
}

func (s *OAuthProviderService) ListScopes(ctx context.Context) ([]*models.OAuthScope, error) {
	scopes, err := s.repo.ListScopes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list scopes: %w", err)
	}
	return scopes, nil
}

func (s *OAuthProviderService) CreateScope(ctx context.Context, scope *models.OAuthScope) error {
	if err := s.repo.CreateScope(ctx, scope); err != nil {
		s.logger.Error("failed to create scope", map[string]interface{}{
			"error": err.Error(),
			"name":  scope.Name,
		})
		return fmt.Errorf("failed to create scope: %w", err)
	}

	s.logger.Info("oauth scope created", map[string]interface{}{
		"name":         scope.Name,
		"display_name": scope.DisplayName,
	})

	return nil
}

func (s *OAuthProviderService) DeleteScope(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteScope(ctx, id); err != nil {
		s.logger.Error("failed to delete scope", map[string]interface{}{
			"error":    err.Error(),
			"scope_id": id.String(),
		})
		return fmt.Errorf("failed to delete scope: %w", err)
	}

	s.logger.Info("oauth scope deleted", map[string]interface{}{
		"scope_id": id.String(),
	})

	return nil
}

func (s *OAuthProviderService) ListClientConsents(ctx context.Context, clientID uuid.UUID) ([]*models.UserConsent, error) {
	consents, err := s.repo.ListClientConsents(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to list client consents: %w", err)
	}
	return consents, nil
}

func (s *OAuthProviderService) logAudit(ctx context.Context, userID *uuid.UUID, action, status string, details map[string]interface{}) {
	var detailsJSON []byte
	if details != nil {
		detailsJSON, _ = json.Marshal(details)
	}

	log := &models.AuditLog{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    action,
		Status:    status,
		Details:   detailsJSON,
		CreatedAt: time.Now(),
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		s.logger.Warn("failed to create audit log", map[string]interface{}{
			"error":  err.Error(),
			"action": action,
		})
	}
}

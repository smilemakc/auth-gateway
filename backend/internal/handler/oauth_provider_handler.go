package handler

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

type OAuthProviderHandler struct {
	service *service.OAuthProviderService
	logger  *logger.Logger
}

func NewOAuthProviderHandler(service *service.OAuthProviderService, logger *logger.Logger) *OAuthProviderHandler {
	return &OAuthProviderHandler{
		service: service,
		logger:  logger,
	}
}

// Authorize handles OAuth 2.0 authorization requests
// @Summary OAuth 2.0 Authorization
// @Description Initiates OAuth 2.0 authorization flow. Redirects user to login if not authenticated, then to consent page if required
// @Tags OAuth Provider
// @Accept json
// @Produce json
// @Param response_type query string true "Response type (code or token)" Enums(code, token)
// @Param client_id query string true "Client ID"
// @Param redirect_uri query string true "Redirect URI"
// @Param scope query string true "Requested scopes (space-separated)"
// @Param state query string true "State parameter for CSRF protection"
// @Param nonce query string false "Nonce for ID token"
// @Param code_challenge query string false "PKCE code challenge"
// @Param code_challenge_method query string false "PKCE code challenge method (S256 or plain)"
// @Param prompt query string false "Prompt behavior (none, login, consent, select_account)"
// @Success 302 {string} string "Redirect to callback with authorization code"
// @Failure 302 {string} string "Redirect with error"
// @Router /oauth/authorize [get]
func (h *OAuthProviderHandler) Authorize(c *gin.Context) {
	var req models.AuthorizeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.redirectError(c, req.RedirectURI, "invalid_request", err.Error(), req.State)
		return
	}

	userID, authenticated := h.getUserIDFromContext(c)
	if !authenticated {
		loginURL := fmt.Sprintf("/login?return_to=%s", url.QueryEscape(c.Request.URL.String()))
		c.Redirect(http.StatusTemporaryRedirect, loginURL)
		return
	}

	authResp, err := h.service.Authorize(c.Request.Context(), &req, userID)
	if err != nil {
		if errors.Is(err, service.ErrConsentRequired) {
			consentURL := h.buildConsentURL(c.Request.URL.Query())
			c.Redirect(http.StatusTemporaryRedirect, consentURL)
			return
		}

		errorCode := h.mapErrorToOAuthCode(err)
		errorDesc := err.Error()
		h.redirectError(c, req.RedirectURI, errorCode, errorDesc, req.State)
		return
	}

	redirectURL := h.buildSuccessRedirect(req.RedirectURI, authResp.Code, authResp.State)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// Token handles OAuth 2.0 token requests
// @Summary OAuth 2.0 Token
// @Description Exchange authorization code for tokens, refresh tokens, or obtain client credentials tokens
// @Tags OAuth Provider
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param grant_type formData string true "Grant type" Enums(authorization_code, refresh_token, client_credentials, urn:ietf:params:oauth:grant-type:device_code)
// @Param code formData string false "Authorization code (for authorization_code grant)"
// @Param redirect_uri formData string false "Redirect URI (for authorization_code grant)"
// @Param client_id formData string true "Client ID"
// @Param client_secret formData string false "Client secret"
// @Param refresh_token formData string false "Refresh token (for refresh_token grant)"
// @Param scope formData string false "Requested scopes"
// @Param code_verifier formData string false "PKCE code verifier"
// @Param device_code formData string false "Device code (for device_code grant)"
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} map[string]string "error and error_description"
// @Failure 401 {object} map[string]string "invalid_client"
// @Router /oauth/token [post]
func (h *OAuthProviderHandler) Token(c *gin.Context) {
	var req models.TokenRequest

	if err := c.ShouldBind(&req); err != nil {
		h.oauthError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	clientID, clientSecret := h.extractClientCredentials(c)
	if clientID != "" {
		req.ClientID = clientID
		if clientSecret != "" {
			req.ClientSecret = &clientSecret
		}
	}

	// Add session context for tracking
	req.IPAddress = utils.GetClientIP(c)
	req.UserAgent = c.Request.UserAgent()

	var resp *models.TokenResponse
	var err error

	switch req.GrantType {
	case string(models.GrantTypeAuthorizationCode):
		resp, err = h.service.ExchangeCode(c.Request.Context(), &req)
	case string(models.GrantTypeRefreshToken):
		resp, err = h.service.RefreshToken(c.Request.Context(), &req)
	case string(models.GrantTypeClientCredentials):
		resp, err = h.service.ClientCredentialsGrant(c.Request.Context(), &req)
	case string(models.GrantTypeDeviceCode):
		resp, err = h.service.PollDeviceToken(c.Request.Context(), &req)
	default:
		h.oauthError(c, http.StatusBadRequest, "unsupported_grant_type", fmt.Sprintf("Grant type '%s' is not supported", req.GrantType))
		return
	}

	if err != nil {
		errorCode := h.mapErrorToOAuthCode(err)
		errorDesc := err.Error()

		if errors.Is(err, service.ErrAuthorizationPending) {
			h.oauthError(c, http.StatusBadRequest, errorCode, "The authorization request is still pending")
			return
		}
		if errors.Is(err, service.ErrSlowDown) {
			h.oauthError(c, http.StatusBadRequest, errorCode, "Polling too frequently, slow down")
			return
		}
		if errors.Is(err, service.ErrExpiredToken) {
			h.oauthError(c, http.StatusBadRequest, errorCode, "The device code has expired")
			return
		}

		statusCode := http.StatusBadRequest
		if errors.Is(err, service.ErrInvalidClient) {
			statusCode = http.StatusUnauthorized
		}

		h.oauthError(c, statusCode, errorCode, errorDesc)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Introspect handles OAuth 2.0 token introspection (RFC 7662)
// @Summary Token Introspection
// @Description Introspect an access or refresh token to determine its validity and metadata
// @Tags OAuth Provider
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param token formData string true "Token to introspect"
// @Param token_type_hint formData string false "Token type hint (access_token or refresh_token)"
// @Param client_id formData string true "Client ID"
// @Param client_secret formData string false "Client secret"
// @Success 200 {object} models.IntrospectionResponse
// @Failure 401 {object} map[string]string "invalid_client"
// @Router /oauth/introspect [post]
func (h *OAuthProviderHandler) Introspect(c *gin.Context) {
	var req models.IntrospectionRequest
	if err := c.ShouldBind(&req); err != nil {
		h.oauthError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	clientID, clientSecret := h.extractClientCredentials(c)
	if clientID != "" {
		req.ClientID = clientID
		if clientSecret != "" {
			req.ClientSecret = &clientSecret
		}
	}

	if req.ClientSecret != nil {
		_, err := h.service.ValidateClientCredentials(c.Request.Context(), req.ClientID, *req.ClientSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
			return
		}
	}

	tokenTypeHint := ""
	if req.TokenTypeHint != nil {
		tokenTypeHint = *req.TokenTypeHint
	}

	resp, err := h.service.IntrospectToken(c.Request.Context(), req.Token, tokenTypeHint, &req.ClientID)
	if err != nil {
		h.logger.Error("token introspection failed", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusOK, models.IntrospectionResponse{Active: false})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Revoke handles OAuth 2.0 token revocation (RFC 7009)
// @Summary Token Revocation
// @Description Revoke an access or refresh token
// @Tags OAuth Provider
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param token formData string true "Token to revoke"
// @Param token_type_hint formData string false "Token type hint (access_token or refresh_token)"
// @Param client_id formData string true "Client ID"
// @Param client_secret formData string false "Client secret"
// @Success 200 {object} map[string]string "Empty response on success"
// @Router /oauth/revoke [post]
func (h *OAuthProviderHandler) Revoke(c *gin.Context) {
	var req models.RevocationRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	clientID, clientSecret := h.extractClientCredentials(c)
	if clientID != "" {
		req.ClientID = clientID
		if clientSecret != "" {
			req.ClientSecret = &clientSecret
		}
	}

	if req.ClientSecret != nil {
		_, err := h.service.ValidateClientCredentials(c.Request.Context(), req.ClientID, *req.ClientSecret)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
	}

	tokenTypeHint := ""
	if req.TokenTypeHint != nil {
		tokenTypeHint = *req.TokenTypeHint
	}

	_ = h.service.RevokeToken(c.Request.Context(), req.Token, tokenTypeHint, &req.ClientID)

	c.JSON(http.StatusOK, gin.H{})
}

// UserInfo handles OIDC UserInfo requests
// @Summary OIDC UserInfo
// @Description Get user information based on access token (OIDC UserInfo endpoint)
// @Tags OAuth Provider
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserInfoResponse
// @Failure 401 {object} map[string]string "invalid_token"
// @Router /oauth/userinfo [get]
func (h *OAuthProviderHandler) UserInfo(c *gin.Context) {
	accessToken := h.extractBearerToken(c)
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "No access token provided",
		})
		return
	}

	userInfo, err := h.service.GetUserInfo(c.Request.Context(), accessToken)
	if err != nil {
		h.logger.Error("failed to get user info", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "Invalid or expired access token",
		})
		return
	}

	c.JSON(http.StatusOK, userInfo)
}

// Discovery handles OIDC Discovery requests
// @Summary OIDC Discovery
// @Description Get OpenID Connect discovery document (/.well-known/openid-configuration)
// @Tags OAuth Provider - Discovery
// @Produce json
// @Success 200 {object} models.OIDCDiscoveryDocument
// @Router /.well-known/openid-configuration [get]
func (h *OAuthProviderHandler) Discovery(c *gin.Context) {
	doc := h.service.GetDiscoveryDocument()
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, doc)
}

// JWKS handles JSON Web Key Set requests
// @Summary JWKS
// @Description Get JSON Web Key Set for token validation (/.well-known/jwks.json)
// @Tags OAuth Provider - Discovery
// @Produce json
// @Success 200 {object} models.JWKSDocument
// @Router /.well-known/jwks.json [get]
func (h *OAuthProviderHandler) JWKS(c *gin.Context) {
	jwks := h.service.GetJWKS()
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, jwks)
}

// DeviceCode handles device authorization requests (RFC 8628)
// @Summary Device Authorization Request
// @Description Initiate device authorization grant flow
// @Tags OAuth Provider - Device Flow
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param client_id formData string true "Client ID"
// @Param scope formData string false "Requested scopes"
// @Success 200 {object} models.DeviceAuthResponse
// @Failure 400 {object} map[string]string "error and error_description"
// @Router /oauth/device/code [post]
func (h *OAuthProviderHandler) DeviceCode(c *gin.Context) {
	var req models.DeviceAuthRequest
	if err := c.ShouldBind(&req); err != nil {
		h.oauthError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	resp, err := h.service.DeviceAuthorization(c.Request.Context(), &req)
	if err != nil {
		errorCode := h.mapErrorToOAuthCode(err)
		errorDesc := err.Error()
		h.oauthError(c, http.StatusBadRequest, errorCode, errorDesc)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeviceToken handles device token polling (RFC 8628)
// @Summary Device Token Polling
// @Description Poll for tokens after user authorization (device flow)
// @Tags OAuth Provider - Device Flow
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param grant_type formData string true "Grant type (urn:ietf:params:oauth:grant-type:device_code)"
// @Param device_code formData string true "Device code from authorization request"
// @Param client_id formData string true "Client ID"
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} map[string]string "authorization_pending, slow_down, or expired_token"
// @Router /oauth/device/token [post]
func (h *OAuthProviderHandler) DeviceToken(c *gin.Context) {
	h.Token(c)
}

// DeviceVerification renders the device verification page
// @Summary Device Verification Page
// @Description HTML page where users enter the device code to authorize
// @Tags OAuth Provider - Device Flow
// @Produce html
// @Param user_code query string false "Pre-filled user code"
// @Success 200 {string} string "HTML verification page"
// @Router /oauth/device [get]
func (h *OAuthProviderHandler) DeviceVerification(c *gin.Context) {
	userCode := c.Query("user_code")

	html := h.renderDeviceVerificationPage(userCode)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// DeviceApprove handles device code approval
// @Summary Approve Device Code
// @Description Approve or deny a device code (requires authentication)
// @Tags OAuth Provider - Device Flow
// @Security BearerAuth
// @Accept application/x-www-form-urlencoded
// @Produce html
// @Param user_code formData string true "User code from device"
// @Param approve formData string true "Approval decision (true or false)"
// @Success 200 {string} string "HTML success page"
// @Failure 400 {string} string "HTML error page"
// @Failure 302 {string} string "Redirect to login"
// @Router /oauth/device/approve [post]
func (h *OAuthProviderHandler) DeviceApprove(c *gin.Context) {
	userID, authenticated := h.getUserIDFromContext(c)
	if !authenticated {
		loginURL := fmt.Sprintf("/login?return_to=%s", url.QueryEscape(c.Request.URL.String()))
		c.Redirect(http.StatusTemporaryRedirect, loginURL)
		return
	}

	userCode := c.PostForm("user_code")
	approve := c.PostForm("approve") == "true"

	if userCode == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":   "Invalid Request",
			"message": "User code is required",
		})
		return
	}

	err := h.service.ApproveDeviceCode(c.Request.Context(), userID, userCode, approve)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":   "Approval Failed",
			"message": err.Error(),
		})
		return
	}

	message := "Device authorization denied"
	if approve {
		message = "Device authorized successfully! You can close this window."
	}

	html := h.renderDeviceSuccessPage(message)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ConsentPage renders the OAuth consent page
// @Summary OAuth Consent Page
// @Description HTML page for user to approve/deny OAuth authorization request
// @Tags OAuth Provider - Consent
// @Security BearerAuth
// @Produce html
// @Param client_id query string true "Client ID"
// @Param scope query string true "Requested scopes"
// @Param redirect_uri query string true "Redirect URI"
// @Param state query string false "State parameter"
// @Success 200 {string} string "HTML consent page"
// @Failure 302 {string} string "Redirect to login"
// @Failure 400 {string} string "HTML error page"
// @Router /oauth/consent [get]
func (h *OAuthProviderHandler) ConsentPage(c *gin.Context) {
	userID, authenticated := h.getUserIDFromContext(c)
	if !authenticated {
		loginURL := fmt.Sprintf("/login?return_to=%s", url.QueryEscape(c.Request.URL.String()))
		c.Redirect(http.StatusTemporaryRedirect, loginURL)
		return
	}

	clientID := c.Query("client_id")
	scope := c.Query("scope")
	redirectURI := c.Query("redirect_uri")
	// state is passed through query params directly to consent form

	if clientID == "" || scope == "" || redirectURI == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":   "Invalid Request",
			"message": "Missing required parameters",
		})
		return
	}

	scopes := strings.Split(scope, " ")
	consentInfo, err := h.service.GetConsentInfo(c.Request.Context(), clientID, scopes)
	if err != nil {
		h.logger.Error("failed to get consent info", map[string]interface{}{
			"error":     err.Error(),
			"user_id":   userID.String(),
			"client_id": clientID,
		})
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error":   "Server Error",
			"message": "Failed to load consent information",
		})
		return
	}

	html := h.renderConsentPage(consentInfo, c.Request.URL.Query())
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ConsentSubmit handles consent form submission
// @Summary Submit Consent Decision
// @Description Process user consent decision and redirect with authorization code
// @Tags OAuth Provider - Consent
// @Security BearerAuth
// @Accept application/x-www-form-urlencoded
// @Param client_id formData string true "Client ID"
// @Param redirect_uri formData string true "Redirect URI"
// @Param scope formData string true "Requested scopes"
// @Param state formData string true "State parameter"
// @Param approve formData string true "Approval decision (true or false)"
// @Param response_type formData string false "Response type"
// @Param nonce formData string false "Nonce"
// @Param code_challenge formData string false "PKCE code challenge"
// @Param code_challenge_method formData string false "PKCE code challenge method"
// @Success 302 {string} string "Redirect to callback with code or error"
// @Router /oauth/consent [post]
func (h *OAuthProviderHandler) ConsentSubmit(c *gin.Context) {
	userID, authenticated := h.getUserIDFromContext(c)
	if !authenticated {
		loginURL := fmt.Sprintf("/login?return_to=%s", url.QueryEscape(c.Request.URL.String()))
		c.Redirect(http.StatusTemporaryRedirect, loginURL)
		return
	}

	clientID := c.PostForm("client_id")
	redirectURI := c.PostForm("redirect_uri")
	scope := c.PostForm("scope")
	state := c.PostForm("state")
	approve := c.PostForm("approve") == "true"

	responseType := c.PostForm("response_type")
	nonce := c.PostForm("nonce")
	codeChallenge := c.PostForm("code_challenge")
	codeChallengeMethod := c.PostForm("code_challenge_method")

	if !approve {
		errorURL := h.buildErrorRedirect(redirectURI, "access_denied", "User denied consent", state)
		c.Redirect(http.StatusTemporaryRedirect, errorURL)
		return
	}

	scopes := strings.Split(scope, " ")
	err := h.service.GrantConsent(c.Request.Context(), userID, clientID, scopes)
	if err != nil {
		h.logger.Error("failed to grant consent", map[string]interface{}{
			"error":     err.Error(),
			"user_id":   userID.String(),
			"client_id": clientID,
		})
		errorURL := h.buildErrorRedirect(redirectURI, "server_error", "Failed to grant consent", state)
		c.Redirect(http.StatusTemporaryRedirect, errorURL)
		return
	}

	var req models.AuthorizeRequest
	req.ResponseType = responseType
	req.ClientID = clientID
	req.RedirectURI = redirectURI
	req.Scope = scope
	req.State = state

	if nonce != "" {
		req.Nonce = &nonce
	}
	if codeChallenge != "" {
		req.CodeChallenge = &codeChallenge
		if codeChallengeMethod != "" {
			req.CodeChallengeMethod = &codeChallengeMethod
		}
	}

	authResp, err := h.service.Authorize(c.Request.Context(), &req, userID)
	if err != nil {
		errorCode := h.mapErrorToOAuthCode(err)
		errorDesc := err.Error()
		errorURL := h.buildErrorRedirect(redirectURI, errorCode, errorDesc, state)
		c.Redirect(http.StatusTemporaryRedirect, errorURL)
		return
	}

	successURL := h.buildSuccessRedirect(redirectURI, authResp.Code, authResp.State)
	c.Redirect(http.StatusTemporaryRedirect, successURL)
}

func (h *OAuthProviderHandler) extractClientCredentials(c *gin.Context) (clientID, clientSecret string) {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Basic ") {
		payload := strings.TrimPrefix(authHeader, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err == nil {
			parts := strings.SplitN(string(decoded), ":", 2)
			if len(parts) == 2 {
				return parts[0], parts[1]
			}
		}
	}

	clientID = c.PostForm("client_id")
	clientSecret = c.PostForm("client_secret")
	return clientID, clientSecret
}

func (h *OAuthProviderHandler) extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

func (h *OAuthProviderHandler) oauthError(c *gin.Context, status int, errorCode, errorDesc string) {
	c.JSON(status, gin.H{
		"error":             errorCode,
		"error_description": errorDesc,
	})
}

func (h *OAuthProviderHandler) getUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userIDPtr, exists := utils.GetUserIDFromContext(c)
	if !exists || userIDPtr == nil {
		return uuid.Nil, false
	}
	return *userIDPtr, true
}

func (h *OAuthProviderHandler) mapErrorToOAuthCode(err error) string {
	switch err {
	case service.ErrInvalidClient:
		return "invalid_client"
	case service.ErrInvalidGrant:
		return "invalid_grant"
	case service.ErrInvalidScope:
		return "invalid_scope"
	case service.ErrInvalidRequest:
		return "invalid_request"
	case service.ErrUnauthorizedClient:
		return "unauthorized_client"
	case service.ErrAccessDenied:
		return "access_denied"
	case service.ErrUnsupportedGrantType:
		return "unsupported_grant_type"
	case service.ErrUnsupportedResponseType:
		return "unsupported_response_type"
	case service.ErrServerError:
		return "server_error"
	case service.ErrConsentRequired:
		return "consent_required"
	case service.ErrLoginRequired:
		return "login_required"
	case service.ErrAuthorizationPending:
		return "authorization_pending"
	case service.ErrSlowDown:
		return "slow_down"
	case service.ErrExpiredToken:
		return "expired_token"
	default:
		return "server_error"
	}
}

func (h *OAuthProviderHandler) redirectError(c *gin.Context, redirectURI, errorCode, errorDesc, state string) {
	if redirectURI == "" {
		h.oauthError(c, http.StatusBadRequest, errorCode, errorDesc)
		return
	}

	errorURL := h.buildErrorRedirect(redirectURI, errorCode, errorDesc, state)
	c.Redirect(http.StatusTemporaryRedirect, errorURL)
}

func (h *OAuthProviderHandler) buildErrorRedirect(redirectURI, errorCode, errorDesc, state string) string {
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

func (h *OAuthProviderHandler) buildSuccessRedirect(redirectURI, code, state string) string {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return redirectURI
	}

	q := u.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

func (h *OAuthProviderHandler) buildConsentURL(params url.Values) string {
	u := &url.URL{
		Path: "/oauth/consent",
	}
	u.RawQuery = params.Encode()
	return u.String()
}

func (h *OAuthProviderHandler) renderDeviceVerificationPage(prefilledCode string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Device Verification</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 500px; margin: 50px auto; padding: 20px; }
        .container { background: #f5f5f5; padding: 30px; border-radius: 8px; }
        h1 { color: #333; }
        input { width: 100%%; padding: 10px; margin: 10px 0; font-size: 16px; }
        button { width: 100%%; padding: 12px; background: #007bff; color: white; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; }
        button:hover { background: #0056b3; }
        .error { color: red; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Device Verification</h1>
        <p>Enter the code displayed on your device:</p>
        <form method="POST" action="/oauth/device/approve">
            <input type="text" name="user_code" placeholder="XXXX-XXXX" value="%s" required maxlength="9" style="text-transform: uppercase;">
            <input type="hidden" name="approve" value="true">
            <button type="submit">Verify</button>
        </form>
    </div>
</body>
</html>`, prefilledCode)
}

func (h *OAuthProviderHandler) renderDeviceSuccessPage(message string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Success</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 500px; margin: 50px auto; padding: 20px; text-align: center; }
        .container { background: #f5f5f5; padding: 30px; border-radius: 8px; }
        h1 { color: #28a745; }
        p { font-size: 18px; color: #333; }
    </style>
</head>
<body>
    <div class="container">
        <h1>✓ Success</h1>
        <p>%s</p>
    </div>
</body>
</html>`, message)
}

func (h *OAuthProviderHandler) renderConsentPage(info *service.ConsentInfo, params url.Values) string {
	scopesList := ""
	for _, scope := range info.RequestedScopes {
		scopesList += fmt.Sprintf(`<li><strong>%s</strong>: %s</li>`, scope.DisplayName, scope.Description)
	}

	hiddenFields := ""
	for key, values := range params {
		for _, value := range values {
			hiddenFields += fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, key, value)
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Authorization Request</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .container { background: #f5f5f5; padding: 30px; border-radius: 8px; }
        h1 { color: #333; }
        .client-info { background: white; padding: 15px; border-radius: 4px; margin: 20px 0; }
        .scopes { background: white; padding: 15px; border-radius: 4px; margin: 20px 0; }
        ul { list-style: none; padding: 0; }
        li { padding: 8px 0; border-bottom: 1px solid #eee; }
        .buttons { display: flex; gap: 10px; margin-top: 20px; }
        button { flex: 1; padding: 12px; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; }
        .approve { background: #28a745; color: white; }
        .approve:hover { background: #218838; }
        .deny { background: #dc3545; color: white; }
        .deny:hover { background: #c82333; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Authorization Request</h1>
        <div class="client-info">
            <h3>%s</h3>
            <p>%s</p>
        </div>
        <div class="scopes">
            <h3>This application is requesting access to:</h3>
            <ul>%s</ul>
        </div>
        <form method="POST" action="/oauth/consent">
            %s
            <div class="buttons">
                <button type="submit" name="approve" value="false" class="deny">Deny</button>
                <button type="submit" name="approve" value="true" class="approve">Allow</button>
            </div>
        </form>
    </div>
</body>
</html>`, info.Client.Name, info.Client.Description, scopesList, hiddenFields)
}

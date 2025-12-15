package models

import "time"

// OIDCDiscovery represents OpenID Connect discovery document (RFC 8414)
type OIDCDiscovery struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	RegistrationEndpoint              string   `json:"registration_endpoint,omitempty"`
	RevocationEndpoint                string   `json:"revocation_endpoint,omitempty"`
	IntrospectionEndpoint             string   `json:"introspection_endpoint,omitempty"`
	DeviceAuthorizationEndpoint       string   `json:"device_authorization_endpoint,omitempty"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported,omitempty"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`
	ClaimsSupported                   []string `json:"claims_supported,omitempty"`
	GrantTypesSupported               []string `json:"grant_types_supported,omitempty"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported,omitempty"`
}

// JWKS represents JSON Web Key Set (RFC 7517)
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key (RFC 7517)
type JWK struct {
	Kty string   `json:"kty"`           // Key type (e.g., "RSA", "EC")
	Use string   `json:"use,omitempty"` // "sig" or "enc"
	Kid string   `json:"kid,omitempty"` // Key ID
	Alg string   `json:"alg,omitempty"` // Algorithm
	N   string   `json:"n,omitempty"`   // RSA modulus
	E   string   `json:"e,omitempty"`   // RSA exponent
	X   string   `json:"x,omitempty"`   // EC x coordinate
	Y   string   `json:"y,omitempty"`   // EC y coordinate
	Crv string   `json:"crv,omitempty"` // EC curve
	X5c []string `json:"x5c,omitempty"` // X.509 certificate chain
}

// OAuthTokenResponse represents an OAuth 2.0 token response (RFC 6749)
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"` // Usually "Bearer"
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"` // For OIDC
}

// TokenIntrospectionResponse represents token introspection response (RFC 7662)
type TokenIntrospectionResponse struct {
	Active    bool   `json:"active"`
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Username  string `json:"username,omitempty"`
	TokenType string `json:"token_type,omitempty"`
	Exp       int64  `json:"exp,omitempty"`
	Iat       int64  `json:"iat,omitempty"`
	Nbf       int64  `json:"nbf,omitempty"`
	Sub       string `json:"sub,omitempty"`
	Aud       string `json:"aud,omitempty"`
	Iss       string `json:"iss,omitempty"`
	Jti       string `json:"jti,omitempty"`
}

// UserInfo represents OIDC UserInfo response
type UserInfo struct {
	Sub                 string `json:"sub"`
	Name                string `json:"name,omitempty"`
	GivenName           string `json:"given_name,omitempty"`
	FamilyName          string `json:"family_name,omitempty"`
	MiddleName          string `json:"middle_name,omitempty"`
	Nickname            string `json:"nickname,omitempty"`
	PreferredUsername   string `json:"preferred_username,omitempty"`
	Profile             string `json:"profile,omitempty"`
	Picture             string `json:"picture,omitempty"`
	Website             string `json:"website,omitempty"`
	Email               string `json:"email,omitempty"`
	EmailVerified       bool   `json:"email_verified,omitempty"`
	Gender              string `json:"gender,omitempty"`
	Birthdate           string `json:"birthdate,omitempty"`
	Zoneinfo            string `json:"zoneinfo,omitempty"`
	Locale              string `json:"locale,omitempty"`
	PhoneNumber         string `json:"phone_number,omitempty"`
	PhoneNumberVerified bool   `json:"phone_number_verified,omitempty"`
	UpdatedAt           int64  `json:"updated_at,omitempty"`
}

// IDTokenClaims represents the claims in an ID token.
type IDTokenClaims struct {
	Iss               string   `json:"iss"`
	Sub               string   `json:"sub"`
	Aud               string   `json:"aud"`
	Exp               int64    `json:"exp"`
	Iat               int64    `json:"iat"`
	AuthTime          int64    `json:"auth_time,omitempty"`
	Nonce             string   `json:"nonce,omitempty"`
	Acr               string   `json:"acr,omitempty"`
	Amr               []string `json:"amr,omitempty"`
	Azp               string   `json:"azp,omitempty"`
	Name              string   `json:"name,omitempty"`
	GivenName         string   `json:"given_name,omitempty"`
	FamilyName        string   `json:"family_name,omitempty"`
	PreferredUsername string   `json:"preferred_username,omitempty"`
	Picture           string   `json:"picture,omitempty"`
	Email             string   `json:"email,omitempty"`
	EmailVerified     bool     `json:"email_verified,omitempty"`
}

// DeviceAuthResponse represents device authorization response (RFC 8628)
type DeviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval,omitempty"` // Minimum polling interval in seconds
}

// OAuthError represents an OAuth 2.0 error response (RFC 6749)
type OAuthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

// String returns the error message.
func (e *OAuthError) String() string {
	if e.ErrorDescription != "" {
		return e.Error + ": " + e.ErrorDescription
	}
	return e.Error
}

// PKCEParams contains PKCE parameters for authorization
type PKCEParams struct {
	CodeVerifier        string
	CodeChallenge       string
	CodeChallengeMethod string
}

// ClientType represents OAuth 2.0 client type.
type ClientType string

const (
	ClientTypeConfidential ClientType = "confidential"
	ClientTypePublic       ClientType = "public"
)

// GrantType represents OAuth 2.0 grant types.
type GrantType string

const (
	GrantTypeAuthorizationCode GrantType = "authorization_code"
	GrantTypeClientCredentials GrantType = "client_credentials"
	GrantTypeRefreshToken      GrantType = "refresh_token"
	GrantTypeDeviceCode        GrantType = "urn:ietf:params:oauth:grant-type:device_code"
	GrantTypePassword          GrantType = "password"
	GrantTypeImplicit          GrantType = "implicit"
)

// OAuthClient represents an OAuth 2.0 client application.
type OAuthClient struct {
	ID                string    `json:"id"`
	ClientID          string    `json:"client_id"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	LogoURL           string    `json:"logo_url,omitempty"`
	ClientType        string    `json:"client_type"`
	RedirectURIs      []string  `json:"redirect_uris"`
	AllowedGrantTypes []string  `json:"allowed_grant_types"`
	AllowedScopes     []string  `json:"allowed_scopes"`
	DefaultScopes     []string  `json:"default_scopes"`
	AccessTokenTTL    int       `json:"access_token_ttl"`
	RefreshTokenTTL   int       `json:"refresh_token_ttl"`
	IDTokenTTL        int       `json:"id_token_ttl"`
	RequirePKCE       bool      `json:"require_pkce"`
	RequireConsent    bool      `json:"require_consent"`
	FirstParty        bool      `json:"first_party"`
	IsActive          bool      `json:"is_active"`
	OwnerID           *string   `json:"owner_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CreateOAuthClientRequest is the request body for creating an OAuth client.
type CreateOAuthClientRequest struct {
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	LogoURL           string   `json:"logo_url,omitempty"`
	ClientType        string   `json:"client_type,omitempty"`
	RedirectURIs      []string `json:"redirect_uris"`
	AllowedGrantTypes []string `json:"allowed_grant_types,omitempty"`
	AllowedScopes     []string `json:"allowed_scopes,omitempty"`
	DefaultScopes     []string `json:"default_scopes,omitempty"`
	AccessTokenTTL    *int     `json:"access_token_ttl,omitempty"`
	RefreshTokenTTL   *int     `json:"refresh_token_ttl,omitempty"`
	IDTokenTTL        *int     `json:"id_token_ttl,omitempty"`
	RequirePKCE       *bool    `json:"require_pkce,omitempty"`
	RequireConsent    *bool    `json:"require_consent,omitempty"`
	FirstParty        *bool    `json:"first_party,omitempty"`
}

// CreateOAuthClientResponse is returned when creating an OAuth client.
type CreateOAuthClientResponse struct {
	Client       *OAuthClient `json:"client"`
	ClientSecret string       `json:"client_secret,omitempty"`
}

// UpdateOAuthClientRequest is the request body for updating an OAuth client.
type UpdateOAuthClientRequest struct {
	Name              *string  `json:"name,omitempty"`
	Description       *string  `json:"description,omitempty"`
	LogoURL           *string  `json:"logo_url,omitempty"`
	RedirectURIs      []string `json:"redirect_uris,omitempty"`
	AllowedGrantTypes []string `json:"allowed_grant_types,omitempty"`
	AllowedScopes     []string `json:"allowed_scopes,omitempty"`
	DefaultScopes     []string `json:"default_scopes,omitempty"`
	AccessTokenTTL    *int     `json:"access_token_ttl,omitempty"`
	RefreshTokenTTL   *int     `json:"refresh_token_ttl,omitempty"`
	IDTokenTTL        *int     `json:"id_token_ttl,omitempty"`
	RequirePKCE       *bool    `json:"require_pkce,omitempty"`
	RequireConsent    *bool    `json:"require_consent,omitempty"`
	FirstParty        *bool    `json:"first_party,omitempty"`
	IsActive          *bool    `json:"is_active,omitempty"`
}

// RotateSecretResponse is returned when rotating a client secret.
type RotateSecretResponse struct {
	ClientSecret string `json:"client_secret"`
}

// ListOAuthClientsResponse is the paginated list of OAuth clients.
type ListOAuthClientsResponse struct {
	Clients []OAuthClient `json:"clients"`
	Total   int           `json:"total"`
	Page    int           `json:"page"`
	PerPage int           `json:"per_page"`
}

// OAuthScope represents an OAuth scope.
type OAuthScope struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description,omitempty"`
	IsDefault   bool      `json:"is_default"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateScopeRequest is the request body for creating a scope.
type CreateScopeRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
}

// ListScopesResponse is the list of OAuth scopes.
type ListScopesResponse struct {
	Scopes []OAuthScope `json:"scopes"`
}

// UserConsent represents a user's consent for an OAuth client.
type UserConsent struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	ClientID  string     `json:"client_id"`
	Scopes    []string   `json:"scopes"`
	GrantedAt time.Time  `json:"granted_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// ListConsentsResponse is the list of user consents.
type ListConsentsResponse struct {
	Consents []UserConsent `json:"consents"`
}

// DeviceAuthRequest is the device authorization request.
type DeviceAuthRequest struct {
	ClientID string `json:"client_id"`
	Scope    string `json:"scope,omitempty"`
}

// Standard OAuth Scopes.
const (
	ScopeOpenID        = "openid"
	ScopeProfile       = "profile"
	ScopeEmail         = "email"
	ScopeAddress       = "address"
	ScopePhone         = "phone"
	ScopeOfflineAccess = "offline_access"
)

// ListOAuthClientsParams contains parameters for listing OAuth clients.
type ListOAuthClientsParams struct {
	Page     int    `url:"page,omitempty"`
	Limit    int    `url:"limit,omitempty"`
	Search   string `url:"search,omitempty"`
	IsActive *bool  `url:"is_active,omitempty"`
}

// ListScopesParams contains parameters for listing OAuth scopes.
type ListScopesParams struct {
	Page   int   `url:"page,omitempty"`
	Limit  int   `url:"limit,omitempty"`
	System *bool `url:"system,omitempty"`
}

// ListConsentsParams contains parameters for listing user consents.
type ListConsentsParams struct {
	Page     int    `url:"page,omitempty"`
	Limit    int    `url:"limit,omitempty"`
	ClientID string `url:"client_id,omitempty"`
}

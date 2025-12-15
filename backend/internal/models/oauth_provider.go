package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// OAuthClient represents a registered OAuth 2.0 client application
type OAuthClient struct {
	bun.BaseModel `bun:"table:oauth_clients"`

	ID                uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	ClientID          string     `json:"client_id" bun:"client_id,notnull,unique" example:"my_client_app_123"`
	ClientSecretHash  *string    `json:"-" bun:"client_secret_hash"`
	Name              string     `json:"name" bun:"name,notnull" example:"My Application"`
	Description       string     `json:"description,omitempty" bun:"description" example:"My OAuth client application"`
	LogoURL           string     `json:"logo_url,omitempty" bun:"logo_url" example:"https://example.com/logo.png"`
	ClientType        string     `json:"client_type" bun:"client_type,notnull,default:'confidential'" example:"confidential"`
	RedirectURIs      []string   `json:"redirect_uris" bun:"redirect_uris,type:jsonb,default:'[]'" example:"https://example.com/callback"`
	AllowedGrantTypes []string   `json:"allowed_grant_types" bun:"allowed_grant_types,type:jsonb" example:"authorization_code,refresh_token"`
	AllowedScopes     []string   `json:"allowed_scopes" bun:"allowed_scopes,type:jsonb" example:"openid,profile,email"`
	DefaultScopes     []string   `json:"default_scopes" bun:"default_scopes,type:jsonb" example:"openid,profile"`
	AccessTokenTTL    int        `json:"access_token_ttl" bun:"access_token_ttl,default:900" example:"900"`
	RefreshTokenTTL   int        `json:"refresh_token_ttl" bun:"refresh_token_ttl,default:604800" example:"604800"`
	IDTokenTTL        int        `json:"id_token_ttl" bun:"id_token_ttl,default:3600" example:"3600"`
	RequirePKCE       bool       `json:"require_pkce" bun:"require_pkce,default:false" example:"true"`
	RequireConsent    bool       `json:"require_consent" bun:"require_consent,default:true" example:"true"`
	FirstParty        bool       `json:"first_party" bun:"first_party,default:false" example:"false"`
	OwnerID           *uuid.UUID `json:"owner_id,omitempty" bun:"owner_id,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	Owner             *User      `json:"owner,omitempty" bun:"rel:belongs-to,join:owner_id=id"`
	IsActive          bool       `json:"is_active" bun:"is_active,default:true" example:"true"`
	CreatedAt         time.Time  `json:"created_at" bun:"created_at,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
	UpdatedAt         time.Time  `json:"updated_at" bun:"updated_at,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
}

// ClientType represents the OAuth 2.0 client type
type ClientType string

const (
	ClientTypeConfidential ClientType = "confidential"
	ClientTypePublic       ClientType = "public"
)

// GrantType represents OAuth 2.0 grant types
type GrantType string

const (
	GrantTypeAuthorizationCode GrantType = "authorization_code"
	GrantTypeClientCredentials GrantType = "client_credentials"
	GrantTypeRefreshToken      GrantType = "refresh_token"
	GrantTypeDeviceCode        GrantType = "urn:ietf:params:oauth:grant-type:device_code"
	GrantTypePassword          GrantType = "password"
	GrantTypeImplicit          GrantType = "implicit"
)

// AuthorizationCode represents a temporary authorization code for OAuth 2.0 authorization code flow
type AuthorizationCode struct {
	bun.BaseModel `bun:"table:authorization_codes"`

	ID                  uuid.UUID    `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	CodeHash            string       `json:"-" bun:"code_hash,notnull,unique"`
	ClientID            uuid.UUID    `json:"client_id" bun:"client_id,type:uuid,notnull" example:"123e4567-e89b-12d3-a456-426614174000"`
	Client              *OAuthClient `json:"-" bun:"rel:belongs-to,join:client_id=id"`
	UserID              uuid.UUID    `json:"user_id" bun:"user_id,type:uuid,notnull" example:"123e4567-e89b-12d3-a456-426614174000"`
	User                *User        `json:"-" bun:"rel:belongs-to,join:user_id=id"`
	RedirectURI         string       `json:"redirect_uri" bun:"redirect_uri,notnull" example:"https://example.com/callback"`
	Scope               string       `json:"scope" bun:"scope,notnull" example:"openid profile email"`
	CodeChallenge       *string      `json:"-" bun:"code_challenge"`
	CodeChallengeMethod *string      `json:"-" bun:"code_challenge_method"`
	Nonce               *string      `json:"-" bun:"nonce"`
	Used                bool         `json:"used" bun:"used,default:false" example:"false"`
	ExpiresAt           time.Time    `json:"expires_at" bun:"expires_at,notnull" example:"2024-01-15T10:40:00Z"`
	CreatedAt           time.Time    `json:"created_at" bun:"created_at,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
}

// OAuthAccessToken represents an OAuth 2.0 access token
type OAuthAccessToken struct {
	bun.BaseModel `bun:"table:oauth_access_tokens"`

	ID        uuid.UUID    `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	TokenHash string       `json:"-" bun:"token_hash,notnull,unique"`
	ClientID  uuid.UUID    `json:"client_id" bun:"client_id,type:uuid,notnull" example:"123e4567-e89b-12d3-a456-426614174000"`
	Client    *OAuthClient `json:"-" bun:"rel:belongs-to,join:client_id=id"`
	UserID    *uuid.UUID   `json:"user_id,omitempty" bun:"user_id,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	User      *User        `json:"-" bun:"rel:belongs-to,join:user_id=id"`
	Scope     string       `json:"scope" bun:"scope,notnull" example:"openid profile email"`
	IsActive  bool         `json:"is_active" bun:"is_active,default:true" example:"true"`
	ExpiresAt time.Time    `json:"expires_at" bun:"expires_at,notnull" example:"2024-01-15T10:45:00Z"`
	CreatedAt time.Time    `json:"created_at" bun:"created_at,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
	RevokedAt *time.Time   `json:"revoked_at,omitempty" bun:"revoked_at" example:"2024-01-15T11:00:00Z"`
}

// OAuthRefreshToken represents an OAuth 2.0 refresh token
type OAuthRefreshToken struct {
	bun.BaseModel `bun:"table:oauth_refresh_tokens"`

	ID            uuid.UUID         `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	TokenHash     string            `json:"-" bun:"token_hash,notnull,unique"`
	AccessTokenID *uuid.UUID        `json:"access_token_id,omitempty" bun:"access_token_id,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	AccessToken   *OAuthAccessToken `json:"-" bun:"rel:belongs-to,join:access_token_id=id"`
	ClientID      uuid.UUID         `json:"client_id" bun:"client_id,type:uuid,notnull" example:"123e4567-e89b-12d3-a456-426614174000"`
	Client        *OAuthClient      `json:"-" bun:"rel:belongs-to,join:client_id=id"`
	UserID        uuid.UUID         `json:"user_id" bun:"user_id,type:uuid,notnull" example:"123e4567-e89b-12d3-a456-426614174000"`
	User          *User             `json:"-" bun:"rel:belongs-to,join:user_id=id"`
	Scope         string            `json:"scope" bun:"scope,notnull" example:"openid profile email"`
	IsActive      bool              `json:"is_active" bun:"is_active,default:true" example:"true"`
	ExpiresAt     time.Time         `json:"expires_at" bun:"expires_at,notnull" example:"2024-01-22T10:30:00Z"`
	CreatedAt     time.Time         `json:"created_at" bun:"created_at,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
	RevokedAt     *time.Time        `json:"revoked_at,omitempty" bun:"revoked_at" example:"2024-01-15T11:00:00Z"`
}

// UserConsent represents a user's consent to an OAuth client accessing their data
type UserConsent struct {
	bun.BaseModel `bun:"table:user_consents"`

	ID        uuid.UUID    `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID    uuid.UUID    `json:"user_id" bun:"user_id,type:uuid,notnull" example:"123e4567-e89b-12d3-a456-426614174000"`
	User      *User        `json:"-" bun:"rel:belongs-to,join:user_id=id"`
	ClientID  uuid.UUID    `json:"client_id" bun:"client_id,type:uuid,notnull" example:"123e4567-e89b-12d3-a456-426614174000"`
	Client    *OAuthClient `json:"-" bun:"rel:belongs-to,join:client_id=id"`
	Scopes    []string     `json:"scopes" bun:"scopes,type:jsonb,notnull" example:"openid,profile,email"`
	GrantedAt time.Time    `json:"granted_at" bun:"granted_at,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
	RevokedAt *time.Time   `json:"revoked_at,omitempty" bun:"revoked_at" example:"2024-01-15T11:00:00Z"`
}

// DeviceCodeStatus represents the status of a device code
type DeviceCodeStatus string

const (
	DeviceCodeStatusPending    DeviceCodeStatus = "pending"
	DeviceCodeStatusAuthorized DeviceCodeStatus = "authorized"
	DeviceCodeStatusDenied     DeviceCodeStatus = "denied"
	DeviceCodeStatusExpired    DeviceCodeStatus = "expired"
)

// DeviceCode represents a device authorization grant (RFC 8628)
type DeviceCode struct {
	bun.BaseModel `bun:"table:device_codes"`

	ID                      uuid.UUID        `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	DeviceCodeHash          string           `json:"-" bun:"device_code_hash,notnull,unique"`
	UserCode                string           `json:"user_code" bun:"user_code,notnull,unique" example:"ABCD-EFGH"`
	ClientID                uuid.UUID        `json:"client_id" bun:"client_id,type:uuid,notnull" example:"123e4567-e89b-12d3-a456-426614174000"`
	Client                  *OAuthClient     `json:"-" bun:"rel:belongs-to,join:client_id=id"`
	UserID                  *uuid.UUID       `json:"user_id,omitempty" bun:"user_id,type:uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	User                    *User            `json:"-" bun:"rel:belongs-to,join:user_id=id"`
	Scope                   string           `json:"scope" bun:"scope,notnull" example:"openid profile email"`
	Status                  DeviceCodeStatus `json:"status" bun:"status,default:'pending'" example:"pending"`
	VerificationURI         string           `json:"verification_uri" bun:"verification_uri,notnull" example:"https://auth.example.com/device"`
	VerificationURIComplete string           `json:"verification_uri_complete,omitempty" bun:"verification_uri_complete" example:"https://auth.example.com/device?user_code=ABCD-EFGH"`
	ExpiresAt               time.Time        `json:"expires_at" bun:"expires_at,notnull" example:"2024-01-15T10:45:00Z"`
	Interval                int              `json:"interval" bun:"interval,default:5" example:"5"`
	CreatedAt               time.Time        `json:"created_at" bun:"created_at,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
}

// OAuthScope represents a defined OAuth 2.0 scope
type OAuthScope struct {
	bun.BaseModel `bun:"table:oauth_scopes"`

	ID          uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string    `json:"name" bun:"name,notnull,unique" example:"profile"`
	DisplayName string    `json:"display_name" bun:"display_name,notnull" example:"Profile Information"`
	Description string    `json:"description,omitempty" bun:"description" example:"Access to basic profile information"`
	IsDefault   bool      `json:"is_default" bun:"is_default,default:false" example:"true"`
	IsSystem    bool      `json:"is_system" bun:"is_system,default:true" example:"true"`
	CreatedAt   time.Time `json:"created_at" bun:"created_at,default:current_timestamp" example:"2024-01-15T10:30:00Z"`
}

// Standard OIDC scopes
const (
	ScopeOpenID        = "openid"
	ScopeProfile       = "profile"
	ScopeEmail         = "email"
	ScopeAddress       = "address"
	ScopePhone         = "phone"
	ScopeOfflineAccess = "offline_access"
)

// CreateOAuthClientRequest represents a request to create a new OAuth client
type CreateOAuthClientRequest struct {
	Name              string   `json:"name" binding:"required,min=3,max=100" example:"My Application"`
	Description       string   `json:"description,omitempty" example:"My OAuth client application"`
	LogoURL           string   `json:"logo_url,omitempty" example:"https://example.com/logo.png"`
	ClientType        string   `json:"client_type" binding:"required,oneof=confidential public" example:"confidential"`
	RedirectURIs      []string `json:"redirect_uris" binding:"dive,url" example:"https://example.com/callback"`
	AllowedGrantTypes []string `json:"allowed_grant_types" binding:"required,min=1" example:"authorization_code,refresh_token"`
	AllowedScopes     []string `json:"allowed_scopes" binding:"required,min=1" example:"openid,profile,email"`
	DefaultScopes     []string `json:"default_scopes,omitempty" example:"openid,profile"`
	AccessTokenTTL    *int     `json:"access_token_ttl,omitempty" example:"900"`
	RefreshTokenTTL   *int     `json:"refresh_token_ttl,omitempty" example:"604800"`
	IDTokenTTL        *int     `json:"id_token_ttl,omitempty" example:"3600"`
	RequirePKCE       *bool    `json:"require_pkce,omitempty" example:"true"`
	RequireConsent    *bool    `json:"require_consent,omitempty" example:"true"`
	FirstParty        *bool    `json:"first_party,omitempty" example:"false"`
}

// CreateOAuthClientResponse represents the response when creating an OAuth client
type CreateOAuthClientResponse struct {
	Client       *OAuthClient `json:"client"`
	ClientSecret string       `json:"client_secret,omitempty" example:"client_secret_abc123xyz789"`
}

// UpdateOAuthClientRequest represents a request to update an OAuth client
type UpdateOAuthClientRequest struct {
	Name              string   `json:"name,omitempty" binding:"omitempty,min=3,max=100" example:"My Updated Application"`
	Description       string   `json:"description,omitempty" example:"Updated description"`
	LogoURL           string   `json:"logo_url,omitempty" example:"https://example.com/new-logo.png"`
	RedirectURIs      []string `json:"redirect_uris,omitempty" binding:"omitempty,dive,url" example:"https://example.com/callback"`
	AllowedGrantTypes []string `json:"allowed_grant_types,omitempty" binding:"omitempty,min=1" example:"authorization_code,refresh_token"`
	AllowedScopes     []string `json:"allowed_scopes,omitempty" binding:"omitempty,min=1" example:"openid,profile,email"`
	DefaultScopes     []string `json:"default_scopes,omitempty" example:"openid,profile"`
	AccessTokenTTL    *int     `json:"access_token_ttl,omitempty" example:"900"`
	RefreshTokenTTL   *int     `json:"refresh_token_ttl,omitempty" example:"604800"`
	IDTokenTTL        *int     `json:"id_token_ttl,omitempty" example:"3600"`
	RequirePKCE       *bool    `json:"require_pkce,omitempty" example:"true"`
	RequireConsent    *bool    `json:"require_consent,omitempty" example:"true"`
	IsActive          *bool    `json:"is_active,omitempty" example:"true"`
}

// AuthorizeRequest represents an OAuth 2.0 authorization request
type AuthorizeRequest struct {
	ResponseType        string  `form:"response_type" binding:"required,oneof=code token" example:"code"`
	ClientID            string  `form:"client_id" binding:"required" example:"my_client_app_123"`
	RedirectURI         string  `form:"redirect_uri" binding:"required,url" example:"https://example.com/callback"`
	Scope               string  `form:"scope" binding:"required" example:"openid profile email"`
	State               string  `form:"state" binding:"required" example:"random_state_string"`
	Nonce               *string `form:"nonce" example:"random_nonce_string"`
	CodeChallenge       *string `form:"code_challenge" example:"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"`
	CodeChallengeMethod *string `form:"code_challenge_method" binding:"omitempty,oneof=S256 plain" example:"S256"`
	Prompt              *string `form:"prompt" binding:"omitempty,oneof=none login consent select_account" example:"consent"`
	MaxAge              *int    `form:"max_age" example:"3600"`
	Display             *string `form:"display" binding:"omitempty,oneof=page popup touch wap" example:"page"`
}

// AuthorizeResponse represents an OAuth 2.0 authorization response
type AuthorizeResponse struct {
	Code  string `json:"code,omitempty" example:"authorization_code_abc123"`
	State string `json:"state" example:"random_state_string"`
}

// TokenRequest represents an OAuth 2.0 token request
type TokenRequest struct {
	GrantType    string  `form:"grant_type" binding:"required" example:"authorization_code"`
	Code         *string `form:"code" example:"authorization_code_abc123"`
	RedirectURI  *string `form:"redirect_uri" example:"https://example.com/callback"`
	ClientID     string  `form:"client_id" binding:"required" example:"my_client_app_123"`
	ClientSecret *string `form:"client_secret" example:"client_secret_abc123xyz789"`
	RefreshToken *string `form:"refresh_token" example:"refresh_token_xyz789"`
	Scope        *string `form:"scope" example:"openid profile email"`
	CodeVerifier *string `form:"code_verifier" example:"dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"`
	DeviceCode   *string `form:"device_code" example:"device_code_abc123"`
	Username     *string `form:"username" example:"user@example.com"`
	Password     *string `form:"password" example:"password123"`

	// Session context (populated by handler, not from form)
	IPAddress string `form:"-" json:"-"`
	UserAgent string `form:"-" json:"-"`
}

// TokenResponse represents an OAuth 2.0 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int    `json:"expires_in" example:"900"`
	RefreshToken string `json:"refresh_token,omitempty" example:"refresh_token_xyz789"`
	IDToken      string `json:"id_token,omitempty" example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Scope        string `json:"scope,omitempty" example:"openid profile email"`
}

// IntrospectionRequest represents an OAuth 2.0 token introspection request
type IntrospectionRequest struct {
	Token         string  `form:"token" binding:"required" example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenTypeHint *string `form:"token_type_hint" example:"access_token"`
	ClientID      string  `form:"client_id" binding:"required" example:"my_client_app_123"`
	ClientSecret  *string `form:"client_secret" example:"client_secret_abc123xyz789"`
}

// IntrospectionResponse represents an OAuth 2.0 token introspection response
type IntrospectionResponse struct {
	Active    bool   `json:"active" example:"true"`
	Scope     string `json:"scope,omitempty" example:"openid profile email"`
	ClientID  string `json:"client_id,omitempty" example:"my_client_app_123"`
	Username  string `json:"username,omitempty" example:"johndoe"`
	TokenType string `json:"token_type,omitempty" example:"Bearer"`
	ExpiresAt int64  `json:"exp,omitempty" example:"1705319400"`
	IssuedAt  int64  `json:"iat,omitempty" example:"1705315800"`
	NotBefore int64  `json:"nbf,omitempty" example:"1705315800"`
	Subject   string `json:"sub,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	Audience  string `json:"aud,omitempty" example:"my_client_app_123"`
	Issuer    string `json:"iss,omitempty" example:"https://auth.example.com"`
	JWTID     string `json:"jti,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// DeviceAuthRequest represents a device authorization request
type DeviceAuthRequest struct {
	ClientID string  `form:"client_id" binding:"required" example:"my_client_app_123"`
	Scope    *string `form:"scope" example:"openid profile email"`
}

// DeviceAuthResponse represents a device authorization response
type DeviceAuthResponse struct {
	DeviceCode              string `json:"device_code" example:"GmRhmhcxhwAzkoEqiMEg_DnyEysNkuNhszIySk9eS"`
	UserCode                string `json:"user_code" example:"ABCD-EFGH"`
	VerificationURI         string `json:"verification_uri" example:"https://auth.example.com/device"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty" example:"https://auth.example.com/device?user_code=ABCD-EFGH"`
	ExpiresIn               int    `json:"expires_in" example:"900"`
	Interval                int    `json:"interval" example:"5"`
}

// UserInfoResponse represents an OIDC UserInfo endpoint response
type UserInfoResponse struct {
	Subject             string  `json:"sub" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name                *string `json:"name,omitempty" example:"John Doe"`
	GivenName           *string `json:"given_name,omitempty" example:"John"`
	FamilyName          *string `json:"family_name,omitempty" example:"Doe"`
	MiddleName          *string `json:"middle_name,omitempty" example:"Michael"`
	Nickname            *string `json:"nickname,omitempty" example:"Johnny"`
	PreferredUsername   *string `json:"preferred_username,omitempty" example:"johndoe"`
	Profile             *string `json:"profile,omitempty" example:"https://example.com/johndoe"`
	Picture             *string `json:"picture,omitempty" example:"https://example.com/avatars/johndoe.jpg"`
	Website             *string `json:"website,omitempty" example:"https://johndoe.com"`
	Email               *string `json:"email,omitempty" example:"john.doe@example.com"`
	EmailVerified       *bool   `json:"email_verified,omitempty" example:"true"`
	Gender              *string `json:"gender,omitempty" example:"male"`
	Birthdate           *string `json:"birthdate,omitempty" example:"1990-01-01"`
	Zoneinfo            *string `json:"zoneinfo,omitempty" example:"America/Los_Angeles"`
	Locale              *string `json:"locale,omitempty" example:"en-US"`
	PhoneNumber         *string `json:"phone_number,omitempty" example:"+1 (555) 123-4567"`
	PhoneNumberVerified *bool   `json:"phone_number_verified,omitempty" example:"false"`
	UpdatedAt           *int64  `json:"updated_at,omitempty" example:"1705315800"`
}

// OIDCDiscoveryDocument represents the OpenID Connect discovery document
type OIDCDiscoveryDocument struct {
	Issuer                                     string   `json:"issuer" example:"https://auth.example.com"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint" example:"https://auth.example.com/oauth2/authorize"`
	TokenEndpoint                              string   `json:"token_endpoint" example:"https://auth.example.com/oauth2/token"`
	UserInfoEndpoint                           string   `json:"userinfo_endpoint" example:"https://auth.example.com/oauth2/userinfo"`
	JwksURI                                    string   `json:"jwks_uri" example:"https://auth.example.com/.well-known/jwks.json"`
	RegistrationEndpoint                       string   `json:"registration_endpoint,omitempty" example:"https://auth.example.com/oauth2/register"`
	RevocationEndpoint                         string   `json:"revocation_endpoint,omitempty" example:"https://auth.example.com/oauth2/revoke"`
	IntrospectionEndpoint                      string   `json:"introspection_endpoint,omitempty" example:"https://auth.example.com/oauth2/introspect"`
	DeviceAuthorizationEndpoint                string   `json:"device_authorization_endpoint,omitempty" example:"https://auth.example.com/oauth2/device/code"`
	EndSessionEndpoint                         string   `json:"end_session_endpoint,omitempty" example:"https://auth.example.com/oauth2/logout"`
	ScopesSupported                            []string `json:"scopes_supported" example:"openid,profile,email,offline_access"`
	ResponseTypesSupported                     []string `json:"response_types_supported" example:"code,token,id_token"`
	ResponseModesSupported                     []string `json:"response_modes_supported,omitempty" example:"query,fragment,form_post"`
	GrantTypesSupported                        []string `json:"grant_types_supported" example:"authorization_code,refresh_token,client_credentials"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported" example:"client_secret_basic,client_secret_post"`
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported,omitempty" example:"RS256,ES256"`
	ServiceDocumentation                       string   `json:"service_documentation,omitempty" example:"https://docs.example.com"`
	UILocalesSupported                         []string `json:"ui_locales_supported,omitempty" example:"en-US,es-ES"`
	OpPolicyURI                                string   `json:"op_policy_uri,omitempty" example:"https://example.com/policy"`
	OpTosURI                                   string   `json:"op_tos_uri,omitempty" example:"https://example.com/tos"`
	ClaimsSupported                            []string `json:"claims_supported,omitempty" example:"sub,name,email,picture"`
	ClaimTypesSupported                        []string `json:"claim_types_supported,omitempty" example:"normal"`
	ClaimsParameterSupported                   bool     `json:"claims_parameter_supported,omitempty" example:"false"`
	RequestParameterSupported                  bool     `json:"request_parameter_supported,omitempty" example:"false"`
	RequestURIParameterSupported               bool     `json:"request_uri_parameter_supported,omitempty" example:"false"`
	RequireRequestURIRegistration              bool     `json:"require_request_uri_registration,omitempty" example:"false"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported,omitempty" example:"S256,plain"`
	IDTokenSigningAlgValuesSupported           []string `json:"id_token_signing_alg_values_supported" example:"RS256,ES256"`
	IDTokenEncryptionAlgValuesSupported        []string `json:"id_token_encryption_alg_values_supported,omitempty" example:"RSA-OAEP,A256KW"`
	IDTokenEncryptionEncValuesSupported        []string `json:"id_token_encryption_enc_values_supported,omitempty" example:"A128CBC-HS256,A256GCM"`
	UserInfoSigningAlgValuesSupported          []string `json:"userinfo_signing_alg_values_supported,omitempty" example:"RS256,ES256"`
	UserInfoEncryptionAlgValuesSupported       []string `json:"userinfo_encryption_alg_values_supported,omitempty" example:"RSA-OAEP,A256KW"`
	UserInfoEncryptionEncValuesSupported       []string `json:"userinfo_encryption_enc_values_supported,omitempty" example:"A128CBC-HS256,A256GCM"`
	SubjectTypesSupported                      []string `json:"subject_types_supported" example:"public,pairwise"`
	DisplayValuesSupported                     []string `json:"display_values_supported,omitempty" example:"page,popup"`
	AcrValuesSupported                         []string `json:"acr_values_supported,omitempty" example:"urn:mace:incommon:iap:silver"`
}

// JWKSDocument represents a JSON Web Key Set
type JWKSDocument struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	KeyType   string   `json:"kty" example:"RSA"`
	Use       string   `json:"use,omitempty" example:"sig"`
	KeyOps    []string `json:"key_ops,omitempty" example:"sign,verify"`
	Algorithm string   `json:"alg,omitempty" example:"RS256"`
	KeyID     string   `json:"kid,omitempty" example:"2023-01-01"`
	X5U       string   `json:"x5u,omitempty" example:"https://example.com/keys/2023-01-01.pem"`
	X5C       []string `json:"x5c,omitempty"`
	X5T       string   `json:"x5t,omitempty" example:"GvnVZS_KnfEhOp7oUdKKPqPFT_8"`
	X5TS256   string   `json:"x5t#S256,omitempty" example:"YmFzZTY0dXJsIGVuY29kZWQgc2hhMjU2IGhhc2g"`
	N         string   `json:"n,omitempty" example:"0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw"`
	E         string   `json:"e,omitempty" example:"AQAB"`
	D         string   `json:"d,omitempty"`
	P         string   `json:"p,omitempty"`
	Q         string   `json:"q,omitempty"`
	DP        string   `json:"dp,omitempty"`
	DQ        string   `json:"dq,omitempty"`
	QI        string   `json:"qi,omitempty"`
	K         string   `json:"k,omitempty"`
	CRV       string   `json:"crv,omitempty" example:"P-256"`
	X         string   `json:"x,omitempty" example:"f83OJ3D2xF1Bg8vub9tLe1gHMzV76e8Tus9uPHvRVEU"`
	Y         string   `json:"y,omitempty" example:"x_FEzRu9m36HLN_tue659LNpXW6pCyStikYjKIWI5a0"`
}

// RevocationRequest represents an OAuth 2.0 token revocation request
type RevocationRequest struct {
	Token         string  `form:"token" binding:"required" example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenTypeHint *string `form:"token_type_hint" example:"refresh_token"`
	ClientID      string  `form:"client_id" binding:"required" example:"my_client_app_123"`
	ClientSecret  *string `form:"client_secret" example:"client_secret_abc123xyz789"`
}

// IsExpired checks if the authorization code is expired
func (ac *AuthorizationCode) IsExpired() bool {
	return time.Now().After(ac.ExpiresAt)
}

// IsValid checks if the authorization code is valid (not used and not expired)
func (ac *AuthorizationCode) IsValid() bool {
	return !ac.Used && !ac.IsExpired()
}

// IsExpired checks if the access token is expired
func (at *OAuthAccessToken) IsExpired() bool {
	return time.Now().After(at.ExpiresAt)
}

// IsValid checks if the access token is valid (active, not revoked, and not expired)
func (at *OAuthAccessToken) IsValid() bool {
	return at.IsActive && at.RevokedAt == nil && !at.IsExpired()
}

// IsExpired checks if the refresh token is expired
func (rt *OAuthRefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid checks if the refresh token is valid (active, not revoked, and not expired)
func (rt *OAuthRefreshToken) IsValid() bool {
	return rt.IsActive && rt.RevokedAt == nil && !rt.IsExpired()
}

// IsExpired checks if the device code is expired
func (dc *DeviceCode) IsExpired() bool {
	return time.Now().After(dc.ExpiresAt)
}

// IsValid checks if the device code is valid (pending or authorized and not expired)
func (dc *DeviceCode) IsValid() bool {
	return (dc.Status == DeviceCodeStatusPending || dc.Status == DeviceCodeStatusAuthorized) && !dc.IsExpired()
}

// IsRevoked checks if the user consent is revoked
func (uc *UserConsent) IsRevoked() bool {
	return uc.RevokedAt != nil
}

// HasScope checks if consent includes a specific scope
func (uc *UserConsent) HasScope(scope string) bool {
	for _, s := range uc.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// IsValidClientType checks if a client type is valid
func IsValidClientType(clientType string) bool {
	return clientType == string(ClientTypeConfidential) || clientType == string(ClientTypePublic)
}

// IsValidGrantType checks if a grant type is valid
func IsValidGrantType(grantType string) bool {
	validGrants := []GrantType{
		GrantTypeAuthorizationCode,
		GrantTypeClientCredentials,
		GrantTypeRefreshToken,
		GrantTypeDeviceCode,
		GrantTypePassword,
		GrantTypeImplicit,
	}
	for _, valid := range validGrants {
		if grantType == string(valid) {
			return true
		}
	}
	return false
}

// IsValidOIDCScope checks if a scope is a valid OIDC standard scope
func IsValidOIDCScope(scope string) bool {
	standardScopes := []string{
		ScopeOpenID,
		ScopeProfile,
		ScopeEmail,
		ScopeAddress,
		ScopePhone,
		ScopeOfflineAccess,
	}
	for _, validScope := range standardScopes {
		if scope == validScope {
			return true
		}
	}
	return false
}

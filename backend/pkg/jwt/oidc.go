package jwt

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/keys"
)

var (
	ErrInvalidSigningKey = errors.New("invalid signing key")
	ErrUnsupportedAlg    = errors.New("unsupported signing algorithm")
	ErrInvalidKID        = errors.New("invalid key ID in token header")
	ErrInvalidIssuer     = errors.New("invalid token issuer")
)

type OIDCService struct {
	keyManager *keys.Manager
	issuer     string
}

type IDTokenClaims struct {
	jwt.RegisteredClaims

	Nonce    string   `json:"nonce,omitempty"`
	AuthTime int64    `json:"auth_time,omitempty"`
	ACR      string   `json:"acr,omitempty"`
	AMR      []string `json:"amr,omitempty"`
	AZP      string   `json:"azp,omitempty"`

	Name              string `json:"name,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	MiddleName        string `json:"middle_name,omitempty"`
	Nickname          string `json:"nickname,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Profile           string `json:"profile,omitempty"`
	Picture           string `json:"picture,omitempty"`
	Website           string `json:"website,omitempty"`
	Gender            string `json:"gender,omitempty"`
	Birthdate         string `json:"birthdate,omitempty"`
	Zoneinfo          string `json:"zoneinfo,omitempty"`
	Locale            string `json:"locale,omitempty"`
	UpdatedAt         int64  `json:"updated_at,omitempty"`

	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`

	Phone         string `json:"phone_number,omitempty"`
	PhoneVerified bool   `json:"phone_number_verified,omitempty"`

	Address *AddressClaim `json:"address,omitempty"`
}

type AddressClaim struct {
	Formatted     string `json:"formatted,omitempty"`
	StreetAddress string `json:"street_address,omitempty"`
	Locality      string `json:"locality,omitempty"`
	Region        string `json:"region,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
}

type OAuthAccessTokenClaims struct {
	jwt.RegisteredClaims

	Scope     string   `json:"scope,omitempty"`
	ClientID  string   `json:"client_id,omitempty"`
	TokenType string   `json:"token_type,omitempty"`
	Roles     []string `json:"roles,omitempty"`
}

func NewOIDCService(keyManager *keys.Manager, issuer string) *OIDCService {
	return &OIDCService{
		keyManager: keyManager,
		issuer:     issuer,
	}
}

func (s *OIDCService) GenerateIDToken(userID uuid.UUID, clientID, nonce string, scopes []string, user *models.User, ttl time.Duration) (string, error) {
	claims := s.BuildIDTokenClaims(userID, clientID, nonce, scopes, user, ttl)

	signingKey, err := s.keyManager.GetCurrentKey()
	if err != nil {
		return "", fmt.Errorf("failed to get signing key: %w", err)
	}

	signingMethod, err := s.getSigningMethod(signingKey.Algorithm)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	token.Header["kid"] = signingKey.KID

	return token.SignedString(signingKey.PrivateKey)
}

func (s *OIDCService) GenerateOAuthAccessToken(userID *uuid.UUID, clientID string, scope string, roles []string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &OAuthAccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		Scope:     scope,
		ClientID:  clientID,
		TokenType: "Bearer",
		Roles:     roles,
	}

	if userID != nil {
		claims.Subject = userID.String()
	}

	signingKey, err := s.keyManager.GetCurrentKey()
	if err != nil {
		return "", fmt.Errorf("failed to get signing key: %w", err)
	}

	signingMethod, err := s.getSigningMethod(signingKey.Algorithm)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	token.Header["kid"] = signingKey.KID

	return token.SignedString(signingKey.PrivateKey)
}

func (s *OIDCService) ValidateIDToken(tokenString string) (*IDTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &IDTokenClaims{}, s.keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*IDTokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	if claims.Issuer != s.issuer {
		return nil, ErrInvalidIssuer
	}

	return claims, nil
}

func (s *OIDCService) ValidateOAuthAccessToken(tokenString string) (*OAuthAccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &OAuthAccessTokenClaims{}, s.keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*OAuthAccessTokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	if claims.Issuer != s.issuer {
		return nil, ErrInvalidIssuer
	}

	return claims, nil
}

func (s *OIDCService) GetSigningKeyID() (string, error) {
	signingKey, err := s.keyManager.GetCurrentKey()
	if err != nil {
		return "", fmt.Errorf("failed to get signing key: %w", err)
	}
	return signingKey.KID, nil
}

func (s *OIDCService) BuildIDTokenClaims(userID uuid.UUID, clientID, nonce string, scopes []string, user *models.User, ttl time.Duration) *IDTokenClaims {
	now := time.Now()
	claims := &IDTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{clientID},
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		AuthTime: now.Unix(),
		AZP:      clientID,
	}

	if nonce != "" {
		claims.Nonce = nonce
	}

	scopeSet := make(map[string]bool)
	for _, scope := range scopes {
		scopeSet[scope] = true
	}

	if scopeSet[models.ScopeProfile] {
		claims.Name = user.FullName
		claims.PreferredUsername = user.Username
		claims.Picture = user.ProfilePictureURL

		if user.UpdatedAt.Unix() > 0 {
			claims.UpdatedAt = user.UpdatedAt.Unix()
		}
	}

	if scopeSet[models.ScopeEmail] {
		claims.Email = user.Email
		claims.EmailVerified = user.EmailVerified
	}

	if scopeSet[models.ScopePhone] {
		if user.Phone != nil {
			claims.Phone = *user.Phone
		}
		claims.PhoneVerified = user.PhoneVerified
	}

	return claims
}

func (s *OIDCService) GetIssuer() string {
	return s.issuer
}

func (s *OIDCService) keyFunc(token *jwt.Token) (interface{}, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok || kid == "" {
		return nil, ErrInvalidKID
	}

	signingKey, err := s.keyManager.GetKey(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get key %s: %w", kid, err)
	}

	switch token.Method.(type) {
	case *jwt.SigningMethodRSA:
		if signingKey.Algorithm != keys.RS256 {
			return nil, ErrUnsupportedAlg
		}
		rsaPublicKey, ok := signingKey.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, ErrInvalidSigningKey
		}
		return rsaPublicKey, nil

	case *jwt.SigningMethodECDSA:
		if signingKey.Algorithm != keys.ES256 {
			return nil, ErrUnsupportedAlg
		}
		ecdsaPublicKey, ok := signingKey.PublicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, ErrInvalidSigningKey
		}
		return ecdsaPublicKey, nil

	default:
		return nil, ErrUnsupportedAlg
	}
}

func (s *OIDCService) getSigningMethod(algorithm keys.Algorithm) (jwt.SigningMethod, error) {
	switch algorithm {
	case keys.RS256:
		return jwt.SigningMethodRS256, nil
	case keys.ES256:
		return jwt.SigningMethodES256, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedAlg, algorithm)
	}
}

func SplitScopes(scope string) []string {
	if scope == "" {
		return []string{}
	}
	scopes := strings.Split(scope, " ")
	result := make([]string, 0, len(scopes))
	for _, s := range scopes {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

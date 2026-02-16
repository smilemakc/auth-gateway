package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService() *Service {
	return NewService("access-secret", "refresh-secret", 15*time.Minute, 7*24*time.Hour)
}

func newTestUser() *models.User {
	return &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		IsActive: true,
		Roles:    []models.Role{{Name: "user"}},
	}
}

// ============================================================
// GenerateAccessToken Tests
// ============================================================

func TestService_GenerateAccessToken_ShouldReturnValidToken_WhenUserProvided(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.IsActive, claims.IsActive)
	assert.Equal(t, []string{"user"}, claims.Roles)
	assert.Equal(t, user.ID.String(), claims.Subject)
}

func TestService_GenerateAccessToken_ShouldSetCorrectExpiration(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	before := time.Now()
	token, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)

	expectedExpiry := before.Add(15 * time.Minute)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 2*time.Second)
	assert.WithinDuration(t, before, claims.IssuedAt.Time, 2*time.Second)
	assert.WithinDuration(t, before, claims.NotBefore.Time, 2*time.Second)
}

func TestService_GenerateAccessToken_ShouldContainUniqueJTI(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token1, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	token2, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	claims1, err := svc.ValidateAccessToken(token1)
	require.NoError(t, err)

	claims2, err := svc.ValidateAccessToken(token2)
	require.NoError(t, err)

	assert.NotEmpty(t, claims1.ID)
	assert.NotEmpty(t, claims2.ID)
	assert.NotEqual(t, claims1.ID, claims2.ID, "each token must have a unique jti")

	_, parseErr := uuid.Parse(claims1.ID)
	assert.NoError(t, parseErr, "jti must be a valid UUID")
}

func TestService_GenerateAccessToken_ShouldUseDefaultRoles_WhenUserHasNoRoles(t *testing.T) {
	svc := newTestService()
	user := &models.User{
		ID:       uuid.New(),
		Email:    "noroles@example.com",
		Username: "noroles",
		IsActive: true,
		Roles:    []models.Role{},
	}

	token, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, []string{"user"}, claims.Roles, "empty roles should default to [\"user\"]")
}

func TestService_GenerateAccessToken_ShouldIncludeMultipleRoles(t *testing.T) {
	svc := newTestService()
	user := &models.User{
		ID:       uuid.New(),
		Email:    "multi@example.com",
		Username: "multirole",
		IsActive: true,
		Roles:    []models.Role{{Name: "admin"}, {Name: "moderator"}, {Name: "user"}},
	}

	token, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, []string{"admin", "moderator", "user"}, claims.Roles)
}

func TestService_GenerateAccessToken_ShouldBindApplicationID_WhenProvided(t *testing.T) {
	svc := newTestService()
	user := newTestUser()
	appID := uuid.New()

	token, err := svc.GenerateAccessToken(user, &appID)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims.ApplicationID)
	assert.Equal(t, appID, *claims.ApplicationID)
}

func TestService_GenerateAccessToken_ShouldNotIncludeApplicationID_WhenNotProvided(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Nil(t, claims.ApplicationID)
}

func TestService_GenerateAccessToken_ShouldIgnoreNilApplicationID(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateAccessToken(user, nil)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Nil(t, claims.ApplicationID)
}

// ============================================================
// GenerateRefreshToken Tests
// ============================================================

func TestService_GenerateRefreshToken_ShouldReturnValidToken(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateRefreshToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateRefreshToken(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Username, claims.Username)
}

func TestService_GenerateRefreshToken_ShouldSetLongerExpiration(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	before := time.Now()
	token, err := svc.GenerateRefreshToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateRefreshToken(token)
	require.NoError(t, err)

	expectedExpiry := before.Add(7 * 24 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 2*time.Second)
}

func TestService_GenerateRefreshToken_ShouldContainUniqueJTI(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateRefreshToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateRefreshToken(token)
	require.NoError(t, err)
	assert.NotEmpty(t, claims.ID)

	_, parseErr := uuid.Parse(claims.ID)
	assert.NoError(t, parseErr, "jti must be a valid UUID")
}

func TestService_GenerateRefreshToken_ShouldDefaultToUserRole_WhenNoRoles(t *testing.T) {
	svc := newTestService()
	user := &models.User{
		ID:       uuid.New(),
		Email:    "noroles@test.com",
		Username: "noroles",
		IsActive: true,
	}

	token, err := svc.GenerateRefreshToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateRefreshToken(token)
	require.NoError(t, err)
	assert.Equal(t, []string{"user"}, claims.Roles)
}

func TestService_GenerateRefreshToken_ShouldBindApplicationID(t *testing.T) {
	svc := newTestService()
	user := newTestUser()
	appID := uuid.New()

	token, err := svc.GenerateRefreshToken(user, &appID)
	require.NoError(t, err)

	claims, err := svc.ValidateRefreshToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims.ApplicationID)
	assert.Equal(t, appID, *claims.ApplicationID)
}

// ============================================================
// GenerateTwoFactorToken Tests
// ============================================================

func TestService_GenerateTwoFactorToken_ShouldReturnValidToken(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateTwoFactorToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// 2FA token uses access secret for validation
	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
}

func TestService_GenerateTwoFactorToken_ShouldExpireIn5Minutes(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	before := time.Now()
	token, err := svc.GenerateTwoFactorToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)

	expectedExpiry := before.Add(5 * time.Minute)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 2*time.Second)
}

func TestService_GenerateTwoFactorToken_ShouldContainUniqueJTI(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateTwoFactorToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.NotEmpty(t, claims.ID)

	_, parseErr := uuid.Parse(claims.ID)
	assert.NoError(t, parseErr, "jti must be a valid UUID")
}

func TestService_GenerateTwoFactorToken_ShouldBindApplicationID(t *testing.T) {
	svc := newTestService()
	user := newTestUser()
	appID := uuid.New()

	token, err := svc.GenerateTwoFactorToken(user, &appID)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims.ApplicationID)
	assert.Equal(t, appID, *claims.ApplicationID)
}

// ============================================================
// ValidateAccessToken Tests
// ============================================================

func TestService_ValidateAccessToken_ShouldReturnClaims_WhenValid(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	assert.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
}

func TestService_ValidateAccessToken_ShouldReturnExpiredError_WhenExpired(t *testing.T) {
	// Create a service with very short expiration
	svc := NewService("access-secret", "refresh-secret", -1*time.Second, 7*24*time.Hour)
	user := newTestUser()

	token, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestService_ValidateAccessToken_ShouldReturnInvalidError_WhenBadSignature(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	// Tamper with the token by changing the last character
	tampered := token[:len(token)-1] + "X"

	claims, err := svc.ValidateAccessToken(tampered)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestService_ValidateAccessToken_ShouldReturnInvalidError_WhenMalformed(t *testing.T) {
	svc := newTestService()

	claims, err := svc.ValidateAccessToken("not.a.valid.jwt")
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestService_ValidateAccessToken_ShouldReturnInvalidError_WhenEmptyString(t *testing.T) {
	svc := newTestService()

	claims, err := svc.ValidateAccessToken("")
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestService_ValidateAccessToken_ShouldRejectRefreshToken(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	// Generate a refresh token (signed with refresh secret)
	refreshToken, err := svc.GenerateRefreshToken(user)
	require.NoError(t, err)

	// Validating as access token should fail (different secret)
	claims, err := svc.ValidateAccessToken(refreshToken)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestService_ValidateAccessToken_ShouldRejectNonHMACSigningMethod(t *testing.T) {
	svc := newTestService()

	// Create a token with an unexpected signing method (none)
	claims := &Claims{
		UserID:   uuid.New(),
		Email:    "test@test.com",
		Username: "test",
		Roles:    []string{"user"},
		IsActive: true,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwtlib.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	result, err := svc.ValidateAccessToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

// ============================================================
// ValidateRefreshToken Tests
// ============================================================

func TestService_ValidateRefreshToken_ShouldReturnClaims_WhenValid(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateRefreshToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateRefreshToken(token)
	assert.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
}

func TestService_ValidateRefreshToken_ShouldReturnExpiredError_WhenExpired(t *testing.T) {
	svc := NewService("access-secret", "refresh-secret", 15*time.Minute, -1*time.Second)
	user := newTestUser()

	token, err := svc.GenerateRefreshToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateRefreshToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestService_ValidateRefreshToken_ShouldRejectAccessToken(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	// Generate an access token (signed with access secret)
	accessToken, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	// Validating as refresh token should fail (different secret)
	claims, err := svc.ValidateRefreshToken(accessToken)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestService_ValidateRefreshToken_ShouldReturnInvalidError_WhenMalformed(t *testing.T) {
	svc := newTestService()

	claims, err := svc.ValidateRefreshToken("garbage")
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

// ============================================================
// ExtractClaims Tests (no validation)
// ============================================================

func TestService_ExtractClaims_ShouldReturnClaims_WhenValidToken(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	claims, err := svc.ExtractClaims(token)
	assert.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Username, claims.Username)
}

func TestService_ExtractClaims_ShouldReturnClaims_WhenExpiredToken(t *testing.T) {
	// Create an expired token
	svc := NewService("access-secret", "refresh-secret", -1*time.Second, 7*24*time.Hour)
	user := newTestUser()

	token, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	// ExtractClaims should work even on expired tokens (it uses ParseUnverified)
	claims, err := svc.ExtractClaims(token)
	assert.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
}

func TestService_ExtractClaims_ShouldReturnError_WhenMalformedToken(t *testing.T) {
	svc := newTestService()

	claims, err := svc.ExtractClaims("not-a-jwt-at-all")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestService_ExtractClaims_ShouldReturnError_WhenEmptyString(t *testing.T) {
	svc := newTestService()

	claims, err := svc.ExtractClaims("")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

// ============================================================
// GetAccessTokenExpiration / GetRefreshTokenExpiration Tests
// ============================================================

func TestService_GetAccessTokenExpiration_ShouldReturnConfiguredDuration(t *testing.T) {
	svc := NewService("secret", "secret", 30*time.Minute, 7*24*time.Hour)

	exp := svc.GetAccessTokenExpiration()
	assert.Equal(t, 30*time.Minute, exp)
}

func TestService_GetRefreshTokenExpiration_ShouldReturnConfiguredDuration(t *testing.T) {
	svc := NewService("secret", "secret", 15*time.Minute, 14*24*time.Hour)

	exp := svc.GetRefreshTokenExpiration()
	assert.Equal(t, 14*24*time.Hour, exp)
}

func TestService_GetExpirations_ShouldReturnDifferentValues(t *testing.T) {
	svc := newTestService()

	accessExp := svc.GetAccessTokenExpiration()
	refreshExp := svc.GetRefreshTokenExpiration()

	assert.Equal(t, 15*time.Minute, accessExp)
	assert.Equal(t, 7*24*time.Hour, refreshExp)
	assert.Less(t, accessExp, refreshExp)
}

// ============================================================
// GenerateAccessTokenWithApp / GenerateRefreshTokenWithApp Tests
// ============================================================

func TestService_GenerateAccessTokenWithApp_ShouldBindApplicationID(t *testing.T) {
	svc := newTestService()
	user := newTestUser()
	appID := uuid.New()

	token, err := svc.GenerateAccessTokenWithApp(user, appID)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims.ApplicationID)
	assert.Equal(t, appID, *claims.ApplicationID)
}

func TestService_GenerateRefreshTokenWithApp_ShouldBindApplicationID(t *testing.T) {
	svc := newTestService()
	user := newTestUser()
	appID := uuid.New()

	token, err := svc.GenerateRefreshTokenWithApp(user, appID)
	require.NoError(t, err)

	claims, err := svc.ValidateRefreshToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims.ApplicationID)
	assert.Equal(t, appID, *claims.ApplicationID)
}

// ============================================================
// GetApplicationIDFromClaims Tests
// ============================================================

func TestGetApplicationIDFromClaims_ShouldReturnAppID_WhenPresent(t *testing.T) {
	appID := uuid.New()
	claims := &Claims{ApplicationID: &appID}

	result := GetApplicationIDFromClaims(claims)
	require.NotNil(t, result)
	assert.Equal(t, appID, *result)
}

func TestGetApplicationIDFromClaims_ShouldReturnNil_WhenAbsent(t *testing.T) {
	claims := &Claims{}

	result := GetApplicationIDFromClaims(claims)
	assert.Nil(t, result)
}

func TestGetApplicationIDFromClaims_ShouldReturnNil_WhenClaimsNil(t *testing.T) {
	result := GetApplicationIDFromClaims(nil)
	assert.Nil(t, result)
}

// ============================================================
// Cross-Secret Validation Tests
// ============================================================

func TestService_ShouldUseDifferentSecrets_ForAccessAndRefreshTokens(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	accessToken, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	refreshToken, err := svc.GenerateRefreshToken(user)
	require.NoError(t, err)

	// Access token should not validate as refresh token
	_, err = svc.ValidateRefreshToken(accessToken)
	assert.Error(t, err)

	// Refresh token should not validate as access token
	_, err = svc.ValidateAccessToken(refreshToken)
	assert.Error(t, err)
}

// ============================================================
// Edge Cases
// ============================================================

func TestService_GenerateAccessToken_ShouldWorkWithInactiveUser(t *testing.T) {
	svc := newTestService()
	user := &models.User{
		ID:       uuid.New(),
		Email:    "inactive@example.com",
		Username: "inactive",
		IsActive: false,
		Roles:    []models.Role{{Name: "user"}},
	}

	token, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.False(t, claims.IsActive)
}

func TestService_NewService_ShouldStoreConfiguration(t *testing.T) {
	accessExp := 30 * time.Minute
	refreshExp := 14 * 24 * time.Hour

	svc := NewService("my-access", "my-refresh", accessExp, refreshExp)

	assert.Equal(t, accessExp, svc.GetAccessTokenExpiration())
	assert.Equal(t, refreshExp, svc.GetRefreshTokenExpiration())
}

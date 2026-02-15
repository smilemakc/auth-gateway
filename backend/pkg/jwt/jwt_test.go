package jwt

import (
	"testing"
	"time"

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

func TestAccessTokenContainsJTI(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateAccessToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)

	assert.NotEmpty(t, claims.ID, "access token must contain a non-empty jti (ID) claim")

	_, parseErr := uuid.Parse(claims.ID)
	assert.NoError(t, parseErr, "jti must be a valid UUID")
}

func TestRefreshTokenContainsJTI(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateRefreshToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateRefreshToken(token)
	require.NoError(t, err)

	assert.NotEmpty(t, claims.ID, "refresh token must contain a non-empty jti (ID) claim")

	_, parseErr := uuid.Parse(claims.ID)
	assert.NoError(t, parseErr, "jti must be a valid UUID")
}

func TestTwoFactorTokenContainsJTI(t *testing.T) {
	svc := newTestService()
	user := newTestUser()

	token, err := svc.GenerateTwoFactorToken(user)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)

	assert.NotEmpty(t, claims.ID, "two-factor token must contain a non-empty jti (ID) claim")

	_, parseErr := uuid.Parse(claims.ID)
	assert.NoError(t, parseErr, "jti must be a valid UUID")
}

func TestEachTokenGetsUniqueJTI(t *testing.T) {
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

	assert.NotEqual(t, claims1.ID, claims2.ID, "each token must have a unique jti")
}

package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAPIKey(t *testing.T) {
	service := &APIKeyService{}

	// Test key generation
	key1, err := service.GenerateAPIKey()
	assert.NoError(t, err)
	assert.NotEmpty(t, key1)
	assert.True(t, len(key1) >= 20)
	assert.True(t, strings.HasPrefix(key1, "agw_"))

	// Test uniqueness
	key2, err := service.GenerateAPIKey()
	assert.NoError(t, err)
	assert.NotEqual(t, key1, key2)

	// Test multiple generations
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, err := service.GenerateAPIKey()
		assert.NoError(t, err)
		if keys[key] {
			t.Errorf("Duplicate key generated: %s", key)
		}
		keys[key] = true
	}
}

func TestAPIKeyService_Create(t *testing.T) {
	mockApiKeyStore := &mockAPIKeyStore{}
	mockUserStore := &mockUserStore{}
	mockAuditLogger := &mockAuditLogger{}

	service := NewAPIKeyService(mockApiKeyStore, mockUserStore, mockAuditLogger)

	ctx := context.Background()
	userID := uuid.New()
	req := &models.CreateAPIKeyRequest{
		Name:   "Test Key",
		Scopes: []string{"users:read"},
	}

	t.Run("Success", func(t *testing.T) {
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID}, nil
		}
		mockApiKeyStore.CreateFunc = func(ctx context.Context, apiKey *models.APIKey) error {
			assert.Equal(t, userID, apiKey.UserID)
			return nil
		}

		resp, err := service.Create(ctx, userID, req, "127.0.0.1", "test-agent")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.PlainKey)
	})

	t.Run("InvalidScope", func(t *testing.T) {
		reqInvalid := &models.CreateAPIKeyRequest{
			Name:   "Test Key",
			Scopes: []string{"invalid:scope"},
		}
		resp, err := service.Create(ctx, userID, reqInvalid, "127.0.0.1", "test-agent")
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, errors.New("user not found")
		}

		resp, err := service.Create(ctx, userID, req, "127.0.0.1", "test-agent")
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestAPIKeyService_ValidateAPIKey(t *testing.T) {
	mockApiKeyStore := &mockAPIKeyStore{}
	mockUserStore := &mockUserStore{}
	mockAuditLogger := &mockAuditLogger{}

	service := NewAPIKeyService(mockApiKeyStore, mockUserStore, mockAuditLogger)
	ctx := context.Background()
	plainKey := "agw_testkey12345"
	userID := uuid.New()

	t.Run("Valid", func(t *testing.T) {
		mockApiKeyStore.GetByKeyHashFunc = func(ctx context.Context, keyHash string) (*models.APIKey, error) {
			return &models.APIKey{
				ID:       uuid.New(),
				UserID:   userID,
				KeyHash:  keyHash,
				IsActive: true,
			}, nil
		}
		mockUserStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID}, nil
		}
		mockApiKeyStore.UpdateLastUsedFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		key, user, err := service.ValidateAPIKey(ctx, plainKey)
		assert.NoError(t, err)
		assert.NotNil(t, key)
		assert.NotNil(t, user)
	})

	t.Run("InvalidKey", func(t *testing.T) {
		mockApiKeyStore.GetByKeyHashFunc = func(ctx context.Context, keyHash string) (*models.APIKey, error) {
			return nil, errors.New("not found")
		}

		key, user, err := service.ValidateAPIKey(ctx, plainKey)
		assert.ErrorIs(t, err, models.ErrInvalidToken)
		assert.Nil(t, key)
		assert.Nil(t, user)
	})

	t.Run("RevokedKey", func(t *testing.T) {
		mockApiKeyStore.GetByKeyHashFunc = func(ctx context.Context, keyHash string) (*models.APIKey, error) {
			return &models.APIKey{
				ID:       uuid.New(),
				UserID:   userID,
				IsActive: false,
			}, nil
		}

		key, user, err := service.ValidateAPIKey(ctx, plainKey)
		assert.Error(t, err) // Should be 401 revoked
		assert.Nil(t, key)
		assert.Nil(t, user)
	})

	t.Run("ExpiredKey", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour)
		mockApiKeyStore.GetByKeyHashFunc = func(ctx context.Context, keyHash string) (*models.APIKey, error) {
			return &models.APIKey{
				ID:        uuid.New(),
				UserID:    userID,
				IsActive:  true,
				ExpiresAt: &past,
			}, nil
		}

		key, user, err := service.ValidateAPIKey(ctx, plainKey)
		assert.Error(t, err) // Should be 401 expired
		assert.Nil(t, key)
		assert.Nil(t, user)
	})
}

func TestAPIKeyService_Revoke(t *testing.T) {
	mockApiKeyStore := &mockAPIKeyStore{}
	mockUserStore := &mockUserStore{}
	mockAuditLogger := &mockAuditLogger{}

	service := NewAPIKeyService(mockApiKeyStore, mockUserStore, mockAuditLogger)
	ctx := context.Background()
	userID := uuid.New()
	keyID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockApiKeyStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return &models.APIKey{ID: keyID, UserID: userID}, nil
		}
		mockApiKeyStore.RevokeFunc = func(ctx context.Context, id uuid.UUID) error {
			assert.Equal(t, keyID, id)
			return nil
		}
		mockAuditLogger.LogWithActionFunc = func(userID *uuid.UUID, action, status, ip, userAgent string, details map[string]interface{}) {
			// Check audit log
		}

		err := service.Revoke(ctx, userID, keyID, "ip", "ua")
		assert.NoError(t, err)
	})

	t.Run("Forbidden", func(t *testing.T) {
		otherUserID := uuid.New()
		mockApiKeyStore.GetByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return &models.APIKey{ID: keyID, UserID: otherUserID}, nil
		}

		err := service.Revoke(ctx, userID, keyID, "ip", "ua")
		assert.ErrorIs(t, err, models.ErrForbidden)
	})
}

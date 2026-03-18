package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Existing tests (preserved)
// ============================================================

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
				KeyHash:  keyHash,
				IsActive: false,
			}, nil
		}

		key, user, err := service.ValidateAPIKey(ctx, plainKey)
		assert.Error(t, err)
		assert.Nil(t, key)
		assert.Nil(t, user)
	})

	t.Run("ExpiredKey", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour)
		mockApiKeyStore.GetByKeyHashFunc = func(ctx context.Context, keyHash string) (*models.APIKey, error) {
			return &models.APIKey{
				ID:        uuid.New(),
				UserID:    userID,
				KeyHash:   keyHash,
				IsActive:  true,
				ExpiresAt: &past,
			}, nil
		}

		key, user, err := service.ValidateAPIKey(ctx, plainKey)
		assert.Error(t, err)
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

// ============================================================
// Commit 9: APIKey Service Additional Tests
// ============================================================

func TestAPIKeyService_GetByID_ShouldReturnKey_WhenOwnerMatches(t *testing.T) {
	// Arrange
	userID := uuid.New()
	keyID := uuid.New()
	scopes, _ := json.Marshal([]string{"users:read"})

	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			assert.Equal(t, keyID, id)
			return &models.APIKey{
				ID:        keyID,
				UserID:    userID,
				Name:      "My Key",
				KeyHash:   "secret-hash",
				KeyPrefix: "agw_abc12345",
				Scopes:    scopes,
				IsActive:  true,
			}, nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	result, err := service.GetByID(ctx, userID, keyID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, keyID, result.ID)
	assert.Equal(t, "My Key", result.Name)
	assert.Empty(t, result.KeyHash, "KeyHash should be stripped by PublicAPIKey")
}

func TestAPIKeyService_GetByID_ShouldReturnError_WhenNotFound(t *testing.T) {
	// Arrange
	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return nil, errors.New("not found")
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	result, err := service.GetByID(ctx, uuid.New(), uuid.New())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAPIKeyService_GetByID_ShouldReturnForbidden_WhenOwnerMismatch(t *testing.T) {
	// Arrange
	userID := uuid.New()
	otherUserID := uuid.New()
	keyID := uuid.New()

	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return &models.APIKey{
				ID:     keyID,
				UserID: otherUserID,
			}, nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	result, err := service.GetByID(ctx, userID, keyID)

	// Assert
	assert.ErrorIs(t, err, models.ErrForbidden)
	assert.Nil(t, result)
}

func TestAPIKeyService_List_ShouldReturnKeys_WhenExists(t *testing.T) {
	// Arrange
	userID := uuid.New()
	scopes, _ := json.Marshal([]string{"users:read"})
	keys := []*models.APIKey{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Name:      "Key 1",
			KeyHash:   "hash1",
			KeyPrefix: "agw_abc",
			Scopes:    scopes,
			IsActive:  true,
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			Name:      "Key 2",
			KeyHash:   "hash2",
			KeyPrefix: "agw_def",
			Scopes:    scopes,
			IsActive:  true,
		},
	}

	mockApiKeyStore := &mockAPIKeyStore{
		GetByUserIDFunc: func(ctx context.Context, uid uuid.UUID, opts ...APIKeyGetOption) ([]*models.APIKey, error) {
			assert.Equal(t, userID, uid)
			return keys, nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	result, err := service.List(ctx, userID)

	// Assert
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "Key 1", result[0].Name)
	assert.Equal(t, "Key 2", result[1].Name)
	// KeyHash should be stripped by PublicAPIKey
	assert.Empty(t, result[0].KeyHash)
	assert.Empty(t, result[1].KeyHash)
}

func TestAPIKeyService_List_ShouldReturnEmpty_WhenNoKeys(t *testing.T) {
	// Arrange
	mockApiKeyStore := &mockAPIKeyStore{
		GetByUserIDFunc: func(ctx context.Context, uid uuid.UUID, opts ...APIKeyGetOption) ([]*models.APIKey, error) {
			return []*models.APIKey{}, nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	result, err := service.List(ctx, uuid.New())

	// Assert
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestAPIKeyService_List_ShouldReturnError_WhenRepoFails(t *testing.T) {
	// Arrange
	mockApiKeyStore := &mockAPIKeyStore{
		GetByUserIDFunc: func(ctx context.Context, uid uuid.UUID, opts ...APIKeyGetOption) ([]*models.APIKey, error) {
			return nil, errors.New("database error")
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	result, err := service.List(ctx, uuid.New())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAPIKeyService_Update_ShouldUpdateFields_WhenValid(t *testing.T) {
	// Arrange
	userID := uuid.New()
	keyID := uuid.New()
	origScopes, _ := json.Marshal([]string{"users:read"})

	var capturedKey *models.APIKey
	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return &models.APIKey{
				ID:          keyID,
				UserID:      userID,
				Name:        "Old Name",
				Description: "Old Desc",
				Scopes:      origScopes,
				IsActive:    true,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, apiKey *models.APIKey) error {
			capturedKey = apiKey
			return nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	isActive := false
	req := &models.UpdateAPIKeyRequest{
		Name:        "New Name",
		Description: "New Desc",
		IsActive:    &isActive,
	}

	// Act
	result, err := service.Update(ctx, userID, keyID, req, "127.0.0.1", "test-agent")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, capturedKey)
	assert.Equal(t, "New Name", capturedKey.Name)
	assert.Equal(t, "New Desc", capturedKey.Description)
	assert.False(t, capturedKey.IsActive)
}

func TestAPIKeyService_Update_ShouldUpdateScopes_WhenNewScopesProvided(t *testing.T) {
	// Arrange
	userID := uuid.New()
	keyID := uuid.New()
	origScopes, _ := json.Marshal([]string{"users:read"})

	var capturedKey *models.APIKey
	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return &models.APIKey{
				ID:       keyID,
				UserID:   userID,
				Name:     "My Key",
				Scopes:   origScopes,
				IsActive: true,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, apiKey *models.APIKey) error {
			capturedKey = apiKey
			return nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	req := &models.UpdateAPIKeyRequest{
		Scopes: []string{"users:read", "users:write"},
	}

	// Act
	result, err := service.Update(ctx, userID, keyID, req, "127.0.0.1", "test-agent")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, capturedKey)

	var newScopes []string
	err = json.Unmarshal(capturedKey.Scopes, &newScopes)
	require.NoError(t, err)
	assert.Equal(t, []string{"users:read", "users:write"}, newScopes)
}

func TestAPIKeyService_Update_ShouldReturnError_WhenNotFound(t *testing.T) {
	// Arrange
	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return nil, errors.New("not found")
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	req := &models.UpdateAPIKeyRequest{Name: "New Name"}

	// Act
	result, err := service.Update(ctx, uuid.New(), uuid.New(), req, "ip", "ua")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAPIKeyService_Update_ShouldReturnForbidden_WhenOwnerMismatch(t *testing.T) {
	// Arrange
	userID := uuid.New()
	otherUserID := uuid.New()
	keyID := uuid.New()

	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return &models.APIKey{
				ID:     keyID,
				UserID: otherUserID,
			}, nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	req := &models.UpdateAPIKeyRequest{Name: "New Name"}

	// Act
	result, err := service.Update(ctx, userID, keyID, req, "ip", "ua")

	// Assert
	assert.ErrorIs(t, err, models.ErrForbidden)
	assert.Nil(t, result)
}

func TestAPIKeyService_Update_ShouldReturnError_WhenInvalidScope(t *testing.T) {
	// Arrange
	userID := uuid.New()
	keyID := uuid.New()
	origScopes, _ := json.Marshal([]string{"users:read"})

	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return &models.APIKey{
				ID:       keyID,
				UserID:   userID,
				Name:     "My Key",
				Scopes:   origScopes,
				IsActive: true,
			}, nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	req := &models.UpdateAPIKeyRequest{
		Scopes: []string{"bogus:scope"},
	}

	// Act
	result, err := service.Update(ctx, userID, keyID, req, "ip", "ua")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid scope")
	assert.Nil(t, result)
}

func TestAPIKeyService_Update_ShouldReturnError_WhenRepoUpdateFails(t *testing.T) {
	// Arrange
	userID := uuid.New()
	keyID := uuid.New()

	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return &models.APIKey{
				ID:       keyID,
				UserID:   userID,
				Name:     "My Key",
				IsActive: true,
			}, nil
		},
		UpdateFunc: func(ctx context.Context, apiKey *models.APIKey) error {
			return errors.New("database error")
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	req := &models.UpdateAPIKeyRequest{Name: "New Name"}

	// Act
	result, err := service.Update(ctx, userID, keyID, req, "ip", "ua")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAPIKeyService_Delete_ShouldSucceed_WhenOwnerMatches(t *testing.T) {
	// Arrange
	userID := uuid.New()
	keyID := uuid.New()

	var deletedID uuid.UUID
	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return &models.APIKey{
				ID:     keyID,
				UserID: userID,
				Name:   "Delete Me",
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			deletedID = id
			return nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	err := service.Delete(ctx, userID, keyID, "127.0.0.1", "test-agent")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, keyID, deletedID)
}

func TestAPIKeyService_Delete_ShouldReturnError_WhenNotFound(t *testing.T) {
	// Arrange
	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return nil, errors.New("not found")
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	err := service.Delete(ctx, uuid.New(), uuid.New(), "ip", "ua")

	// Assert
	assert.Error(t, err)
}

func TestAPIKeyService_Delete_ShouldReturnForbidden_WhenOwnerMismatch(t *testing.T) {
	// Arrange
	userID := uuid.New()
	otherUserID := uuid.New()
	keyID := uuid.New()

	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return &models.APIKey{
				ID:     keyID,
				UserID: otherUserID,
			}, nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	err := service.Delete(ctx, userID, keyID, "ip", "ua")

	// Assert
	assert.ErrorIs(t, err, models.ErrForbidden)
}

func TestAPIKeyService_Delete_ShouldReturnError_WhenRepoDeleteFails(t *testing.T) {
	// Arrange
	userID := uuid.New()
	keyID := uuid.New()

	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return &models.APIKey{
				ID:     keyID,
				UserID: userID,
			}, nil
		},
		DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("database error")
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	err := service.Delete(ctx, userID, keyID, "ip", "ua")

	// Assert
	assert.Error(t, err)
}

func TestAPIKeyService_Revoke_ShouldReturnError_WhenNotFound(t *testing.T) {
	// Arrange
	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return nil, errors.New("not found")
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	err := service.Revoke(ctx, uuid.New(), uuid.New(), "ip", "ua")

	// Assert
	assert.Error(t, err)
}

func TestAPIKeyService_Revoke_ShouldReturnError_WhenRepoRevokeFails(t *testing.T) {
	// Arrange
	userID := uuid.New()
	keyID := uuid.New()

	mockApiKeyStore := &mockAPIKeyStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
			return &models.APIKey{ID: keyID, UserID: userID}, nil
		},
		RevokeFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("database error")
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, &mockUserStore{}, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	err := service.Revoke(ctx, userID, keyID, "ip", "ua")

	// Assert
	assert.Error(t, err)
}

func TestAPIKeyService_ValidateAPIKey_ShouldReturnError_WhenUserNotFound(t *testing.T) {
	// Arrange
	userID := uuid.New()
	plainKey := "agw_testkey12345"

	mockApiKeyStore := &mockAPIKeyStore{
		GetByKeyHashFunc: func(ctx context.Context, keyHash string) (*models.APIKey, error) {
			return &models.APIKey{
				ID:       uuid.New(),
				UserID:   userID,
				KeyHash:  keyHash,
				IsActive: true,
			}, nil
		},
		UpdateLastUsedFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	mockUserStore := &mockUserStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, errors.New("user not found")
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, mockUserStore, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	key, user, err := service.ValidateAPIKey(ctx, plainKey)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, key)
	assert.Nil(t, user)
}

func TestAPIKeyService_ValidateAPIKey_ShouldReturnKeyAndUser_WhenNotExpired(t *testing.T) {
	// Arrange
	userID := uuid.New()
	plainKey := "agw_testkey12345"
	future := time.Now().Add(24 * time.Hour)

	mockApiKeyStore := &mockAPIKeyStore{
		GetByKeyHashFunc: func(ctx context.Context, keyHash string) (*models.APIKey, error) {
			return &models.APIKey{
				ID:        uuid.New(),
				UserID:    userID,
				KeyHash:   keyHash,
				IsActive:  true,
				ExpiresAt: &future,
			}, nil
		},
		UpdateLastUsedFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	mockUserStore := &mockUserStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID}, nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, mockUserStore, &mockAuditLogger{})
	ctx := context.Background()

	// Act
	key, user, err := service.ValidateAPIKey(ctx, plainKey)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, key)
	require.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
}

func TestAPIKeyService_HasScope_ShouldReturnTrue_WhenScopeExists(t *testing.T) {
	service := &APIKeyService{}

	scopes, _ := json.Marshal([]string{"users:read", "token:validate"})
	apiKey := &models.APIKey{Scopes: scopes}

	assert.True(t, service.HasScope(apiKey, models.ScopeReadUsers))
	assert.True(t, service.HasScope(apiKey, models.ScopeValidateToken))
}

func TestAPIKeyService_HasScope_ShouldReturnFalse_WhenScopeAbsent(t *testing.T) {
	service := &APIKeyService{}

	scopes, _ := json.Marshal([]string{"users:read"})
	apiKey := &models.APIKey{Scopes: scopes}

	assert.False(t, service.HasScope(apiKey, models.ScopeWriteUsers))
	assert.False(t, service.HasScope(apiKey, models.ScopeAdmin))
}

func TestAPIKeyService_HasScope_ShouldReturnTrue_WhenAllScope(t *testing.T) {
	service := &APIKeyService{}

	scopes, _ := json.Marshal([]string{"all"})
	apiKey := &models.APIKey{Scopes: scopes}

	assert.True(t, service.HasScope(apiKey, models.ScopeReadUsers))
	assert.True(t, service.HasScope(apiKey, models.ScopeWriteUsers))
	assert.True(t, service.HasScope(apiKey, models.ScopeAdmin))
}

func TestAPIKeyService_HasScope_ShouldReturnFalse_WhenInvalidJSON(t *testing.T) {
	service := &APIKeyService{}

	apiKey := &models.APIKey{Scopes: []byte("invalid-json")}

	assert.False(t, service.HasScope(apiKey, models.ScopeReadUsers))
}

func TestAPIKeyService_Create_ShouldStoreKeyHashNotPlainKey(t *testing.T) {
	// Arrange
	userID := uuid.New()
	var storedKey *models.APIKey

	mockApiKeyStore := &mockAPIKeyStore{
		CreateFunc: func(ctx context.Context, apiKey *models.APIKey) error {
			storedKey = apiKey
			return nil
		},
	}
	mockUserStore := &mockUserStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID}, nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, mockUserStore, &mockAuditLogger{})
	ctx := context.Background()

	req := &models.CreateAPIKeyRequest{
		Name:   "Test Key",
		Scopes: []string{"users:read"},
	}

	// Act
	resp, err := service.Create(ctx, userID, req, "ip", "ua")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, storedKey)

	// The stored hash should NOT equal the plain key
	assert.NotEqual(t, resp.PlainKey, storedKey.KeyHash)
	assert.NotEmpty(t, storedKey.KeyHash)
	// Prefix should be first 12 chars of the plain key
	assert.Equal(t, resp.PlainKey[:12], storedKey.KeyPrefix)
	assert.True(t, storedKey.IsActive)
}

func TestAPIKeyService_Create_ShouldReturnError_WhenRepoCreateFails(t *testing.T) {
	// Arrange
	userID := uuid.New()

	mockApiKeyStore := &mockAPIKeyStore{
		CreateFunc: func(ctx context.Context, apiKey *models.APIKey) error {
			return errors.New("unique constraint violation")
		},
	}
	mockUserStore := &mockUserStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID}, nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, mockUserStore, &mockAuditLogger{})
	ctx := context.Background()

	req := &models.CreateAPIKeyRequest{
		Name:   "Test Key",
		Scopes: []string{"users:read"},
	}

	// Act
	resp, err := service.Create(ctx, userID, req, "ip", "ua")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestAPIKeyService_Create_ShouldSetExpiresAt_WhenProvided(t *testing.T) {
	// Arrange
	userID := uuid.New()
	future := time.Now().Add(30 * 24 * time.Hour)
	var storedKey *models.APIKey

	mockApiKeyStore := &mockAPIKeyStore{
		CreateFunc: func(ctx context.Context, apiKey *models.APIKey) error {
			storedKey = apiKey
			return nil
		},
	}
	mockUserStore := &mockUserStore{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID}, nil
		},
	}
	service := NewAPIKeyService(mockApiKeyStore, mockUserStore, &mockAuditLogger{})
	ctx := context.Background()

	req := &models.CreateAPIKeyRequest{
		Name:      "Test Key",
		Scopes:    []string{"users:read"},
		ExpiresAt: &future,
	}

	// Act
	resp, err := service.Create(ctx, userID, req, "ip", "ua")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, storedKey)
	require.NotNil(t, storedKey.ExpiresAt)
	assert.WithinDuration(t, future, *storedKey.ExpiresAt, time.Second)
}

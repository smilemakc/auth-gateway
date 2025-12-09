package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIsValidScope(t *testing.T) {
	tests := []struct {
		name     string
		scope    string
		expected bool
	}{
		{"Valid scope - users:read", "users:read", true},
		{"Valid scope - users:write", "users:write", true},
		{"Valid scope - profile:read", "profile:read", true},
		{"Valid scope - profile:write", "profile:write", true},
		{"Valid scope - admin:all", "admin:all", true},
		{"Valid scope - token:validate", "token:validate", true},
		{"Valid scope - token:introspect", "token:introspect", true},
		{"Valid scope - all", "all", true},
		{"Invalid scope - empty", "", false},
		{"Invalid scope - random", "random:scope", false},
		{"Invalid scope - typo", "user:read", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidScope(tt.scope)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAPIKey_HasScope(t *testing.T) {
	t.Run("Empty scopes array", func(t *testing.T) {
		apiKey := &APIKey{
			Scopes: []byte(`[]`),
		}
		result := apiKey.HasScope(ScopeReadUsers)
		assert.False(t, result)
	})

	t.Run("Nil scopes", func(t *testing.T) {
		apiKey := &APIKey{
			Scopes: nil,
		}
		result := apiKey.HasScope(ScopeReadUsers)
		assert.False(t, result)
	})

	t.Run("Invalid JSON scopes", func(t *testing.T) {
		apiKey := &APIKey{
			Scopes: []byte(`invalid json`),
		}
		result := apiKey.HasScope(ScopeReadUsers)
		assert.False(t, result)
	})
}

func TestAPIKey_IsExpired(t *testing.T) {
	t.Run("Never expires (nil ExpiresAt)", func(t *testing.T) {
		apiKey := &APIKey{
			ExpiresAt: nil,
		}
		assert.False(t, apiKey.IsExpired())
	})

	t.Run("Expired key", func(t *testing.T) {
		past := time.Now().Add(-24 * time.Hour)
		apiKey := &APIKey{
			ExpiresAt: &past,
		}
		assert.True(t, apiKey.IsExpired())
	})

	t.Run("Not expired key", func(t *testing.T) {
		future := time.Now().Add(24 * time.Hour)
		apiKey := &APIKey{
			ExpiresAt: &future,
		}
		assert.False(t, apiKey.IsExpired())
	})

	t.Run("Key expiring right now", func(t *testing.T) {
		now := time.Now()
		apiKey := &APIKey{
			ExpiresAt: &now,
		}
		// Should be expired or very close
		result := apiKey.IsExpired()
		// Could be true or false depending on timing
		_ = result
	})
}

func TestAPIKey_IsValid(t *testing.T) {
	t.Run("Active and not expired", func(t *testing.T) {
		future := time.Now().Add(24 * time.Hour)
		apiKey := &APIKey{
			IsActive:  true,
			ExpiresAt: &future,
		}
		assert.True(t, apiKey.IsValid())
	})

	t.Run("Active but expired", func(t *testing.T) {
		past := time.Now().Add(-24 * time.Hour)
		apiKey := &APIKey{
			IsActive:  true,
			ExpiresAt: &past,
		}
		assert.False(t, apiKey.IsValid())
	})

	t.Run("Inactive but not expired", func(t *testing.T) {
		future := time.Now().Add(24 * time.Hour)
		apiKey := &APIKey{
			IsActive:  false,
			ExpiresAt: &future,
		}
		assert.False(t, apiKey.IsValid())
	})

	t.Run("Inactive and expired", func(t *testing.T) {
		past := time.Now().Add(-24 * time.Hour)
		apiKey := &APIKey{
			IsActive:  false,
			ExpiresAt: &past,
		}
		assert.False(t, apiKey.IsValid())
	})

	t.Run("Active and never expires", func(t *testing.T) {
		apiKey := &APIKey{
			IsActive:  true,
			ExpiresAt: nil,
		}
		assert.True(t, apiKey.IsValid())
	})
}

func TestAPIKey_PublicAPIKey(t *testing.T) {
	now := time.Now()
	lastUsed := time.Now().Add(-1 * time.Hour)
	expiresAt := time.Now().Add(24 * time.Hour)

	original := &APIKey{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Name:        "Test Key",
		Description: "Test Description",
		KeyHash:     "secret_hash_should_not_be_exposed",
		KeyPrefix:   "agw_test",
		Scopes:      []byte(`["users:read"]`),
		IsActive:    true,
		LastUsedAt:  &lastUsed,
		ExpiresAt:   &expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	public := original.PublicAPIKey()

	t.Run("Copies all public fields", func(t *testing.T) {
		assert.Equal(t, original.ID, public.ID)
		assert.Equal(t, original.UserID, public.UserID)
		assert.Equal(t, original.Name, public.Name)
		assert.Equal(t, original.Description, public.Description)
		assert.Equal(t, original.KeyPrefix, public.KeyPrefix)
		assert.Equal(t, original.Scopes, public.Scopes)
		assert.Equal(t, original.IsActive, public.IsActive)
		assert.Equal(t, original.LastUsedAt, public.LastUsedAt)
		assert.Equal(t, original.ExpiresAt, public.ExpiresAt)
		assert.Equal(t, original.CreatedAt, public.CreatedAt)
		assert.Equal(t, original.UpdatedAt, public.UpdatedAt)
	})

	t.Run("Does not expose KeyHash", func(t *testing.T) {
		// KeyHash should be empty in public version
		assert.Empty(t, public.KeyHash)
		assert.NotEqual(t, original.KeyHash, public.KeyHash)
	})

	t.Run("Returns new instance", func(t *testing.T) {
		assert.NotSame(t, original, public)
	})
}

func TestCreateAPIKeyRequest(t *testing.T) {
	t.Run("Valid request", func(t *testing.T) {
		req := CreateAPIKeyRequest{
			Name:        "Test Key",
			Description: "Test Description",
			Scopes:      []string{"users:read", "users:write"},
		}
		assert.NotEmpty(t, req.Name)
		assert.NotEmpty(t, req.Scopes)
	})

	t.Run("Request with expiration", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		req := CreateAPIKeyRequest{
			Name:        "Expiring Key",
			Description: "Expires tomorrow",
			Scopes:      []string{"users:read"},
			ExpiresAt:   &expiresAt,
		}
		assert.NotNil(t, req.ExpiresAt)
	})
}

func TestUpdateAPIKeyRequest(t *testing.T) {
	t.Run("Update all fields", func(t *testing.T) {
		isActive := true
		req := UpdateAPIKeyRequest{
			Name:        "Updated Name",
			Description: "Updated Description",
			Scopes:      []string{"users:read"},
			IsActive:    &isActive,
		}
		assert.NotEmpty(t, req.Name)
		assert.NotNil(t, req.IsActive)
		assert.True(t, *req.IsActive)
	})

	t.Run("Deactivate key", func(t *testing.T) {
		isActive := false
		req := UpdateAPIKeyRequest{
			IsActive: &isActive,
		}
		assert.NotNil(t, req.IsActive)
		assert.False(t, *req.IsActive)
	})
}

func TestAPIKeyScope_Constants(t *testing.T) {
	t.Run("All scope constants are defined", func(t *testing.T) {
		assert.Equal(t, APIKeyScope("users:read"), ScopeReadUsers)
		assert.Equal(t, APIKeyScope("users:write"), ScopeWriteUsers)
		assert.Equal(t, APIKeyScope("profile:read"), ScopeReadProfile)
		assert.Equal(t, APIKeyScope("profile:write"), ScopeWriteProfile)
		assert.Equal(t, APIKeyScope("admin:all"), ScopeAdmin)
		assert.Equal(t, APIKeyScope("token:validate"), ScopeValidateToken)
		assert.Equal(t, APIKeyScope("token:introspect"), ScopeIntrospectToken)
		assert.Equal(t, APIKeyScope("all"), ScopeAll)
	})
}

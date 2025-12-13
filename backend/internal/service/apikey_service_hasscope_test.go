package service

import (
	"encoding/json"
	"testing"

	"github.com/smilemakc/auth-gateway/internal/models"
)

func TestAPIKeyService_HasScope(t *testing.T) {
	service := &APIKeyService{}

	t.Run("Has specific scope", func(t *testing.T) {
		scopes, _ := json.Marshal([]string{"users:read", "users:write"})
		apiKey := &models.APIKey{
			Scopes: scopes,
		}

		result := service.HasScope(apiKey, models.ScopeReadUsers)
		if !result {
			t.Error("Expected HasScope to return true for users:read")
		}
	})

	t.Run("Has all scope", func(t *testing.T) {
		scopes, _ := json.Marshal([]string{"all"})
		apiKey := &models.APIKey{
			Scopes: scopes,
		}

		result := service.HasScope(apiKey, models.ScopeReadUsers)
		if !result {
			t.Error("Expected HasScope to return true for 'all' scope")
		}
	})

	t.Run("Does not have scope", func(t *testing.T) {
		scopes, _ := json.Marshal([]string{"users:read"})
		apiKey := &models.APIKey{
			Scopes: scopes,
		}

		result := service.HasScope(apiKey, models.ScopeWriteUsers)
		if result {
			t.Error("Expected HasScope to return false for users:write")
		}
	})

	t.Run("Invalid JSON scopes", func(t *testing.T) {
		apiKey := &models.APIKey{
			Scopes: []byte("invalid json"),
		}

		result := service.HasScope(apiKey, models.ScopeReadUsers)
		if result {
			t.Error("Expected HasScope to return false for invalid JSON")
		}
	})

	t.Run("Empty scopes", func(t *testing.T) {
		scopes, _ := json.Marshal([]string{})
		apiKey := &models.APIKey{
			Scopes: scopes,
		}

		result := service.HasScope(apiKey, models.ScopeReadUsers)
		if result {
			t.Error("Expected HasScope to return false for empty scopes")
		}
	})

	t.Run("Nil scopes", func(t *testing.T) {
		apiKey := &models.APIKey{
			Scopes: nil,
		}

		result := service.HasScope(apiKey, models.ScopeReadUsers)
		if result {
			t.Error("Expected HasScope to return false for nil scopes")
		}
	})

	t.Run("Multiple scopes with match", func(t *testing.T) {
		scopes, _ := json.Marshal([]string{"users:read", "profile:read", "token:validate"})
		apiKey := &models.APIKey{
			Scopes: scopes,
		}

		result := service.HasScope(apiKey, models.ScopeValidateToken)
		if !result {
			t.Error("Expected HasScope to return true for token:validate")
		}
	})
}

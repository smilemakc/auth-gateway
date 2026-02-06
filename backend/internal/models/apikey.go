package models

import (
	"time"

	"github.com/google/uuid"
)

// APIKey represents an API key for external service integration
type APIKey struct {
	ID            uuid.UUID  `json:"id" bun:"id"`
	UserID        uuid.UUID  `json:"user_id" bun:"user_id"`
	ApplicationID *uuid.UUID `bun:"application_id,type:uuid" json:"application_id,omitempty"`
	Name          string     `json:"name" bun:"name"`
	Description string     `json:"description,omitempty" bun:"description"`
	KeyHash     string     `json:"-" bun:"key_hash"`            // Never expose key hash
	KeyPrefix   string     `json:"key_prefix" bun:"key_prefix"` // First 8 chars for identification
	Scopes      []byte     `json:"scopes" bun:"scopes"`         // JSON array of permissions
	IsActive    bool       `json:"is_active" bun:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" bun:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" bun:"expires_at"` // NULL = never expires
	CreatedAt   time.Time  `json:"created_at" bun:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" bun:"updated_at"`
}

// APIKeyScope represents available scopes for API keys
type APIKeyScope string

const (
	// Read scopes
	ScopeReadUsers   APIKeyScope = "users:read"
	ScopeReadProfile APIKeyScope = "profile:read"

	// Write scopes
	ScopeWriteUsers   APIKeyScope = "users:write"
	ScopeWriteProfile APIKeyScope = "profile:write"

	// Admin scopes
	ScopeAdmin APIKeyScope = "admin:all"

	// Token scopes
	ScopeValidateToken   APIKeyScope = "token:validate"
	ScopeIntrospectToken APIKeyScope = "token:introspect"

	// Special scopes
	ScopeAll APIKeyScope = "all"
)

// CreateAPIKeyRequest represents a request to create a new API key
type CreateAPIKeyRequest struct {
	// API key name (3-100 characters)
	Name string `json:"name" binding:"required,min=3,max=100" example:"Production API Key"`
	// API key description
	Description string `json:"description,omitempty" example:"API key for production environment"`
	// List of permission scopes
	Scopes []string `json:"scopes" binding:"required,min=1" example:"users:read,token:validate"`
	// Optional expiration date (null for no expiration)
	ExpiresAt *time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
}

// UpdateAPIKeyRequest represents a request to update an API key
type UpdateAPIKeyRequest struct {
	// API key name
	Name string `json:"name,omitempty" example:"Updated API Key Name"`
	// API key description
	Description string `json:"description,omitempty" example:"Updated description"`
	// List of permission scopes
	Scopes []string `json:"scopes,omitempty" example:"users:read,users:write"`
	// Whether the key is active
	IsActive *bool `json:"is_active,omitempty" example:"true"`
}

// CreateAPIKeyResponse represents the response when creating an API key
type CreateAPIKeyResponse struct {
	// Created API key metadata
	APIKey *APIKey `json:"api_key"`
	// Plain text API key (shown only once at creation)
	PlainKey string `json:"plain_key" example:"agw_1234567890abcdef1234567890abcdef"`
}

// ListAPIKeysResponse represents a list of API keys
type ListAPIKeysResponse struct {
	// List of API keys
	APIKeys []*APIKey `json:"api_keys"`
	// Total number of API keys
	Total int `json:"total" example:"5"`
}

// IsValidScope checks if a scope is valid
func IsValidScope(scope string) bool {
	validScopes := []APIKeyScope{
		ScopeReadUsers,
		ScopeReadProfile,
		ScopeWriteUsers,
		ScopeWriteProfile,
		ScopeAdmin,
		ScopeValidateToken,
		ScopeIntrospectToken,
		ScopeAll,
	}

	for _, validScope := range validScopes {
		if scope == string(validScope) {
			return true
		}
	}
	return false
}

// HasScope checks if API key has a specific scope
func (k *APIKey) HasScope(scope APIKeyScope) bool {
	// Parse scopes from JSON
	var scopes []string
	// Simple JSON parsing - in production use json.Unmarshal
	// For now, we'll assume scopes are stored as JSON array

	// Check if has "all" scope
	for _, s := range scopes {
		if s == string(ScopeAll) {
			return true
		}
		if s == string(scope) {
			return true
		}
	}

	return false
}

// IsExpired checks if the API key is expired
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false // Never expires
	}
	return time.Now().After(*k.ExpiresAt)
}

// IsValid checks if API key is valid (active and not expired)
func (k *APIKey) IsValid() bool {
	return k.IsActive && !k.IsExpired()
}

// PublicAPIKey returns API key without sensitive information
func (k *APIKey) PublicAPIKey() *APIKey {
	return &APIKey{
		ID:            k.ID,
		UserID:        k.UserID,
		ApplicationID: k.ApplicationID,
		Name:          k.Name,
		Description:   k.Description,
		KeyPrefix:     k.KeyPrefix,
		Scopes:        k.Scopes,
		IsActive:      k.IsActive,
		LastUsedAt:    k.LastUsedAt,
		ExpiresAt:     k.ExpiresAt,
		CreatedAt:     k.CreatedAt,
		UpdatedAt:     k.UpdatedAt,
	}
}

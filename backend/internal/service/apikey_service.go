package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

const (
	APIKeyPrefix = "agw"
	APIKeyLength = 32
)

type APIKeyService struct {
	apiKeyRepo   *repository.APIKeyRepository
	userRepo     *repository.UserRepository
	auditService *AuditService
}

func NewAPIKeyService(
	apiKeyRepo *repository.APIKeyRepository,
	userRepo *repository.UserRepository,
	auditService *AuditService,
) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo:   apiKeyRepo,
		userRepo:     userRepo,
		auditService: auditService,
	}
}

// GenerateAPIKey generates a new random API key
func (s *APIKeyService) GenerateAPIKey() (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, APIKeyLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64
	randomPart := base64.URLEncoding.EncodeToString(randomBytes)

	// Format: agw_<random>
	apiKey := fmt.Sprintf("%s_%s", APIKeyPrefix, randomPart)

	return apiKey, nil
}

// Create creates a new API key for a user
func (s *APIKeyService) Create(ctx context.Context, userID uuid.UUID, req *models.CreateAPIKeyRequest, ip, userAgent string) (*models.CreateAPIKeyResponse, error) {
	// Validate scopes
	for _, scope := range req.Scopes {
		if !models.IsValidScope(scope) {
			return nil, models.NewAppError(400, fmt.Sprintf("Invalid scope: %s", scope))
		}
	}

	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Generate API key
	plainKey, err := s.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash the key
	keyHash := utils.HashToken(plainKey)

	// Get key prefix (first 12 chars: "agw_" + first 8 of random part)
	keyPrefix := plainKey[:12]

	// Convert scopes to JSON
	scopesJSON, err := json.Marshal(req.Scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scopes: %w", err)
	}

	// Create API key record
	apiKey := &models.APIKey{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		KeyHash:     keyHash,
		KeyPrefix:   keyPrefix,
		Scopes:      scopesJSON,
		IsActive:    true,
		ExpiresAt:   req.ExpiresAt,
	}

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		s.logAudit(&userID, "api_key_create", "failed", ip, userAgent, map[string]interface{}{
			"name":  req.Name,
			"error": err.Error(),
		})
		return nil, err
	}

	// Log successful creation
	s.logAudit(&userID, "api_key_create", "success", ip, userAgent, map[string]interface{}{
		"api_key_id": apiKey.ID.String(),
		"name":       req.Name,
		"scopes":     req.Scopes,
	})

	return &models.CreateAPIKeyResponse{
		APIKey:   apiKey.PublicAPIKey(),
		PlainKey: plainKey, // Only returned once!
	}, nil
}

// ValidateAPIKey validates an API key and returns the associated user
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, plainKey string) (*models.APIKey, *models.User, error) {
	// Hash the key
	keyHash := utils.HashToken(plainKey)

	// Get API key from database
	apiKey, err := s.apiKeyRepo.GetByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, nil, models.ErrInvalidToken
	}

	// Check if active
	if !apiKey.IsActive {
		return nil, nil, models.NewAppError(401, "API key is revoked")
	}

	// Check if expired
	if apiKey.IsExpired() {
		return nil, nil, models.NewAppError(401, "API key is expired")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, apiKey.UserID)
	if err != nil {
		return nil, nil, err
	}

	// Update last used timestamp (async)
	go func() {
		_ = s.apiKeyRepo.UpdateLastUsed(ctx, apiKey.ID)
	}()

	return apiKey, user, nil
}

// GetByID retrieves an API key by ID
func (s *APIKeyService) GetByID(ctx context.Context, userID, apiKeyID uuid.UUID) (*models.APIKey, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, apiKeyID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if apiKey.UserID != userID {
		return nil, models.ErrForbidden
	}

	return apiKey.PublicAPIKey(), nil
}

// List retrieves all API keys for a user
func (s *APIKeyService) List(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error) {
	apiKeys, err := s.apiKeyRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Return public keys
	publicKeys := make([]*models.APIKey, len(apiKeys))
	for i, key := range apiKeys {
		publicKeys[i] = key.PublicAPIKey()
	}

	return publicKeys, nil
}

// Update updates an API key
func (s *APIKeyService) Update(ctx context.Context, userID, apiKeyID uuid.UUID, req *models.UpdateAPIKeyRequest, ip, userAgent string) (*models.APIKey, error) {
	// Get existing API key
	apiKey, err := s.apiKeyRepo.GetByID(ctx, apiKeyID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if apiKey.UserID != userID {
		return nil, models.ErrForbidden
	}

	// Update fields
	if req.Name != "" {
		apiKey.Name = req.Name
	}
	if req.Description != "" {
		apiKey.Description = req.Description
	}
	if req.IsActive != nil {
		apiKey.IsActive = *req.IsActive
	}
	if req.Scopes != nil {
		// Validate scopes
		for _, scope := range req.Scopes {
			if !models.IsValidScope(scope) {
				return nil, models.NewAppError(400, fmt.Sprintf("Invalid scope: %s", scope))
			}
		}

		scopesJSON, err := json.Marshal(req.Scopes)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal scopes: %w", err)
		}
		apiKey.Scopes = scopesJSON
	}

	// Save changes
	if err := s.apiKeyRepo.Update(ctx, apiKey); err != nil {
		s.logAudit(&userID, "api_key_update", "failed", ip, userAgent, map[string]interface{}{
			"api_key_id": apiKeyID.String(),
			"error":      err.Error(),
		})
		return nil, err
	}

	// Log successful update
	s.logAudit(&userID, "api_key_update", "success", ip, userAgent, map[string]interface{}{
		"api_key_id": apiKey.ID.String(),
		"name":       apiKey.Name,
	})

	return apiKey.PublicAPIKey(), nil
}

// Revoke revokes an API key
func (s *APIKeyService) Revoke(ctx context.Context, userID, apiKeyID uuid.UUID, ip, userAgent string) error {
	// Get API key
	apiKey, err := s.apiKeyRepo.GetByID(ctx, apiKeyID)
	if err != nil {
		return err
	}

	// Verify ownership
	if apiKey.UserID != userID {
		return models.ErrForbidden
	}

	// Revoke
	if err := s.apiKeyRepo.Revoke(ctx, apiKeyID); err != nil {
		s.logAudit(&userID, "api_key_revoke", "failed", ip, userAgent, map[string]interface{}{
			"api_key_id": apiKeyID.String(),
			"error":      err.Error(),
		})
		return err
	}

	// Log successful revocation
	s.logAudit(&userID, "api_key_revoke", "success", ip, userAgent, map[string]interface{}{
		"api_key_id": apiKey.ID.String(),
		"name":       apiKey.Name,
	})

	return nil
}

// Delete permanently deletes an API key
func (s *APIKeyService) Delete(ctx context.Context, userID, apiKeyID uuid.UUID, ip, userAgent string) error {
	// Get API key
	apiKey, err := s.apiKeyRepo.GetByID(ctx, apiKeyID)
	if err != nil {
		return err
	}

	// Verify ownership
	if apiKey.UserID != userID {
		return models.ErrForbidden
	}

	// Delete
	if err := s.apiKeyRepo.Delete(ctx, apiKeyID); err != nil {
		s.logAudit(&userID, "api_key_delete", "failed", ip, userAgent, map[string]interface{}{
			"api_key_id": apiKeyID.String(),
			"error":      err.Error(),
		})
		return err
	}

	// Log successful deletion
	s.logAudit(&userID, "api_key_delete", "success", ip, userAgent, map[string]interface{}{
		"api_key_id": apiKey.ID.String(),
		"name":       apiKey.Name,
	})

	return nil
}

// HasScope checks if an API key has a specific scope
func (s *APIKeyService) HasScope(apiKey *models.APIKey, scope models.APIKeyScope) bool {
	var scopes []string
	if err := json.Unmarshal(apiKey.Scopes, &scopes); err != nil {
		return false
	}

	// Check for "all" scope
	for _, s := range scopes {
		if s == string(models.ScopeAll) {
			return true
		}
		if s == string(scope) {
			return true
		}
	}

	return false
}

func (s *APIKeyService) logAudit(userID *uuid.UUID, action, status, ip, userAgent string, details map[string]interface{}) {
	s.auditService.LogWithAction(userID, action, status, ip, userAgent, details)
}

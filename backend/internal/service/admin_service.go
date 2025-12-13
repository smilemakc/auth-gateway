package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
)

// AdminService provides admin operations
type AdminService struct {
	userRepo   *repository.UserRepository
	apiKeyRepo *repository.APIKeyRepository
	auditRepo  *repository.AuditRepository
	oauthRepo  *repository.OAuthRepository
}

// NewAdminService creates a new admin service
func NewAdminService(
	userRepo *repository.UserRepository,
	apiKeyRepo *repository.APIKeyRepository,
	auditRepo *repository.AuditRepository,
	oauthRepo *repository.OAuthRepository,
) *AdminService {
	return &AdminService{
		userRepo:   userRepo,
		apiKeyRepo: apiKeyRepo,
		auditRepo:  auditRepo,
		oauthRepo:  oauthRepo,
	}
}

// GetStats returns system statistics
func (s *AdminService) GetStats(ctx context.Context) (*models.AdminStatsResponse, error) {
	stats := &models.AdminStatsResponse{
		UsersByRole: make(map[string]int),
	}

	// Total users
	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	stats.TotalUsers = totalUsers

	// Get all users for detailed stats
	users, err := s.userRepo.List(ctx, 10000, 0) // Get all users
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	for _, user := range users {
		if user.IsActive {
			stats.ActiveUsers++
		}
		if user.EmailVerified {
			stats.VerifiedEmailUsers++
		}
		if user.PhoneVerified {
			stats.VerifiedPhoneUsers++
		}
		if user.TOTPEnabled {
			stats.Users2FAEnabled++
		}

		// Count by role
		stats.UsersByRole[user.Role]++

		// Recent signups
		if user.CreatedAt.After(yesterday) {
			stats.RecentSignups++
		}
	}

	// API keys stats
	allAPIKeys, err := s.apiKeyRepo.ListAll(ctx)
	if err == nil {
		stats.TotalAPIKeys = len(allAPIKeys)
		for _, key := range allAPIKeys {
			if key.IsActive {
				stats.ActiveAPIKeys++
			}
		}
	}

	// OAuth accounts stats
	oauthAccounts, err := s.oauthRepo.ListAll(ctx)
	if err == nil {
		stats.TotalOAuthAccounts = len(oauthAccounts)
	}

	// Recent logins from audit logs
	recentLogins, err := s.auditRepo.CountByActionSince(ctx, models.ActionSignIn, yesterday)
	if err == nil {
		stats.RecentLogins = recentLogins
	}

	return stats, nil
}

// ListUsers returns paginated list of users with admin info
func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int) (*models.AdminUserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	users, err := s.userRepo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	adminUsers := make([]*models.AdminUserResponse, 0, len(users))
	for _, user := range users {
		adminUser := s.userToAdminResponse(user)

		// Count API keys
		apiKeys, err := s.apiKeyRepo.GetByUserID(ctx, user.ID)
		if err == nil {
			adminUser.APIKeysCount = len(apiKeys)
		}

		// Count OAuth accounts
		oauthAccounts, err := s.oauthRepo.GetByUserID(ctx, user.ID)
		if err == nil {
			adminUser.OAuthAccountsCount = len(oauthAccounts)
		}

		adminUsers = append(adminUsers, adminUser)
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &models.AdminUserListResponse{
		Users:      adminUsers,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUser returns detailed user information
func (s *AdminService) GetUser(ctx context.Context, userID uuid.UUID) (*models.AdminUserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	adminUser := s.userToAdminResponse(user)

	// Count API keys
	apiKeys, err := s.apiKeyRepo.GetByUserID(ctx, userID)
	if err == nil {
		adminUser.APIKeysCount = len(apiKeys)
	}

	// Count OAuth accounts
	oauthAccounts, err := s.oauthRepo.GetByUserID(ctx, userID)
	if err == nil {
		adminUser.OAuthAccountsCount = len(oauthAccounts)
	}

	return adminUser, nil
}

// UpdateUser updates user information (admin only)
func (s *AdminService) UpdateUser(ctx context.Context, userID uuid.UUID, req *models.AdminUpdateUserRequest) (*models.AdminUserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update role if provided
	if req.Role != nil {
		if !models.IsValidRole(*req.Role) {
			return nil, models.NewAppError(400, "Invalid role")
		}
		user.Role = *req.Role
	}

	// Update active status if provided
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.GetUser(ctx, userID)
}

// DeleteUser deletes a user (soft delete by setting is_active = false)
func (s *AdminService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Prevent deleting the last admin
	if user.Role == string(models.RoleAdmin) {
		adminCount := 0
		users, err := s.userRepo.List(ctx, 10000, 0)
		if err == nil {
			for _, u := range users {
				if u.Role == string(models.RoleAdmin) && u.IsActive {
					adminCount++
				}
			}
			if adminCount <= 1 {
				return models.NewAppError(400, "Cannot delete the last admin user")
			}
		}
	}

	user.IsActive = false
	return s.userRepo.Update(ctx, user)
}

// ListAPIKeys returns all API keys with user information
func (s *AdminService) ListAPIKeys(ctx context.Context, page, pageSize int) ([]*models.AdminAPIKeyResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	apiKeys, err := s.apiKeyRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(apiKeys) {
		return []*models.AdminAPIKeyResponse{}, nil
	}
	if end > len(apiKeys) {
		end = len(apiKeys)
	}

	adminAPIKeys := make([]*models.AdminAPIKeyResponse, 0, end-start)
	for i := start; i < end; i++ {
		key := apiKeys[i]
		user, err := s.userRepo.GetByID(ctx, key.UserID)
		if err != nil {
			continue
		}

		var scopes []string
		if err := json.Unmarshal(key.Scopes, &scopes); err != nil {
			scopes = []string{}
		}

		adminAPIKeys = append(adminAPIKeys, &models.AdminAPIKeyResponse{
			ID:         key.ID,
			UserID:     key.UserID,
			Username:   user.Username,
			Name:       key.Name,
			Prefix:     key.KeyPrefix,
			Scopes:     scopes,
			ExpiresAt:  key.ExpiresAt,
			LastUsedAt: key.LastUsedAt,
			IsRevoked:  !key.IsActive,
			RevokedAt:  nil, // Not tracked in current schema
			CreatedAt:  key.CreatedAt,
		})
	}

	return adminAPIKeys, nil
}

// RevokeAPIKey revokes an API key
func (s *AdminService) RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error {
	return s.apiKeyRepo.Revoke(ctx, keyID)
}

// ListAuditLogs returns paginated audit logs
func (s *AdminService) ListAuditLogs(ctx context.Context, page, pageSize int, userID *uuid.UUID) ([]*models.AdminAuditLogResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	var logs []*models.AuditLog
	var err error

	if userID != nil {
		logs, err = s.auditRepo.GetByUserID(ctx, *userID, pageSize, offset)
	} else {
		logs, err = s.auditRepo.List(ctx, pageSize, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}

	adminLogs := make([]*models.AdminAuditLogResponse, 0, len(logs))
	for _, log := range logs {
		var details map[string]interface{}
		if log.Details != nil {
			json.Unmarshal(log.Details, &details)
		}

		adminLogs = append(adminLogs, &models.AdminAuditLogResponse{
			ID:        log.ID,
			UserID:    log.UserID,
			Action:    string(log.Action),
			Status:    string(log.Status),
			IP:        log.IPAddress,
			UserAgent: log.UserAgent,
			Details:   details,
			CreatedAt: log.CreatedAt,
		})
	}

	return adminLogs, nil
}

// userToAdminResponse converts User to AdminUserResponse
func (s *AdminService) userToAdminResponse(user *models.User) *models.AdminUserResponse {
	return &models.AdminUserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Phone:         user.Phone,
		Username:      user.Username,
		FullName:      user.FullName,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		PhoneVerified: user.PhoneVerified,
		IsActive:      user.IsActive,
		TOTPEnabled:   user.TOTPEnabled,
		TOTPEnabledAt: user.TOTPEnabledAt,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}

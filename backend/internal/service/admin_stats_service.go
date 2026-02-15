package service

import (
	"context"
	"fmt"
	"time"

	"github.com/smilemakc/auth-gateway/internal/models"
)

type AdminStatsService struct {
	userRepo   UserStore
	auditRepo  AuditStore
	apiKeyRepo APIKeyStore
	oauthRepo  OAuthStore
	rbacRepo   RBACStore
}

func (s *AdminStatsService) GetStats(ctx context.Context) (*models.AdminStatsResponse, error) {
	stats := &models.AdminStatsResponse{
		UsersByRole: make(map[string]int),
	}

	totalUsers, err := s.userRepo.Count(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	stats.TotalUsers = totalUsers

	users, err := s.userRepo.List(ctx, UserListLimit(10000), UserListOffset(0))
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

		if user.CreatedAt.After(yesterday) {
			stats.RecentSignups++
		}
	}

	stats.UsersByRole = make(map[string]int)
	for _, user := range users {
		roles, err := s.rbacRepo.GetUserRoles(ctx, user.ID)
		if err == nil {
			for _, role := range roles {
				stats.UsersByRole[role.Name]++
			}
		}
	}

	allAPIKeys, err := s.apiKeyRepo.ListAll(ctx)
	if err == nil {
		stats.TotalAPIKeys = len(allAPIKeys)
		for _, key := range allAPIKeys {
			if key.IsActive {
				stats.ActiveAPIKeys++
			}
		}
	}

	oauthAccounts, err := s.oauthRepo.ListAll(ctx)
	if err == nil {
		stats.TotalOAuthAccounts = len(oauthAccounts)
	}

	recentLogins, err := s.auditRepo.CountByActionSince(ctx, models.ActionSignIn, yesterday)
	if err == nil {
		stats.RecentLogins = recentLogins
	}

	return stats, nil
}

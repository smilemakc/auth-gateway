package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// BlacklistService provides unified token blacklist operations.
// It combines Redis (fast, primary) and PostgreSQL (persistent, fallback) storage.
type BlacklistService struct {
	redis       CacheService
	tokenRepo   TokenStore
	jwtService  TokenService
	logger      *logger.Logger
	auditLogger AuditLogger
	syncStats   SyncStats // Statistics for monitoring sync discrepancies
}

// SyncStats tracks synchronization statistics
type SyncStats struct {
	LastSyncTime   time.Time
	TotalSynced    int64
	SyncErrors     int64
	RedisErrors    int64
	DatabaseErrors int64
	Discrepancies  int64
}

// NewBlacklistService creates a new blacklist service
func NewBlacklistService(redis CacheService, tokenRepo TokenStore, jwtService TokenService, logger *logger.Logger, auditLogger AuditLogger) *BlacklistService {
	return &BlacklistService{
		redis:       redis,
		tokenRepo:   tokenRepo,
		jwtService:  jwtService,
		logger:      logger,
		auditLogger: auditLogger,
		syncStats: SyncStats{
			LastSyncTime: time.Time{},
		},
	}
}

// SyncFromDatabase synchronizes blacklist entries from PostgreSQL to Redis
// This should be called on startup to ensure Redis cache is populated
func (s *BlacklistService) SyncFromDatabase(ctx context.Context) error {
	s.logger.Info("Starting blacklist synchronization from database to Redis")

	// Get all active blacklist entries from database
	tokenRepo, ok := s.tokenRepo.(*repository.TokenRepository)
	if !ok {
		return fmt.Errorf("token repository does not support GetAllActiveBlacklistEntries")
	}

	entries, err := tokenRepo.GetAllActiveBlacklistEntries(ctx)
	if err != nil {
		s.syncStats.DatabaseErrors++
		s.logger.Error("Failed to get blacklist entries from database", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to get blacklist entries: %w", err)
	}

	synced := int64(0)
	errors := int64(0)

	// Sync each entry to Redis
	for _, entry := range entries {
		ttl := time.Until(entry.ExpiresAt)
		if ttl <= 0 {
			// Skip expired entries
			continue
		}

		// Check if already in Redis
		exists, redisErr := s.redis.IsBlacklisted(ctx, entry.TokenHash)
		if redisErr != nil {
			s.syncStats.RedisErrors++
			errors++
			s.logger.Warn("Failed to check Redis during sync", map[string]interface{}{
				"token_hash": entry.TokenHash,
				"error":      redisErr.Error(),
			})
			continue
		}

		if !exists {
			// Add to Redis
			if err := s.redis.AddToBlacklist(ctx, entry.TokenHash, ttl); err != nil {
				s.syncStats.RedisErrors++
				errors++
				s.logger.Warn("Failed to sync entry to Redis", map[string]interface{}{
					"token_hash": entry.TokenHash,
					"error":      err.Error(),
				})
				continue
			}
			synced++
		}
	}

	s.syncStats.LastSyncTime = time.Now()
	s.syncStats.TotalSynced += synced
	s.syncStats.SyncErrors += errors

	s.logger.Info("Blacklist synchronization completed", map[string]interface{}{
		"total_entries": len(entries),
		"synced":        synced,
		"errors":        errors,
	})

	return nil
}

// GetSyncStats returns current synchronization statistics
func (s *BlacklistService) GetSyncStats() SyncStats {
	return s.syncStats
}

// IsBlacklisted checks if a token hash is blacklisted.
// First checks Redis (fast), then falls back to PostgreSQL if needed.
// If found in DB but not in Redis, restores the entry to Redis cache.
func (s *BlacklistService) IsBlacklisted(ctx context.Context, tokenHash string) bool {
	// First check Redis (fast)
	blacklisted, redisErr := s.redis.IsBlacklisted(ctx, tokenHash)
	if blacklisted {
		return true
	}

	// If not found in Redis or Redis error, check database as fallback
	if redisErr != nil {
		dbBlacklisted, dbErr := s.tokenRepo.IsBlacklisted(ctx, tokenHash)
		if dbErr != nil {
			s.logger.Warn("Failed to check blacklist in DB", map[string]interface{}{
				"error": dbErr.Error(),
			})
			return false
		}

		if dbBlacklisted {
			// Re-populate Redis cache if DB says blacklisted
			ttl := s.jwtService.GetAccessTokenExpiration()
			if err := s.redis.AddToBlacklist(ctx, tokenHash, ttl); err != nil {
				s.logger.Warn("Failed to restore blacklist entry to Redis", map[string]interface{}{
					"error": err.Error(),
				})
			}
			return true
		}
	}

	return false
}

// AddToBlacklist adds a token to both Redis and PostgreSQL blacklists.
// Returns error only if both storages fail.
func (s *BlacklistService) AddToBlacklist(ctx context.Context, tokenHash string, userID *uuid.UUID, ttl time.Duration) error {
	var lastErr error

	// Skip if TTL is zero or negative
	if ttl <= 0 {
		return nil
	}

	// Add to Redis (primary, fast check)
	if err := s.redis.AddToBlacklist(ctx, tokenHash, ttl); err != nil {
		s.logger.Error("Failed to add token to Redis blacklist", map[string]interface{}{
			"error": err.Error(),
		})
		lastErr = err
	}

	// Add to PostgreSQL (persistent, survives Redis restart)
	blacklistEntry := &models.TokenBlacklist{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: time.Now().Add(ttl),
	}

	if err := s.tokenRepo.AddToBlacklist(ctx, blacklistEntry); err != nil {
		s.logger.Error("Failed to add token to DB blacklist", map[string]interface{}{
			"error": err.Error(),
		})
		// If Redis succeeded, don't return error - Redis is primary
		if lastErr != nil {
			return lastErr // Both failed
		}
	}

	s.logger.Debug("Token added to blacklist", map[string]interface{}{
		"ttl": ttl.String(),
	})

	return nil
}

// AddAccessToken adds an access token hash to blacklist with access token TTL
func (s *BlacklistService) AddAccessToken(ctx context.Context, tokenHash string, userID *uuid.UUID) error {
	return s.AddToBlacklist(ctx, tokenHash, userID, s.jwtService.GetAccessTokenExpiration())
}

// AddRefreshToken adds a refresh token hash to blacklist with refresh token TTL
func (s *BlacklistService) AddRefreshToken(ctx context.Context, tokenHash string, userID *uuid.UUID) error {
	return s.AddToBlacklist(ctx, tokenHash, userID, s.jwtService.GetRefreshTokenExpiration())
}

// BlacklistSessionTokens adds both access and refresh tokens from a session to blacklist.
// This is used when revoking a session to ensure both tokens become invalid.
func (s *BlacklistService) BlacklistSessionTokens(ctx context.Context, session *models.Session) error {
	var lastErr error

	// Blacklist access token
	if session.AccessTokenHash != "" {
		if err := s.AddAccessToken(ctx, session.AccessTokenHash, &session.UserID); err != nil {
			lastErr = err
			s.logger.Error("Failed to blacklist access token", map[string]interface{}{
				"session_id": session.ID,
				"error":      err.Error(),
			})
		} else {
			s.logger.Info("Access token blacklisted", map[string]interface{}{
				"session_id": session.ID,
				"user_id":    session.UserID,
			})
		}
	}

	// Blacklist refresh token
	if session.TokenHash != "" {
		if err := s.AddRefreshToken(ctx, session.TokenHash, &session.UserID); err != nil {
			lastErr = err
			s.logger.Error("Failed to blacklist refresh token", map[string]interface{}{
				"session_id": session.ID,
				"error":      err.Error(),
			})
		} else {
			s.logger.Info("Refresh token blacklisted", map[string]interface{}{
				"session_id": session.ID,
				"user_id":    session.UserID,
			})
		}
	}

	if session.AccessTokenHash == "" && session.TokenHash == "" {
		s.logger.Warn("Session has no tokens to blacklist", map[string]interface{}{
			"session_id": session.ID,
		})
	}
	s.auditLogger.Log(AuditLogParams{
		UserID:    &session.UserID,
		Action:    models.ActionSessionRevoked,
		Status:    models.StatusSuccess,
		IP:        session.IPAddress,
		UserAgent: session.UserAgent,
		Details: map[string]interface{}{
			"session": session.ID.String(),
			"device":  session.DeviceType,
			"os":      session.OS,
		},
	})
	return lastErr
}

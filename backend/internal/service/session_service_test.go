package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// setupSessionService creates a SessionService with mocks for testing
func setupSessionService() (*SessionService, *mockSessionStore, *mockTokenStore, *mockCacheService, *mockTokenService, *BlacklistService) {
	mockSession := &mockSessionStore{}
	mockToken := &mockTokenStore{}
	mockCache := &mockCacheService{}
	mAudit := &mockAuditLogger{}
	mockJWT := &mockTokenService{
		GetAccessTokenExpirationFunc:  func() time.Duration { return 15 * time.Minute },
		GetRefreshTokenExpirationFunc: func() time.Duration { return 7 * 24 * time.Hour },
	}
	log := logger.New("session-test", logger.DebugLevel, false)

	// Create BlacklistService with mocks
	blacklistSvc := NewBlacklistService(mockCache, mockToken, mockSession, mockJWT, log, mAudit)

	svc := NewSessionService(mockSession, blacklistSvc, log)
	return svc, mockSession, mockToken, mockCache, mockJWT, blacklistSvc
}

func TestSessionService_CreateSessionWithParams(t *testing.T) {
	svc, mockStore, _, _, _, _ := setupSessionService()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		tokenHash := "hash"
		accessTokenHash := "access_hash"
		ip := "127.0.0.1"
		ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"
		expiresAt := time.Now().Add(time.Hour)

		mockStore.CreateSessionFunc = func(ctx context.Context, session *models.Session) error {
			assert.Equal(t, userID, session.UserID)
			assert.Equal(t, tokenHash, session.TokenHash)
			assert.Equal(t, accessTokenHash, session.AccessTokenHash)
			assert.Equal(t, ip, session.IPAddress)
			assert.Equal(t, "macOS 10.15.7", session.OS)
			assert.Equal(t, "Chrome 91.0.4472.114", session.Browser)
			return nil
		}

		session, err := svc.CreateSessionWithParams(ctx, SessionCreationParams{
			UserID:          userID,
			TokenHash:       tokenHash,
			AccessTokenHash: accessTokenHash,
			IPAddress:       ip,
			UserAgent:       ua,
			ExpiresAt:       expiresAt,
		})
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, userID, session.UserID)
	})

	t.Run("ValidationError_MissingUserID", func(t *testing.T) {
		_, err := svc.CreateSessionWithParams(ctx, SessionCreationParams{
			TokenHash: "hash",
			ExpiresAt: time.Now().Add(time.Hour),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user_id is required")
	})

	t.Run("ValidationError_MissingTokenHash", func(t *testing.T) {
		_, err := svc.CreateSessionWithParams(ctx, SessionCreationParams{
			UserID:    uuid.New(),
			ExpiresAt: time.Now().Add(time.Hour),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token_hash is required")
	})
}

func TestSessionService_GetUserSessions(t *testing.T) {
	svc, mockStore, _, _, _, _ := setupSessionService()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		mockStore.GetUserSessionsPaginatedFunc = func(ctx context.Context, uid uuid.UUID, page, perPage int) ([]models.Session, int, error) {
			return []models.Session{
				{ID: uuid.New(), UserID: userID, DeviceType: "Desktop"},
			}, 1, nil
		}

		resp, err := svc.GetUserSessions(ctx, userID, 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, resp.Total)
		assert.Equal(t, 1, len(resp.Sessions))
		assert.Equal(t, "Desktop", resp.Sessions[0].DeviceType)
	})
}

func TestSessionService_RevokeSession(t *testing.T) {
	svc, mockStore, mockToken, mockCache, _, _ := setupSessionService()
	ctx := context.Background()

	t.Run("Success_with_blacklist", func(t *testing.T) {
		userID := uuid.New()
		sessionID := uuid.New()
		tokenHash := "test_token_hash"
		expiresAt := time.Now().Add(time.Hour)

		// Mock GetSessionByID
		mockStore.GetSessionByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Session, error) {
			return &models.Session{
				ID:        sessionID,
				UserID:    userID,
				TokenHash: tokenHash,
				ExpiresAt: expiresAt,
			}, nil
		}

		// Mock Redis blacklist
		mockCache.AddToBlacklistFunc = func(ctx context.Context, hash string, exp time.Duration) error {
			assert.Equal(t, tokenHash, hash)
			return nil
		}

		// Mock DB blacklist
		mockToken.AddToBlacklistFunc = func(ctx context.Context, token *models.TokenBlacklist) error {
			assert.Equal(t, tokenHash, token.TokenHash)
			return nil
		}

		// Mock RevokeUserSession
		mockStore.RevokeUserSessionFunc = func(ctx context.Context, uid, sid uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, sessionID, sid)
			return nil
		}

		err := svc.RevokeSession(ctx, userID, sessionID)
		assert.NoError(t, err)
	})

	t.Run("Fail_session_not_owned", func(t *testing.T) {
		userID := uuid.New()
		otherUserID := uuid.New()
		sessionID := uuid.New()

		mockStore.GetSessionByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Session, error) {
			return &models.Session{
				ID:     sessionID,
				UserID: otherUserID, // Different user
			}, nil
		}

		err := svc.RevokeSession(ctx, userID, sessionID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not belong to user")
	})
}

func TestSessionService_AdminRevokeSession(t *testing.T) {
	svc, mockStore, mockToken, mockCache, _, _ := setupSessionService()
	ctx := context.Background()

	t.Run("Success_with_blacklist", func(t *testing.T) {
		sessionID := uuid.New()
		userID := uuid.New()
		tokenHash := "admin_test_token"
		expiresAt := time.Now().Add(time.Hour)

		mockStore.GetSessionByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.Session, error) {
			return &models.Session{
				ID:        sessionID,
				UserID:    userID,
				TokenHash: tokenHash,
				ExpiresAt: expiresAt,
			}, nil
		}

		mockCache.AddToBlacklistFunc = func(ctx context.Context, hash string, exp time.Duration) error {
			assert.Equal(t, tokenHash, hash)
			return nil
		}

		mockToken.AddToBlacklistFunc = func(ctx context.Context, token *models.TokenBlacklist) error {
			assert.Equal(t, tokenHash, token.TokenHash)
			return nil
		}

		mockStore.RevokeSessionFunc = func(ctx context.Context, id uuid.UUID) error {
			assert.Equal(t, sessionID, id)
			return nil
		}

		err := svc.AdminRevokeSession(ctx, sessionID)
		assert.NoError(t, err)
	})
}

func TestSessionService_RevokeAllUserSessions(t *testing.T) {
	svc, mockStore, _, _, _, _ := setupSessionService()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		mockStore.RevokeAllUserSessionsFunc = func(ctx context.Context, uid uuid.UUID, except *uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Nil(t, except)
			return nil
		}

		err := svc.RevokeAllUserSessions(ctx, userID, nil)
		assert.NoError(t, err)
	})
}

func TestSessionService_UpdateSessionName(t *testing.T) {
	svc, mockStore, _, _, _, _ := setupSessionService()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		sessionID := uuid.New()
		name := "My Session"
		mockStore.UpdateSessionNameFunc = func(ctx context.Context, sid uuid.UUID, n string) error {
			assert.Equal(t, sessionID, sid)
			assert.Equal(t, name, n)
			return nil
		}

		err := svc.UpdateSessionName(ctx, sessionID, name)
		assert.NoError(t, err)
	})
}

func TestSessionService_GetSessionStats(t *testing.T) {
	svc, mockStore, _, _, _, _ := setupSessionService()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockStore.GetSessionStatsFunc = func(ctx context.Context) (*models.SessionStats, error) {
			return &models.SessionStats{TotalActiveSessions: 100}, nil
		}

		stats, err := svc.GetSessionStats(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, 100, stats.TotalActiveSessions)
	})
}

func TestSessionService_CleanupExpiredSessions(t *testing.T) {
	svc, mockStore, _, _, _, _ := setupSessionService()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockStore.DeleteExpiredSessionsFunc = func(ctx context.Context, olderThan time.Duration) error {
			assert.Equal(t, 7*24*time.Hour, olderThan)
			return nil
		}

		err := svc.CleanupExpiredSessions(ctx)
		assert.NoError(t, err)
	})
}

func TestSessionService_RefreshSession(t *testing.T) {
	svc, mockStore, _, _, _, _ := setupSessionService()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		oldRefreshHash := "old_refresh_hash"
		newRefreshHash := "new_refresh_hash"
		newAccessHash := "new_access_hash"
		newExpiresAt := time.Now().Add(7 * 24 * time.Hour)

		mockStore.RefreshSessionTokensFunc = func(ctx context.Context, oldHash, newHash, newAccess string, exp time.Time) error {
			assert.Equal(t, oldRefreshHash, oldHash)
			assert.Equal(t, newRefreshHash, newHash)
			assert.Equal(t, newAccessHash, newAccess)
			return nil
		}

		err := svc.RefreshSession(ctx, SessionRefreshParams{
			OldRefreshTokenHash: oldRefreshHash,
			NewRefreshTokenHash: newRefreshHash,
			NewAccessTokenHash:  newAccessHash,
			NewExpiresAt:        newExpiresAt,
		})
		assert.NoError(t, err)
	})

	t.Run("ValidationError_MissingOldHash", func(t *testing.T) {
		err := svc.RefreshSession(ctx, SessionRefreshParams{
			NewRefreshTokenHash: "new_hash",
			NewExpiresAt:        time.Now().Add(time.Hour),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "old_refresh_token_hash is required")
	})

	t.Run("ValidationError_MissingNewHash", func(t *testing.T) {
		err := svc.RefreshSession(ctx, SessionRefreshParams{
			OldRefreshTokenHash: "old_hash",
			NewExpiresAt:        time.Now().Add(time.Hour),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "new_refresh_token_hash is required")
	})

	t.Run("ValidationError_MissingExpiration", func(t *testing.T) {
		err := svc.RefreshSession(ctx, SessionRefreshParams{
			OldRefreshTokenHash: "old_hash",
			NewRefreshTokenHash: "new_hash",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "new_expires_at is required")
	})
}

func TestSessionService_RefreshSessionNonFatal(t *testing.T) {
	svc, mockStore, _, _, _, _ := setupSessionService()
	ctx := context.Background()

	t.Run("Success_ReturnsTrue", func(t *testing.T) {
		mockStore.RefreshSessionTokensFunc = func(ctx context.Context, oldHash, newHash, newAccess string, exp time.Time) error {
			return nil
		}

		result := svc.RefreshSessionNonFatal(ctx, SessionRefreshParams{
			OldRefreshTokenHash: "old_hash",
			NewRefreshTokenHash: "new_hash",
			NewAccessTokenHash:  "new_access",
			NewExpiresAt:        time.Now().Add(time.Hour),
		})
		assert.True(t, result)
	})

	t.Run("Failure_ReturnsFalse", func(t *testing.T) {
		// Validation will fail because of missing required field
		result := svc.RefreshSessionNonFatal(ctx, SessionRefreshParams{
			OldRefreshTokenHash: "old_hash",
			// Missing NewRefreshTokenHash
			NewExpiresAt: time.Now().Add(time.Hour),
		})
		assert.False(t, result)
	})
}

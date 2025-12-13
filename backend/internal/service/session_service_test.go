package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSessionService_CreateSession(t *testing.T) {
	mockStore := &mockSessionStore{}
	svc := NewSessionService(mockStore)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		tokenHash := "hash"
		ip := "127.0.0.1"
		ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"
		expiresAt := time.Now().Add(time.Hour)

		mockStore.CreateSessionFunc = func(ctx context.Context, session *models.Session) error {
			assert.Equal(t, userID, session.UserID)
			assert.Equal(t, tokenHash, session.TokenHash)
			assert.Equal(t, ip, session.IPAddress)
			assert.Equal(t, "macOS 10.15.7", session.OS)
			assert.Equal(t, "Chrome 91.0.4472.114", session.Browser)
			return nil
		}

		session, err := svc.CreateSession(ctx, userID, tokenHash, ip, ua, expiresAt)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, userID, session.UserID)
	})
}

func TestSessionService_GetUserSessions(t *testing.T) {
	mockStore := &mockSessionStore{}
	svc := NewSessionService(mockStore)
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
	mockStore := &mockSessionStore{}
	svc := NewSessionService(mockStore)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		sessionID := uuid.New()
		mockStore.RevokeUserSessionFunc = func(ctx context.Context, uid, sid uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, sessionID, sid)
			return nil
		}

		err := svc.RevokeSession(ctx, userID, sessionID)
		assert.NoError(t, err)
	})
}

func TestSessionService_RevokeAllUserSessions(t *testing.T) {
	mockStore := &mockSessionStore{}
	svc := NewSessionService(mockStore)
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
	mockStore := &mockSessionStore{}
	svc := NewSessionService(mockStore)
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
	mockStore := &mockSessionStore{}
	svc := NewSessionService(mockStore)
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
	mockStore := &mockSessionStore{}
	svc := NewSessionService(mockStore)
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

package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRefreshToken_IsExpired(t *testing.T) {
	t.Run("Expired", func(t *testing.T) {
		rt := &RefreshToken{
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		assert.True(t, rt.IsExpired())
	})

	t.Run("Not Expired", func(t *testing.T) {
		rt := &RefreshToken{
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		assert.False(t, rt.IsExpired())
	})
}

func TestRefreshToken_IsRevoked(t *testing.T) {
	t.Run("Revoked", func(t *testing.T) {
		now := time.Now()
		rt := &RefreshToken{
			RevokedAt: &now,
		}
		assert.True(t, rt.IsRevoked())
	})

	t.Run("Not Revoked", func(t *testing.T) {
		rt := &RefreshToken{
			RevokedAt: nil,
		}
		assert.False(t, rt.IsRevoked())
	})
}

package models

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_BeforeInsert(t *testing.T) {
	t.Run("Sets CreatedAt and UpdatedAt if zero", func(t *testing.T) {
		u := &User{}
		err := u.BeforeInsert(context.Background())
		assert.NoError(t, err)
		assert.False(t, u.CreatedAt.IsZero())
		assert.False(t, u.UpdatedAt.IsZero())
		assert.WithinDuration(t, time.Now(), u.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), u.UpdatedAt, time.Second)
	})

	t.Run("Keeps existing CreatedAt", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour)
		u := &User{
			CreatedAt: past,
		}
		err := u.BeforeInsert(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, past, u.CreatedAt)
		assert.False(t, u.UpdatedAt.IsZero())
	})
}

func TestUser_BeforeUpdate(t *testing.T) {
	t.Run("Updates UpdatedAt", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour)
		u := &User{
			UpdatedAt: past,
		}
		err := u.BeforeUpdate(context.Background())
		assert.NoError(t, err)
		assert.True(t, u.UpdatedAt.After(past))
		assert.WithinDuration(t, time.Now(), u.UpdatedAt, time.Second)
	})
}

func TestIsValidAccountType(t *testing.T) {
	tests := []struct {
		accountType string
		expected    bool
	}{
		{"human", true},
		{"service", true},
		{"invalid", false},
		{"HUMAN", false}, // Case sensitive check
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.accountType, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsValidAccountType(tt.accountType))
		})
	}
}

func TestUser_PublicUser(t *testing.T) {
	secret := "secret_totp"
	phone := "1234567890"
	u := &User{
		ID:                uuid.New(),
		Email:             "test@example.com",
		Phone:             &phone,
		Username:          "testuser",
		PasswordHash:      "hashed_password",
		FullName:          "Test User",
		ProfilePictureURL: "http://example.com/pic.jpg",
		AccountType:       "human",
		EmailVerified:     true,
		PhoneVerified:     true,
		IsActive:          true,
		TOTPSecret:        &secret,
		TOTPEnabled:       true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	public := u.PublicUser()

	assert.Equal(t, u.ID, public.ID)
	assert.Equal(t, u.Email, public.Email)
	assert.Equal(t, u.Username, public.Username)
	assert.Equal(t, u.FullName, public.FullName)
	assert.Equal(t, u.TOTPEnabled, public.TOTPEnabled)

	// Sensitive fields cleared
	assert.Empty(t, public.PasswordHash)
	assert.Nil(t, public.TOTPSecret)

	// Check new instance
	assert.NotSame(t, u, public)
}

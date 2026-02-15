package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := "01234567890123456789012345678901"
	plaintext := "my-secret-ldap-password"

	encrypted, err := EncryptAESGCM(plaintext, key)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := DecryptAESGCM(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecrypt_DifferentCiphertexts(t *testing.T) {
	key := "01234567890123456789012345678901"
	plaintext := "same-password-twice"

	encrypted1, err := EncryptAESGCM(plaintext, key)
	require.NoError(t, err)

	encrypted2, err := EncryptAESGCM(plaintext, key)
	require.NoError(t, err)

	assert.NotEqual(t, encrypted1, encrypted2)

	decrypted1, err := DecryptAESGCM(encrypted1, key)
	require.NoError(t, err)

	decrypted2, err := DecryptAESGCM(encrypted2, key)
	require.NoError(t, err)

	assert.Equal(t, decrypted1, decrypted2)
}

func TestDecrypt_WrongKey_Fails(t *testing.T) {
	key := "01234567890123456789012345678901"
	wrongKey := "10987654321098765432109876543210"
	plaintext := "secret-password"

	encrypted, err := EncryptAESGCM(plaintext, key)
	require.NoError(t, err)

	_, err = DecryptAESGCM(encrypted, wrongKey)
	assert.Error(t, err)
}

func TestEncrypt_InvalidKeyLength(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "too short", key: "short"},
		{name: "too long", key: "01234567890123456789012345678901extra"},
		{name: "empty", key: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := EncryptAESGCM("plaintext", tc.key)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "encryption key must be exactly 32 bytes")

			_, err = DecryptAESGCM("dGVzdA==", tc.key)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "encryption key must be exactly 32 bytes")
		})
	}
}

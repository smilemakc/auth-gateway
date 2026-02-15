package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashToken(t *testing.T) {
	token := "test-token"
	expectedHash := sha256.Sum256([]byte(token))
	expected := hex.EncodeToString(expectedHash[:])

	result := HashToken(token)

	assert.Equal(t, expected, result)
}

func TestCompareHashConstantTime(t *testing.T) {
	hash1 := HashToken("secret-key-1")
	hash2 := HashToken("secret-key-1")
	hash3 := HashToken("secret-key-2")

	assert.True(t, CompareHashConstantTime(hash1, hash2), "same hashes should match")
	assert.False(t, CompareHashConstantTime(hash1, hash3), "different hashes should not match")
	assert.False(t, CompareHashConstantTime(hash1, ""), "empty string should not match")
	assert.True(t, CompareHashConstantTime("", ""), "two empty strings should match")
}

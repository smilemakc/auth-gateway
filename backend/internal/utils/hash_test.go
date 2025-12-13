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

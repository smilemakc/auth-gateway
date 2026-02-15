package utils

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// HashToken hashes a token using SHA-256
// Used for storing refresh tokens and blacklisted tokens
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CompareHashConstantTime compares two hash strings using constant-time comparison
// to prevent timing side-channel attacks.
func CompareHashConstantTime(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

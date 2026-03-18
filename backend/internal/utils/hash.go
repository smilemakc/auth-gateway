package utils

import (
	"crypto/hmac"
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

// HMACHash computes HMAC-SHA256 of data with the given secret key.
// Used for OTP verification where bcrypt is too slow for brute-force resistance
// on short codes (6 digits = 10^6 possibilities).
func HMACHash(data, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// HMACVerify compares an HMAC hash with a computed HMAC of the data.
// Uses constant-time comparison to prevent timing attacks.
func HMACVerify(data, secret, expectedHash string) bool {
	computed := HMACHash(data, secret)
	return hmac.Equal([]byte(computed), []byte(expectedHash))
}

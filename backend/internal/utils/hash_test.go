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

func TestHMACHash(t *testing.T) {
	secret := "test-secret-key"
	data := "123456"

	hash1 := HMACHash(data, secret)
	hash2 := HMACHash(data, secret)

	assert.Equal(t, hash1, hash2, "same data and secret should produce same hash")
	assert.NotEmpty(t, hash1, "hash should not be empty")
	assert.Len(t, hash1, 64, "HMAC-SHA256 hex string should be 64 characters")
}

func TestHMACHashDifferentInputs(t *testing.T) {
	secret := "test-secret-key"

	hash1 := HMACHash("123456", secret)
	hash2 := HMACHash("654321", secret)
	hash3 := HMACHash("123456", "different-secret")

	assert.NotEqual(t, hash1, hash2, "different data should produce different hashes")
	assert.NotEqual(t, hash1, hash3, "different secrets should produce different hashes")
}

func TestHMACVerify(t *testing.T) {
	secret := "test-secret-key"
	data := "123456"

	hash := HMACHash(data, secret)

	assert.True(t, HMACVerify(data, secret, hash), "valid data and secret should verify")
	assert.False(t, HMACVerify("wrong-data", secret, hash), "wrong data should not verify")
	assert.False(t, HMACVerify(data, "wrong-secret", hash), "wrong secret should not verify")
	assert.False(t, HMACVerify(data, secret, "wrong-hash"), "wrong hash should not verify")
}

func TestHMACVerifyOTPUseCase(t *testing.T) {
	secret := "otp-hmac-secret-for-production"

	otpCode := "123456"
	storedHash := HMACHash(otpCode, secret)

	assert.True(t, HMACVerify(otpCode, secret, storedHash), "correct OTP should verify")
	assert.False(t, HMACVerify("123457", secret, storedHash), "incorrect OTP should not verify")
	assert.False(t, HMACVerify("999999", secret, storedHash), "brute-force attempt should fail")
}

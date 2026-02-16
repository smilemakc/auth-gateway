package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
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

func TestHashToken_ShouldReturnConsistentHash_WhenCalledMultipleTimes(t *testing.T) {
	token := "consistent-token"
	hash1 := HashToken(token)
	hash2 := HashToken(token)
	assert.Equal(t, hash1, hash2)
}

func TestHashToken_ShouldReturnValidHex_WhenInputIsEmpty(t *testing.T) {
	hash := HashToken("")
	assert.Len(t, hash, 64, "SHA-256 hex string should always be 64 characters")
	assert.NotEmpty(t, hash)

	// SHA-256 of empty string is a well-known constant
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	assert.Equal(t, expected, hash)
}

func TestHashToken_ShouldProduceDifferentHashes_WhenInputsDiffer(t *testing.T) {
	hash1 := HashToken("token-a")
	hash2 := HashToken("token-b")
	assert.NotEqual(t, hash1, hash2)
}

func TestHashToken_ShouldHandleVeryLongInput(t *testing.T) {
	longToken := strings.Repeat("x", 100000)
	hash := HashToken(longToken)
	assert.Len(t, hash, 64)
	assert.NotEmpty(t, hash)
}

func TestHashToken_ShouldHandleUnicodeInput(t *testing.T) {
	hash := HashToken("токен-с-юникодом-🔑")
	assert.Len(t, hash, 64)
	assert.NotEmpty(t, hash)
}

func TestHashToken_ShouldHandleNullBytes(t *testing.T) {
	hash := HashToken("token\x00with\x00nulls")
	assert.Len(t, hash, 64)

	// Should differ from the version without null bytes
	hashWithout := HashToken("tokenwithnulls")
	assert.NotEqual(t, hash, hashWithout)
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

func TestCompareHashConstantTime_ShouldBeFalse_WhenHashesDifferByOneChar(t *testing.T) {
	hash := HashToken("test")
	// Flip last character
	modified := hash[:63] + "0"
	if hash[63] == '0' {
		modified = hash[:63] + "1"
	}
	assert.False(t, CompareHashConstantTime(hash, modified))
}

func TestCompareHashConstantTime_ShouldBeFalse_WhenLengthsDiffer(t *testing.T) {
	assert.False(t, CompareHashConstantTime("abc", "abcd"))
	assert.False(t, CompareHashConstantTime("abcd", "abc"))
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

func TestHMACHash_ShouldReturnValidHash_WhenInputIsEmpty(t *testing.T) {
	hash := HMACHash("", "secret")
	assert.Len(t, hash, 64)
	assert.NotEmpty(t, hash)
}

func TestHMACHash_ShouldReturnValidHash_WhenSecretIsEmpty(t *testing.T) {
	hash := HMACHash("data", "")
	assert.Len(t, hash, 64)
	assert.NotEmpty(t, hash)
}

func TestHMACHash_ShouldReturnValidHash_WhenBothEmpty(t *testing.T) {
	hash := HMACHash("", "")
	assert.Len(t, hash, 64)
	assert.NotEmpty(t, hash)
}

func TestHMACHash_ShouldHandleVeryLongInput(t *testing.T) {
	longData := strings.Repeat("a", 1000000)
	hash := HMACHash(longData, "secret")
	assert.Len(t, hash, 64)
}

func TestHMACHash_ShouldHandleVeryLongSecret(t *testing.T) {
	longSecret := strings.Repeat("s", 1000000)
	hash := HMACHash("data", longSecret)
	assert.Len(t, hash, 64)
}

func TestHMACHash_ShouldHandleUnicodeData(t *testing.T) {
	hash := HMACHash("данные-🔑", "секрет")
	assert.Len(t, hash, 64)
}

func TestHMACHashDifferentInputs(t *testing.T) {
	secret := "test-secret-key"

	hash1 := HMACHash("123456", secret)
	hash2 := HMACHash("654321", secret)
	hash3 := HMACHash("123456", "different-secret")

	assert.NotEqual(t, hash1, hash2, "different data should produce different hashes")
	assert.NotEqual(t, hash1, hash3, "different secrets should produce different hashes")
}

func TestHMACHash_ShouldDiffer_WhenSecretsDifferByOneChar(t *testing.T) {
	hash1 := HMACHash("data", "secret1")
	hash2 := HMACHash("data", "secret2")
	assert.NotEqual(t, hash1, hash2)
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

func TestHMACVerify_ShouldFail_WhenHashIsEmpty(t *testing.T) {
	assert.False(t, HMACVerify("data", "secret", ""))
}

func TestHMACVerify_ShouldFail_WhenHashIsTruncated(t *testing.T) {
	hash := HMACHash("data", "secret")
	truncated := hash[:32] // half the expected length
	assert.False(t, HMACVerify("data", "secret", truncated))
}

func TestHMACVerify_ShouldFail_WhenHashHasExtraChars(t *testing.T) {
	hash := HMACHash("data", "secret")
	extended := hash + "extra"
	assert.False(t, HMACVerify("data", "secret", extended))
}

func TestHMACVerifyOTPUseCase(t *testing.T) {
	secret := "otp-hmac-secret-for-production"

	otpCode := "123456"
	storedHash := HMACHash(otpCode, secret)

	assert.True(t, HMACVerify(otpCode, secret, storedHash), "correct OTP should verify")
	assert.False(t, HMACVerify("123457", secret, storedHash), "incorrect OTP should not verify")
	assert.False(t, HMACVerify("999999", secret, storedHash), "brute-force attempt should fail")
}

func TestHMACVerify_ShouldVerifyAllSixDigitOTPCodes(t *testing.T) {
	secret := "otp-secret"
	// Verify a few representative OTP codes
	codes := []string{"000000", "000001", "123456", "999999", "500000"}
	for _, code := range codes {
		hash := HMACHash(code, secret)
		assert.True(t, HMACVerify(code, secret, hash), "OTP code %s should verify", code)
		// Adjacent code should not verify
		if code != "999999" {
			assert.False(t, HMACVerify(code, secret, HMACHash("999999", secret)),
				"OTP code %s should not verify against hash of 999999", code)
		}
	}
}

func TestHashToken_ShouldDifferFromHMACHash(t *testing.T) {
	// HashToken uses plain SHA-256, HMACHash uses HMAC-SHA-256
	// Even with the same input they should differ (unless secret happens to produce same result)
	data := "test-data"
	tokenHash := HashToken(data)
	hmacHash := HMACHash(data, "any-secret")
	assert.NotEqual(t, tokenHash, hmacHash, "SHA-256 and HMAC-SHA-256 should produce different outputs")
}

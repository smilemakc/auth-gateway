package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	password := "testPassword123"
	cost := 10

	hash, err := HashPassword(password, cost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not equal plain password")
	}
}

func TestHashPassword_ShouldProduceDifferentHashes_WhenSamePasswordHashedTwice(t *testing.T) {
	password := "samePasswordTwice"
	hash1, err := HashPassword(password, 4)
	require.NoError(t, err)

	hash2, err := HashPassword(password, 4)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2, "bcrypt should produce different hashes due to random salt")
}

func TestHashPassword_ShouldSucceed_WhenPasswordContainsUnicode(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"Emoji characters", "pass🔥word🎉ab"},
		{"Chinese characters", "密码abcdefgh"},
		{"Arabic characters", "كلمةالمرور1a"},
		{"Combining diacritical marks", "pàsswörd"},
		{"Mixed RTL and LTR", "abcדהוword"},
		{"Japanese hiragana", "パスワードabcd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password, 4)
			require.NoError(t, err)
			assert.NotEmpty(t, hash)

			err = CheckPassword(hash, tt.password)
			assert.NoError(t, err, "should verify unicode password correctly")
		})
	}
}

func TestHashPassword_ShouldTruncateAtBcryptLimit_WhenPasswordExceeds72Bytes(t *testing.T) {
	// bcrypt silently truncates passwords at 72 bytes
	base := strings.Repeat("a", 72)
	extended := base + "EXTRA_CHARS"

	hash, err := HashPassword(base, 4)
	require.NoError(t, err)

	// Both should match because bcrypt truncates at 72 bytes
	err = CheckPassword(hash, extended)
	assert.NoError(t, err, "bcrypt truncates at 72 bytes, so extra chars are ignored")

	err = CheckPassword(hash, base)
	assert.NoError(t, err)
}

func TestHashPassword_ShouldSucceed_WhenPasswordIsEmpty(t *testing.T) {
	hash, err := HashPassword("", 4)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	err = CheckPassword(hash, "")
	assert.NoError(t, err)

	err = CheckPassword(hash, "notempty")
	assert.Error(t, err)
}

func TestHashPassword_ShouldSucceed_WhenPasswordContainsOnlySpaces(t *testing.T) {
	password := "        " // 8 spaces
	hash, err := HashPassword(password, 4)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	err = CheckPassword(hash, password)
	assert.NoError(t, err)

	err = CheckPassword(hash, "notspaces")
	assert.Error(t, err)
}

func TestHashPassword_ShouldSucceed_WhenPasswordContainsSpecialChars(t *testing.T) {
	password := `p@$$w0rd!#%^&*()_+-=[]{}|;:'"<>,.?/~`
	hash, err := HashPassword(password, 4)
	require.NoError(t, err)

	err = CheckPassword(hash, password)
	assert.NoError(t, err)
}

func TestHashPassword_ShouldFail_WhenCostIsTooHigh(t *testing.T) {
	// bcrypt max cost is 31
	_, err := HashPassword("test", 32)
	assert.Error(t, err)
}

func TestHashPassword_ShouldSucceed_WhenCostIsMinimum(t *testing.T) {
	hash, err := HashPassword("testpassword", 4) // bcrypt minimum cost
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	err = CheckPassword(hash, "testpassword")
	assert.NoError(t, err)
}

func TestHashPassword_ShouldContainNullBytes(t *testing.T) {
	password := "pass\x00word"
	hash, err := HashPassword(password, 4)
	require.NoError(t, err)

	err = CheckPassword(hash, password)
	assert.NoError(t, err)
}

func TestCheckPassword(t *testing.T) {
	password := "testPassword123"
	cost := 10

	hash, err := HashPassword(password, cost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test correct password
	err = CheckPassword(hash, password)
	if err != nil {
		t.Errorf("CheckPassword failed for correct password: %v", err)
	}

	// Test incorrect password
	err = CheckPassword(hash, "wrongPassword")
	if err == nil {
		t.Error("CheckPassword should fail for incorrect password")
	}
}

func TestCheckPassword_ShouldFail_WhenHashIsInvalid(t *testing.T) {
	err := CheckPassword("not-a-bcrypt-hash", "password")
	assert.Error(t, err)
}

func TestCheckPassword_ShouldFail_WhenHashIsEmpty(t *testing.T) {
	err := CheckPassword("", "password")
	assert.Error(t, err)
}

func TestIsPasswordValid(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"Valid password", "testPassword123", true},
		{"Valid minimum length", "abcdefgh", true},
		{"Too short", "abcdefg", false},
		{"Empty", "", false},
		{"Only digits no lowercase", "12345678", false},
		{"Mixed valid", "pass1234", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPasswordValid(tt.password)
			if got != tt.want {
				t.Errorf("IsPasswordValid(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestDefaultPasswordPolicy_ShouldReturnExpectedDefaults(t *testing.T) {
	policy := DefaultPasswordPolicy()

	assert.Equal(t, 8, policy.MinLength)
	assert.False(t, policy.RequireUppercase)
	assert.True(t, policy.RequireLowercase)
	assert.False(t, policy.RequireNumbers)
	assert.False(t, policy.RequireSpecial)
	assert.Equal(t, 0, policy.MaxLength)
}

func TestValidatePassword_ShouldEnforceMinLength(t *testing.T) {
	policy := PasswordPolicy{MinLength: 10}

	err := ValidatePassword("short", policy)
	assert.Error(t, err)
	var pvErr *PasswordValidationError
	assert.ErrorAs(t, err, &pvErr)
	assert.Equal(t, 10, pvErr.MinLength)

	err = ValidatePassword("longEnough", policy)
	assert.NoError(t, err)

	err = ValidatePassword("exactlyten", policy)
	assert.NoError(t, err)
}

func TestValidatePassword_ShouldEnforceMaxLength(t *testing.T) {
	policy := PasswordPolicy{MaxLength: 20}

	err := ValidatePassword(strings.Repeat("a", 21), policy)
	assert.Error(t, err)
	var pvErr *PasswordValidationError
	assert.ErrorAs(t, err, &pvErr)
	assert.Equal(t, 20, pvErr.MaxLength)

	err = ValidatePassword(strings.Repeat("a", 20), policy)
	assert.NoError(t, err)

	err = ValidatePassword("short", policy)
	assert.NoError(t, err)
}

func TestValidatePassword_ShouldEnforceMaxLengthZeroMeansNoMax(t *testing.T) {
	policy := PasswordPolicy{MaxLength: 0}

	err := ValidatePassword(strings.Repeat("a", 1000), policy)
	assert.NoError(t, err, "MaxLength 0 means no maximum")
}

func TestValidatePassword_ShouldRequireUppercase(t *testing.T) {
	policy := PasswordPolicy{RequireUppercase: true}

	err := ValidatePassword("alllowercase", policy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uppercase")

	err = ValidatePassword("hasUppercase", policy)
	assert.NoError(t, err)
}

func TestValidatePassword_ShouldRequireLowercase(t *testing.T) {
	policy := PasswordPolicy{RequireLowercase: true}

	err := ValidatePassword("ALLUPPERCASE123", policy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lowercase")

	err = ValidatePassword("HasLowercase", policy)
	assert.NoError(t, err)
}

func TestValidatePassword_ShouldRequireNumbers(t *testing.T) {
	policy := PasswordPolicy{RequireNumbers: true}

	err := ValidatePassword("nodigitshere", policy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "number")

	err = ValidatePassword("has1number", policy)
	assert.NoError(t, err)
}

func TestValidatePassword_ShouldRequireSpecialChars(t *testing.T) {
	policy := PasswordPolicy{RequireSpecial: true}

	err := ValidatePassword("nospecialchars123", policy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "special")

	err = ValidatePassword("has!special", policy)
	assert.NoError(t, err)

	err = ValidatePassword("has@special", policy)
	assert.NoError(t, err)

	err = ValidatePassword("has$special", policy)
	assert.NoError(t, err)
}

func TestValidatePassword_ShouldEnforceAllPolicyRulesCombined(t *testing.T) {
	policy := PasswordPolicy{
		MinLength:        8,
		MaxLength:        64,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSpecial:   true,
	}

	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{"All requirements met", "Passw0rd!", false},
		{"Missing uppercase", "passw0rd!", true},
		{"Missing lowercase", "PASSW0RD!", true},
		{"Missing number", "Password!", true},
		{"Missing special", "Passw0rdx", true},
		{"Too short", "Pw0!", true},
		{"Too long", strings.Repeat("A", 30) + strings.Repeat("a", 30) + "1!12345", true},
		{"Exactly min length", "Aa1!xxxx", false},
		{"Exactly max length", strings.Repeat("A", 15) + strings.Repeat("a", 15) + strings.Repeat("1", 15) + strings.Repeat("!", 15) + "Aa1!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, policy)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword_ShouldAcceptUnicodeUpperLower(t *testing.T) {
	// unicode.IsUpper/IsLower recognize international characters
	policy := PasswordPolicy{
		RequireUppercase: true,
		RequireLowercase: true,
	}

	// German uppercase/lowercase
	err := ValidatePassword("Übergrößen", policy)
	assert.NoError(t, err, "unicode upper/lower should satisfy policy")
}

func TestValidatePassword_ShouldCountByteLength_NotRuneLength(t *testing.T) {
	// ValidatePassword uses len(password) which counts bytes, not runes
	// 3 emoji = 12 bytes (4 bytes each), but only 3 runes
	policy := PasswordPolicy{MinLength: 8}

	threeEmoji := "🔥🎉🌍"
	assert.True(t, len(threeEmoji) >= 8, "3 emoji should be >= 8 bytes")
	err := ValidatePassword(threeEmoji, policy)
	assert.NoError(t, err, "byte length should pass MinLength check")
}

func TestPasswordValidationError_ShouldFormatMessage_WhenMinLengthSet(t *testing.T) {
	err := &PasswordValidationError{
		Message:   "Password must be at least %d characters long",
		MinLength: 8,
	}
	// The Error() method replaces %d with the character corresponding to the integer value
	result := err.Error()
	assert.NotEmpty(t, result)
	assert.NotEqual(t, "Password must be at least %d characters long", result)
}

func TestPasswordValidationError_ShouldFormatMessage_WhenMaxLengthSet(t *testing.T) {
	err := &PasswordValidationError{
		Message:   "Password must be at most %d characters long",
		MaxLength: 64,
	}
	result := err.Error()
	assert.NotEmpty(t, result)
	assert.NotEqual(t, "Password must be at most %d characters long", result)
}

func TestPasswordValidationError_ShouldReturnRawMessage_WhenNoLengthSet(t *testing.T) {
	err := &PasswordValidationError{
		Message: "Password must contain at least one uppercase letter",
	}
	assert.Equal(t, "Password must contain at least one uppercase letter", err.Error())
}

func TestGetDummyPasswordHash_ShouldReturnNonEmptyHash(t *testing.T) {
	hash := GetDummyPasswordHash()
	assert.NotEmpty(t, hash)
	// Should be a valid bcrypt hash (starts with $2a$ or $2b$)
	assert.True(t, strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$"),
		"dummy hash should be a valid bcrypt hash")
}

func TestGetDummyPasswordHash_ShouldReturnSameValueOnMultipleCalls(t *testing.T) {
	hash1 := GetDummyPasswordHash()
	hash2 := GetDummyPasswordHash()
	assert.Equal(t, hash1, hash2, "dummy hash should be computed once and reused")
}

func TestGetDummyPasswordHash_ShouldNotMatchArbitraryPasswords(t *testing.T) {
	hash := GetDummyPasswordHash()
	err := CheckPassword(hash, "password")
	assert.Error(t, err, "dummy hash should not match common passwords")

	err = CheckPassword(hash, "")
	assert.Error(t, err, "dummy hash should not match empty string")
}

func TestCommonPasswords_ShouldContainKnownWeakPasswords(t *testing.T) {
	assert.Contains(t, CommonPasswords, "password")
	assert.Contains(t, CommonPasswords, "12345678")
	assert.Contains(t, CommonPasswords, "qwerty")
	assert.Greater(t, len(CommonPasswords), 10, "should contain a reasonable number of common passwords")
}

func TestValidatePassword_ShouldRejectEmptyPassword_WhenMinLengthPositive(t *testing.T) {
	policy := PasswordPolicy{MinLength: 1}
	err := ValidatePassword("", policy)
	assert.Error(t, err)
}

func TestValidatePassword_ShouldAcceptEmptyPassword_WhenMinLengthZero(t *testing.T) {
	policy := PasswordPolicy{MinLength: 0}
	err := ValidatePassword("", policy)
	assert.NoError(t, err)
}

func TestValidatePassword_ShouldCheckMinLengthBeforeOtherRules(t *testing.T) {
	// When password is too short, the error should mention length, not missing characters
	policy := PasswordPolicy{
		MinLength:        20,
		RequireUppercase: true,
		RequireNumbers:   true,
		RequireSpecial:   true,
	}

	err := ValidatePassword("abc", policy)
	require.Error(t, err)
	var pvErr *PasswordValidationError
	assert.ErrorAs(t, err, &pvErr)
	assert.Equal(t, 20, pvErr.MinLength, "min length error should be returned first")
}

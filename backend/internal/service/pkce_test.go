package service

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCodeVerifier_ShouldReturnValidVerifier_WhenCalled(t *testing.T) {
	// Act
	verifier, err := GenerateCodeVerifier()

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, verifier)
	assert.GreaterOrEqual(t, len(verifier), MinCodeVerifierLength)
	assert.LessOrEqual(t, len(verifier), MaxCodeVerifierLength)
	assert.True(t, IsValidCodeVerifier(verifier))
}

func TestGenerateCodeVerifier_ShouldReturnDifferentValues_WhenCalledMultipleTimes(t *testing.T) {
	// Act
	verifier1, err1 := GenerateCodeVerifier()
	verifier2, err2 := GenerateCodeVerifier()

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, verifier1, verifier2, "Two generated verifiers should be different")
}

func TestGenerateCodeVerifier_ShouldOnlyContainUnreservedCharacters(t *testing.T) {
	// Arrange & Act
	verifier, err := GenerateCodeVerifier()

	// Assert
	require.NoError(t, err)
	for _, c := range verifier {
		assert.True(t, isUnreservedChar(c), "Character %c should be unreserved", c)
	}
}

func TestGenerateCodeChallenge_ShouldReturnS256Hash_WhenMethodIsS256(t *testing.T) {
	// Arrange
	verifier := strings.Repeat("a", 43) // Valid verifier

	// Act
	challenge, err := GenerateCodeChallenge(verifier, CodeChallengeMethodS256)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, challenge)

	// Verify the challenge is correct SHA256 hash
	hash := sha256.Sum256([]byte(verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])
	assert.Equal(t, expectedChallenge, challenge)
}

func TestGenerateCodeChallenge_ShouldReturnVerifier_WhenMethodIsPlain(t *testing.T) {
	// Arrange
	verifier := strings.Repeat("b", 50)

	// Act
	challenge, err := GenerateCodeChallenge(verifier, CodeChallengeMethodPlain)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, verifier, challenge, "Plain challenge should be identical to verifier")
}

func TestGenerateCodeChallenge_ShouldReturnError_WhenVerifierTooShort(t *testing.T) {
	// Arrange
	shortVerifier := strings.Repeat("a", MinCodeVerifierLength-1)

	// Act
	challenge, err := GenerateCodeChallenge(shortVerifier, CodeChallengeMethodS256)

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCodeVerifier)
	assert.Empty(t, challenge)
}

func TestGenerateCodeChallenge_ShouldReturnError_WhenVerifierTooLong(t *testing.T) {
	// Arrange
	longVerifier := strings.Repeat("a", MaxCodeVerifierLength+1)

	// Act
	challenge, err := GenerateCodeChallenge(longVerifier, CodeChallengeMethodS256)

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCodeVerifier)
	assert.Empty(t, challenge)
}

func TestGenerateCodeChallenge_ShouldReturnError_WhenMethodIsInvalid(t *testing.T) {
	// Arrange
	verifier := strings.Repeat("a", 43)

	// Act
	challenge, err := GenerateCodeChallenge(verifier, "invalid_method")

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCodeChallengeMethod)
	assert.Empty(t, challenge)
}

func TestGenerateCodeChallenge_ShouldReturnError_WhenVerifierContainsInvalidChars(t *testing.T) {
	// Arrange
	invalidVerifier := strings.Repeat("a", 40) + "!@#" // Contains invalid characters

	// Act
	challenge, err := GenerateCodeChallenge(invalidVerifier, CodeChallengeMethodS256)

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCodeVerifier)
	assert.Empty(t, challenge)
}

func TestValidateCodeChallenge_ShouldSucceed_WhenS256ChallengeIsValid(t *testing.T) {
	// Arrange
	verifier := strings.Repeat("x", 50)
	challenge, _ := GenerateCodeChallenge(verifier, CodeChallengeMethodS256)

	// Act
	err := ValidateCodeChallenge(verifier, challenge, CodeChallengeMethodS256)

	// Assert
	assert.NoError(t, err)
}

func TestValidateCodeChallenge_ShouldSucceed_WhenPlainChallengeIsValid(t *testing.T) {
	// Arrange
	verifier := strings.Repeat("y", 60)
	challenge := verifier // Plain method means challenge equals verifier

	// Act
	err := ValidateCodeChallenge(verifier, challenge, CodeChallengeMethodPlain)

	// Assert
	assert.NoError(t, err)
}

func TestValidateCodeChallenge_ShouldFail_WhenChallengeMismatch(t *testing.T) {
	// Arrange
	verifier := strings.Repeat("z", 43)
	wrongChallenge := "wrong_challenge_value_that_does_not_match"

	// Act
	err := ValidateCodeChallenge(verifier, wrongChallenge, CodeChallengeMethodS256)

	// Assert
	assert.ErrorIs(t, err, ErrCodeChallengeMismatch)
}

func TestValidateCodeChallenge_ShouldFail_WhenVerifierIsInvalid(t *testing.T) {
	// Arrange
	invalidVerifier := "too_short"
	challenge := "some_challenge"

	// Act
	err := ValidateCodeChallenge(invalidVerifier, challenge, CodeChallengeMethodS256)

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCodeVerifier)
}

func TestValidateCodeChallenge_ShouldFail_WhenChallengeIsEmpty(t *testing.T) {
	// Arrange
	verifier := strings.Repeat("a", 43)

	// Act
	err := ValidateCodeChallenge(verifier, "", CodeChallengeMethodS256)

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCodeChallenge)
}

func TestValidateCodeChallenge_ShouldFail_WhenMethodIsInvalid(t *testing.T) {
	// Arrange
	verifier := strings.Repeat("a", 43)
	challenge := verifier

	// Act
	err := ValidateCodeChallenge(verifier, challenge, "invalid_method")

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCodeChallengeMethod)
}

func TestIsValidCodeVerifier_ShouldReturnTrue_WhenVerifierIsValid(t *testing.T) {
	testCases := []struct {
		name     string
		verifier string
		expected bool
	}{
		{
			name:     "minimum length verifier",
			verifier: strings.Repeat("a", MinCodeVerifierLength),
			expected: true,
		},
		{
			name:     "maximum length verifier",
			verifier: strings.Repeat("a", MaxCodeVerifierLength),
			expected: true,
		},
		{
			name:     "mixed unreserved characters",
			verifier: "abcdefghijklmnopqrstuvwxyzABCDEF0123456789-._~",
			expected: true,
		},
		{
			name:     "generated verifier",
			verifier: func() string { v, _ := GenerateCodeVerifier(); return v }(),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidCodeVerifier(tc.verifier)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsValidCodeVerifier_ShouldReturnFalse_WhenVerifierIsInvalid(t *testing.T) {
	testCases := []struct {
		name     string
		verifier string
	}{
		{
			name:     "empty verifier",
			verifier: "",
		},
		{
			name:     "too short verifier",
			verifier: strings.Repeat("a", MinCodeVerifierLength-1),
		},
		{
			name:     "too long verifier",
			verifier: strings.Repeat("a", MaxCodeVerifierLength+1),
		},
		{
			name:     "contains space",
			verifier: strings.Repeat("a", 42) + " ",
		},
		{
			name:     "contains special character",
			verifier: strings.Repeat("a", 42) + "!",
		},
		{
			name:     "contains slash",
			verifier: strings.Repeat("a", 42) + "/",
		},
		{
			name:     "contains plus",
			verifier: strings.Repeat("a", 42) + "+",
		},
		{
			name:     "contains equals",
			verifier: strings.Repeat("a", 42) + "=",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidCodeVerifier(tc.verifier)
			assert.False(t, result)
		})
	}
}

func TestIsValidCodeChallengeMethod_ShouldReturnTrue_WhenMethodIsValid(t *testing.T) {
	testCases := []struct {
		name   string
		method string
	}{
		{name: "S256 method", method: CodeChallengeMethodS256},
		{name: "plain method", method: CodeChallengeMethodPlain},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidCodeChallengeMethod(tc.method)
			assert.True(t, result)
		})
	}
}

func TestIsValidCodeChallengeMethod_ShouldReturnFalse_WhenMethodIsInvalid(t *testing.T) {
	testCases := []struct {
		name   string
		method string
	}{
		{name: "empty method", method: ""},
		{name: "lowercase s256", method: "s256"},
		{name: "sha256", method: "sha256"},
		{name: "random string", method: "random"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidCodeChallengeMethod(tc.method)
			assert.False(t, result)
		})
	}
}

func TestNewPKCEParams_ShouldReturnValidParams_WhenCalled(t *testing.T) {
	// Act
	params, err := NewPKCEParams()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, params)
	assert.NotEmpty(t, params.CodeVerifier)
	assert.NotEmpty(t, params.CodeChallenge)
	assert.Equal(t, CodeChallengeMethodS256, params.CodeChallengeMethod)
	assert.True(t, IsValidCodeVerifier(params.CodeVerifier))

	// Verify challenge matches verifier
	err = ValidateCodeChallenge(params.CodeVerifier, params.CodeChallenge, params.CodeChallengeMethod)
	assert.NoError(t, err)
}

func TestNewPKCEParams_ShouldReturnDifferentParams_WhenCalledMultipleTimes(t *testing.T) {
	// Act
	params1, err1 := NewPKCEParams()
	params2, err2 := NewPKCEParams()

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, params1.CodeVerifier, params2.CodeVerifier)
	assert.NotEqual(t, params1.CodeChallenge, params2.CodeChallenge)
}

func TestValidatePKCERequest_ShouldSucceed_WhenPKCEIsRequired_AndParamsProvided(t *testing.T) {
	// Arrange
	challenge := "valid_challenge_string"
	method := CodeChallengeMethodS256

	// Act
	err := ValidatePKCERequest(challenge, method, true)

	// Assert
	assert.NoError(t, err)
}

func TestValidatePKCERequest_ShouldFail_WhenPKCEIsRequired_AndChallengeIsMissing(t *testing.T) {
	// Arrange
	method := CodeChallengeMethodS256

	// Act
	err := ValidatePKCERequest("", method, true)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code_challenge is required")
}

func TestValidatePKCERequest_ShouldFail_WhenPKCEIsRequired_AndMethodIsMissing(t *testing.T) {
	// Arrange
	challenge := "valid_challenge_string"

	// Act
	err := ValidatePKCERequest(challenge, "", true)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code_challenge_method is required")
}

func TestValidatePKCERequest_ShouldSucceed_WhenPKCEIsNotRequired_AndParamsNotProvided(t *testing.T) {
	// Act
	err := ValidatePKCERequest("", "", false)

	// Assert
	assert.NoError(t, err)
}

func TestValidatePKCERequest_ShouldFail_WhenChallengeProvided_WithInvalidMethod(t *testing.T) {
	// Arrange
	challenge := "valid_challenge_string"
	invalidMethod := "invalid_method"

	// Act
	err := ValidatePKCERequest(challenge, invalidMethod, false)

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCodeChallengeMethod)
}

func TestValidatePKCERequest_ShouldDefaultToPlain_WhenChallengeProvided_WithoutMethod(t *testing.T) {
	// Arrange
	challenge := "valid_challenge_string"

	// Act
	err := ValidatePKCERequest(challenge, "", false)

	// Assert
	assert.NoError(t, err) // Should succeed with plain as default
}

func TestIsUnreservedChar_ShouldReturnTrue_ForValidCharacters(t *testing.T) {
	// Test all valid unreserved characters per RFC 7636
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"

	for _, c := range validChars {
		t.Run(string(c), func(t *testing.T) {
			assert.True(t, isUnreservedChar(c), "Character %c should be unreserved", c)
		})
	}
}

func TestIsUnreservedChar_ShouldReturnFalse_ForInvalidCharacters(t *testing.T) {
	invalidChars := "!@#$%^&*()+=/\\|<>?[]{}:;'\"` "

	for _, c := range invalidChars {
		t.Run(string(c), func(t *testing.T) {
			assert.False(t, isUnreservedChar(c), "Character %c should be reserved", c)
		})
	}
}

func TestSecureCompare_ShouldReturnTrue_WhenStringsAreEqual(t *testing.T) {
	// Arrange
	str := "test_string_for_comparison"

	// Act & Assert
	assert.True(t, secureCompare(str, str))
}

func TestSecureCompare_ShouldReturnFalse_WhenStringsAreDifferent(t *testing.T) {
	testCases := []struct {
		name string
		a    string
		b    string
	}{
		{
			name: "completely different",
			a:    "string_a",
			b:    "string_b",
		},
		{
			name: "different lengths",
			a:    "short",
			b:    "longer_string",
		},
		{
			name: "one character different",
			a:    "test_stringA",
			b:    "test_stringB",
		},
		{
			name: "case difference",
			a:    "TestString",
			b:    "teststring",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.False(t, secureCompare(tc.a, tc.b))
		})
	}
}

func TestSecureCompare_ShouldReturnFalse_WhenDifferentLengths(t *testing.T) {
	// Assert
	assert.False(t, secureCompare("a", "ab"))
	assert.False(t, secureCompare("ab", "a"))
	assert.False(t, secureCompare("", "a"))
	assert.False(t, secureCompare("a", ""))
}

func TestSecureCompare_ShouldReturnTrue_ForEmptyStrings(t *testing.T) {
	assert.True(t, secureCompare("", ""))
}

// Integration test: Full PKCE flow
func TestPKCEIntegration_FullFlow(t *testing.T) {
	// Step 1: Generate PKCE parameters (client-side)
	params, err := NewPKCEParams()
	require.NoError(t, err)

	// Step 2: Validate the code challenge method is supported
	assert.True(t, IsValidCodeChallengeMethod(params.CodeChallengeMethod))

	// Step 3: Client sends code_challenge to authorization server
	// Server validates the request
	err = ValidatePKCERequest(params.CodeChallenge, params.CodeChallengeMethod, true)
	require.NoError(t, err)

	// Step 4: Later, client sends code_verifier during token exchange
	// Server validates it matches the stored challenge
	err = ValidateCodeChallenge(params.CodeVerifier, params.CodeChallenge, params.CodeChallengeMethod)
	assert.NoError(t, err, "PKCE validation should succeed in full flow")
}

// Test edge case: boundary lengths
func TestCodeVerifier_BoundaryLengths(t *testing.T) {
	testCases := []struct {
		name     string
		length   int
		expected bool
	}{
		{"one below minimum", MinCodeVerifierLength - 1, false},
		{"exactly minimum", MinCodeVerifierLength, true},
		{"one above minimum", MinCodeVerifierLength + 1, true},
		{"one below maximum", MaxCodeVerifierLength - 1, true},
		{"exactly maximum", MaxCodeVerifierLength, true},
		{"one above maximum", MaxCodeVerifierLength + 1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			verifier := strings.Repeat("a", tc.length)
			result := IsValidCodeVerifier(verifier)
			assert.Equal(t, tc.expected, result)
		})
	}
}

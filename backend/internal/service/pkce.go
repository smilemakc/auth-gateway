package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

var (
	ErrInvalidCodeVerifier        = errors.New("invalid code verifier: must be 43-128 characters")
	ErrInvalidCodeChallenge       = errors.New("invalid code challenge")
	ErrInvalidCodeChallengeMethod = errors.New("invalid code challenge method: must be 'plain' or 'S256'")
	ErrCodeChallengeMismatch      = errors.New("code challenge verification failed")
)

const (
	CodeChallengeMethodPlain = "plain"
	CodeChallengeMethodS256  = "S256"

	MinCodeVerifierLength = 43
	MaxCodeVerifierLength = 128
)

func GenerateCodeVerifier() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func GenerateCodeChallenge(verifier string, method string) (string, error) {
	if !IsValidCodeVerifier(verifier) {
		return "", ErrInvalidCodeVerifier
	}

	switch method {
	case CodeChallengeMethodS256:
		hash := sha256.Sum256([]byte(verifier))
		return base64.RawURLEncoding.EncodeToString(hash[:]), nil
	case CodeChallengeMethodPlain:
		return verifier, nil
	default:
		return "", ErrInvalidCodeChallengeMethod
	}
}

func ValidateCodeChallenge(verifier, challenge, method string) error {
	if !IsValidCodeVerifier(verifier) {
		return ErrInvalidCodeVerifier
	}

	if challenge == "" {
		return ErrInvalidCodeChallenge
	}

	expectedChallenge, err := GenerateCodeChallenge(verifier, method)
	if err != nil {
		return err
	}

	if !secureCompare(expectedChallenge, challenge) {
		return ErrCodeChallengeMismatch
	}

	return nil
}

func IsValidCodeVerifier(verifier string) bool {
	length := len(verifier)
	if length < MinCodeVerifierLength || length > MaxCodeVerifierLength {
		return false
	}

	for _, c := range verifier {
		if !isUnreservedChar(c) {
			return false
		}
	}
	return true
}

func IsValidCodeChallengeMethod(method string) bool {
	return method == CodeChallengeMethodPlain || method == CodeChallengeMethodS256
}

func isUnreservedChar(c rune) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '.' || c == '_' || c == '~'
}

func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

type PKCEParams struct {
	CodeVerifier        string
	CodeChallenge       string
	CodeChallengeMethod string
}

func NewPKCEParams() (*PKCEParams, error) {
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		return nil, err
	}

	challenge, err := GenerateCodeChallenge(verifier, CodeChallengeMethodS256)
	if err != nil {
		return nil, err
	}

	return &PKCEParams{
		CodeVerifier:        verifier,
		CodeChallenge:       challenge,
		CodeChallengeMethod: CodeChallengeMethodS256,
	}, nil
}

func ValidatePKCERequest(codeChallenge, codeChallengeMethod string, requirePKCE bool) error {
	if requirePKCE {
		if codeChallenge == "" {
			return errors.New("code_challenge is required")
		}
		if codeChallengeMethod == "" {
			return errors.New("code_challenge_method is required")
		}
	}

	if codeChallenge != "" {
		if codeChallengeMethod == "" {
			codeChallengeMethod = CodeChallengeMethodPlain
		}
		if !IsValidCodeChallengeMethod(codeChallengeMethod) {
			return ErrInvalidCodeChallengeMethod
		}
	}

	return nil
}

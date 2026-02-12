package utils

import (
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// PasswordPolicy defines password validation rules
type PasswordPolicy struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumbers   bool
	RequireSpecial   bool
	MaxLength        int // 0 means no maximum
}

// DefaultPasswordPolicy returns a default password policy
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:        8,
		RequireUppercase: false,
		RequireLowercase: true,
		RequireNumbers:   false,
		RequireSpecial:   false,
		MaxLength:        0,
	}
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a hashed password with a plaintext password
func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// IsPasswordValid checks if a password meets the minimum requirements (backward compatibility)
func IsPasswordValid(password string) bool {
	policy := DefaultPasswordPolicy()
	return ValidatePassword(password, policy) == nil
}

// ValidatePassword validates a password against a policy
func ValidatePassword(password string, policy PasswordPolicy) error {
	if len(password) < policy.MinLength {
		return &PasswordValidationError{
			Message:   "Password must be at least %d characters long",
			MinLength: policy.MinLength,
		}
	}

	if policy.MaxLength > 0 && len(password) > policy.MaxLength {
		return &PasswordValidationError{
			Message:   "Password must be at most %d characters long",
			MaxLength: policy.MaxLength,
		}
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if policy.RequireUppercase && !hasUpper {
		return &PasswordValidationError{
			Message: "Password must contain at least one uppercase letter",
		}
	}

	if policy.RequireLowercase && !hasLower {
		return &PasswordValidationError{
			Message: "Password must contain at least one lowercase letter",
		}
	}

	if policy.RequireNumbers && !hasNumber {
		return &PasswordValidationError{
			Message: "Password must contain at least one number",
		}
	}

	if policy.RequireSpecial && !hasSpecial {
		return &PasswordValidationError{
			Message: "Password must contain at least one special character",
		}
	}

	return nil
}

// PasswordValidationError represents a password validation error
type PasswordValidationError struct {
	Message   string
	MinLength int
	MaxLength int
}

func (e *PasswordValidationError) Error() string {
	if e.MinLength > 0 {
		return strings.Replace(e.Message, "%d", string(rune(e.MinLength)), 1)
	}
	if e.MaxLength > 0 {
		return strings.Replace(e.Message, "%d", string(rune(e.MaxLength)), 1)
	}
	return e.Message
}

// CommonPasswords is a list of common passwords to check against (optional)
var CommonPasswords = []string{
	"password", "12345678", "123456789", "1234567890",
	"qwerty", "abc123", "monkey", "1234567",
	"letmein", "trustno1", "dragon", "baseball",
	"iloveyou", "master", "sunshine", "ashley",
	"bailey", "passw0rd", "shadow", "123123",
	"654321", "superman", "qazwsx", "michael",
}

// dummyPasswordHash is a bcrypt hash of a dummy password used to prevent timing attacks
// This hash is computed once and reused to avoid performance impact
var dummyPasswordHash = func() string {
	// Generate a dummy hash that will never match any real password
	// Using a fixed dummy password "dummy_password_for_timing_attack_protection_12345"
	hash, _ := HashPassword("dummy_password_for_timing_attack_protection_12345", 10)
	return hash
}()

// GetDummyPasswordHash returns a dummy password hash for timing attack protection
// This hash is used when a user is not found, ensuring bcrypt comparison always takes place
func GetDummyPasswordHash() string {
	return dummyPasswordHash
}

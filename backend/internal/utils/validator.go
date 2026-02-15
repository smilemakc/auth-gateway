package utils

import (
	"errors"
	"regexp"
	"strings"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,100}$`)
	phoneRegex    = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`) // E.164 format
)

// IsValidEmail checks if an email is valid
func IsValidEmail(email string) bool {
	if len(email) > 255 {
		return false
	}
	return emailRegex.MatchString(email)
}

// ValidateEmail performs comprehensive email validation including RFC 5321 limits
// and header injection prevention.
func ValidateEmail(email string) error {
	if len(email) > 254 {
		return errors.New("email too long")
	}

	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || len(parts[0]) > 64 {
		return errors.New("invalid email format")
	}

	if strings.ContainsAny(email, "\r\n") {
		return errors.New("email contains invalid characters")
	}

	if !IsValidEmail(email) {
		return errors.New("invalid email format")
	}

	return nil
}

// IsValidUsername checks if a username is valid
// Username must be 3-100 characters and contain only letters, numbers, underscores, and hyphens
func IsValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

// NormalizeEmail normalizes an email address (lowercase, trimmed)
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizeUsername normalizes a username (lowercase, trimmed)
func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// IsValidPhone checks if a phone number is valid (E.164 format)
// Accepts: +79991234567 or 79991234567
func IsValidPhone(phone string) bool {
	normalized := NormalizePhone(phone)
	return phoneRegex.MatchString(normalized)
}

// NormalizePhone normalizes a phone number by removing non-digit characters
// (except '+') and ensuring it starts with '+'
func NormalizePhone(phone string) string {
	// Remove everything except digits and '+'
	re := regexp.MustCompile(`[^\d+]`)
	normalized := re.ReplaceAllString(phone, "")

	// Add '+' if missing and first character is a digit
	if normalized != "" && normalized[0] >= '0' && normalized[0] <= '9' {
		return "+" + normalized
	}

	return normalized
}

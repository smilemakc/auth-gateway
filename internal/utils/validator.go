package utils

import (
	"regexp"
	"strings"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,100}$`)
)

// IsValidEmail checks if an email is valid
func IsValidEmail(email string) bool {
	if len(email) > 255 {
		return false
	}
	return emailRegex.MatchString(email)
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

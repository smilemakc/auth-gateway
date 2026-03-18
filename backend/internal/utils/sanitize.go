package utils

import (
	"html"
	"strings"
)

// SanitizeHTML escapes HTML entities and trims whitespace.
// Use on user-provided text fields to prevent XSS.
func SanitizeHTML(input string) string {
	return html.EscapeString(strings.TrimSpace(input))
}

// SanitizeUsername sanitizes a username: trims whitespace, removes control characters.
func SanitizeUsername(input string) string {
	trimmed := strings.TrimSpace(input)
	// Remove control characters
	return strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, trimmed)
}

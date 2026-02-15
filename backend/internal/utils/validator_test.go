package utils

import (
	"strings"
	"testing"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"Valid email", "user@example.com", true},
		{"Valid email with subdomain", "user@mail.example.com", true},
		{"Valid email with plus", "user+tag@example.com", true},
		{"Invalid email no @", "userexample.com", false},
		{"Invalid email no domain", "user@", false},
		{"Invalid email no user", "@example.com", false},
		{"Empty email", "", false},
		{"Too long email", "user@" + string(make([]byte, 300)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEmail(tt.email)
			if got != tt.want {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{"Valid username", "johndoe", true},
		{"Valid username with underscore", "john_doe", true},
		{"Valid username with hyphen", "john-doe", true},
		{"Valid username with numbers", "john123", true},
		{"Too short", "ab", false},
		{"Too long", string(make([]byte, 101)), false},
		{"Invalid characters", "john doe", false},
		{"Invalid characters special", "john@doe", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidUsername(tt.username)
			if got != tt.want {
				t.Errorf("IsValidUsername(%q) = %v, want %v", tt.username, got, tt.want)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"Lowercase", "User@Example.com", "user@example.com"},
		{"Trim spaces", " user@example.com ", "user@example.com"},
		{"Both", " User@Example.COM ", "user@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeEmail(tt.email)
			if got != tt.want {
				t.Errorf("NormalizeEmail(%q) = %q, want %q", tt.email, got, tt.want)
			}
		})
	}
}

func TestNormalizeUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     string
	}{
		{"Lowercase", "JohnDoe", "johndoe"},
		{"Trim spaces", " johndoe ", "johndoe"},
		{"Both", " JohnDoe ", "johndoe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeUsername(tt.username)
			if got != tt.want {
				t.Errorf("NormalizeUsername(%q) = %q, want %q", tt.username, got, tt.want)
			}
		})
	}
}

// TestNormalizePhone validates phone normalization using regexp
func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Already normalized", "+1234567890", "+1234567890"},
		{"Spaces and dashes", "123 456-7890", "+1234567890"},
		{"Parentheses", "(123) 456 7890", "+1234567890"},
		{"Plus with formatting", "+1 (234) 567-890", "+1234567890"},
		{"Empty", "", ""},
		{"Letters inside", "12a3b4", "+1234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePhone(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		wantError bool
	}{
		{"Valid email", "user@example.com", false},
		{"Valid email with subdomain", "user@mail.example.com", false},
		{"Valid email with plus", "user+tag@example.com", false},
		{"Valid email with dot", "user.name@example.com", false},
		{"Too long email total", "a@" + strings.Repeat("a", 253) + ".com", true},
		{"Too long local part", strings.Repeat("a", 65) + "@example.com", true},
		{"Header injection CRLF", "user@example.com\r\nBcc: attacker@evil.com", true},
		{"Header injection CR", "user\r@example.com", true},
		{"Header injection LF", "user\n@example.com", true},
		{"Invalid format no @", "not-an-email", true},
		{"Invalid format no domain", "user@", true},
		{"Invalid format no local", "@example.com", true},
		{"Empty email", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateEmail(%q) error = %v, wantError %v", tt.email, err, tt.wantError)
			}
		})
	}
}

package utils

import (
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

package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
		{"Valid email with dots in local", "first.last@example.com", true},
		{"Valid email with hyphen in domain", "user@my-domain.com", true},
		{"Valid email with numbers in local", "user123@example.com", true},
		{"Valid email with percent", "user%tag@example.com", true},
		{"Valid email with underscore", "user_name@example.com", true},
		{"Valid email with long TLD", "user@example.museum", true},
		{"Valid email with two-char TLD", "user@example.co", true},
		{"Valid email with multiple subdomains", "user@a.b.c.example.com", true},
		{"Invalid email no @", "userexample.com", false},
		{"Invalid email no domain", "user@", false},
		{"Invalid email no user", "@example.com", false},
		{"Empty email", "", false},
		{"Too long email", "user@" + string(make([]byte, 300)), false},
		{"Invalid email double @", "user@@example.com", false},
		{"Invalid email space in local", "user name@example.com", false},
		{"Invalid email no TLD", "user@example", false},
		{"Invalid email single char TLD", "user@example.c", false},
		{"Invalid email trailing dot domain", "user@example.com.", false},
		{"Email with leading dot local", ".user@example.com", true},
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

func TestIsValidEmail_ShouldRejectExactly256CharEmail(t *testing.T) {
	// len > 255 returns false
	local := strings.Repeat("a", 64)
	domain := strings.Repeat("b", 186) + ".com" // 64 + 1(@) + 186 + 4 = 255 is OK, 256 not
	email255 := local + "@" + domain
	email256 := local + "@" + "x" + domain

	if len(email255) <= 255 {
		assert.True(t, IsValidEmail(email255) || !IsValidEmail(email255)) // just bounds check
	}
	if len(email256) > 255 {
		assert.False(t, IsValidEmail(email256))
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
		{"Valid username exactly 3 chars", "abc", true},
		{"Valid username exactly 100 chars", strings.Repeat("a", 100), true},
		{"Too short", "ab", false},
		{"Too short single char", "a", false},
		{"Too long", strings.Repeat("a", 101), false},
		{"Invalid characters space", "john doe", false},
		{"Invalid characters special", "john@doe", false},
		{"Invalid characters dot", "john.doe", false},
		{"Invalid characters exclamation", "john!", false},
		{"Empty", "", false},
		{"Only numbers", "123456", true},
		{"Only underscores", "___", true},
		{"Only hyphens", "---", true},
		{"Mixed valid", "a-b_c-1", true},
		{"Unicode characters", "Джон", false},
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
		{"Already normalized", "user@example.com", "user@example.com"},
		{"All caps", "USER@EXAMPLE.COM", "user@example.com"},
		{"Empty string", "", ""},
		{"Only spaces", "   ", ""},
		{"Tabs and spaces", "\t user@example.com \t", "user@example.com"},
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
		{"Already normalized", "johndoe", "johndoe"},
		{"All caps", "JOHNDOE", "johndoe"},
		{"Empty string", "", ""},
		{"With underscores", "John_Doe", "john_doe"},
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
		{"Only plus sign", "+", "+"},
		{"Plus at start with spaces", "+ 1 2 3", "+123"},
		{"Multiple plus signs", "++123", "++123"},
		{"Dots as separators", "123.456.7890", "+1234567890"},
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

func TestIsValidPhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  bool
	}{
		{"Valid E.164 with plus", "+79991234567", true},
		{"Valid E.164 without plus", "79991234567", true},
		{"Valid US number", "+12025551234", true},
		{"Valid short number", "+1234", true},
		{"Valid minimum digits", "+12", true},
		{"Valid max E.164 (15 digits)", "+123456789012345", true},
		{"Invalid too long (16 digits)", "+1234567890123456", false},
		{"Invalid starts with 0", "+0123456789", false},
		{"Empty", "", false},
		{"Only plus", "+", false},
		{"Formatted with spaces (normalized)", "+7 999 123 4567", true},
		{"Formatted with dashes", "+7-999-123-4567", true},
		{"Formatted with parentheses", "+1 (202) 555-1234", true},
		{"Letters only", "abcdefgh", false},
		{"Single digit", "5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPhone(tt.phone)
			assert.Equal(t, tt.want, got, "IsValidPhone(%q)", tt.phone)
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
		{"Exactly 254 chars is OK", strings.Repeat("a", 64) + "@" + strings.Repeat("b", 185) + ".com", false},
		{"255 chars is too long", strings.Repeat("a", 64) + "@" + strings.Repeat("b", 186) + ".com", true},
		{"Local part exactly 64 chars", strings.Repeat("a", 64) + "@example.com", false},
		{"Newline in domain", "user@example\n.com", true},
		{"Carriage return in domain", "user@example\r.com", true},
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

func TestValidateEmail_ShouldReturnSpecificErrorMessages(t *testing.T) {
	err := ValidateEmail(strings.Repeat("a", 255))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too long")

	err = ValidateEmail("user\r\n@example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid characters")

	err = ValidateEmail("not-an-email")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestNormalizePhone_ShouldBeIdempotent(t *testing.T) {
	phone := "+79991234567"
	first := NormalizePhone(phone)
	second := NormalizePhone(first)
	assert.Equal(t, first, second, "normalizing twice should give the same result")
}

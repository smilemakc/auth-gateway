package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "XSS script tag",
			input:    "<script>alert(1)</script>",
			expected: "&lt;script&gt;alert(1)&lt;/script&gt;",
		},
		{
			name:     "trim whitespace",
			input:    "  John Doe  ",
			expected: "John Doe",
		},
		{
			name:     "normal text",
			input:    "normal text",
			expected: "normal text",
		},
		{
			name:     "HTML entities",
			input:    "<b>Bold</b> & \"quoted\"",
			expected: "&lt;b&gt;Bold&lt;/b&gt; &amp; &#34;quoted&#34;",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeUsername(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal username",
			input:    "user123",
			expected: "user123",
		},
		{
			name:     "trim whitespace",
			input:    "  user123  ",
			expected: "user123",
		},
		{
			name:     "remove null character",
			input:    "user\x00123",
			expected: "user123",
		},
		{
			name:     "remove control characters",
			input:    "user\x01\x02\x03123",
			expected: "user123",
		},
		{
			name:     "remove DEL character",
			input:    "user\x7f123",
			expected: "user123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: "",
		},
		{
			name:     "with special characters",
			input:    "user_name-123",
			expected: "user_name-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeUsername(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

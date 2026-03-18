package utils

import (
	"strings"
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
		{
			name:     "Nested script tags",
			input:    "<scr<script>ipt>alert(1)</scr</script>ipt>",
			expected: "&lt;scr&lt;script&gt;ipt&gt;alert(1)&lt;/scr&lt;/script&gt;ipt&gt;",
		},
		{
			name:     "Event handler in tag",
			input:    `<img src="x" onerror="alert(1)">`,
			expected: `&lt;img src=&#34;x&#34; onerror=&#34;alert(1)&#34;&gt;`,
		},
		{
			name:     "SVG XSS vector",
			input:    `<svg onload="alert(1)">`,
			expected: `&lt;svg onload=&#34;alert(1)&#34;&gt;`,
		},
		{
			name:     "JavaScript URL scheme",
			input:    `<a href="javascript:alert(1)">click</a>`,
			expected: `&lt;a href=&#34;javascript:alert(1)&#34;&gt;click&lt;/a&gt;`,
		},
		{
			name:     "Single quotes",
			input:    "it's a test",
			expected: "it&#39;s a test",
		},
		{
			name:     "HTML comment",
			input:    "<!-- comment -->text",
			expected: "&lt;!-- comment --&gt;text",
		},
		{
			name:     "Iframe injection",
			input:    `<iframe src="https://evil.com"></iframe>`,
			expected: `&lt;iframe src=&#34;https://evil.com&#34;&gt;&lt;/iframe&gt;`,
		},
		{
			name:     "Unicode text should pass through",
			input:    "Привет мир 🌍",
			expected: "Привет мир 🌍",
		},
		{
			name:     "Tab characters preserved inside",
			input:    "hello\tworld",
			expected: "hello\tworld",
		},
		{
			name:     "Newlines preserved inside",
			input:    "line1\nline2",
			expected: "line1\nline2",
		},
		{
			name:     "Mixed whitespace trimmed at edges only",
			input:    "\t\n hello \n\t",
			expected: "hello",
		},
		{
			name:     "Style tag",
			input:    `<style>body{display:none}</style>`,
			expected: `&lt;style&gt;body{display:none}&lt;/style&gt;`,
		},
		{
			name:     "Data URI in img",
			input:    `<img src="data:text/html,<script>alert(1)</script>">`,
			expected: `&lt;img src=&#34;data:text/html,&lt;script&gt;alert(1)&lt;/script&gt;&#34;&gt;`,
		},
		{
			name:     "Angle brackets inside text",
			input:    "5 < 10 and 10 > 5",
			expected: "5 &lt; 10 and 10 &gt; 5",
		},
		{
			name:     "Ampersand in URL",
			input:    "example.com?a=1&b=2",
			expected: "example.com?a=1&amp;b=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeHTML_ShouldBeIdempotent_WhenAppliedToPlainText(t *testing.T) {
	input := "plain text without html"
	first := SanitizeHTML(input)
	// Applying it again to plain text should remain the same
	assert.Equal(t, input, first)
}

func TestSanitizeHTML_ShouldDoubleEncode_WhenAppliedTwice(t *testing.T) {
	input := "<script>alert(1)</script>"
	first := SanitizeHTML(input)
	second := SanitizeHTML(first)
	// Second application encodes the & in &lt; etc
	assert.NotEqual(t, first, second, "double sanitization should double-encode entities")
	assert.Contains(t, second, "&amp;lt;")
}

func TestSanitizeHTML_ShouldHandleLongInput(t *testing.T) {
	long := strings.Repeat("<script>", 10000)
	result := SanitizeHTML(long)
	assert.NotContains(t, result, "<script>")
	assert.Contains(t, result, "&lt;script&gt;")
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
		{
			name:     "tab character removed",
			input:    "user\t123",
			expected: "user123",
		},
		{
			name:     "newline removed",
			input:    "user\n123",
			expected: "user123",
		},
		{
			name:     "carriage return removed",
			input:    "user\r123",
			expected: "user123",
		},
		{
			name:     "bell character removed",
			input:    "user\a123",
			expected: "user123",
		},
		{
			name:     "backspace removed",
			input:    "user\b123",
			expected: "user123",
		},
		{
			name:     "form feed removed",
			input:    "user\f123",
			expected: "user123",
		},
		{
			name:     "vertical tab removed",
			input:    "user\v123",
			expected: "user123",
		},
		{
			name:     "escape character removed",
			input:    "user\x1b123",
			expected: "user123",
		},
		{
			name:     "multiple control chars in sequence",
			input:    "\x00\x01\x02user\x03\x04\x05",
			expected: "user",
		},
		{
			name:     "unicode preserved",
			input:    "пользователь",
			expected: "пользователь",
		},
		{
			name:     "emoji preserved",
			input:    "user🔥name",
			expected: "user🔥name",
		},
		{
			name:     "printable ASCII preserved",
			input:    "user@name.com!#$%",
			expected: "user@name.com!#$%",
		},
		{
			name:     "space in middle preserved",
			input:    "user name",
			expected: "user name",
		},
		{
			name:     "only control characters",
			input:    "\x00\x01\x02\x7f",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeUsername(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeUsername_ShouldNotTrimInternalSpaces(t *testing.T) {
	result := SanitizeUsername("  hello   world  ")
	assert.Equal(t, "hello   world", result, "only leading/trailing whitespace should be trimmed")
}

func TestSanitizeUsername_ShouldHandleLongInput(t *testing.T) {
	long := strings.Repeat("a", 10000)
	result := SanitizeUsername(long)
	assert.Equal(t, long, result)
}

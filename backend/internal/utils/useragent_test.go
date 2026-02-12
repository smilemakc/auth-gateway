package utils

import (
	"testing"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestParseUserAgent(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  models.DeviceInfo
	}{
		{
			name:      "Empty",
			userAgent: "",
			expected: models.DeviceInfo{
				DeviceType: "unknown",
				OS:         "unknown",
				Browser:    "unknown",
				IsBot:      false,
			},
		},
		{
			name:      "Chrome on macOS",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			expected: models.DeviceInfo{
				DeviceType: "desktop",
				OS:         "macOS 10.15.7",
				Browser:    "Chrome 120.0.0.0",
				IsBot:      false,
			},
		},
		{
			name:      "Safari on iPhone",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			expected: models.DeviceInfo{
				DeviceType: "mobile",
				OS:         "iOS 17.0",
				Browser:    "Safari 17.0",
				IsBot:      false,
			},
		},
		{
			name:      "GoogleBot",
			userAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			expected: models.DeviceInfo{
				DeviceType: "bot",
				OS:         "unknown",
				Browser:    "unknown",
				IsBot:      true,
			},
		},
		{
			name:      "Firefox on Linux",
			userAgent: "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/110.0",
			expected: models.DeviceInfo{
				DeviceType: "desktop",
				OS:         "Ubuntu",
				Browser:    "Firefox 110.0",
				IsBot:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseUserAgent(tt.userAgent)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSessionName(t *testing.T) {
	tests := []struct {
		name       string
		deviceInfo models.DeviceInfo
		expected   string
	}{
		{
			name: "Desktop",
			deviceInfo: models.DeviceInfo{
				Browser: "Chrome",
				OS:      "macOS",
				IsBot:   false,
			},
			expected: "Chrome on macOS",
		},
		{
			name: "Bot",
			deviceInfo: models.DeviceInfo{
				Browser: "Chrome",
				OS:      "Linux",
				IsBot:   true,
			},
			expected: "Chrome (Bot)",
		},
		{
			name: "Unknown",
			deviceInfo: models.DeviceInfo{
				Browser: "unknown",
				OS:      "unknown",
				IsBot:   false,
			},
			expected: "Unknown Browser on Unknown OS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSessionName(tt.deviceInfo)
			assert.Equal(t, tt.expected, result)
		})
	}
}

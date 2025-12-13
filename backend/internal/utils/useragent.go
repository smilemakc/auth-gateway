package utils

import (
	"github.com/smilemakc/auth-gateway/internal/models"
	"regexp"
	"strings"
)

// ParseUserAgent parses a user agent string and extracts device information
func ParseUserAgent(userAgent string) models.DeviceInfo {
	if userAgent == "" {
		return models.DeviceInfo{
			DeviceType: "unknown",
			OS:         "unknown",
			Browser:    "unknown",
			IsBot:      false,
		}
	}

	info := models.DeviceInfo{
		DeviceType: detectDeviceType(userAgent),
		OS:         detectOS(userAgent),
		Browser:    detectBrowser(userAgent),
		IsBot:      detectBot(userAgent),
	}

	return info
}

// GenerateSessionName creates a human-readable session name from device info
func GenerateSessionName(deviceInfo models.DeviceInfo) string {
	browser := deviceInfo.Browser
	if browser == "unknown" {
		browser = "Unknown Browser"
	}

	os := deviceInfo.OS
	if os == "unknown" {
		os = "Unknown OS"
	}

	if deviceInfo.IsBot {
		return browser + " (Bot)"
	}

	return browser + " on " + os
}

// detectDeviceType detects the device type from user agent
func detectDeviceType(ua string) string {
	ua = strings.ToLower(ua)

	// Bot detection
	if detectBot(ua) {
		return "bot"
	}

	// Mobile devices
	if strings.Contains(ua, "mobile") ||
		strings.Contains(ua, "android") ||
		strings.Contains(ua, "iphone") ||
		strings.Contains(ua, "ipod") ||
		strings.Contains(ua, "windows phone") ||
		strings.Contains(ua, "blackberry") {
		return "mobile"
	}

	// Tablets
	if strings.Contains(ua, "tablet") ||
		strings.Contains(ua, "ipad") ||
		(strings.Contains(ua, "android") && !strings.Contains(ua, "mobile")) {
		return "tablet"
	}

	// Desktop
	if strings.Contains(ua, "windows") ||
		strings.Contains(ua, "macintosh") ||
		strings.Contains(ua, "linux") ||
		strings.Contains(ua, "x11") {
		return "desktop"
	}

	return "unknown"
}

// detectOS detects the operating system from user agent
func detectOS(ua string) string {
	ua = strings.ToLower(ua)

	// Windows
	if strings.Contains(ua, "windows nt 10.0") {
		return "Windows 10/11"
	} else if strings.Contains(ua, "windows nt 6.3") {
		return "Windows 8.1"
	} else if strings.Contains(ua, "windows nt 6.2") {
		return "Windows 8"
	} else if strings.Contains(ua, "windows nt 6.1") {
		return "Windows 7"
	} else if strings.Contains(ua, "windows") {
		return "Windows"
	}

	// macOS
	if strings.Contains(ua, "mac os x") {
		re := regexp.MustCompile(`mac os x ([\d_]+)`)
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			version := strings.Replace(matches[1], "_", ".", -1)
			return "macOS " + version
		}
		return "macOS"
	}

	// iOS
	if strings.Contains(ua, "iphone os") || strings.Contains(ua, "ipad") {
		re := regexp.MustCompile(`(?:iphone|cpu) os ([\d_]+)`)
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			version := strings.Replace(matches[1], "_", ".", -1)
			return "iOS " + version
		}
		return "iOS"
	}

	// Android
	if strings.Contains(ua, "android") {
		re := regexp.MustCompile(`android ([\d.]+)`)
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			return "Android " + matches[1]
		}
		return "Android"
	}

	// Linux
	if strings.Contains(ua, "linux") {
		if strings.Contains(ua, "ubuntu") {
			return "Ubuntu"
		} else if strings.Contains(ua, "fedora") {
			return "Fedora"
		} else if strings.Contains(ua, "debian") {
			return "Debian"
		}
		return "Linux"
	}

	return "unknown"
}

// detectBrowser detects the browser from user agent
func detectBrowser(ua string) string {
	ua = strings.ToLower(ua)

	// Edge (Chromium-based)
	if strings.Contains(ua, "edg/") {
		re := regexp.MustCompile(`edg/([\d.]+)`)
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			return "Edge " + matches[1]
		}
		return "Edge"
	}

	// Opera
	if strings.Contains(ua, "opr/") || strings.Contains(ua, "opera/") {
		re := regexp.MustCompile(`(?:opr|opera)/([\d.]+)`)
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			return "Opera " + matches[1]
		}
		return "Opera"
	}

	// Chrome
	if strings.Contains(ua, "chrome/") && !strings.Contains(ua, "edg/") {
		re := regexp.MustCompile(`chrome/([\d.]+)`)
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			return "Chrome " + matches[1]
		}
		return "Chrome"
	}

	// Safari
	if strings.Contains(ua, "safari/") && !strings.Contains(ua, "chrome/") {
		re := regexp.MustCompile(`version/([\d.]+)`)
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			return "Safari " + matches[1]
		}
		return "Safari"
	}

	// Firefox
	if strings.Contains(ua, "firefox/") {
		re := regexp.MustCompile(`firefox/([\d.]+)`)
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			return "Firefox " + matches[1]
		}
		return "Firefox"
	}

	// Internet Explorer
	if strings.Contains(ua, "msie") || strings.Contains(ua, "trident/") {
		re := regexp.MustCompile(`(?:msie |rv:)([\d.]+)`)
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			return "IE " + matches[1]
		}
		return "Internet Explorer"
	}

	return "unknown"
}

// detectBot detects if the user agent is a bot/crawler
func detectBot(ua string) bool {
	ua = strings.ToLower(ua)

	botKeywords := []string{
		"bot", "crawler", "spider", "scraper", "curl", "wget",
		"python-requests", "http", "library", "monitor", "checker",
		"googlebot", "bingbot", "slackbot", "facebookexternalhit",
		"twitterbot", "linkedinbot", "discordbot", "telegrambot",
	}

	for _, keyword := range botKeywords {
		if strings.Contains(ua, keyword) {
			return true
		}
	}

	return false
}

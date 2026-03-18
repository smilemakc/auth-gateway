package service

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PasswordChecker struct {
	httpClient *http.Client
	enabled    bool
}

func NewPasswordChecker(enabled bool) *PasswordChecker {
	return &PasswordChecker{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		enabled:    enabled,
	}
}

func (pc *PasswordChecker) IsCompromised(ctx context.Context, password string) (bool, int) {
	if !pc.enabled {
		return false, 0
	}

	hash := fmt.Sprintf("%X", sha1.Sum([]byte(password)))
	prefix := hash[:5]
	suffix := hash[5:]

	url := fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", prefix)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, 0
	}
	req.Header.Set("User-Agent", "AuthGateway-PasswordCheck")

	resp, err := pc.httpClient.Do(req)
	if err != nil {
		return false, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, 0
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, 0
	}

	for _, line := range strings.Split(string(body), "\r\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.EqualFold(parts[0], suffix) {
			count, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			return true, count
		}
	}

	return false, 0
}

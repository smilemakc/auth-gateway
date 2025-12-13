package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TwilioProvider implements SMSProvider for Twilio
type TwilioProvider struct {
	accountSID string
	authToken  string
	fromNumber string
	httpClient *http.Client
}

// TwilioConfig holds Twilio configuration
type TwilioConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
}

// NewTwilioProvider creates a new Twilio SMS provider
func NewTwilioProvider(config TwilioConfig) (*TwilioProvider, error) {
	provider := &TwilioProvider{
		accountSID: config.AccountSID,
		authToken:  config.AuthToken,
		fromNumber: config.FromNumber,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, err
	}

	return provider, nil
}

// SendSMS sends an SMS via Twilio
func (t *TwilioProvider) SendSMS(ctx context.Context, to, message string) (string, error) {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.accountSID)

	data := url.Values{}
	data.Set("To", to)
	data.Set("From", t.fromNumber)
	data.Set("Body", message)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	req.SetBasicAuth(t.accountSID, t.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: failed to read response: %v", ErrSendFailed, err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: twilio returned status %d: %s", ErrSendFailed, resp.StatusCode, string(body))
	}

	// Parse response to get message SID
	var twilioResp struct {
		SID          string  `json:"sid"`
		Status       string  `json:"status"`
		ErrorCode    *int    `json:"error_code"`
		ErrorMessage *string `json:"error_message"`
	}

	if err := json.Unmarshal(body, &twilioResp); err != nil {
		return "", fmt.Errorf("%w: failed to parse response: %v", ErrSendFailed, err)
	}

	if twilioResp.ErrorCode != nil {
		return "", fmt.Errorf("%w: twilio error %d: %s", ErrSendFailed, *twilioResp.ErrorCode, *twilioResp.ErrorMessage)
	}

	return twilioResp.SID, nil
}

// GetProviderName returns the provider name
func (t *TwilioProvider) GetProviderName() string {
	return string(ProviderTwilio)
}

// ValidateConfig validates the Twilio configuration
func (t *TwilioProvider) ValidateConfig() error {
	if t.accountSID == "" {
		return fmt.Errorf("%w: account SID is required", ErrProviderNotConfigured)
	}
	if t.authToken == "" {
		return fmt.Errorf("%w: auth token is required", ErrProviderNotConfigured)
	}
	if t.fromNumber == "" {
		return fmt.Errorf("%w: from number is required", ErrProviderNotConfigured)
	}
	return nil
}

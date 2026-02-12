package sms

import (
	"context"
	"errors"
)

// SMSProvider defines the interface for SMS providers
type SMSProvider interface {
	// SendSMS sends an SMS message to the specified phone number
	SendSMS(ctx context.Context, to, message string) (messageID string, err error)

	// GetProviderName returns the name of the SMS provider
	GetProviderName() string

	// ValidateConfig validates the provider configuration
	ValidateConfig() error
}

// SMSMessage represents an SMS message to be sent
type SMSMessage struct {
	To      string
	Message string
	From    string // Optional: override default from number
}

// SMSResponse represents the response from sending an SMS
type SMSResponse struct {
	MessageID string
	Provider  string
	Status    string
	Cost      *float64 // Optional: cost of sending the message
}

var (
	// ErrProviderNotConfigured is returned when the SMS provider is not properly configured
	ErrProviderNotConfigured = errors.New("sms provider not configured")

	// ErrInvalidPhoneNumber is returned when the phone number format is invalid
	ErrInvalidPhoneNumber = errors.New("invalid phone number format")

	// ErrSendFailed is returned when sending SMS fails
	ErrSendFailed = errors.New("failed to send SMS")

	// ErrProviderUnavailable is returned when the provider is unavailable
	ErrProviderUnavailable = errors.New("sms provider unavailable")

	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = errors.New("sms rate limit exceeded")
)

// ProviderType represents the type of SMS provider
type ProviderType string

const (
	ProviderTwilio ProviderType = "twilio"
	ProviderAWSSNS ProviderType = "aws_sns"
	ProviderVonage ProviderType = "vonage"
	ProviderMock   ProviderType = "mock" // For testing
)

// IsValid checks if the provider type is valid
func (p ProviderType) IsValid() bool {
	switch p {
	case ProviderTwilio, ProviderAWSSNS, ProviderVonage, ProviderMock:
		return true
	}
	return false
}

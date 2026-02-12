package sms

import (
	"context"
	"fmt"
)

// ProviderConfig holds common provider configuration
type ProviderConfig struct {
	Provider        ProviderType
	TwilioConfig    *TwilioConfig
	AWSSNSConfig    *AWSSNSConfig
	EnableMockInDev bool
}

// NewProvider creates a new SMS provider based on configuration
func NewProvider(ctx context.Context, config ProviderConfig) (SMSProvider, error) {
	// Force mock provider in development if enabled
	if config.EnableMockInDev && config.Provider == ProviderMock {
		return NewMockProvider(), nil
	}

	switch config.Provider {
	case ProviderTwilio:
		if config.TwilioConfig == nil {
			return nil, fmt.Errorf("%w: Twilio configuration is required", ErrProviderNotConfigured)
		}
		return NewTwilioProvider(*config.TwilioConfig)

	case ProviderAWSSNS:
		if config.AWSSNSConfig == nil {
			return nil, fmt.Errorf("%w: AWS SNS configuration is required", ErrProviderNotConfigured)
		}
		return NewAWSSNSProvider(ctx, *config.AWSSNSConfig)

	case ProviderMock:
		return NewMockProvider(), nil

	default:
		return nil, fmt.Errorf("unsupported SMS provider: %s", config.Provider)
	}
}

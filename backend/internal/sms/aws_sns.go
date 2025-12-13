package sms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
)

// AWSSNSProvider implements SMSProvider for AWS SNS
type AWSSNSProvider struct {
	client     *sns.Client
	fromNumber string
}

// AWSSNSConfig holds AWS SNS configuration
type AWSSNSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	FromNumber      string // Optional: for sender ID
}

// NewAWSSNSProvider creates a new AWS SNS provider
func NewAWSSNSProvider(ctx context.Context, cfg AWSSNSConfig) (*AWSSNSProvider, error) {
	var awsConfig aws.Config
	var err error

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		// Use static credentials
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		)
	} else {
		// Use default credential chain (IAM roles, environment variables, etc.)
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: failed to load AWS config: %v", ErrProviderNotConfigured, err)
	}

	provider := &AWSSNSProvider{
		client:     sns.NewFromConfig(awsConfig),
		fromNumber: cfg.FromNumber,
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, err
	}

	return provider, nil
}

// SendSMS sends an SMS via AWS SNS
func (a *AWSSNSProvider) SendSMS(ctx context.Context, to, message string) (string, error) {
	input := &sns.PublishInput{
		Message:     aws.String(message),
		PhoneNumber: aws.String(to),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"AWS.SNS.SMS.SMSType": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Transactional"), // For OTP codes, use Transactional for higher reliability
			},
		},
	}

	// Add sender ID if configured
	if a.fromNumber != "" {
		input.MessageAttributes["AWS.SNS.SMS.SenderID"] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(a.fromNumber),
		}
	}

	result, err := a.client.Publish(ctx, input)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	if result.MessageId == nil {
		return "", fmt.Errorf("%w: no message ID returned", ErrSendFailed)
	}

	return *result.MessageId, nil
}

// GetProviderName returns the provider name
func (a *AWSSNSProvider) GetProviderName() string {
	return string(ProviderAWSSNS)
}

// ValidateConfig validates the AWS SNS configuration
func (a *AWSSNSProvider) ValidateConfig() error {
	if a.client == nil {
		return fmt.Errorf("%w: SNS client is not initialized", ErrProviderNotConfigured)
	}
	return nil
}

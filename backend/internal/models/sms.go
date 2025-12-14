package models

import (
	"time"

	"github.com/google/uuid"
)

// SMSSettings represents SMS provider settings in the database
type SMSSettings struct {
	ID                 uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Provider           string     `json:"provider" bun:"provider"`
	Enabled            bool       `json:"enabled" bun:"enabled"`
	AccountSID         *string    `json:"account_sid,omitempty" bun:"account_sid"`
	AuthToken          *string    `json:"-" bun:"auth_token"` // Never expose in JSON
	FromNumber         *string    `json:"from_number,omitempty" bun:"from_number"`
	AWSRegion          *string    `json:"aws_region,omitempty" bun:"aws_region"`
	AWSAccessKeyID     *string    `json:"aws_access_key_id,omitempty" bun:"aws_access_key_id"`
	AWSSecretAccessKey *string    `json:"-" bun:"aws_secret_access_key"` // Never expose in JSON
	AWSSenderID        *string    `json:"aws_sender_id,omitempty" bun:"aws_sender_id"`
	MaxPerHour         int        `json:"max_per_hour" bun:"max_per_hour"`
	MaxPerDay          int        `json:"max_per_day" bun:"max_per_day"`
	MaxPerNumber       int        `json:"max_per_number" bun:"max_per_number"`
	CreatedAt          time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt          time.Time  `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	CreatedBy          *uuid.UUID `json:"created_by,omitempty" bun:"created_by,type:uuid"`
}

// SMSLog represents a log of sent SMS messages
type SMSLog struct {
	ID           uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Phone        string     `json:"phone" bun:"phone"`
	Message      string     `json:"message" bun:"message"`
	Type         OTPType    `json:"type" bun:"type"`
	Provider     string     `json:"provider" bun:"provider"`
	MessageID    *string    `json:"message_id,omitempty" bun:"message_id"`
	Status       SMSStatus  `json:"status" bun:"status"`
	ErrorMessage *string    `json:"error_message,omitempty" bun:"error_message"`
	SentAt       *time.Time `json:"sent_at,omitempty" bun:"sent_at"`
	CreatedAt    time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UserID       *uuid.UUID `json:"user_id,omitempty" bun:"user_id,type:uuid"`
	IPAddress    *string    `json:"ip_address,omitempty" bun:"ip_address"`
}

// SMSStatus represents the status of an SMS message
type SMSStatus string

const (
	SMSStatusPending   SMSStatus = "pending"
	SMSStatusSent      SMSStatus = "sent"
	SMSStatusFailed    SMSStatus = "failed"
	SMSStatusDelivered SMSStatus = "delivered" // If delivery status is available
)

// CreateSMSSettingsRequest represents the request to create SMS settings
type CreateSMSSettingsRequest struct {
	Provider           string  `json:"provider" binding:"required,oneof=twilio aws_sns vonage mock"`
	Enabled            bool    `json:"enabled"`
	AccountSID         *string `json:"account_sid,omitempty"`
	AuthToken          *string `json:"auth_token,omitempty"`
	FromNumber         *string `json:"from_number,omitempty"`
	AWSRegion          *string `json:"aws_region,omitempty"`
	AWSAccessKeyID     *string `json:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey *string `json:"aws_secret_access_key,omitempty"`
	AWSSenderID        *string `json:"aws_sender_id,omitempty"`
	MaxPerHour         *int    `json:"max_per_hour,omitempty"`
	MaxPerDay          *int    `json:"max_per_day,omitempty"`
	MaxPerNumber       *int    `json:"max_per_number,omitempty"`
}

// UpdateSMSSettingsRequest represents the request to update SMS settings
type UpdateSMSSettingsRequest struct {
	Provider           *string `json:"provider,omitempty" binding:"omitempty,oneof=twilio aws_sns vonage mock"`
	Enabled            *bool   `json:"enabled,omitempty"`
	AccountSID         *string `json:"account_sid,omitempty"`
	AuthToken          *string `json:"auth_token,omitempty"`
	FromNumber         *string `json:"from_number,omitempty"`
	AWSRegion          *string `json:"aws_region,omitempty"`
	AWSAccessKeyID     *string `json:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey *string `json:"aws_secret_access_key,omitempty"`
	AWSSenderID        *string `json:"aws_sender_id,omitempty"`
	MaxPerHour         *int    `json:"max_per_hour,omitempty"`
	MaxPerDay          *int    `json:"max_per_day,omitempty"`
	MaxPerNumber       *int    `json:"max_per_number,omitempty"`
}

// SendSMSRequest represents the request to send an SMS
type SendSMSRequest struct {
	// Phone number to send SMS to
	Phone string `json:"phone" binding:"required" example:"+1234567890"`
	// OTP type: verification, password_reset, 2fa, or login
	Type OTPType `json:"type" binding:"required,oneof=verification password_reset 2fa login" example:"verification"`
}

// SendSMSResponse represents the response after sending an SMS
type SendSMSResponse struct {
	// Whether SMS was sent successfully
	Success bool `json:"success" example:"true"`
	// Message ID from SMS provider
	MessageID *string `json:"message_id,omitempty" example:"SM1234567890abcdef"`
	// OTP expiration timestamp
	ExpiresAt time.Time `json:"expires_at" example:"2024-01-15T10:40:00Z"`
}

// VerifySMSOTPRequest represents the request to verify SMS OTP
type VerifySMSOTPRequest struct {
	// Phone number that received the OTP
	Phone string `json:"phone" binding:"required" example:"+1234567890"`
	// 6-digit OTP code
	Code string `json:"code" binding:"required,len=6" example:"123456"`
	// OTP type: verification, password_reset, 2fa, or login
	Type OTPType `json:"type" binding:"required,oneof=verification password_reset 2fa login" example:"verification"`
}

// VerifySMSOTPResponse represents the response after verifying SMS OTP
type VerifySMSOTPResponse struct {
	// Whether the OTP is valid
	Valid bool `json:"valid" example:"true"`
	// Response message
	Message string `json:"message,omitempty" example:"OTP verified successfully"`
	// User information (if verification was for login)
	User *User `json:"user,omitempty"`
}

// SMSStatsResponse represents SMS statistics
type SMSStatsResponse struct {
	TotalSent      int64            `json:"total_sent"`
	TotalFailed    int64            `json:"total_failed"`
	SentToday      int64            `json:"sent_today"`
	SentThisHour   int64            `json:"sent_this_hour"`
	ByType         map[string]int64 `json:"by_type"`
	ByStatus       map[string]int64 `json:"by_status"`
	RecentMessages []SMSLog         `json:"recent_messages"`
}

# SMS Provider Documentation

This document describes the SMS provider system for sending OTP codes and phone verification codes in the Auth Gateway.

## Overview

The SMS provider system allows the Auth Gateway to send one-time passwords (OTP) via SMS for:
- Phone number verification
- Password reset
- Two-factor authentication (2FA)
- Passwordless login

## Features

- **Multiple Providers**: Support for Twilio, AWS SNS, and mock provider for testing
- **Rate Limiting**: Configurable rate limits at multiple levels (per phone, per hour, per day)
- **Logging**: Complete SMS log tracking with status and error messages
- **Settings Management**: Admin panel for managing SMS provider settings
- **Statistics**: Detailed SMS statistics and reporting

## Supported Providers

### 1. Twilio

Twilio is a popular cloud communications platform.

**Configuration:**
```env
SMS_PROVIDER=twilio
SMS_ENABLED=true
TWILIO_ACCOUNT_SID=your_account_sid
TWILIO_AUTH_TOKEN=your_auth_token
TWILIO_FROM_NUMBER=+1234567890
```

**How to get credentials:**
1. Sign up at [twilio.com](https://www.twilio.com)
2. Get your Account SID and Auth Token from the console
3. Purchase a phone number for sending SMS

### 2. AWS SNS (Simple Notification Service)

AWS SNS is Amazon's messaging service.

**Configuration:**
```env
SMS_PROVIDER=aws_sns
SMS_ENABLED=true
AWS_SNS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_SNS_SENDER_ID=YourAppName
```

**How to get credentials:**
1. Create an AWS account
2. Create an IAM user with SNS permissions
3. Get the Access Key ID and Secret Access Key
4. Ensure SMS spending limits are configured in SNS settings

### 3. Mock Provider (Testing)

The mock provider logs SMS messages without actually sending them. Perfect for development and testing.

**Configuration:**
```env
SMS_PROVIDER=mock
SMS_ENABLED=true
```

## Rate Limiting

The SMS system implements multiple layers of rate limiting to prevent abuse:

### Configuration

```env
SMS_MAX_PER_HOUR=10      # Global SMS limit per hour
SMS_MAX_PER_DAY=50       # Global SMS limit per day
SMS_MAX_PER_NUMBER=5     # Max SMS per phone number per hour
```

### Rate Limit Layers

1. **Per Phone Number**: Limits SMS sent to a specific phone number (default: 5 per hour)
2. **Per OTP Type**: Limits same OTP type to same phone (default: 3 per hour)
3. **Global Hourly**: System-wide hourly limit (default: 10 per hour)
4. **Global Daily**: System-wide daily limit (default: 50 per day)

## API Endpoints

### Public Endpoints

#### Send SMS OTP

**POST** `/sms/send`

Send an OTP code via SMS.

**Request:**
```json
{
  "phone": "+1234567890",
  "type": "verification"
}
```

**Response:**
```json
{
  "success": true,
  "message_id": "SM1234567890abcdef",
  "expires_at": "2024-01-15T10:20:00Z"
}
```

**OTP Types:**
- `verification` - Phone number verification
- `password_reset` - Password reset
- `2fa` - Two-factor authentication
- `login` - Passwordless login

#### Verify SMS OTP

**POST** `/sms/verify`

Verify an OTP code sent via SMS.

**Request:**
```json
{
  "phone": "+1234567890",
  "code": "123456",
  "type": "verification"
}
```

**Response:**
```json
{
  "valid": true,
  "message": "OTP verified successfully",
  "user": {
    "id": "uuid",
    "phone": "+1234567890",
    "phone_verified": true
  }
}
```

### Admin Endpoints (Requires Authentication)

#### Get SMS Statistics

**GET** `/sms/stats`

Requires: Admin role

**Response:**
```json
{
  "total_sent": 1234,
  "total_failed": 12,
  "sent_today": 45,
  "sent_this_hour": 5,
  "by_type": {
    "verification": 800,
    "password_reset": 234,
    "2fa": 150,
    "login": 50
  },
  "by_status": {
    "sent": 1234,
    "failed": 12,
    "pending": 2
  },
  "recent_messages": [...]
}
```

#### Manage SMS Settings

**POST** `/admin/sms/settings`

Create new SMS provider settings.

**GET** `/admin/sms/settings`

Get all SMS settings.

**GET** `/admin/sms/settings/active`

Get currently active SMS settings.

**GET** `/admin/sms/settings/{id}`

Get SMS settings by ID.

**PUT** `/admin/sms/settings/{id}`

Update SMS settings.

**DELETE** `/admin/sms/settings/{id}`

Delete SMS settings.

## Database Schema

### OTPs Table (Extended)

```sql
CREATE TABLE otps (
    id UUID PRIMARY KEY,
    email VARCHAR(255),           -- Either email or phone is required
    phone VARCHAR(20),             -- Either email or phone is required
    code VARCHAR(255) NOT NULL,    -- Bcrypt hashed
    type VARCHAR(50) NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### SMS Settings Table

```sql
CREATE TABLE sms_settings (
    id UUID PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    enabled BOOLEAN DEFAULT FALSE,
    account_sid VARCHAR(255),
    auth_token VARCHAR(255),
    from_number VARCHAR(20),
    aws_region VARCHAR(50),
    aws_access_key_id VARCHAR(255),
    aws_secret_access_key VARCHAR(255),
    aws_sender_id VARCHAR(50),
    max_per_hour INTEGER DEFAULT 10,
    max_per_day INTEGER DEFAULT 50,
    max_per_number INTEGER DEFAULT 5,
    created_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### SMS Logs Table

```sql
CREATE TABLE sms_logs (
    id UUID PRIMARY KEY,
    phone VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    message_id VARCHAR(255),
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    sent_at TIMESTAMP,
    user_id UUID,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Usage Examples

### Phone Verification Flow

1. **User signs up with phone number**
```bash
curl -X POST http://localhost:3000/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "phone": "+1234567890",
    "password": "SecurePass123"
  }'
```

2. **Send verification SMS**
```bash
curl -X POST http://localhost:3000/sms/send \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1234567890",
    "type": "verification"
  }'
```

3. **Verify OTP code**
```bash
curl -X POST http://localhost:3000/sms/verify \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1234567890",
    "code": "123456",
    "type": "verification"
  }'
```

### Password Reset Flow

1. **Request password reset**
```bash
curl -X POST http://localhost:3000/sms/send \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1234567890",
    "type": "password_reset"
  }'
```

2. **Verify OTP and reset password**
```bash
curl -X POST http://localhost:3000/sms/verify \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1234567890",
    "code": "123456",
    "type": "password_reset"
  }'
```

## Security Considerations

### OTP Code Storage

- OTP codes are **never** stored in plain text
- All codes are hashed using bcrypt before storage
- Codes expire after 10 minutes
- Used codes are marked and cannot be reused

### Rate Limiting

- Multiple layers of rate limiting prevent abuse
- IP-based and phone-based limits
- Automatic cleanup of old rate limit counters via Redis TTL

### Phone Number Validation

- All phone numbers are validated against E.164 format
- Numbers are normalized before storage (e.g., +1234567890)
- Duplicate phone numbers are rejected

### Sensitive Data

- SMS provider credentials are never exposed in API responses
- Auth tokens and API keys use JSON tag `-` to prevent serialization
- Admin-only endpoints require authentication and authorization

## Troubleshooting

### SMS Not Sending

1. **Check provider configuration**
   - Verify credentials are correct
   - Ensure SMS_ENABLED=true
   - Check provider is set correctly

2. **Check rate limits**
   - Verify rate limits in configuration
   - Check Redis for rate limit counters
   - Review SMS logs for error messages

3. **Check SMS logs**
```bash
curl -X GET http://localhost:3000/sms/stats \
  -H "Authorization: Bearer your-admin-token"
```

### Rate Limit Exceeded

If you're hitting rate limits:

1. **Increase limits in configuration**
```env
SMS_MAX_PER_HOUR=20
SMS_MAX_PER_DAY=100
SMS_MAX_PER_NUMBER=10
```

2. **Clear Redis rate limit keys** (development only)
```bash
redis-cli KEYS "sms:limit:*" | xargs redis-cli DEL
```

### OTP Verification Failing

1. **Check OTP expiration** (10 minutes)
2. **Ensure code hasn't been used already**
3. **Verify phone number format matches**
4. **Check OTP type matches**

## Migration

To add SMS support to an existing deployment:

1. **Run migration**
```bash
migrate -path ./migrations -database "postgresql://user:pass@localhost/db" up
```

2. **Update environment variables**
```bash
cp .env.example .env
# Edit .env with your SMS provider credentials
```

3. **Restart the application**
```bash
./auth-gateway
```

## Best Practices

1. **Use Twilio or AWS SNS in production** - Mock provider is for testing only
2. **Set appropriate rate limits** - Balance security with user experience
3. **Monitor SMS costs** - Track usage via provider dashboard
4. **Enable SMS only when needed** - Keep SMS_ENABLED=false if not using
5. **Rotate credentials regularly** - Update provider credentials periodically
6. **Monitor SMS logs** - Review failed messages and error patterns
7. **Clean up old logs** - Implement periodic cleanup of SMS logs (>30 days)

## Cost Optimization

### Twilio
- Use alphanumeric sender ID where supported (cheaper)
- Monitor usage dashboard
- Set up spending limits

### AWS SNS
- Use transactional SMS type for OTP (higher reliability)
- Monitor CloudWatch metrics
- Set SNS spending limits

### General
- Implement aggressive rate limiting
- Add CAPTCHA before SMS sending
- Use email as primary, SMS as fallback
- Cache verification status to reduce redundant sends

## Support

For issues or questions:
- Check the troubleshooting section above
- Review SMS logs at `/sms/stats`
- Check provider dashboard for delivery status
- Contact support with message_id for provider-specific issues

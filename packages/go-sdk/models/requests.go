package models

import "time"

// SignUpRequest contains data for user registration.
type SignUpRequest struct {
	Email       string  `json:"email"`
	Phone       *string `json:"phone,omitempty"`
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	FullName    string  `json:"full_name,omitempty"`
	AccountType string  `json:"account_type,omitempty"` // "human" or "service"
}

// SignInRequest contains data for user login.
type SignInRequest struct {
	Email    string  `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Password string  `json:"password"`
}

// RefreshTokenRequest contains the refresh token for token renewal.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ChangePasswordRequest contains data for password change.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ForgotPasswordRequest initiates password reset.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest completes password reset.
type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// UpdateProfileRequest contains data for profile update.
type UpdateProfileRequest struct {
	FullName          string `json:"full_name,omitempty"`
	ProfilePictureURL string `json:"profile_picture_url,omitempty"`
}

// SendOTPRequest initiates OTP sending.
type SendOTPRequest struct {
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
	Type  string  `json:"type"` // "verification", "password_reset", "passwordless", "2fa"
}

// VerifyOTPRequest verifies an OTP code.
type VerifyOTPRequest struct {
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
	Code  string  `json:"code"`
}

// VerifyEmailRequest verifies email with OTP.
type VerifyEmailRequest struct {
	Code string `json:"code"`
}

// TwoFactorLoginVerifyRequest verifies 2FA during login.
type TwoFactorLoginVerifyRequest struct {
	TwoFactorToken string `json:"two_factor_token"`
	Code           string `json:"code"`
}

// VerifyTwoFARequest verifies and enables 2FA.
type VerifyTwoFARequest struct {
	Code string `json:"code"`
}

// DisableTwoFARequest disables 2FA.
type DisableTwoFARequest struct {
	Password string `json:"password"`
}

// CreateAPIKeyRequest creates a new API key.
type CreateAPIKeyRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Scopes      []string   `json:"scopes"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// UpdateAPIKeyRequest updates an API key.
type UpdateAPIKeyRequest struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

// PasswordlessRequest initiates passwordless login.
type PasswordlessRequest struct {
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
}

// PasswordlessVerifyRequest completes passwordless login.
type PasswordlessVerifyRequest struct {
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
	Code  string  `json:"code"`
}

// InitPasswordlessRegistrationRequest initiates passwordless registration (two-step signup).
type InitPasswordlessRegistrationRequest struct {
	Email    *string `json:"email,omitempty"`    // Optional if phone is provided
	Phone    *string `json:"phone,omitempty"`    // Optional if email is provided
	Username string  `json:"username,omitempty"` // Optional, auto-generated if not provided
	FullName string  `json:"full_name,omitempty"`
}

// CompletePasswordlessRegistrationRequest completes passwordless registration with OTP.
type CompletePasswordlessRegistrationRequest struct {
	Email *string `json:"email,omitempty"` // Must match init request
	Phone *string `json:"phone,omitempty"` // Must match init request
	Code  string  `json:"code"`            // 6-digit OTP code
}

// CreateUserRequest is for admin user creation.
type CreateUserRequest struct {
	Email       string  `json:"email"`
	Phone       *string `json:"phone,omitempty"`
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	FullName    string  `json:"full_name,omitempty"`
	AccountType string  `json:"account_type,omitempty"`
}

// UpdateUserRequest is for admin user update.
type UpdateUserRequest struct {
	Email             *string `json:"email,omitempty"`
	Username          *string `json:"username,omitempty"`
	FullName          *string `json:"full_name,omitempty"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty"`
	IsActive          *bool   `json:"is_active,omitempty"`
}

// CreateRoleRequest creates a new role.
type CreateRoleRequest struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"` // Permission IDs
}

// UpdateRoleRequest updates a role.
type UpdateRoleRequest struct {
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"` // Permission IDs
}

// CreatePermissionRequest creates a new permission.
type CreatePermissionRequest struct {
	Name        string `json:"name"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description,omitempty"`
}

// AssignRoleRequest assigns a role to a user.
type AssignRoleRequest struct {
	RoleID string `json:"role_id"`
}

// CreateIPFilterRequest creates an IP filter.
type CreateIPFilterRequest struct {
	IPAddress   string `json:"ip_address"`
	Type        string `json:"type"` // "whitelist" or "blacklist"
	Description string `json:"description,omitempty"`
}

// UpdateBrandingRequest updates branding settings.
type UpdateBrandingRequest struct {
	LogoURL        string `json:"logo_url,omitempty"`
	PrimaryColor   string `json:"primary_color,omitempty"`
	SecondaryColor string `json:"secondary_color,omitempty"`
	CompanyName    string `json:"company_name,omitempty"`
	SupportEmail   string `json:"support_email,omitempty"`
	TermsURL       string `json:"terms_url,omitempty"`
	PrivacyURL     string `json:"privacy_url,omitempty"`
}

// MaintenanceModeRequest enables/disables maintenance mode.
type MaintenanceModeRequest struct {
	Enabled bool   `json:"enabled"`
	Message string `json:"message,omitempty"`
}

// ListUsersParams contains parameters for listing users.
type ListUsersParams struct {
	Page     int    `url:"page,omitempty"`
	Limit    int    `url:"limit,omitempty"`
	Search   string `url:"search,omitempty"`
	IsActive *bool  `url:"is_active,omitempty"`
	Role     string `url:"role,omitempty"`
}

// ListAuditLogsParams contains parameters for listing audit logs.
type ListAuditLogsParams struct {
	Page     int    `url:"page,omitempty"`
	Limit    int    `url:"limit,omitempty"`
	UserID   string `url:"user_id,omitempty"`
	Action   string `url:"action,omitempty"`
	Resource string `url:"resource,omitempty"`
	Status   string `url:"status,omitempty"`
	From     string `url:"from,omitempty"`
	To       string `url:"to,omitempty"`
}

// ListSessionsParams contains parameters for listing sessions.
type ListSessionsParams struct {
	Page   int    `url:"page,omitempty"`
	Limit  int    `url:"limit,omitempty"`
	UserID string `url:"user_id,omitempty"`
}

// CreateAppOAuthProviderRequest creates a per-app OAuth provider
type CreateAppOAuthProviderRequest struct {
	Provider     string   `json:"provider"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	CallbackURL  string   `json:"callback_url"`
	Scopes       []string `json:"scopes,omitempty"`
	AuthURL      string   `json:"auth_url,omitempty"`
	TokenURL     string   `json:"token_url,omitempty"`
	UserInfoURL  string   `json:"user_info_url,omitempty"`
	IsActive     *bool    `json:"is_active,omitempty"`
}

// UpdateAppOAuthProviderRequest updates a per-app OAuth provider
type UpdateAppOAuthProviderRequest struct {
	ClientID     *string  `json:"client_id,omitempty"`
	ClientSecret *string  `json:"client_secret,omitempty"`
	CallbackURL  *string  `json:"callback_url,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	AuthURL      *string  `json:"auth_url,omitempty"`
	TokenURL     *string  `json:"token_url,omitempty"`
	UserInfoURL  *string  `json:"user_info_url,omitempty"`
	IsActive     *bool    `json:"is_active,omitempty"`
}

// CreateTelegramBotRequest creates a Telegram bot for an app
type CreateTelegramBotRequest struct {
	BotToken    string `json:"bot_token"`
	BotUsername string `json:"bot_username"`
	DisplayName string `json:"display_name"`
	IsAuthBot   *bool  `json:"is_auth_bot,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

// UpdateTelegramBotRequest updates a Telegram bot
type UpdateTelegramBotRequest struct {
	BotToken    *string `json:"bot_token,omitempty"`
	BotUsername *string `json:"bot_username,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	IsAuthBot   *bool   `json:"is_auth_bot,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

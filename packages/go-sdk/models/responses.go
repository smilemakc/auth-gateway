package models

// AuthResponse contains authentication tokens and user info.
type AuthResponse struct {
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
	User           *User  `json:"user,omitempty"`
	ExpiresIn      int64  `json:"expires_in"`
	Requires2FA    bool   `json:"requires_2fa"`
	TwoFactorToken string `json:"two_factor_token,omitempty"`
}

// TokenResponse contains only tokens.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// TwoFASetupResponse contains 2FA setup information.
type TwoFASetupResponse struct {
	Secret      string   `json:"secret"`
	QRCode      string   `json:"qr_code"` // base64 encoded PNG
	BackupCodes []string `json:"backup_codes"`
}

// TwoFAStatusResponse contains 2FA status.
type TwoFAStatusResponse struct {
	Enabled              bool   `json:"enabled"`
	EnabledAt            string `json:"enabled_at,omitempty"`
	BackupCodesRemaining int    `json:"backup_codes_remaining"`
}

// BackupCodesResponse contains regenerated backup codes.
type BackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
}

// CreateAPIKeyResponse contains created API key with plain key.
type CreateAPIKeyResponse struct {
	APIKey   *APIKey `json:"api_key"`
	PlainKey string  `json:"plain_key"` // Only returned once at creation
}

// MessageResponse contains a simple message.
type MessageResponse struct {
	Message string `json:"message"`
}

// OTPResponse contains OTP sending result.
type OTPResponse struct {
	Message   string `json:"message"`
	ExpiresIn int    `json:"expires_in"`
}

// InitPasswordlessRegistrationResponse contains the result of passwordless registration initiation.
type InitPasswordlessRegistrationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PermissionMatrixResponse contains the permission matrix for UI.
type PermissionMatrixResponse struct {
	Roles       []Role              `json:"roles"`
	Permissions []Permission        `json:"permissions"`
	Matrix      map[string][]string `json:"matrix"` // role_id -> []permission_id
}

// ErrorResponse represents an API error.
type ErrorResponse struct {
	Error      string            `json:"error"`
	Message    string            `json:"message"`
	Code       string            `json:"code,omitempty"`
	Details    map[string]string `json:"details,omitempty"`
	StatusCode int               `json:"-"`
}

func (e *ErrorResponse) String() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Error
}

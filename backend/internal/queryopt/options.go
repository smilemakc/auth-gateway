package queryopt

import "github.com/google/uuid"

// --- UserStore options ---

// UserGetOptions holds options for UserStore Get methods.
type UserGetOptions struct {
	WithRoles bool
}

// UserGetOption configures UserGetOptions.
type UserGetOption func(*UserGetOptions)

// UserGetWithRoles loads user roles via Relation("Roles").
func UserGetWithRoles() UserGetOption {
	return func(o *UserGetOptions) { o.WithRoles = true }
}

func BuildUserGetOptions(opts []UserGetOption) UserGetOptions {
	var o UserGetOptions
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// UserListOptions holds options for UserStore List method.
type UserListOptions struct {
	Limit     int
	Offset    int
	IsActive  *bool
	AppID     *uuid.UUID
	WithRoles bool
}

// UserListOption configures UserListOptions.
type UserListOption func(*UserListOptions)

// UserListWithRoles loads user roles via Relation("Roles").
func UserListWithRoles() UserListOption {
	return func(o *UserListOptions) { o.WithRoles = true }
}

// UserListLimit sets the limit for List queries.
func UserListLimit(n int) UserListOption {
	return func(o *UserListOptions) { o.Limit = n }
}

// UserListOffset sets the offset for List queries.
func UserListOffset(n int) UserListOption {
	return func(o *UserListOptions) { o.Offset = n }
}

// UserListActive filters by is_active status.
func UserListActive(v *bool) UserListOption {
	return func(o *UserListOptions) { o.IsActive = v }
}

// UserListAppID filters by application ID.
func UserListAppID(id uuid.UUID) UserListOption {
	return func(o *UserListOptions) { o.AppID = &id }
}

func BuildUserListOptions(opts []UserListOption) UserListOptions {
	var o UserListOptions
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// --- OAuthClientStore options ---

// OAuthClientListOptions holds options for OAuthProviderStore ListClients method.
type OAuthClientListOptions struct {
	IsActive *bool
	OwnerID  *uuid.UUID
}

// OAuthClientListOption configures OAuthClientListOptions.
type OAuthClientListOption func(*OAuthClientListOptions)

// OAuthClientListActive filters by is_active status.
func OAuthClientListActive(v *bool) OAuthClientListOption {
	return func(o *OAuthClientListOptions) { o.IsActive = v }
}

// OAuthClientListOwner filters by owner ID.
func OAuthClientListOwner(id uuid.UUID) OAuthClientListOption {
	return func(o *OAuthClientListOptions) { o.OwnerID = &id }
}

func BuildOAuthClientListOptions(opts []OAuthClientListOption) OAuthClientListOptions {
	var o OAuthClientListOptions
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// --- APIKeyStore options ---

// APIKeyGetOptions holds options for APIKeyStore Get/Count methods.
type APIKeyGetOptions struct {
	ActiveOnly bool
}

// APIKeyGetOption configures APIKeyGetOptions.
type APIKeyGetOption func(*APIKeyGetOptions)

// APIKeyActiveOnly filters for active (non-revoked, non-expired) keys.
func APIKeyActiveOnly() APIKeyGetOption {
	return func(o *APIKeyGetOptions) { o.ActiveOnly = true }
}

func BuildAPIKeyGetOptions(opts []APIKeyGetOption) APIKeyGetOptions {
	var o APIKeyGetOptions
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

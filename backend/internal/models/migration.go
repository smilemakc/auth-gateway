package models

type ImportUsersRequest struct {
	ApplicationID string            `json:"application_id" binding:"required"`
	Users         []ImportUserEntry `json:"users" binding:"required,min=1"`
}

type ImportUserEntry struct {
	Email      string         `json:"email" binding:"required,email"`
	Username   string         `json:"username"`
	FullName   string         `json:"full_name"`
	Password   string         `json:"password"`
	HashFormat string         `json:"hash_format"`
	Roles      []string       `json:"roles"`
	IsActive   *bool          `json:"is_active"`
	Metadata   map[string]any `json:"metadata"`
}

type ImportOAuthAccountsRequest struct {
	Accounts []ImportOAuthEntry `json:"accounts" binding:"required,min=1"`
}

type ImportOAuthEntry struct {
	Email          string `json:"email" binding:"required,email"`
	Provider       string `json:"provider" binding:"required"`
	ProviderUserID string `json:"provider_user_id" binding:"required"`
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
}

type ImportRolesRequest struct {
	ApplicationID string            `json:"application_id" binding:"required"`
	Roles         []ImportRoleEntry `json:"roles" binding:"required,min=1"`
}

type ImportRoleEntry struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	AssignTo    []string `json:"assign_to"`
}

type ImportResult struct {
	Total   int      `json:"total"`
	Created int      `json:"created"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}

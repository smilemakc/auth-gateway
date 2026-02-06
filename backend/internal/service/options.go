package service

import "github.com/smilemakc/auth-gateway/internal/queryopt"

// Re-export option types from queryopt to avoid import cycles.
// The queryopt package is imported by both service and repository.

// --- UserStore options ---

type UserGetOptions = queryopt.UserGetOptions
type UserGetOption = queryopt.UserGetOption

var (
	UserGetWithRoles    = queryopt.UserGetWithRoles
	BuildUserGetOptions = queryopt.BuildUserGetOptions
)

type UserListOptions = queryopt.UserListOptions
type UserListOption = queryopt.UserListOption

var (
	UserListWithRoles    = queryopt.UserListWithRoles
	UserListLimit        = queryopt.UserListLimit
	UserListOffset       = queryopt.UserListOffset
	UserListActive       = queryopt.UserListActive
	UserListAppID        = queryopt.UserListAppID
	BuildUserListOptions = queryopt.BuildUserListOptions
)

// --- APIKeyStore options ---

type APIKeyGetOptions = queryopt.APIKeyGetOptions
type APIKeyGetOption = queryopt.APIKeyGetOption

var (
	APIKeyActiveOnly     = queryopt.APIKeyActiveOnly
	BuildAPIKeyGetOptions = queryopt.BuildAPIKeyGetOptions
)

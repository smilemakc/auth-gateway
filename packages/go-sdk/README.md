# Auth Gateway Go SDK

A comprehensive Go SDK for the Auth Gateway authentication and authorization service.

## Installation

```bash
go get github.com/smilemakc/auth-gateway/packages/go-sdk
```

## Quick Start

### REST API Client

```go
package main

import (
    "context"
    "fmt"
    "log"

    authgateway "github.com/smilemakc/auth-gateway/packages/go-sdk"
    "github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

func main() {
    // Create a client
    client := authgateway.NewClient(authgateway.Config{
        BaseURL:     "http://localhost:3000",
        AutoRefresh: true,
    })

    ctx := context.Background()

    // Sign in
    resp, err := client.Auth.SignInWithEmail(ctx, "user@example.com", "password")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Signed in as: %s\n", resp.User.Email)

    // Get profile
    user, err := client.Profile.Get(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Hello, %s!\n", user.FullName)
}
```

### gRPC Client (Server-to-Server)

```go
package main

import (
    "context"
    "fmt"
    "log"

    authgateway "github.com/smilemakc/auth-gateway/packages/go-sdk"
)

func main() {
    // Create gRPC client
    grpcClient, err := authgateway.NewGRPCClient(authgateway.GRPCConfig{
        Address:  "localhost:50051",
        Insecure: true, // false for production
    })
    if err != nil {
        log.Fatal(err)
    }
    defer grpcClient.Close()

    ctx := context.Background()

    // Validate token
    resp, err := grpcClient.ValidateToken(ctx, "jwt_token_here")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Token valid: %t, User: %s\n", resp.Valid, resp.Email)
}
```

### API Key Authentication

```go
client := authgateway.NewClient(authgateway.Config{
    BaseURL: "http://localhost:3000",
    APIKey:  "agw_your_api_key_here",
})
```

## Features

### Authentication
- Email/password login
- Phone/password login
- OAuth (Google, GitHub, Yandex, Instagram, Telegram)
- Passwordless login (OTP via email/SMS)
- Two-factor authentication (TOTP)
- Automatic token refresh

### User Management
- User registration
- Profile management
- Password change/reset
- Session management

### API Keys
- Create/list/update/revoke API keys
- Scope-based permissions

### Admin Operations
- User management
- Role-based access control (RBAC)
- Audit logs
- IP filtering
- System statistics
- Session management

### gRPC API
- Token validation
- User lookup
- Permission checking
- Token introspection

## Services

### Auth Service
```go
client.Auth.SignUp(ctx, &models.SignUpRequest{...})
client.Auth.SignIn(ctx, &models.SignInRequest{...})
client.Auth.SignInWithEmail(ctx, email, password)
client.Auth.Verify2FA(ctx, twoFactorToken, code)
client.Auth.RefreshTokens(ctx)
client.Auth.Logout(ctx)
client.Auth.RequestPasswordReset(ctx, email)
client.Auth.ResetPassword(ctx, email, code, newPassword)
client.Auth.VerifyEmail(ctx, code)
```

### Profile Service
```go
client.Profile.Get(ctx)
client.Profile.Update(ctx, &models.UpdateProfileRequest{...})
client.Profile.ChangePassword(ctx, oldPassword, newPassword)
```

### Two-Factor Service
```go
client.TwoFactor.Setup(ctx)
client.TwoFactor.Verify(ctx, code)
client.TwoFactor.Disable(ctx, password)
client.TwoFactor.Status(ctx)
client.TwoFactor.RegenerateBackupCodes(ctx)
```

### API Keys Service
```go
client.APIKeys.Create(ctx, &models.CreateAPIKeyRequest{...})
client.APIKeys.List(ctx)
client.APIKeys.Get(ctx, id)
client.APIKeys.Update(ctx, id, &models.UpdateAPIKeyRequest{...})
client.APIKeys.Revoke(ctx, id)
client.APIKeys.Delete(ctx, id)
```

### Sessions Service
```go
client.Sessions.List(ctx)
client.Sessions.Revoke(ctx, sessionID)
client.Sessions.RevokeAll(ctx)
```

### OTP Service
```go
client.OTP.SendToEmail(ctx, email, otpType)
client.OTP.SendToPhone(ctx, phone, otpType)
client.OTP.VerifyEmail(ctx, email, code)
client.OTP.VerifyPhone(ctx, phone, code)
```

### Passwordless Service
```go
client.Passwordless.RequestWithEmail(ctx, email)
client.Passwordless.RequestWithPhone(ctx, phone)
client.Passwordless.VerifyWithEmail(ctx, email, code)
client.Passwordless.VerifyWithPhone(ctx, phone, code)
```

### OAuth Service
```go
client.OAuth.GetProviders(ctx)
client.OAuth.GetAuthURL(provider)
```

### Admin Service
```go
// Statistics
client.Admin.GetStats(ctx)
client.Admin.GetSessionStats(ctx)
client.Admin.GetGeoDistribution(ctx)

// User Management
client.Admin.ListUsers(ctx, &models.ListUsersParams{...})
client.Admin.CreateUser(ctx, &models.CreateUserRequest{...})
client.Admin.GetUser(ctx, userID)
client.Admin.UpdateUser(ctx, userID, &models.UpdateUserRequest{...})
client.Admin.DeleteUser(ctx, userID)
client.Admin.AssignRole(ctx, userID, roleID)
client.Admin.RemoveRole(ctx, userID, roleID)

// RBAC
client.Admin.ListRoles(ctx)
client.Admin.CreateRole(ctx, &models.CreateRoleRequest{...})
client.Admin.GetRole(ctx, roleID)
client.Admin.UpdateRole(ctx, roleID, &models.UpdateRoleRequest{...})
client.Admin.DeleteRole(ctx, roleID)
client.Admin.ListPermissions(ctx)
client.Admin.CreatePermission(ctx, &models.CreatePermissionRequest{...})

// Audit Logs
client.Admin.ListAuditLogs(ctx, &models.ListAuditLogsParams{...})

// IP Filters
client.Admin.ListIPFilters(ctx)
client.Admin.CreateIPFilter(ctx, &models.CreateIPFilterRequest{...})
client.Admin.DeleteIPFilter(ctx, filterID)

// System
client.Admin.SetMaintenanceMode(ctx, &models.MaintenanceModeRequest{...})
client.Admin.GetSystemHealth(ctx)
```

## Error Handling

```go
resp, err := client.Auth.SignIn(ctx, req)
if err != nil {
    // Check for 2FA requirement
    if tfaErr, ok := err.(*authgateway.TwoFactorRequiredError); ok {
        // Handle 2FA
        resp, err = client.Auth.Verify2FA(ctx, tfaErr.TwoFactorToken, userCode)
    }

    // Check for API errors
    if apiErr, ok := err.(*authgateway.APIError); ok {
        fmt.Printf("Error: %s (code: %s, status: %d)\n",
            apiErr.Message, apiErr.Code, apiErr.StatusCode)

        if apiErr.IsUnauthorized() {
            // Handle unauthorized
        }
        if apiErr.IsTooManyRequests() {
            // Handle rate limiting
        }
    }

    // Check for network errors
    if netErr, ok := err.(*authgateway.NetworkError); ok {
        fmt.Printf("Network error: %v\n", netErr.Unwrap())
    }
}
```

## Configuration Options

```go
client := authgateway.NewClient(authgateway.Config{
    // Base URL of the Auth Gateway server
    BaseURL: "http://localhost:3000",

    // Custom HTTP client (optional)
    HTTPClient: &http.Client{},

    // Request timeout (default: 30s)
    Timeout: 30 * time.Second,

    // API key for authentication (alternative to JWT)
    APIKey: "agw_your_api_key",

    // Auto-refresh tokens when they expire (default: false)
    AutoRefresh: true,
})
```

## Token Management

```go
// Set tokens manually
client.SetTokens(accessToken, refreshToken, expiresIn)

// Get current tokens
accessToken := client.GetAccessToken()
refreshToken := client.GetRefreshToken()

// Check authentication state
if client.IsAuthenticated() {
    // ...
}

// Check if token is expired
if client.IsTokenExpired() {
    // Token will be auto-refreshed if AutoRefresh is enabled
}

// Clear tokens (logout)
client.ClearTokens()
```

## Examples

See the `examples/` directory for complete examples:

- `examples/basic/` - Basic authentication flow
- `examples/apikey/` - API key authentication
- `examples/grpc/` - gRPC client usage
- `examples/admin/` - Admin operations

## License

MIT

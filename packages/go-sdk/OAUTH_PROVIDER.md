# OAuth Provider Client

This guide shows how to use Auth Gateway as an OAuth 2.0 / OpenID Connect provider for your Go applications.

## Overview

The `OAuthProviderClient` allows your application to integrate with Auth Gateway as an identity provider, supporting:

- **Authorization Code Flow** with PKCE (recommended for web/mobile apps)
- **Device Authorization Flow** for devices with limited input (TVs, IoT devices)
- **Client Credentials Flow** for service-to-service authentication
- **Token introspection and revocation** (RFC 7662, RFC 7009)
- **OIDC UserInfo endpoint**
- **Automatic OIDC discovery** (RFC 8414)

## Installation

```bash
go get github.com/smilemakc/auth-gateway/packages/go-sdk
```

## Quick Start

```go
import (
    authgateway "github.com/smilemakc/auth-gateway/packages/go-sdk"
)

// Create OAuth provider client
client := authgateway.NewOAuthProviderClient(authgateway.OAuthProviderConfig{
    Issuer:       "https://auth.example.com",
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    RedirectURI:  "https://yourapp.com/callback",
    Scopes:       []string{"openid", "profile", "email"},
    UsePKCE:      true, // Enabled by default
})
```

## Authorization Code Flow (with PKCE)

This is the **recommended flow** for web and mobile applications. PKCE is enabled by default for security.

### Step 1: Generate Authorization URL

```go
authURL, err := client.GetAuthorizationURL(ctx, &authgateway.AuthorizationURLOptions{
    Prompt: "consent", // Optional: "none", "login", "consent", "select_account"
})
if err != nil {
    log.Fatal(err)
}

// Store these for later verification
state := authURL.State
nonce := authURL.Nonce
codeVerifier := authURL.CodeVerifier

// Redirect user to authURL.URL
fmt.Printf("Visit: %s\n", authURL.URL)
```

### Step 2: Handle Callback

After user authorizes, they'll be redirected to your `redirect_uri` with a `code` parameter:

```
https://yourapp.com/callback?code=AUTH_CODE&state=STATE
```

### Step 3: Exchange Code for Tokens

```go
// Verify state matches what you stored
if receivedState != state {
    // Handle CSRF attack
}

tokens, err := client.ExchangeCode(ctx, code, codeVerifier)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Access Token: %s\n", tokens.AccessToken)
fmt.Printf("Refresh Token: %s\n", tokens.RefreshToken)
fmt.Printf("ID Token: %s\n", tokens.IDToken)
fmt.Printf("Expires in: %d seconds\n", tokens.ExpiresIn)
```

### Step 4: Get User Information

```go
userInfo, err := client.GetUserInfo(ctx, tokens.AccessToken)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("User ID: %s\n", userInfo.Sub)
fmt.Printf("Email: %s (verified: %t)\n", userInfo.Email, userInfo.EmailVerified)
fmt.Printf("Name: %s\n", userInfo.Name)
```

## Device Authorization Flow

Perfect for devices with limited input (smart TVs, CLI tools, IoT devices).

```go
// Step 1: Request device code
deviceResp, err := client.RequestDeviceCode(ctx, []string{"openid", "profile"})
if err != nil {
    log.Fatal(err)
}

// Step 2: Show user the verification URL and code
fmt.Printf("Visit: %s\n", deviceResp.VerificationURI)
fmt.Printf("Enter code: %s\n", deviceResp.UserCode)

// Or use the complete URL (includes code)
fmt.Printf("Or visit: %s\n", deviceResp.VerificationURIComplete)

// Step 3: Poll for authorization
interval := deviceResp.Interval
if interval == 0 {
    interval = 5 // Default to 5 seconds
}

ticker := time.NewTicker(time.Duration(interval) * time.Second)
defer ticker.Stop()

for {
    <-ticker.C

    tokens, err := client.PollDeviceToken(ctx, deviceResp.DeviceCode)
    if err != nil {
        switch err {
        case authgateway.ErrAuthorizationPending:
            continue // Keep polling
        case authgateway.ErrSlowDown:
            // Increase polling interval
            interval += 5
            ticker.Reset(time.Duration(interval) * time.Second)
            continue
        case authgateway.ErrAccessDenied:
            fmt.Println("User denied authorization")
            return
        case authgateway.ErrExpiredToken:
            fmt.Println("Device code expired")
            return
        default:
            log.Fatal(err)
        }
    }

    // Success! User authorized
    fmt.Printf("Access Token: %s\n", tokens.AccessToken)
    break
}
```

## Client Credentials Flow

For **service-to-service** authentication (machine-to-machine). Requires a confidential client with `client_secret`.

```go
tokens, err := client.ClientCredentialsGrant(ctx, []string{"api:read", "api:write"})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Access Token: %s\n", tokens.AccessToken)
fmt.Printf("Scope: %s\n", tokens.Scope)
```

## Token Management

### Refresh Access Token

```go
newTokens, err := client.RefreshTokens(ctx, refreshToken)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("New Access Token: %s\n", newTokens.AccessToken)
```

### Introspect Token (RFC 7662)

Check if a token is valid and get its metadata:

```go
result, err := client.IntrospectToken(ctx, accessToken)
if err != nil {
    log.Fatal(err)
}

if result.Active {
    fmt.Printf("Token is active\n")
    fmt.Printf("Username: %s\n", result.Username)
    fmt.Printf("Scope: %s\n", result.Scope)
    fmt.Printf("Expires at: %d\n", result.Exp)
} else {
    fmt.Println("Token is inactive/invalid")
}
```

### Revoke Token (RFC 7009)

```go
// Revoke access token
err := client.RevokeToken(ctx, accessToken, "access_token")

// Revoke refresh token
err = client.RevokeToken(ctx, refreshToken, "refresh_token")
```

## OIDC Discovery

The client automatically discovers endpoints via the OIDC discovery document:

```go
discovery, err := client.GetDiscovery(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Issuer: %s\n", discovery.Issuer)
fmt.Printf("Authorization Endpoint: %s\n", discovery.AuthorizationEndpoint)
fmt.Printf("Token Endpoint: %s\n", discovery.TokenEndpoint)
fmt.Printf("Supported Scopes: %v\n", discovery.ScopesSupported)
```

### Get JWKS (JSON Web Key Set)

Useful for validating ID tokens:

```go
jwks, err := client.GetJWKS(ctx)
if err != nil {
    log.Fatal(err)
}

for _, key := range jwks.Keys {
    fmt.Printf("Key ID: %s, Type: %s, Use: %s\n", key.Kid, key.Kty, key.Use)
}
```

## Security Best Practices

### 1. Always Use PKCE

PKCE is enabled by default. Never disable it for public clients:

```go
config := authgateway.OAuthProviderConfig{
    // ...
    UsePKCE: true, // Default, recommended
}
```

### 2. Validate State Parameter

Always verify the `state` parameter to prevent CSRF attacks:

```go
authURL, _ := client.GetAuthorizationURL(ctx, nil)

// Store state in session
session.Set("oauth_state", authURL.State)

// On callback, verify it matches
if receivedState != session.Get("oauth_state") {
    // Reject: possible CSRF attack
}
```

### 3. Validate Nonce in ID Token

The `nonce` should be validated when decoding the ID token to prevent replay attacks.

### 4. Use HTTPS in Production

Always use HTTPS for `Issuer` and `RedirectURI` in production:

```go
config := authgateway.OAuthProviderConfig{
    Issuer:      "https://auth.example.com", // Not http://
    RedirectURI: "https://app.example.com/callback",
}
```

### 5. Store Client Secret Securely

Never commit client secrets to version control. Use environment variables:

```go
config := authgateway.OAuthProviderConfig{
    ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
}
```

### 6. Validate Redirect URIs

On the Auth Gateway side, ensure only whitelisted redirect URIs are allowed for each client.

## Configuration Options

```go
type OAuthProviderConfig struct {
    // Required: Auth Gateway server URL
    Issuer string

    // Required: OAuth client ID from Auth Gateway
    ClientID string

    // Required for confidential clients (web apps, backend services)
    // Optional for public clients (SPAs, mobile apps)
    ClientSecret string

    // Required: Where Auth Gateway redirects after authorization
    RedirectURI string

    // Optional: Default scopes to request
    // Default: ["openid"]
    Scopes []string

    // Optional: Enable PKCE (Proof Key for Code Exchange)
    // Default: true (recommended)
    UsePKCE bool

    // Optional: Custom HTTP client
    // Default: 30 second timeout
    HTTPClient *http.Client
}
```

## Available Scopes

Standard OIDC scopes:
- `openid` - Required for OIDC
- `profile` - User's profile information (name, username, picture)
- `email` - User's email and email_verified
- `phone` - User's phone number

Custom scopes (defined in Auth Gateway):
- `offline_access` - Enables refresh tokens
- Your custom scopes

## Error Handling

```go
tokens, err := client.ExchangeCode(ctx, code, verifier)
if err != nil {
    switch {
    case errors.Is(err, authgateway.ErrAuthorizationPending):
        // Device flow: user hasn't authorized yet
    case errors.Is(err, authgateway.ErrSlowDown):
        // Device flow: polling too fast
    case errors.Is(err, authgateway.ErrAccessDenied):
        // User denied authorization
    case errors.Is(err, authgateway.ErrExpiredToken):
        // Token/code expired
    default:
        // Other error
        log.Printf("OAuth error: %v", err)
    }
}
```

## Complete Example: Web Application

```go
package main

import (
    "context"
    "fmt"
    "net/http"

    authgateway "github.com/smilemakc/auth-gateway/packages/go-sdk"
)

var oauthClient *authgateway.OAuthProviderClient

func init() {
    oauthClient = authgateway.NewOAuthProviderClient(authgateway.OAuthProviderConfig{
        Issuer:       "https://auth.example.com",
        ClientID:     "webapp-client-id",
        ClientSecret: "webapp-client-secret",
        RedirectURI:  "https://webapp.example.com/callback",
        Scopes:       []string{"openid", "profile", "email"},
    })
}

func main() {
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/callback", callbackHandler)
    http.ListenAndServe(":8080", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    authURL, err := oauthClient.GetAuthorizationURL(ctx, nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Store state and nonce in session
    session, _ := store.Get(r, "session")
    session.Values["state"] = authURL.State
    session.Values["nonce"] = authURL.Nonce
    session.Values["code_verifier"] = authURL.CodeVerifier
    session.Save(r, w)

    http.Redirect(w, r, authURL.URL, http.StatusFound)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Verify state
    session, _ := store.Get(r, "session")
    if r.URL.Query().Get("state") != session.Values["state"] {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }

    code := r.URL.Query().Get("code")
    codeVerifier := session.Values["code_verifier"].(string)

    tokens, err := oauthClient.ExchangeCode(ctx, code, codeVerifier)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Store tokens in session
    session.Values["access_token"] = tokens.AccessToken
    session.Values["refresh_token"] = tokens.RefreshToken
    session.Save(r, w)

    // Get user info
    userInfo, _ := oauthClient.GetUserInfo(ctx, tokens.AccessToken)

    fmt.Fprintf(w, "Welcome, %s!", userInfo.Name)
}
```

## Testing

For testing, you can point to a local Auth Gateway instance:

```go
client := authgateway.NewOAuthProviderClient(authgateway.OAuthProviderConfig{
    Issuer:      "http://localhost:8811",
    ClientID:    "test-client",
    RedirectURI: "http://localhost:8080/callback",
})
```

## Resources

- [OAuth 2.0 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [PKCE RFC 7636](https://tools.ietf.org/html/rfc7636)
- [Device Flow RFC 8628](https://tools.ietf.org/html/rfc8628)
- [Token Introspection RFC 7662](https://tools.ietf.org/html/rfc7662)
- [Token Revocation RFC 7009](https://tools.ietf.org/html/rfc7009)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)

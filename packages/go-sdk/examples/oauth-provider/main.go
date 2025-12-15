// Example: Using Auth Gateway as an OAuth/OIDC provider for your application
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	authgateway "github.com/smilemakc/auth-gateway/packages/go-sdk"
)

func main() {
	ctx := context.Background()

	oauthClient := authgateway.NewOAuthProviderClient(authgateway.OAuthProviderConfig{
		Issuer:       "http://localhost:8811",
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		RedirectURI:  "http://localhost:8811/callback",
		Scopes:       []string{"openid", "profile", "email"},
		UsePKCE:      true,
	})

	exampleAuthorizationCodeFlow(ctx, oauthClient)

	exampleDeviceFlow(ctx, oauthClient)

	exampleClientCredentials(ctx, oauthClient)

	exampleTokenManagement(ctx, oauthClient)
}

func exampleAuthorizationCodeFlow(ctx context.Context, client *authgateway.OAuthProviderClient) {
	fmt.Println("=== Authorization Code Flow with PKCE ===")

	authURL, err := client.GetAuthorizationURL(ctx, &authgateway.AuthorizationURLOptions{
		Prompt: "consent",
	})
	if err != nil {
		log.Fatalf("Failed to generate authorization URL: %v", err)
	}

	fmt.Printf("1. Visit this URL to authorize:\n%s\n\n", authURL.URL)
	fmt.Printf("State: %s\n", authURL.State)
	fmt.Printf("Nonce: %s\n", authURL.Nonce)
	fmt.Printf("Code Verifier (for PKCE): %s\n\n", authURL.CodeVerifier)

	fmt.Println("2. After authorization, you'll be redirected to your redirect_uri with a 'code' parameter")
	fmt.Println("3. Exchange the code for tokens:")

	code := "authorization-code-from-callback"
	tokenResp, err := client.ExchangeCode(ctx, code, authURL.CodeVerifier)
	if err != nil {
		log.Printf("Token exchange failed: %v", err)
		return
	}

	fmt.Printf("Access Token: %s\n", tokenResp.AccessToken)
	fmt.Printf("Refresh Token: %s\n", tokenResp.RefreshToken)
	fmt.Printf("ID Token: %s\n", tokenResp.IDToken)
	fmt.Printf("Expires in: %d seconds\n\n", tokenResp.ExpiresIn)

	userInfo, err := client.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		return
	}

	fmt.Printf("User Info:\n")
	fmt.Printf("  Subject: %s\n", userInfo.Sub)
	fmt.Printf("  Email: %s (verified: %t)\n", userInfo.Email, userInfo.EmailVerified)
	fmt.Printf("  Name: %s\n", userInfo.Name)
	fmt.Printf("  Picture: %s\n\n", userInfo.Picture)
}

func exampleDeviceFlow(ctx context.Context, client *authgateway.OAuthProviderClient) {
	fmt.Println("=== Device Authorization Flow ===")

	deviceResp, err := client.RequestDeviceCode(ctx, []string{"openid", "profile"})
	if err != nil {
		log.Printf("Device authorization request failed: %v", err)
		return
	}

	fmt.Printf("Please visit: %s\n", deviceResp.VerificationURI)
	fmt.Printf("And enter code: %s\n", deviceResp.UserCode)
	fmt.Printf("Or visit directly: %s\n\n", deviceResp.VerificationURIComplete)

	interval := deviceResp.Interval
	if interval == 0 {
		interval = 5
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	timeout := time.After(time.Duration(deviceResp.ExpiresIn) * time.Second)

	fmt.Println("Polling for authorization...")
	for {
		select {
		case <-timeout:
			fmt.Println("Device flow timed out")
			return

		case <-ticker.C:
			tokenResp, err := client.PollDeviceToken(ctx, deviceResp.DeviceCode)
			if err != nil {
				if err == authgateway.ErrAuthorizationPending {
					continue
				}
				if err == authgateway.ErrSlowDown {
					ticker.Reset(time.Duration(interval+5) * time.Second)
					continue
				}
				log.Printf("Device flow error: %v", err)
				return
			}

			fmt.Printf("Authorization successful!\n")
			fmt.Printf("Access Token: %s\n\n", tokenResp.AccessToken)
			return
		}
	}
}

func exampleClientCredentials(ctx context.Context, client *authgateway.OAuthProviderClient) {
	fmt.Println("=== Client Credentials Flow ===")

	tokenResp, err := client.ClientCredentialsGrant(ctx, []string{"api:read", "api:write"})
	if err != nil {
		log.Printf("Client credentials grant failed: %v", err)
		return
	}

	fmt.Printf("Access Token: %s\n", tokenResp.AccessToken)
	fmt.Printf("Token Type: %s\n", tokenResp.TokenType)
	fmt.Printf("Expires in: %d seconds\n", tokenResp.ExpiresIn)
	fmt.Printf("Scope: %s\n\n", tokenResp.Scope)
}

func exampleTokenManagement(ctx context.Context, client *authgateway.OAuthProviderClient) {
	fmt.Println("=== Token Management ===")

	accessToken := "your-access-token"
	refreshToken := "your-refresh-token"

	introspection, err := client.IntrospectToken(ctx, accessToken)
	if err != nil {
		log.Printf("Token introspection failed: %v", err)
		return
	}

	fmt.Printf("Token Introspection:\n")
	fmt.Printf("  Active: %t\n", introspection.Active)
	if introspection.Active {
		fmt.Printf("  Username: %s\n", introspection.Username)
		fmt.Printf("  Scope: %s\n", introspection.Scope)
		fmt.Printf("  Expires at: %d\n", introspection.Exp)
		fmt.Printf("  Client ID: %s\n\n", introspection.ClientID)
	}

	newTokens, err := client.RefreshTokens(ctx, refreshToken)
	if err != nil {
		log.Printf("Token refresh failed: %v", err)
		return
	}

	fmt.Printf("Refreshed Tokens:\n")
	fmt.Printf("  New Access Token: %s\n", newTokens.AccessToken)
	fmt.Printf("  New Refresh Token: %s\n\n", newTokens.RefreshToken)

	err = client.RevokeToken(ctx, accessToken, "access_token")
	if err != nil {
		log.Printf("Token revocation failed: %v", err)
		return
	}

	fmt.Println("Token revoked successfully")
}

func exampleDiscovery(ctx context.Context, client *authgateway.OAuthProviderClient) {
	fmt.Println("=== OIDC Discovery ===")

	discovery, err := client.GetDiscovery(ctx)
	if err != nil {
		log.Printf("Discovery failed: %v", err)
		return
	}

	fmt.Printf("Issuer: %s\n", discovery.Issuer)
	fmt.Printf("Authorization Endpoint: %s\n", discovery.AuthorizationEndpoint)
	fmt.Printf("Token Endpoint: %s\n", discovery.TokenEndpoint)
	fmt.Printf("UserInfo Endpoint: %s\n", discovery.UserInfoEndpoint)
	fmt.Printf("JWKS URI: %s\n", discovery.JwksURI)
	fmt.Printf("Supported Grant Types: %v\n", discovery.GrantTypesSupported)
	fmt.Printf("Supported Scopes: %v\n\n", discovery.ScopesSupported)

	jwks, err := client.GetJWKS(ctx)
	if err != nil {
		log.Printf("JWKS fetch failed: %v", err)
		return
	}

	fmt.Printf("Number of keys in JWKS: %d\n", len(jwks.Keys))
	for i, key := range jwks.Keys {
		fmt.Printf("Key %d:\n", i+1)
		fmt.Printf("  Type: %s\n", key.Kty)
		fmt.Printf("  Use: %s\n", key.Use)
		fmt.Printf("  Algorithm: %s\n", key.Alg)
		fmt.Printf("  Key ID: %s\n", key.Kid)
	}
}

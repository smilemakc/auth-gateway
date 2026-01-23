// Package authgateway provides a Go SDK for Auth Gateway - a centralized
// authentication and authorization service.
//
// # Installation
//
// Install the latest version:
//
//	go get github.com/smilemakc/auth-gateway/packages/go-sdk@latest
//
// Install a specific version:
//
//	go get github.com/smilemakc/auth-gateway/packages/go-sdk@v0.1.0
//
// # Quick Start
//
// The SDK provides three main clients for different use cases:
//
// # REST Client
//
// For general API access (authentication, user management, etc.):
//
//	client := authgateway.NewClient(authgateway.Config{
//	    BaseURL: "https://auth.example.com",
//	})
//
//	// Sign in
//	resp, err := client.SignIn(ctx, &authgateway.SignInRequest{
//	    Email:    "user@example.com",
//	    Password: "password",
//	})
//
// # gRPC Client
//
// For server-to-server communication with token validation:
//
//	grpcClient, err := authgateway.NewGRPCClient(authgateway.GRPCConfig{
//	    Address: "auth.example.com:50051",
//	})
//
//	// Validate token
//	result, err := grpcClient.ValidateToken(ctx, accessToken)
//
// # OAuth Provider Client
//
// For using Auth Gateway as an OAuth 2.0 / OpenID Connect provider:
//
//	oauthClient := authgateway.NewOAuthProviderClient(authgateway.OAuthProviderConfig{
//	    Issuer:       "https://auth.example.com",
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	    RedirectURI:  "https://yourapp.com/callback",
//	    Scopes:       []string{"openid", "profile", "email"},
//	})
//
//	// Get authorization URL
//	authURL, err := oauthClient.GetAuthorizationURL(ctx, nil)
//
// For complete documentation and examples, see:
// https://github.com/smilemakc/auth-gateway/tree/main/packages/go-sdk
package authgateway

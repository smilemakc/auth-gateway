// Package main Auth Gateway API
//
// This is a comprehensive authentication and authorization gateway for microservices ecosystems.
// It provides both REST API and gRPC interfaces for user authentication, authorization, and management.
//
// Features:
//   - Email/Password Authentication
//   - OAuth 2.0 (Google, Yandex, GitHub, Instagram, Telegram)
//   - OTP via Email and SMS
//   - TOTP Two-Factor Authentication (Google Authenticator)
//   - API Keys for service-to-service authentication
//   - Role-Based Access Control (RBAC)
//   - Session Management
//   - IP Filtering
//   - Rate Limiting
//   - Audit Logging with Geo-location
//
// Terms Of Service:
//
// https://auth-gateway.example.com/terms
//
// Schemes: http, https
// Host: localhost:3000
// BasePath: /
// Version: 1.0.0
// Contact: Auth Gateway Team<support@auth-gateway.example.com>
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// SecurityDefinitions:
//
//	BearerAuth:
//	  type: apiKey
//	  name: Authorization
//	  in: header
//	  description: JWT Bearer token. Format: "Bearer {token}"
//	ApiKeyAuth:
//	  type: apiKey
//	  name: X-API-Key
//	  in: header
//	  description: API Key for service-to-service authentication. Format: "agw_{key}"
//
// swagger:meta
package main

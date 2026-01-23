# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-01-23

### Added
- REST client for Auth Gateway API
  - Authentication (sign up, sign in, sign out, token refresh)
  - OAuth 2.0 integration (Google, Yandex, GitHub, Instagram, Telegram)
  - Two-factor authentication (TOTP setup, verify, disable)
  - OTP via SMS and Email
  - Session management
  - API Keys management
  - User profile management
- gRPC client for server-to-server communication
  - Token validation
  - User retrieval
  - Permission checking
  - Token introspection
- OAuth Provider client for using Auth Gateway as identity provider
  - Authorization Code flow with PKCE
  - Device Authorization flow
  - Client Credentials flow
  - Token management (refresh, revoke)
  - UserInfo endpoint support
- Admin API for user and role management
  - User CRUD operations
  - Role and permission management
  - Audit logs access
- Automatic token refresh handling
- Comprehensive error types with HTTP status codes
- Examples for all major use cases

# OIDC Signing Keys - Quick Start Guide

## 1. Generate Keys (First Time)

```bash
cd backend
make keys-generate
```

Output:
```
Keys generated in ./keys with ID: key-20231215
```

## 2. Add to .env

```env
OIDC_SIGNING_KEY_PATH=./keys/rsa_private_key-20231215.pem
OIDC_SIGNING_KEY_ID=key-20231215
OIDC_SIGNING_ALGORITHM=RS256
```

## 3. Initialize in Code

```go
import "auth-gateway/backend/pkg/keys"

// In your main.go or initialization code
configs := []keys.KeyConfig{
    {
        ID:             os.Getenv("OIDC_SIGNING_KEY_ID"),
        Algorithm:      keys.Algorithm(os.Getenv("OIDC_SIGNING_ALGORITHM")),
        PrivateKeyPath: os.Getenv("OIDC_SIGNING_KEY_PATH"),
    },
}

keyManager, err := keys.NewManager(configs, os.Getenv("OIDC_SIGNING_KEY_ID"))
if err != nil {
    log.Fatal(err)
}
```

## 4. Use for Signing

```go
// Sign JWT payload
payload := []byte(jwtHeader + "." + jwtPayload)
signature, kid, err := keyManager.Sign(payload)

// Include kid in JWT header
header := map[string]interface{}{
    "alg": "RS256",
    "typ": "JWT",
    "kid": kid,
}
```

## 5. Add JWKS Endpoint

```go
// In your router setup
r.GET("/.well-known/jwks.json", func(c *gin.Context) {
    jwks := keyManager.GetJWKS()
    c.JSON(http.StatusOK, jwks)
})
```

## Test It

```bash
# Run example
make keys-example

# Check JWKS endpoint
curl http://localhost:3000/.well-known/jwks.json | jq
```

## Common Commands

```bash
# Generate keys
make keys-generate

# Generate keys with custom location
make keys-generate DIR=/etc/keys ID=prod-key

# Run tests
cd pkg/keys && go test -v

# Run example
make keys-example
```

## Key Rotation (When Needed)

```bash
# 1. Generate new key
make keys-generate ID=key-20240101

# 2. Update .env
OIDC_SIGNING_KEY_ID=key-20240101
OIDC_SIGNING_KEY_PATH=./keys/rsa_private_key-20240101.pem
OIDC_ADDITIONAL_KEYS=key-20231215:./keys/rsa_private_key-20231215.pem:RS256

# 3. Deploy and wait for old tokens to expire

# 4. Remove old key from config
```

## Files Reference

- **Implementation**: `/backend/pkg/keys/manager.go`
- **Tests**: `/backend/pkg/keys/manager_test.go`
- **Script**: `/backend/scripts/generate-keys.sh`
- **Example**: `/backend/examples/keys-example/main.go`
- **Full Docs**: `/backend/docs/OIDC_SIGNING_KEYS.md`

## Algorithm Choice

- **RS256**: Use for public-facing OIDC (recommended)
- **ES256**: Use for internal microservices (faster)

## Security Reminders

1. Never commit `*.pem` files (already in `.gitignore`)
2. Set file permissions: `chmod 600 keys/*.pem`
3. Use secret management in production
4. Rotate keys every 6-12 months

## Need Help?

1. Check full documentation: `docs/OIDC_SIGNING_KEYS.md`
2. Review implementation notes: `pkg/keys/IMPLEMENTATION_NOTES.md`
3. Run the example: `make keys-example`
4. Check logs for detailed errors

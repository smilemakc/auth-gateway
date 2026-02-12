# OIDC Signing Keys Configuration

This document describes how to configure and manage OIDC signing keys for Auth Gateway.

## Overview

Auth Gateway uses asymmetric cryptography to sign OIDC tokens (ID tokens and access tokens). This provides:

- **Security**: Private keys remain on the auth server, public keys are distributed
- **Verification**: Clients can verify tokens without calling the auth server
- **Standard Compliance**: Follows OIDC/OAuth 2.0 specifications
- **Key Rotation**: Support for multiple keys enables zero-downtime rotation

## Quick Start

### 1. Generate Keys

```bash
cd backend
make keys-generate
```

This creates keys in `./keys/` directory with today's date as the key ID.

Custom location and ID:
```bash
make keys-generate DIR=/etc/auth-gateway/keys ID=prod-key-2023
```

### 2. Configure Environment Variables

Add to your `.env` file:

```env
# OIDC Signing Keys
OIDC_SIGNING_KEY_PATH=./keys/rsa_private_key-20231215.pem
OIDC_SIGNING_KEY_ID=key-20231215
OIDC_SIGNING_ALGORITHM=RS256
```

For multiple keys (rotation):
```env
# Primary key (used for signing)
OIDC_SIGNING_KEY_PATH=./keys/rsa_private_key-new.pem
OIDC_SIGNING_KEY_ID=key-new
OIDC_SIGNING_ALGORITHM=RS256

# Additional keys (comma-separated, used for verification only)
OIDC_ADDITIONAL_KEYS=key-old:./keys/rsa_private_key-old.pem:RS256
```

### 3. Update Application Code

Initialize the key manager in your main application:

```go
import (
    "auth-gateway/backend/pkg/keys"
    "os"
    "strings"
)

func initializeKeyManager() (*keys.Manager, error) {
    configs := []keys.KeyConfig{
        {
            ID:             os.Getenv("OIDC_SIGNING_KEY_ID"),
            Algorithm:      keys.Algorithm(os.Getenv("OIDC_SIGNING_ALGORITHM")),
            PrivateKeyPath: os.Getenv("OIDC_SIGNING_KEY_PATH"),
        },
    }

    // Add additional keys for rotation
    if additionalKeys := os.Getenv("OIDC_ADDITIONAL_KEYS"); additionalKeys != "" {
        for _, keySpec := range strings.Split(additionalKeys, ",") {
            parts := strings.Split(keySpec, ":")
            if len(parts) == 3 {
                configs = append(configs, keys.KeyConfig{
                    ID:             parts[0],
                    Algorithm:      keys.Algorithm(parts[2]),
                    PrivateKeyPath: parts[1],
                })
            }
        }
    }

    return keys.NewManager(configs, os.Getenv("OIDC_SIGNING_KEY_ID"))
}
```

## Algorithm Selection

### RS256 (Recommended)

**RSA with SHA-256**

- ✓ Industry standard, universally supported
- ✓ Maximum compatibility with clients
- ✓ Well-tested and proven
- ✗ Larger key size (2048+ bits)
- ✗ Slower signing operations

**Use for:**
- Public-facing OIDC provider
- Mobile apps
- Third-party integrations
- Maximum compatibility requirements

```env
OIDC_SIGNING_ALGORITHM=RS256
```

### ES256 (Modern Alternative)

**ECDSA with P-256 and SHA-256**

- ✓ Smaller keys (256 bits)
- ✓ Faster signing and verification
- ✓ Smaller signature size
- ✗ May not be supported by older clients
- ✗ Less common in enterprise

**Use for:**
- Internal microservices
- Modern applications
- Performance-critical scenarios
- Cloud-native deployments

```env
OIDC_SIGNING_ALGORITHM=ES256
```

## Key Rotation

### Why Rotate Keys?

- **Security**: Limit exposure if a key is compromised
- **Compliance**: Meet regulatory requirements
- **Best Practice**: Industry standard (rotate every 6-12 months)

### Rotation Process

#### Step 1: Generate New Key

```bash
make keys-generate ID=key-20240101
```

#### Step 2: Configure Both Keys

Update `.env` to include both old and new keys:

```env
# New key (primary)
OIDC_SIGNING_KEY_PATH=./keys/rsa_private_key-20240101.pem
OIDC_SIGNING_KEY_ID=key-20240101
OIDC_SIGNING_ALGORITHM=RS256

# Old key (kept for verification)
OIDC_ADDITIONAL_KEYS=key-20231215:./keys/rsa_private_key-20231215.pem:RS256
```

#### Step 3: Deploy

Deploy the updated configuration. The service will:
- Sign new tokens with `key-20240101`
- Verify tokens signed with either key
- Publish both public keys in JWKS

#### Step 4: Wait for Token Expiration

Wait until all tokens signed with the old key have expired:
- ID tokens: typically 1 hour
- Access tokens: typically 15 minutes
- Refresh tokens: typically 7 days

**Wait time = longest refresh token TTL + safety margin**

#### Step 5: Remove Old Key

After waiting, remove the old key from configuration:

```env
OIDC_SIGNING_KEY_PATH=./keys/rsa_private_key-20240101.pem
OIDC_SIGNING_KEY_ID=key-20240101
OIDC_SIGNING_ALGORITHM=RS256
# OIDC_ADDITIONAL_KEYS removed
```

#### Step 6: Archive Old Key

Move old key to archive (don't delete immediately):

```bash
mkdir -p keys/archive
mv keys/rsa_private_key-20231215.pem keys/archive/
mv keys/rsa_public_key-20231215.pem keys/archive/
```

## JWKS Endpoint

Public keys are automatically published at:

```
GET /.well-known/jwks.json
```

Example response:
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "key-20240101",
      "n": "xGOr-H7A...",
      "e": "AQAB"
    },
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "key-20231215",
      "n": "yKPs-I8B...",
      "e": "AQAB"
    }
  ]
}
```

Clients use this endpoint to:
1. Discover current public keys
2. Verify token signatures
3. Cache keys (with refresh)

## Production Deployment

### Key Storage Options

#### Development (Local Files)

```env
OIDC_SIGNING_KEY_PATH=./keys/rsa_private_key.pem
```

✓ Simple
✓ Good for development
✗ Not secure for production
✗ Hard to rotate across multiple servers

#### Production (Secret Management)

**AWS Secrets Manager:**
```go
import "github.com/aws/aws-sdk-go/service/secretsmanager"

func loadKeyFromAWS(secretName string) (string, error) {
    svc := secretsmanager.New(session.New())
    result, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretName),
    })
    if err != nil {
        return "", err
    }
    return *result.SecretString, nil
}
```

**HashiCorp Vault:**
```go
import "github.com/hashicorp/vault/api"

func loadKeyFromVault(path string) (string, error) {
    client, _ := api.NewClient(api.DefaultConfig())
    secret, err := client.Logical().Read(path)
    if err != nil {
        return "", err
    }
    return secret.Data["private_key"].(string), nil
}
```

**Kubernetes Secrets:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: oidc-signing-keys
type: Opaque
data:
  private-key: <base64-encoded-pem>
```

Mount as file:
```yaml
volumeMounts:
  - name: signing-keys
    mountPath: /etc/keys
    readOnly: true
```

### Security Best Practices

1. **File Permissions**
   ```bash
   chmod 600 keys/*.pem
   chown app:app keys/*.pem
   ```

2. **Never Commit Keys**
   - `*.pem` is in `.gitignore`
   - Use secret management for production
   - Rotate if accidentally committed

3. **Key Backup**
   - Keep encrypted backups of keys
   - Store in secure location (not version control)
   - Document recovery procedure

4. **Access Control**
   - Limit who can generate keys
   - Audit key access
   - Use separate keys per environment

5. **Monitoring**
   - Alert on key loading failures
   - Monitor JWKS endpoint access
   - Track signature verification failures

## Troubleshooting

### Error: "failed to decode PEM block"

**Cause:** Invalid PEM file format

**Solution:**
```bash
# Verify PEM format
openssl rsa -in key.pem -text -noout

# Re-generate if corrupted
make keys-generate
```

### Error: "current key ID not found"

**Cause:** `OIDC_SIGNING_KEY_ID` doesn't match any loaded key

**Solution:** Verify environment configuration:
```bash
# Check current config
echo $OIDC_SIGNING_KEY_ID
echo $OIDC_SIGNING_KEY_PATH

# Ensure key file exists
ls -l $OIDC_SIGNING_KEY_PATH
```

### Error: "signature verification failed"

**Cause:** Token signed with unknown key or key mismatch

**Solution:**
1. Check JWKS endpoint includes the key
2. Verify token's `kid` header matches available keys
3. Ensure key rotation was completed properly

### JWKS Endpoint Returns Empty Keys

**Cause:** No keys loaded or service not initialized

**Solution:**
1. Check logs for key loading errors
2. Verify `OIDC_SIGNING_KEY_PATH` is correct
3. Ensure file permissions allow reading

## Testing

### Generate Test Keys

```bash
make keys-generate DIR=./test/keys ID=test-key
```

### Run Example

```bash
KEY_DIR=./keys KEY_ID=key-20231215 make keys-example
```

### Verify JWKS

```bash
curl http://localhost:3000/.well-known/jwks.json | jq
```

### Test Token Signing

```go
package main

import (
    "auth-gateway/backend/pkg/keys"
    "testing"
)

func TestTokenSigning(t *testing.T) {
    manager, err := keys.NewManager([]keys.KeyConfig{
        {
            ID:             "test-key",
            Algorithm:      keys.RS256,
            PrivateKeyPath: "./test/keys/rsa_private_test-key.pem",
        },
    }, "test-key")

    if err != nil {
        t.Fatalf("Failed to create manager: %v", err)
    }

    payload := []byte("test.payload.data")
    signature, kid, err := manager.Sign(payload)

    if err != nil {
        t.Fatalf("Failed to sign: %v", err)
    }

    err = manager.Verify(payload, signature, kid)
    if err != nil {
        t.Fatalf("Failed to verify: %v", err)
    }
}
```

## Migration Guide

### From Symmetric Keys (HS256)

If you're currently using symmetric keys (HS256):

1. **Generate asymmetric keys**
   ```bash
   make keys-generate
   ```

2. **Update token service to use new manager**
   ```go
   // Old (HS256)
   token := jwt.SignedString(symmetricKey)

   // New (RS256)
   signature, kid, _ := keyManager.Sign(payload)
   ```

3. **Deploy with both algorithms supported**
   - Sign new tokens with RS256
   - Verify both HS256 and RS256

4. **Wait for old tokens to expire**

5. **Remove HS256 support**

### From External Key Management

If you're loading keys from AWS/Vault:

1. **Load key content into temporary file**
   ```go
   keyContent, _ := loadFromVault()
   tmpFile, _ := os.CreateTemp("", "key-*.pem")
   tmpFile.WriteString(keyContent)
   tmpFile.Close()
   ```

2. **Initialize manager with temp file**
   ```go
   manager, _ := keys.NewManager([]keys.KeyConfig{
       {
           ID:             "vault-key",
           Algorithm:      keys.RS256,
           PrivateKeyPath: tmpFile.Name(),
       },
   }, "vault-key")
   ```

3. **Clean up temp file on shutdown**
   ```go
   defer os.Remove(tmpFile.Name())
   ```

## Reference

- [RFC 7517 - JSON Web Key (JWK)](https://tools.ietf.org/html/rfc7517)
- [RFC 7518 - JSON Web Algorithms (JWA)](https://tools.ietf.org/html/rfc7518)
- [OIDC Core Specification](https://openid.net/specs/openid-connect-core-1_0.html)
- [OAuth 2.0 Token Introspection](https://tools.ietf.org/html/rfc7662)

## Support

For issues or questions:
1. Check logs for detailed error messages
2. Verify configuration matches this guide
3. Run example to test key manager in isolation
4. Review key file permissions and paths

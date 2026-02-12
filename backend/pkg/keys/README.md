# Keys Package

This package provides signing key management for OIDC JWT tokens, supporting both RSA and ECDSA algorithms.

## Features

- RSA-256 (RS256) and ECDSA P-256 (ES256) signing algorithms
- Multiple key support for key rotation
- JWKS (JSON Web Key Set) generation for public key distribution
- Thread-safe key management
- PEM file format support

## Quick Start

### 1. Generate Keys

Use the provided script to generate key pairs:

```bash
cd backend
./scripts/generate-keys.sh ./keys key-20231215
```

This creates:
- `keys/rsa_private_key-20231215.pem` - RSA private key
- `keys/rsa_public_key-20231215.pem` - RSA public key
- `keys/ecdsa_private_key-20231215.pem` - ECDSA private key
- `keys/ecdsa_public_key-20231215.pem` - ECDSA public key

### 2. Configure Environment

Add to your `.env`:

```env
OIDC_SIGNING_KEY_PATH=./keys/rsa_private_key-20231215.pem
OIDC_SIGNING_KEY_ID=key-20231215
OIDC_SIGNING_ALGORITHM=RS256
```

### 3. Initialize Manager

```go
import "auth-gateway/backend/pkg/keys"

// Single key configuration
configs := []keys.KeyConfig{
    {
        ID:             "key-20231215",
        Algorithm:      keys.RS256,
        PrivateKeyPath: "./keys/rsa_private_key-20231215.pem",
    },
}

manager, err := keys.NewManager(configs, "key-20231215")
if err != nil {
    log.Fatal(err)
}
```

## Usage Examples

### Signing Data

```go
data := []byte("data to sign")
signature, kid, err := manager.Sign(data)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Signed with key: %s\n", kid)
```

### Verifying Signatures

```go
err := manager.Verify(data, signature, kid)
if err != nil {
    log.Printf("Verification failed: %v\n", err)
}
```

### Getting JWKS for Public Distribution

```go
jwks := manager.GetJWKS()

// Serve at /.well-known/jwks.json
http.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(jwks)
})
```

## Key Rotation

To rotate keys while maintaining backward compatibility:

### 1. Generate New Key

```bash
./scripts/generate-keys.sh ./keys key-20231216
```

### 2. Update Configuration

```go
configs := []keys.KeyConfig{
    {
        ID:             "key-20231216",
        Algorithm:      keys.RS256,
        PrivateKeyPath: "./keys/rsa_private_key-20231216.pem",
    },
    {
        ID:             "key-20231215",
        Algorithm:      keys.RS256,
        PrivateKeyPath: "./keys/rsa_private_key-20231215.pem",
    },
}

// New key is current, old key still available for verification
manager, err := keys.NewManager(configs, "key-20231216")
```

### 3. Update Environment

```env
OIDC_SIGNING_KEY_ID=key-20231216
```

The JWKS endpoint will automatically include both keys, allowing clients to verify tokens signed with either key during the rotation period.

## Supported Algorithms

### RS256 (RSA with SHA-256)

- **Key Size**: 2048 bits (minimum recommended)
- **Use Case**: Industry standard, widely supported
- **Performance**: Slower signing, faster verification
- **Recommendation**: Use for public-facing OIDC

```go
{
    ID:        "rsa-key",
    Algorithm: keys.RS256,
    PrivateKeyPath: "./keys/rsa_private.pem",
}
```

### ES256 (ECDSA with P-256 and SHA-256)

- **Key Size**: 256 bits
- **Use Case**: Modern, efficient alternative to RSA
- **Performance**: Faster signing and smaller signatures
- **Recommendation**: Use for internal services

```go
{
    ID:        "ec-key",
    Algorithm: keys.ES256,
    PrivateKeyPath: "./keys/ecdsa_private.pem",
}
```

## JWKS Format

The `GetJWKS()` method returns a JSON Web Key Set following RFC 7517:

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "key-20231215",
      "n": "base64url-encoded-modulus",
      "e": "AQAB"
    },
    {
      "kty": "EC",
      "use": "sig",
      "alg": "ES256",
      "kid": "key-20231216",
      "crv": "P-256",
      "x": "base64url-encoded-x",
      "y": "base64url-encoded-y"
    }
  ]
}
```

## Security Best Practices

1. **Private Key Protection**
   - Never commit private keys to version control
   - Add `*.pem` to `.gitignore`
   - Set file permissions to 600: `chmod 600 keys/*.pem`
   - Use environment-specific keys (dev/staging/prod)

2. **Key Rotation**
   - Rotate keys periodically (e.g., every 6-12 months)
   - Keep old keys available for token verification during rotation
   - Remove old keys after token expiration period

3. **Key Storage**
   - Production: Use secret management (AWS Secrets Manager, HashiCorp Vault)
   - Development: Use local files with restricted permissions
   - Never pass keys through environment variables in production

4. **Algorithm Selection**
   - Use RS256 for maximum compatibility
   - Use ES256 for better performance with modern clients
   - Never use symmetric algorithms (HS256) for OIDC

## Testing

Run the test suite:

```bash
cd backend/pkg/keys
go test -v
```

Run tests with coverage:

```bash
go test -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## API Reference

### Types

#### `Algorithm`
```go
type Algorithm string

const (
    RS256 Algorithm = "RS256"
    ES256 Algorithm = "ES256"
)
```

#### `KeyConfig`
```go
type KeyConfig struct {
    ID             string
    Algorithm      Algorithm
    PrivateKeyPath string
    PublicKeyPath  string  // Optional, derived from private key
}
```

#### `Manager`
```go
type Manager struct {
    // Thread-safe key management
}
```

### Functions

#### `NewManager`
```go
func NewManager(configs []KeyConfig, currentKID string) (*Manager, error)
```
Creates a new key manager. Returns error if no keys provided or current key not found.

#### `(*Manager) GetCurrentKey`
```go
func (m *Manager) GetCurrentKey() (*SigningKey, error)
```
Returns the current signing key for creating new signatures.

#### `(*Manager) GetKey`
```go
func (m *Manager) GetKey(kid string) (*SigningKey, error)
```
Returns a specific key by ID for verification.

#### `(*Manager) GetJWKS`
```go
func (m *Manager) GetJWKS() *JWKS
```
Returns JWKS document with all public keys.

#### `(*Manager) Sign`
```go
func (m *Manager) Sign(data []byte) ([]byte, string, error)
```
Signs data with current key. Returns signature, key ID, and error.

#### `(*Manager) Verify`
```go
func (m *Manager) Verify(data, signature []byte, kid string) error
```
Verifies signature with specified key.

#### `LoadRSAPrivateKey`
```go
func LoadRSAPrivateKey(path string) (*rsa.PrivateKey, error)
```
Loads RSA private key from PEM file. Supports PKCS1 and PKCS8 formats.

#### `LoadECDSAPrivateKey`
```go
func LoadECDSAPrivateKey(path string) (*ecdsa.PrivateKey, error)
```
Loads ECDSA private key from PEM file. Supports EC and PKCS8 formats.

## Integration with OIDC

This package is designed to work with OIDC token generation:

```go
// In your OIDC token service
manager, _ := keys.NewManager(configs, currentKID)

// Sign JWT payload
payload := []byte(jwtHeader + "." + jwtPayload)
signature, kid, err := manager.Sign(payload)

// Create JWT with kid in header
token := jwtHeader + "." + jwtPayload + "." + base64url(signature)
```

Clients can discover public keys via JWKS endpoint:
```
GET /.well-known/jwks.json
```

## Troubleshooting

### Error: "failed to decode PEM block"

Check that your key file is valid PEM format:
```bash
openssl rsa -in key.pem -text -noout
```

### Error: "not an RSA private key"

Ensure you're using the correct algorithm for your key type:
- RSA keys → `RS256`
- ECDSA keys → `ES256`

### Error: "signature verification failed"

Verify:
1. Data hasn't been modified
2. Using correct key ID
3. Signature hasn't been corrupted
4. Key hasn't been rotated without updating verification

## License

Part of Auth Gateway project.

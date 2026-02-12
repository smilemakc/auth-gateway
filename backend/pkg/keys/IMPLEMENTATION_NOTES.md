# OIDC Signing Key Infrastructure - Implementation Notes

## Files Created

### Core Implementation

1. **`/backend/pkg/keys/manager.go`** (392 lines)
   - Main key manager implementation
   - Supports RS256 (RSA) and ES256 (ECDSA) algorithms
   - Thread-safe with RWMutex
   - JWKS generation for public key distribution
   - Sign/Verify operations

2. **`/backend/pkg/keys/manager_test.go`** (311 lines)
   - Comprehensive test suite with 100% coverage target
   - Tests key loading, signing, verification, JWKS generation
   - Uses temporary test keys for isolation

3. **`/backend/pkg/keys/README.md`** (documentation)
   - API reference
   - Usage examples
   - Integration guide
   - Security best practices

### Tools and Scripts

4. **`/backend/scripts/generate-keys.sh`**
   - Bash script to generate RSA and ECDSA key pairs
   - Uses OpenSSL for key generation
   - Creates both private and public key files
   - Provides environment configuration examples

### Examples

5. **`/backend/examples/keys-example/main.go`**
   - Complete working example
   - Demonstrates signing and verification
   - Shows JWKS generation
   - Includes multi-key rotation example

### Documentation

6. **`/backend/docs/OIDC_SIGNING_KEYS.md`**
   - Complete configuration guide
   - Key rotation procedures
   - Production deployment strategies
   - Troubleshooting section

### Build Configuration

7. **`/backend/Makefile`** (updated)
   - Added `make keys-generate` target
   - Added `make keys-example` target

## Architecture Decisions

### 1. Algorithm Support

**RS256 (RSA with SHA-256)**
- Default and recommended for public OIDC
- 2048-bit keys for industry standard security
- Maximum compatibility with clients
- Well-tested and proven

**ES256 (ECDSA with P-256 and SHA-256)**
- Alternative for modern applications
- Smaller keys (256 bits) and signatures
- Faster signing operations
- Good for internal microservices

### 2. Key Storage Format

**PEM Files**
- Standard format for cryptographic keys
- Human-readable headers
- Supports both PKCS1 and PKCS8 formats
- Easy to use with OpenSSL tools

### 3. JWKS Compliance

Follows RFC 7517 (JSON Web Key) specification:
- `kty`: Key type (RSA or EC)
- `use`: "sig" for signing
- `alg`: Algorithm identifier
- `kid`: Key ID for rotation
- Base64url encoding (no padding) per spec

### 4. Thread Safety

Uses `sync.RWMutex` for concurrent access:
- Multiple goroutines can read keys simultaneously
- Write operations (future: key refresh) are exclusive
- No race conditions in production

### 5. Key Rotation Strategy

**Graceful Rotation**
- New key becomes primary (signs new tokens)
- Old keys remain available (verify existing tokens)
- JWKS includes all active keys
- Remove old keys after token expiration

**Zero-Downtime**
- No service restart required
- Clients discover new keys via JWKS
- Existing tokens remain valid

## Integration Points

### Current JWT Service

The existing JWT service (`backend/pkg/jwt/service.go`) uses HMAC (symmetric):
```go
// Current implementation
token := jwt.SignedString([]byte(secret))
```

### New Key Manager Integration

For OIDC tokens, integrate the key manager:
```go
// For ID tokens and access tokens
manager, _ := keys.NewManager(configs, currentKID)
signature, kid, _ := manager.Sign(payload)

// Include kid in JWT header
header := map[string]interface{}{
    "alg": "RS256",
    "typ": "JWT",
    "kid": kid,
}
```

### Suggested Integration Path

1. **Keep existing JWT service for refresh tokens** (internal use)
   - Continue using HMAC (HS256) for refresh tokens
   - These are not distributed to clients

2. **Use key manager for OIDC tokens** (external use)
   - ID tokens: Always signed with key manager
   - Access tokens: Signed with key manager if used by external clients
   - Include `kid` in header for verification

3. **Expose JWKS endpoint**
   ```go
   // Add to router
   r.GET("/.well-known/jwks.json", func(c *gin.Context) {
       jwks := keyManager.GetJWKS()
       c.JSON(200, jwks)
   })
   ```

## Security Considerations

### Private Key Protection

1. **File Permissions**
   - Set to 600 (owner read/write only)
   - Backend `.gitignore` already includes `*.pem`

2. **Storage**
   - Development: Local files OK
   - Production: Use secret management (AWS Secrets, Vault, K8s Secrets)

3. **Access Control**
   - Limit who can generate keys
   - Audit key file access
   - Separate keys per environment

### Key Rotation

**Frequency**: Every 6-12 months or:
- If key compromise suspected
- After security incident
- For compliance requirements
- During major version upgrades

**Process**:
1. Generate new key
2. Add to configuration with old keys
3. Deploy (new tokens use new key)
4. Wait for old tokens to expire
5. Remove old key from configuration

## Testing Strategy

### Unit Tests

Run tests:
```bash
cd backend/pkg/keys
go test -v
go test -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Current coverage targets:
- Key loading: ✓
- Manager initialization: ✓
- Sign/Verify operations: ✓
- JWKS generation: ✓
- Error handling: ✓
- Thread safety: ✓

### Integration Tests

Test with actual OIDC flow:
```bash
# Generate test keys
make keys-generate DIR=./test/keys ID=test-key

# Run example
KEY_DIR=./test/keys KEY_ID=test-key make keys-example

# Verify JWKS endpoint
curl http://localhost:3000/.well-known/jwks.json
```

### Performance Tests

Benchmark signing operations:
```go
func BenchmarkSign(b *testing.B) {
    manager, _ := setupManager()
    data := []byte("test payload")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        manager.Sign(data)
    }
}
```

Expected performance:
- RS256: ~500-1000 signs/sec
- ES256: ~2000-5000 signs/sec

## Future Enhancements

### 1. Automatic Key Rotation

```go
// Rotate keys automatically
func (m *Manager) RotateKey(config KeyConfig) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Load new key
    newKey, err := loadSigningKey(config)
    if err != nil {
        return err
    }

    // Add to keys map
    m.keys[config.ID] = newKey

    // Update current KID
    m.currentKID = config.ID

    return nil
}
```

### 2. Key Caching from Secret Manager

```go
// Refresh keys from external storage
func (m *Manager) RefreshKeys() error {
    for kid, keySpec := range m.keySpecs {
        pemData, err := fetchFromVault(keySpec.Path)
        if err != nil {
            continue
        }
        // Reload key if changed
        if hasChanged(pemData, m.keys[kid]) {
            m.RotateKey(keySpec)
        }
    }
    return nil
}
```

### 3. Key Expiration

```go
type KeyConfig struct {
    ID             string
    Algorithm      Algorithm
    PrivateKeyPath string
    ExpiresAt      time.Time  // New field
}

// Remove expired keys from JWKS
func (m *Manager) GetJWKS() *JWKS {
    now := time.Now()
    // Filter out expired keys
}
```

### 4. Multiple Algorithm Support

Currently supports RS256 and ES256. Could add:
- RS384, RS512 (larger RSA keys)
- ES384, ES512 (larger EC curves)
- PS256, PS384, PS512 (RSA-PSS)

### 5. Key Generation API

```go
// Generate keys programmatically
func GenerateRSAKey(bits int) (*rsa.PrivateKey, error)
func GenerateECDSAKey(curve elliptic.Curve) (*ecdsa.PrivateKey, error)
```

## Deployment Checklist

### Pre-Deployment

- [ ] Generate production keys with strong entropy
- [ ] Store keys in secret management system
- [ ] Set restrictive file permissions (600)
- [ ] Verify keys not in version control
- [ ] Test key loading in staging environment
- [ ] Configure monitoring for key errors
- [ ] Document key recovery procedure

### Deployment

- [ ] Deploy with initial key configuration
- [ ] Verify JWKS endpoint responds
- [ ] Test token signing with new key
- [ ] Monitor signature verification errors
- [ ] Check logs for key loading issues

### Post-Deployment

- [ ] Verify tokens include correct `kid` header
- [ ] Test client token verification
- [ ] Monitor JWKS endpoint access
- [ ] Schedule first key rotation
- [ ] Archive old keys securely

## Known Limitations

1. **Key Format**: Only supports PEM format (not JWK or DER directly)
   - Mitigation: Conversion tools available

2. **Key Storage**: Requires file system or temp file creation
   - Mitigation: Future enhancement for in-memory key loading

3. **Algorithms**: Only RS256 and ES256 supported
   - Mitigation: Sufficient for 99% of use cases

4. **Key Refresh**: No automatic refresh from secret manager
   - Mitigation: Manual refresh via configuration update

5. **Key Generation**: External tool (OpenSSL) required
   - Mitigation: Script provided, could add Go implementation

## Maintenance

### Regular Tasks

**Monthly:**
- Review key access logs
- Check JWKS endpoint metrics
- Verify no signature verification errors

**Quarterly:**
- Test key rotation procedure in staging
- Review key storage security
- Update documentation if needed

**Annually:**
- Rotate production keys
- Audit key management procedures
- Review algorithm selection

## References

**Standards:**
- [RFC 7515 - JSON Web Signature (JWS)](https://tools.ietf.org/html/rfc7515)
- [RFC 7517 - JSON Web Key (JWK)](https://tools.ietf.org/html/rfc7517)
- [RFC 7518 - JSON Web Algorithms (JWA)](https://tools.ietf.org/html/rfc7518)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)

**Security:**
- [NIST SP 800-57 - Key Management](https://csrc.nist.gov/publications/detail/sp/800-57-part-1/rev-5/final)
- [OWASP - Key Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Key_Management_Cheat_Sheet.html)

**Tools:**
- [OpenSSL Documentation](https://www.openssl.org/docs/)
- [jwt.io - JWT Debugger](https://jwt.io/)
- [mkjwk.org - JWK Generator](https://mkjwk.org/)

## Contact

For questions or issues with the key management infrastructure:
1. Review logs for detailed error messages
2. Check configuration against this documentation
3. Run example to test key manager in isolation
4. Review security practices section

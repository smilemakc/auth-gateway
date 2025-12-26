# SAML 2.0 Integration Guide

## Overview

Auth Gateway implements SAML 2.0 Identity Provider (IdP) functionality, allowing it to act as a Single Sign-On (SSO) provider for Service Providers (SPs) like enterprise applications, VPNs, and other SAML-compatible services.

## Features

- **SAML 2.0 IdP**: Full SAML 2.0 Identity Provider implementation
- **SSO (Single Sign-On)**: Users authenticate once and access multiple services
- **SLO (Single Logout)**: Logout from all services with one action
- **Metadata Endpoint**: Automatic metadata generation for SPs
- **Multiple SPs**: Support for multiple Service Providers
- **Attribute Mapping**: Customizable SAML attributes in assertions
- **Signed Assertions**: Cryptographically signed SAML assertions

## Configuration

### Environment Variables

```bash
# Enable SAML IdP
SAML_ENABLED=true

# SAML Entity ID (unique identifier for this IdP)
SAML_ISSUER=https://auth.example.com

# SSO and SLO endpoints
SAML_SSO_URL=https://auth.example.com/saml/sso
SAML_SLO_URL=https://auth.example.com/saml/slo

# Certificate and private key paths for signing assertions
SAML_CERTIFICATE_PATH=/path/to/certificate.pem
SAML_PRIVATE_KEY_PATH=/path/to/private-key.pem

# Metadata URL
SAML_METADATA_URL=https://auth.example.com/saml/metadata
```

### Service Provider Configuration

Configure Service Providers via the Admin API:

#### Create Service Provider

```bash
POST /api/admin/saml/sp
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "Corporate VPN",
  "entity_id": "https://vpn.corporate.com/saml",
  "acs_url": "https://vpn.corporate.com/saml/acs",
  "slo_url": "https://vpn.corporate.com/saml/slo",
  "certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
  "enabled": true,
  "attribute_mappings": {
    "email": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
    "username": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
    "roles": "http://schemas.microsoft.com/ws/2008/06/identity/claims/role"
  }
}
```

## Endpoints

### Metadata Endpoint

Service Providers use this endpoint to discover IdP capabilities:

```http
GET /saml/metadata
```

Returns SAML metadata XML containing:
- Entity ID
- SSO and SLO endpoints
- Signing certificate
- Supported bindings

### SSO Endpoint

Service Providers redirect users here for authentication:

```http
POST /saml/sso
Content-Type: application/x-www-form-urlencoded

SAMLRequest=<base64-encoded-SAML-request>
RelayState=<optional-relay-state>
```

Flow:
1. SP redirects user to `/saml/sso` with SAML request
2. User authenticates (if not already)
3. Auth Gateway generates SAML assertion
4. User is redirected back to SP with SAML response

### SLO Endpoint (Optional)

Single Logout endpoint:

```http
POST /saml/slo
Content-Type: application/x-www-form-urlencoded

SAMLRequest=<base64-encoded-logout-request>
RelayState=<optional-relay-state>
```

## SAML Assertions

### Standard Attributes

Auth Gateway includes these attributes in SAML assertions by default:

- **Email**: User's email address
- **Username**: User's username
- **Roles**: User's roles (if configured)

### Custom Attribute Mappings

Configure attribute mappings per Service Provider:

```json
{
  "attribute_mappings": {
    "email": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
    "username": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
    "first_name": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
    "last_name": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
    "roles": "http://schemas.microsoft.com/ws/2008/06/identity/claims/role",
    "groups": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/groups"
  }
}
```

## Integration Examples

### VPN Integration (OpenVPN, etc.)

1. **Configure VPN as SP**:
   - Entity ID: `https://vpn.corporate.com/saml`
   - ACS URL: `https://vpn.corporate.com/saml/acs`
   - Configure attribute mappings

2. **Download IdP Metadata**:
   ```bash
   curl https://auth.example.com/saml/metadata > idp-metadata.xml
   ```

3. **Configure VPN**:
   - Import IdP metadata
   - Configure attribute mappings
   - Test SSO flow

### Enterprise Application Integration

#### Example: Jira/Confluence

1. In Jira/Confluence, go to Administration → Security → SAML Single Sign-On
2. Configure:
   - IdP Entity ID: `https://auth.example.com`
   - IdP SSO URL: `https://auth.example.com/saml/sso`
   - IdP Certificate: Download from metadata or certificate file
3. Configure attribute mappings:
   - Username: `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name`
   - Email: `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress`
4. Test SSO

#### Example: Microsoft 365 / Azure AD

If integrating Auth Gateway as IdP for Microsoft services:

1. Configure Auth Gateway as custom IdP in Azure AD
2. Upload IdP metadata
3. Configure claim mappings
4. Assign to applications

## Certificate Management

### Generate Self-Signed Certificate

```bash
# Generate private key
openssl genrsa -out saml-private-key.pem 2048

# Generate certificate
openssl req -new -x509 -key saml-private-key.pem -out saml-certificate.pem -days 365

# For production, use certificates from a trusted CA
```

### Certificate Requirements

- **Algorithm**: RSA 2048-bit or higher (recommended: RSA 4096-bit)
- **Format**: PEM format
- **Validity**: Ensure certificates are not expired
- **Key Usage**: Digital signature

## Security Best Practices

1. **Use HTTPS**: Always use HTTPS for SAML endpoints
2. **Certificate Security**: Store private keys securely (use secrets management)
3. **Assertion Expiry**: Configure appropriate assertion validity periods
4. **Replay Protection**: Implement replay attack protection
5. **Signature Validation**: Always validate SAML request signatures
6. **Audit Logging**: Log all SAML operations for security auditing

## Troubleshooting

### Common Issues

1. **Metadata Not Loading**:
   - Verify certificate path is correct
   - Check certificate format (PEM)
   - Ensure certificate is valid

2. **SSO Not Working**:
   - Verify SP entity ID matches configuration
   - Check ACS URL is correct
   - Review SAML request/response in browser developer tools
   - Check assertion signature

3. **Attributes Not Appearing**:
   - Verify attribute mappings in SP configuration
   - Check user has required attributes
   - Review SAML assertion XML

### Debugging

Enable SAML debugging:

1. Check server logs for SAML operations
2. Use browser developer tools to inspect SAML requests/responses
3. Use SAML tracer browser extension
4. Review SAML assertion XML for attribute values

## Testing

### Test with SAML Test Tool

Use online SAML test tools or local SAML test SP:

1. Configure test SP with Auth Gateway metadata
2. Initiate SSO flow
3. Verify SAML response
4. Check attributes in assertion

### Manual Testing

```bash
# Get metadata
curl https://auth.example.com/saml/metadata

# Test SSO (requires SAML request from SP)
# Use browser to navigate to SP and initiate SSO
```

## Compliance

SAML 2.0 implementation follows:
- **SAML 2.0 Core Specification**
- **SAML 2.0 Bindings Specification**
- **SAML 2.0 Profiles Specification**

## Additional Resources

- [SAML 2.0 Technical Overview](http://docs.oasis-open.org/security/saml/Post2.0/sstc-saml-tech-overview-2.0.html)
- [SAML 2.0 Core Specification](http://docs.oasis-open.org/security/saml/v2.0/saml-core-2.0-os.pdf)
- [OneLogin SAML Test Tool](https://www.samltool.com/)


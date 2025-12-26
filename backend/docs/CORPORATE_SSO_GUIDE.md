# Corporate SSO Guide

## Overview

This guide explains how to use Auth Gateway as a corporate Single Sign-On (SSO) solution for your organization's infrastructure, including email, VPN, and other services.

## Architecture

Auth Gateway acts as a central Identity Provider (IdP) that provides:

- **OAuth 2.0 / OIDC Provider**: For modern applications
- **SAML 2.0 IdP**: For enterprise applications
- **LDAP/AD Integration**: Sync with existing directories
- **SCIM 2.0**: For automated user provisioning
- **Groups/Departments**: Organizational structure
- **RBAC**: Role-based access control

## Use Cases

### 1. Email Access (OAuth/OIDC)

Configure email clients and webmail to use Auth Gateway:

**Example: Custom Email Client**

```javascript
// OAuth 2.0 Authorization Code Flow
const authUrl = 'https://auth.example.com/oauth/authorize?' +
  'client_id=email-client&' +
  'redirect_uri=https://email.example.com/callback&' +
  'response_type=code&' +
  'scope=openid email profile';

// After user authenticates, exchange code for tokens
const tokenResponse = await fetch('https://auth.example.com/oauth/token', {
  method: 'POST',
  body: JSON.stringify({
    grant_type: 'authorization_code',
    code: authorizationCode,
    client_id: 'email-client',
    client_secret: 'client-secret',
    redirect_uri: 'https://email.example.com/callback'
  })
});
```

### 2. VPN Access (SAML)

Configure VPN to use Auth Gateway as SAML IdP:

**Example: OpenVPN with SAML**

1. **Configure Auth Gateway**:
   ```bash
   POST /api/admin/saml/sp
   {
     "name": "Corporate VPN",
     "entity_id": "https://vpn.corporate.com/saml",
     "acs_url": "https://vpn.corporate.com/saml/acs",
     "enabled": true
   }
   ```

2. **Configure VPN**:
   - Import IdP metadata from `https://auth.example.com/saml/metadata`
   - Configure attribute mappings
   - Test SSO flow

### 3. Application Access (OIDC)

Configure web applications to use Auth Gateway:

**Example: React Application**

```javascript
import { useAuth } from '@auth-gateway/react-sdk';

function App() {
  const { user, login, logout, isAuthenticated } = useAuth({
    issuer: 'https://auth.example.com',
    clientId: 'my-app',
    redirectUri: window.location.origin + '/callback'
  });

  if (!isAuthenticated) {
    return <button onClick={login}>Login</button>;
  }

  return (
    <div>
      <p>Welcome, {user.email}!</p>
      <button onClick={logout}>Logout</button>
    </div>
  );
}
```

## Setup Steps

### Step 1: Configure LDAP/AD Integration

If you have an existing Active Directory or LDAP:

1. **Enable LDAP**:
   ```bash
   LDAP_ENABLED=true
   LDAP_DEFAULT_SERVER=ad.corporate.com
   LDAP_DEFAULT_BASE_DN=dc=corporate,dc=com
   ```

2. **Create LDAP Configuration**:
   ```bash
   POST /api/admin/ldap/config
   {
     "name": "Corporate AD",
     "server": "ad.corporate.com",
     "user_search_filter": "(&(objectClass=user)(objectCategory=person))",
     "sync_enabled": true
   }
   ```

3. **Test and Sync**:
   ```bash
   POST /api/admin/ldap/test-connection
   POST /api/admin/ldap/config/{id}/sync
   ```

### Step 2: Configure Groups/Departments

Organize users into groups:

```bash
POST /api/admin/groups
{
  "name": "Engineering",
  "display_name": "Engineering Department",
  "description": "Software engineering team"
}

POST /api/admin/groups/{id}/members
{
  "user_ids": ["user-id-1", "user-id-2"]
}
```

### Step 3: Configure OAuth/OIDC Clients

Create OAuth clients for applications:

```bash
POST /api/admin/oauth/clients
{
  "name": "Email Client",
  "client_id": "email-client",
  "client_secret": "secure-secret",
  "redirect_uris": ["https://email.example.com/callback"],
  "grant_types": ["authorization_code", "refresh_token"],
  "scopes": ["openid", "email", "profile"]
}
```

### Step 4: Configure SAML Service Providers

For SAML-based services (VPN, legacy apps):

```bash
POST /api/admin/saml/sp
{
  "name": "Corporate VPN",
  "entity_id": "https://vpn.corporate.com/saml",
  "acs_url": "https://vpn.corporate.com/saml/acs",
  "enabled": true
}
```

### Step 5: Enable JIT Provisioning (Optional)

Automatically create users on first login:

```bash
JIT_PROVISIONING_ENABLED=true
```

## Integration Patterns

### Pattern 1: OAuth 2.0 / OIDC for Modern Apps

**Best for**: Web applications, mobile apps, APIs

**Flow**:
1. User clicks "Login with Corporate SSO"
2. Redirected to Auth Gateway
3. User authenticates
4. Redirected back with authorization code
5. App exchanges code for tokens
6. App uses tokens to access user info and APIs

**Example Services**:
- Custom web applications
- Mobile applications
- REST APIs
- GraphQL APIs

### Pattern 2: SAML 2.0 for Enterprise Apps

**Best for**: VPN, legacy applications, Microsoft services

**Flow**:
1. User accesses service (e.g., VPN)
2. Service redirects to Auth Gateway SSO endpoint
3. User authenticates (if not already)
4. Auth Gateway generates SAML assertion
5. User redirected back to service with assertion
6. Service validates assertion and grants access

**Example Services**:
- VPN (OpenVPN, Cisco AnyConnect)
- Jira/Confluence
- Legacy enterprise applications
- Microsoft 365 (if configured as custom IdP)

### Pattern 3: SCIM 2.0 for User Provisioning

**Best for**: Automated user management from HR systems

**Flow**:
1. HR system creates/updates user
2. HR system calls SCIM API
3. Auth Gateway creates/updates user
4. User can immediately access services

**Example Systems**:
- Okta
- Azure AD
- Workday
- Custom HR systems

### Pattern 4: LDAP Sync for Directory Integration

**Best for**: Syncing with existing Active Directory

**Flow**:
1. Auth Gateway connects to AD/LDAP
2. Syncs users and groups periodically
3. Users can authenticate with AD credentials
4. Changes in AD are reflected in Auth Gateway

## User Management

### Bulk Operations

Create/update/delete multiple users:

```bash
POST /api/admin/users/bulk-create
{
  "users": [
    {
      "email": "user1@corporate.com",
      "username": "user1",
      "first_name": "John",
      "last_name": "Doe"
    },
    {
      "email": "user2@corporate.com",
      "username": "user2",
      "first_name": "Jane",
      "last_name": "Smith"
    }
  ]
}

POST /api/admin/users/bulk-assign-roles
{
  "user_ids": ["id1", "id2"],
  "role_ids": ["role-id"]
}
```

### Role Assignment

Assign roles to users or groups:

```bash
POST /api/admin/users/{id}/roles
{
  "role_ids": ["admin-role-id", "user-role-id"]
}
```

## Security Best Practices

1. **Use HTTPS**: Always use HTTPS for all endpoints
2. **Secure Secrets**: Store client secrets and certificates securely
3. **Token Expiry**: Configure appropriate token TTLs
4. **Rate Limiting**: Enable rate limiting to prevent abuse
5. **Audit Logging**: Monitor all authentication and authorization events
6. **Regular Updates**: Keep Auth Gateway updated
7. **Certificate Rotation**: Rotate certificates regularly
8. **Access Control**: Limit admin API access

## Monitoring

### Health Checks

```bash
GET /health      # Overall health
GET /ready       # Readiness probe
GET /live        # Liveness probe
```

### Metrics (Prometheus)

```bash
GET /metrics     # Prometheus metrics
```

Key metrics:
- `auth_gateway_login_total`: Login attempts
- `auth_gateway_active_sessions`: Active sessions
- `auth_gateway_http_request_duration_seconds`: Request latency
- `auth_gateway_database_connections`: DB connection pool

### Audit Logs

All authentication and authorization events are logged:
- User logins/logouts
- Token validations
- Role assignments
- Admin operations

## Troubleshooting

### Common Issues

1. **Users Can't Login**:
   - Check user is active
   - Verify credentials
   - Check LDAP sync status
   - Review audit logs

2. **OAuth Flow Fails**:
   - Verify redirect URI matches configuration
   - Check client secret
   - Review OAuth logs

3. **SAML SSO Not Working**:
   - Verify SP configuration
   - Check certificate validity
   - Review SAML assertion
   - Test with SAML tracer

4. **LDAP Sync Issues**:
   - Test LDAP connection
   - Check attribute mappings
   - Review sync logs

## Example: Complete Corporate Setup

```bash
# 1. Enable features
LDAP_ENABLED=true
SAML_ENABLED=true
OIDC_ENABLED=true
JIT_PROVISIONING_ENABLED=true

# 2. Configure LDAP
POST /api/admin/ldap/config
{
  "name": "Corporate AD",
  "server": "ad.corporate.com",
  "sync_enabled": true
}

# 3. Create groups
POST /api/admin/groups
{"name": "Engineering", "display_name": "Engineering"}
POST /api/admin/groups
{"name": "Sales", "display_name": "Sales"}

# 4. Create OAuth clients
POST /api/admin/oauth/clients
{"name": "Email", "client_id": "email-client", ...}
POST /api/admin/oauth/clients
{"name": "Intranet", "client_id": "intranet-client", ...}

# 5. Configure SAML SPs
POST /api/admin/saml/sp
{"name": "VPN", "entity_id": "https://vpn.corporate.com/saml", ...}
POST /api/admin/saml/sp
{"name": "Jira", "entity_id": "https://jira.corporate.com/saml", ...}

# 6. Sync LDAP
POST /api/admin/ldap/config/{id}/sync
```

## Next Steps

1. **Test Integration**: Test each service integration
2. **User Onboarding**: Onboard users to new SSO system
3. **Monitor**: Set up monitoring and alerting
4. **Documentation**: Document internal procedures
5. **Training**: Train users and administrators

## Support

For issues or questions:
- Review logs: Check application logs for errors
- Check documentation: See other integration guides
- Audit logs: Review authentication events
- Health checks: Verify service health


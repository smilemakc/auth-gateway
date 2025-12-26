# LDAP/Active Directory Integration Guide

## Overview

Auth Gateway supports integration with LDAP and Active Directory for user authentication and synchronization. This allows organizations to use their existing directory services as the source of truth for user accounts.

## Features

- **LDAP Authentication**: Users can authenticate using their LDAP/AD credentials
- **User Synchronization**: Automatic synchronization of users and groups from LDAP/AD
- **Multiple LDAP Configurations**: Support for multiple LDAP servers/configurations
- **Attribute Mapping**: Flexible mapping of LDAP attributes to user fields
- **Change Detection**: Automatic detection of changes in LDAP directory
- **Sync Logging**: Comprehensive logging of synchronization operations

## Configuration

### Environment Variables

```bash
# Enable LDAP integration
LDAP_ENABLED=true

# Default LDAP server settings (can be overridden per-config in database)
LDAP_DEFAULT_SERVER=ldap.example.com
LDAP_DEFAULT_PORT=389
LDAP_DEFAULT_BASE_DN=dc=example,dc=com
LDAP_DEFAULT_BIND_DN=cn=admin,dc=example,dc=com
LDAP_DEFAULT_BIND_PASSWORD=secret

# Sync settings
LDAP_SYNC_INTERVAL=1h
LDAP_AUTO_SYNC_ENABLED=true
```

### API Configuration

You can create and manage LDAP configurations via the Admin API:

#### Create LDAP Configuration

```bash
POST /api/admin/ldap/config
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "Corporate AD",
  "server": "ldap.corporate.com",
  "port": 389,
  "base_dn": "dc=corporate,dc=com",
  "bind_dn": "cn=auth-gateway,ou=service-accounts,dc=corporate,dc=com",
  "bind_password": "secret",
  "user_search_base": "ou=users,dc=corporate,dc=com",
  "user_search_filter": "(objectClass=person)",
  "user_attributes": {
    "email": "mail",
    "username": "sAMAccountName",
    "first_name": "givenName",
    "last_name": "sn",
    "phone": "telephoneNumber"
  },
  "group_search_base": "ou=groups,dc=corporate,dc=com",
  "group_search_filter": "(objectClass=group)",
  "group_attributes": {
    "name": "cn",
    "description": "description"
  },
  "sync_enabled": true,
  "sync_interval": "1h"
}
```

#### Test Connection

```bash
POST /api/admin/ldap/test-connection
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "server": "ldap.corporate.com",
  "port": 389,
  "bind_dn": "cn=auth-gateway,ou=service-accounts,dc=corporate,dc=com",
  "bind_password": "secret"
}
```

#### Manual Sync

```bash
POST /api/admin/ldap/config/{id}/sync
Authorization: Bearer <admin_token>
```

#### Get Sync Logs

```bash
GET /api/admin/ldap/config/{id}/sync-logs
Authorization: Bearer <admin_token>
```

## Attribute Mapping

### User Attributes

Map LDAP attributes to Auth Gateway user fields:

- `email` → User email (required)
- `username` → Username (required)
- `first_name` → First name
- `last_name` → Last name
- `phone` → Phone number
- `external_id` → External identifier (e.g., employee ID)

### Group Attributes

Map LDAP group attributes:

- `name` → Group name (required)
- `description` → Group description
- `external_id` → External group identifier

## Synchronization

### Automatic Sync

When `LDAP_AUTO_SYNC_ENABLED=true`, Auth Gateway automatically synchronizes users and groups at the configured interval.

### Manual Sync

You can trigger a manual sync via the API:

```bash
POST /api/admin/ldap/config/{id}/sync
```

### Sync Process

1. **Connect to LDAP**: Establish connection using configured credentials
2. **Search Users**: Query users based on search filter
3. **Map Attributes**: Map LDAP attributes to user fields
4. **Create/Update Users**: Create new users or update existing ones
5. **Search Groups**: Query groups based on search filter
6. **Map Group Attributes**: Map LDAP group attributes
7. **Create/Update Groups**: Create new groups or update existing ones
8. **Sync Memberships**: Update user-group relationships
9. **Log Results**: Record sync statistics and errors

### Change Detection

The sync process detects:
- New users in LDAP → Creates users in Auth Gateway
- Updated user attributes → Updates existing users
- Deleted users in LDAP → Optionally deactivates users (configurable)
- New groups → Creates groups
- Updated groups → Updates groups
- Group membership changes → Updates user-group relationships

## Authentication Flow

1. User attempts to login with LDAP credentials
2. Auth Gateway connects to LDAP server
3. Binds with user's credentials
4. If successful, creates or updates user in Auth Gateway
5. Returns JWT tokens for Auth Gateway API

## Best Practices

1. **Use Service Account**: Create a dedicated service account in LDAP/AD for Auth Gateway
2. **Limit Permissions**: Grant only necessary read permissions to the service account
3. **Secure Credentials**: Store LDAP bind password securely (use secrets management)
4. **Test Connection**: Always test LDAP connection before enabling sync
5. **Monitor Sync Logs**: Regularly review sync logs for errors
6. **Sync Frequency**: Balance sync frequency with LDAP server load
7. **Attribute Mapping**: Ensure attribute mappings match your LDAP schema
8. **Backup**: Keep backups of LDAP configurations

## Troubleshooting

### Connection Issues

- Verify LDAP server address and port
- Check firewall rules
- Validate bind DN and password
- Test with `ldapsearch` command-line tool

### Sync Issues

- Review sync logs for specific errors
- Check attribute mappings
- Verify search filters
- Ensure service account has read permissions

### Authentication Issues

- Verify user credentials
- Check user search filter includes the user
- Ensure user is not disabled in LDAP/AD
- Review attribute mappings

## Example: Active Directory Integration

```json
{
  "name": "Corporate Active Directory",
  "server": "ad.corporate.com",
  "port": 389,
  "base_dn": "dc=corporate,dc=com",
  "bind_dn": "CN=AuthGateway,CN=Service Accounts,DC=corporate,DC=com",
  "bind_password": "SecurePassword123",
  "user_search_base": "CN=Users,DC=corporate,DC=com",
  "user_search_filter": "(&(objectClass=user)(objectCategory=person))",
  "user_attributes": {
    "email": "mail",
    "username": "sAMAccountName",
    "first_name": "givenName",
    "last_name": "sn",
    "phone": "telephoneNumber",
    "external_id": "employeeID"
  },
  "group_search_base": "CN=Groups,DC=corporate,DC=com",
  "group_search_filter": "(objectClass=group)",
  "group_attributes": {
    "name": "cn",
    "description": "description",
    "external_id": "objectGUID"
  },
  "sync_enabled": true,
  "sync_interval": "30m"
}
```

## Security Considerations

1. **TLS/SSL**: Use LDAPS (LDAP over SSL) for secure connections
2. **Password Storage**: Store bind passwords encrypted
3. **Access Control**: Limit LDAP service account permissions
4. **Audit Logging**: Enable audit logging for sync operations
5. **Network Security**: Use VPN or private network for LDAP connections


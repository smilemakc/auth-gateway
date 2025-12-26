# SCIM 2.0 Integration Guide

## Overview

Auth Gateway implements the SCIM 2.0 (System for Cross-domain Identity Management) protocol, allowing integration with identity providers and user management systems like Okta, Azure AD, and other SCIM-compatible services.

## Features

- **SCIM 2.0 Compliant**: Full support for SCIM 2.0 specification
- **User Management**: Create, read, update, and delete users via SCIM
- **Group Management**: Manage groups and group memberships
- **Bulk Operations**: Support for bulk operations
- **Filtering & Pagination**: Query users and groups with filters and pagination
- **Service Provider Config**: Exposes service provider capabilities

## Authentication

SCIM endpoints require authentication via Bearer token. You can use:
- OAuth 2.0 access token
- API key with appropriate permissions

## Endpoints

### Service Provider Configuration

#### Get Service Provider Config

```http
GET /scim/v2/ServiceProviderConfig
Authorization: Bearer <token>
```

Returns SCIM service provider capabilities and configuration.

#### Get Schemas

```http
GET /scim/v2/Schemas
Authorization: Bearer <token>
```

Returns available SCIM schemas.

### User Management

#### List Users

```http
GET /scim/v2/Users?filter=userName eq "john.doe"&startIndex=1&count=100
Authorization: Bearer <token>
```

Query parameters:
- `filter`: SCIM filter expression
- `startIndex`: Pagination start index (1-based)
- `count`: Number of results per page

#### Get User

```http
GET /scim/v2/Users/{id}
Authorization: Bearer <token>
```

#### Create User

```http
POST /scim/v2/Users
Authorization: Bearer <token>
Content-Type: application/scim+json

{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "john.doe",
  "name": {
    "givenName": "John",
    "familyName": "Doe"
  },
  "emails": [
    {
      "value": "john.doe@example.com",
      "primary": true
    }
  ],
  "active": true
}
```

#### Update User (Full)

```http
PUT /scim/v2/Users/{id}
Authorization: Bearer <token>
Content-Type: application/scim+json

{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "userName": "john.doe",
  "name": {
    "givenName": "John",
    "familyName": "Doe"
  },
  "emails": [
    {
      "value": "john.doe@example.com",
      "primary": true
    }
  ],
  "active": true
}
```

#### Update User (Partial)

```http
PATCH /scim/v2/Users/{id}
Authorization: Bearer <token>
Content-Type: application/scim+json

{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
  "Operations": [
    {
      "op": "replace",
      "path": "active",
      "value": false
    },
    {
      "op": "add",
      "path": "emails",
      "value": [
        {
          "value": "john.new@example.com",
          "primary": true
        }
      ]
    }
  ]
}
```

Supported operations:
- `add`: Add attribute or value
- `remove`: Remove attribute or value
- `replace`: Replace attribute value

#### Delete User

```http
DELETE /scim/v2/Users/{id}
Authorization: Bearer <token>
```

### Group Management

#### List Groups

```http
GET /scim/v2/Groups?filter=displayName eq "Developers"&startIndex=1&count=100
Authorization: Bearer <token>
```

#### Get Group

```http
GET /scim/v2/Groups/{id}
Authorization: Bearer <token>
```

## Attribute Mapping

### User Attributes

| SCIM Attribute | Auth Gateway Field | Notes |
|---------------|-------------------|-------|
| `id` | `id` | UUID |
| `userName` | `username` | Required, unique |
| `name.givenName` | `first_name` | |
| `name.familyName` | `last_name` | |
| `emails[].value` | `email` | Primary email used |
| `emails[].primary` | - | Used to determine primary email |
| `phoneNumbers[].value` | `phone` | Primary phone used |
| `active` | `is_active` | Boolean |
| `externalId` | `external_id` | External identifier |

### Group Attributes

| SCIM Attribute | Auth Gateway Field | Notes |
|---------------|-------------------|-------|
| `id` | `id` | UUID |
| `displayName` | `name` | Required |
| `members[].value` | User IDs | Group members |

## Filtering

SCIM supports filtering with operators:

- `eq`: Equals
- `ne`: Not equals
- `co`: Contains
- `sw`: Starts with
- `pr`: Present (has value)
- `gt`: Greater than
- `ge`: Greater than or equal
- `lt`: Less than
- `le`: Less than or equal

Examples:
```
userName eq "john.doe"
active eq true
emails.value co "@example.com"
name.familyName sw "Doe"
```

## Pagination

Use `startIndex` and `count` parameters:

```
GET /scim/v2/Users?startIndex=1&count=50
```

Response includes:
- `totalResults`: Total number of results
- `startIndex`: Current start index
- `itemsPerPage`: Number of items per page
- `Resources`: Array of resources

## Bulk Operations

SCIM 2.0 supports bulk operations, but Auth Gateway processes them sequentially for data consistency.

## Integration Examples

### Okta Integration

1. In Okta, go to Applications → Create App Integration
2. Choose SAML 2.0 or OIDC
3. Configure SCIM settings:
   - SCIM connector base URL: `https://auth.example.com/scim/v2`
   - Unique identifier field: `userName`
   - Supported provisioning actions: Create, Update, Deactivate
4. Configure attribute mappings
5. Test connection and sync

### Azure AD Integration

1. In Azure AD, go to Enterprise Applications
2. Add new application → Non-gallery application
3. Configure Provisioning:
   - Provisioning Mode: Automatic
   - Tenant URL: `https://auth.example.com/scim/v2`
   - Secret Token: Generate and configure
4. Configure attribute mappings
5. Start provisioning

## Error Handling

SCIM uses standard HTTP status codes:

- `200 OK`: Success
- `201 Created`: Resource created
- `204 No Content`: Resource deleted
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (e.g., duplicate)
- `500 Internal Server Error`: Server error

Error response format:
```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "detail": "Error message",
  "status": "400"
}
```

## Best Practices

1. **Use External IDs**: Map external identifiers for reliable user matching
2. **Handle Deactivations**: Use `active=false` instead of deleting users
3. **Idempotency**: Ensure operations are idempotent
4. **Rate Limiting**: Respect rate limits for bulk operations
5. **Error Handling**: Implement retry logic for transient errors
6. **Logging**: Monitor SCIM operations for troubleshooting

## Security Considerations

1. **Authentication**: Always use HTTPS and secure tokens
2. **Authorization**: Verify user has appropriate permissions
3. **Input Validation**: Validate all SCIM requests
4. **Audit Logging**: Log all SCIM operations
5. **Rate Limiting**: Implement rate limiting to prevent abuse

## Testing

Use SCIM client libraries or tools:

- **Postman**: Import SCIM collection
- **SCIM Test Client**: Use SCIM 2.0 test clients
- **cURL**: Manual testing with curl commands

Example:
```bash
curl -X GET "https://auth.example.com/scim/v2/Users" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/scim+json"
```


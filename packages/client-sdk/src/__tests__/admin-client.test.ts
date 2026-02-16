import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { HttpClient } from '../core/http';
import { AdminUsersService } from '../services/admin/users';
import { AdminApplicationsService } from '../services/admin/applications';
import { AdminRBACService } from '../services/admin/rbac';
import { AdminSystemService } from '../services/admin/system';
import type { AdminUserResponse, AdminUserListResponse } from '../types/user';
import type {
  Application,
  ApplicationListResponse,
  CreateApplicationResponse,
  SystemHealthResponse,
  GeoDistributionResponse,
  MaintenanceModeResponse,
} from '../types/admin';
import type { Role, Permission, PermissionMatrix } from '../types/rbac';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const TEST_BASE_URL = 'https://api.example.com';

function createHttpClient(): HttpClient {
  return new HttpClient({
    baseUrl: TEST_BASE_URL,
    autoRefreshTokens: false,
    retry: { maxRetries: 0 },
  });
}

function mockFetchJsonResponse(data: unknown, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    headers: new Headers({ 'Content-Type': 'application/json' }),
    json: () => Promise.resolve(data),
  };
}

function createMockAdminUser(overrides: Partial<AdminUserResponse> = {}): AdminUserResponse {
  return {
    id: 'user-1',
    email: 'admin@example.com',
    username: 'admin',
    full_name: 'Admin User',
    roles: [{ id: 'role-admin', name: 'admin', display_name: 'Administrator' }],
    account_type: 'human',
    email_verified: true,
    phone_verified: false,
    is_active: true,
    totp_enabled: false,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    ...overrides,
  };
}

function createMockApplication(overrides: Partial<Application> = {}): Application {
  return {
    id: 'app-1',
    name: 'test-app',
    display_name: 'Test Application',
    description: 'A test application',
    allowed_auth_methods: ['password', 'otp_email'],
    is_active: true,
    is_system: false,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    ...overrides,
  };
}

function createMockRole(overrides: Partial<Role> = {}): Role {
  return {
    id: 'role-1',
    name: 'editor',
    display_name: 'Editor',
    description: 'Can edit content',
    is_system_role: false,
    permissions: [],
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    ...overrides,
  };
}

function createMockPermission(overrides: Partial<Permission> = {}): Permission {
  return {
    id: 'perm-1',
    name: 'articles:read',
    resource: 'articles',
    action: 'read',
    description: 'Read articles',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// AdminUsersService Tests
// ---------------------------------------------------------------------------

describe('AdminUsersService', () => {
  let http: HttpClient;
  let service: AdminUsersService;
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    // @ts-ignore
    global.fetch = fetchMock;
    http = createHttpClient();
    service = new AdminUsersService(http);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('getStats', () => {
    it('should return admin dashboard statistics', async () => {
      const mockStats = {
        total_users: 150,
        active_users: 120,
        new_users_today: 5,
        total_api_keys: 30,
        active_api_keys: 25,
        login_attempts_today: 200,
        failed_login_attempts_today: 10,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockStats));

      const stats = await service.getStats();

      expect(stats.total_users).toBe(150);
      expect(stats.active_users).toBe(120);
      expect(stats.new_users_today).toBe(5);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/stats`);
    });
  });

  describe('list', () => {
    it('should list users with default pagination', async () => {
      const mockList: AdminUserListResponse = {
        users: [createMockAdminUser()],
        total: 1,
        page: 1,
        page_size: 20,
        total_pages: 1,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockList));

      const result = await service.list();

      expect(result.users).toHaveLength(1);
      expect(result.total).toBe(1);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toContain('/api/admin/users');
      expect(url).toContain('page=1');
      expect(url).toContain('page_size=20');
    });

    it('should support custom pagination', async () => {
      const mockList: AdminUserListResponse = {
        users: [],
        total: 0,
        page: 3,
        page_size: 50,
        total_pages: 0,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockList));

      await service.list(3, 50);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toContain('page=3');
      expect(url).toContain('page_size=50');
    });
  });

  describe('get', () => {
    it('should get user by ID', async () => {
      const mockUser = createMockAdminUser({ id: 'user-42' });

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockUser));

      const result = await service.get('user-42');

      expect(result.id).toBe('user-42');

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/users/user-42`);
    });
  });

  describe('create', () => {
    it('should create a new user', async () => {
      const newUser = createMockAdminUser({
        id: 'user-new',
        email: 'new@example.com',
        username: 'newuser',
      });

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(newUser));

      const result = await service.create({
        email: 'new@example.com',
        username: 'newuser',
        password: 'Pass123!',
        full_name: 'New User',
      });

      expect(result.id).toBe('user-new');
      expect(result.email).toBe('new@example.com');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/users`);
      expect(options.method).toBe('POST');
    });
  });

  describe('update', () => {
    it('should update a user', async () => {
      const updatedUser = createMockAdminUser({ is_active: false });

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(updatedUser));

      const result = await service.update('user-1', { is_active: false });

      expect(result.is_active).toBe(false);

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/users/user-1`);
      expect(options.method).toBe('PUT');
    });
  });

  describe('delete', () => {
    it('should delete a user', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ message: 'User deleted' }));

      const result = await service.delete('user-1');

      expect(result.message).toBe('User deleted');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/users/user-1`);
      expect(options.method).toBe('DELETE');
    });
  });

  describe('activate / deactivate', () => {
    it('should activate a user', async () => {
      const activeUser = createMockAdminUser({ is_active: true });
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(activeUser));

      const result = await service.activate('user-1');

      expect(result.is_active).toBe(true);

      const [, options] = fetchMock.mock.calls[0]!;
      expect(JSON.parse(options.body)).toEqual({ is_active: true });
    });

    it('should deactivate a user', async () => {
      const inactiveUser = createMockAdminUser({ is_active: false });
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(inactiveUser));

      const result = await service.deactivate('user-1');

      expect(result.is_active).toBe(false);

      const [, options] = fetchMock.mock.calls[0]!;
      expect(JSON.parse(options.body)).toEqual({ is_active: false });
    });
  });

  describe('setRoles', () => {
    it('should set roles for a user', async () => {
      const updatedUser = createMockAdminUser();
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(updatedUser));

      await service.setRoles('user-1', ['role-a', 'role-b']);

      const [, options] = fetchMock.mock.calls[0]!;
      expect(JSON.parse(options.body)).toEqual({ role_ids: ['role-a', 'role-b'] });
    });
  });

  describe('assignRole / removeRole', () => {
    it('should assign a role to a user', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ message: 'Role assigned' }));

      const result = await service.assignRole('user-1', 'role-admin');

      expect(result.message).toBe('Role assigned');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/users/user-1/roles`);
      expect(options.method).toBe('POST');
      expect(JSON.parse(options.body)).toEqual({ role_id: 'role-admin' });
    });

    it('should remove a role from a user', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ message: 'Role removed' }));

      const result = await service.removeRole('user-1', 'role-admin');

      expect(result.message).toBe('Role removed');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/users/user-1/roles/role-admin`);
      expect(options.method).toBe('DELETE');
    });
  });

  describe('reset2FA', () => {
    it('should reset 2FA for a user', async () => {
      fetchMock.mockResolvedValueOnce(
        mockFetchJsonResponse({ message: '2FA reset', user_id: 'user-1' })
      );

      const result = await service.reset2FA('user-1');

      expect(result.message).toBe('2FA reset');
      expect(result.user_id).toBe('user-1');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/users/user-1/reset-2fa`);
      expect(options.method).toBe('POST');
    });
  });

  describe('sendPasswordReset', () => {
    it('should send password reset email for a user', async () => {
      fetchMock.mockResolvedValueOnce(
        mockFetchJsonResponse({ message: 'Password reset email sent', email: 'admin@example.com' })
      );

      const result = await service.sendPasswordReset('user-1');

      expect(result.message).toBe('Password reset email sent');
      expect(result.email).toBe('admin@example.com');
    });
  });

  describe('getOAuthAccounts', () => {
    it('should return OAuth accounts for a user', async () => {
      const mockAccounts = {
        accounts: [
          { provider: 'google', provider_user_id: 'google-123', email: 'user@gmail.com' },
        ],
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockAccounts));

      const result = await service.getOAuthAccounts('user-1');

      expect(result.accounts).toHaveLength(1);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/users/user-1/oauth-accounts`);
    });
  });

  describe('error handling', () => {
    it('should throw NotFoundError for 404', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ message: 'User not found' }, 404));

      await expect(service.get('nonexistent')).rejects.toThrow('User not found');
    });

    it('should throw AuthorizationError for 403', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ message: 'Forbidden' }, 403));

      await expect(service.getStats()).rejects.toThrow('Forbidden');
    });
  });
});

// ---------------------------------------------------------------------------
// AdminApplicationsService Tests
// ---------------------------------------------------------------------------

describe('AdminApplicationsService', () => {
  let http: HttpClient;
  let service: AdminApplicationsService;
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    // @ts-ignore
    global.fetch = fetchMock;
    http = createHttpClient();
    service = new AdminApplicationsService(http);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('create', () => {
    it('should create a new application', async () => {
      const mockResponse: CreateApplicationResponse = {
        application: createMockApplication({ id: 'app-new' }),
        secret: 'app_secret_abc123',
        warning: 'Store this secret securely',
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockResponse));

      const result = await service.create({
        name: 'my-app',
        display_name: 'My Application',
      });

      expect(result.application.id).toBe('app-new');
      expect(result.secret).toBe('app_secret_abc123');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/applications`);
      expect(options.method).toBe('POST');
    });
  });

  describe('list', () => {
    it('should list applications with default pagination', async () => {
      const mockList: ApplicationListResponse = {
        applications: [createMockApplication()],
        total: 1,
        page: 1,
        page_size: 20,
        total_pages: 1,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockList));

      const result = await service.list();

      expect(result.applications).toHaveLength(1);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toContain('/api/admin/applications');
      expect(url).toContain('page=1');
      expect(url).toContain('page_size=20');
    });

    it('should filter by active status', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({
        applications: [],
        total: 0,
        page: 1,
        page_size: 20,
        total_pages: 0,
      }));

      await service.list(1, 20, true);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toContain('is_active=true');
    });
  });

  describe('getById', () => {
    it('should get an application by ID', async () => {
      const mockApp = createMockApplication({ id: 'app-42' });

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockApp));

      const result = await service.getById('app-42');

      expect(result.id).toBe('app-42');

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/applications/app-42`);
    });
  });

  describe('update', () => {
    it('should update an application', async () => {
      const updatedApp = createMockApplication({ display_name: 'Updated App' });

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(updatedApp));

      const result = await service.update('app-1', { display_name: 'Updated App' });

      expect(result.display_name).toBe('Updated App');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/applications/app-1`);
      expect(options.method).toBe('PUT');
    });
  });

  describe('remove', () => {
    it('should delete an application', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ message: 'Application deleted' }));

      const result = await service.remove('app-1');

      expect(result.message).toBe('Application deleted');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/applications/app-1`);
      expect(options.method).toBe('DELETE');
    });
  });

  describe('rotateSecret', () => {
    it('should rotate the application secret', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({
        secret: 'new_app_secret_xyz',
        warning: 'Old secret is now invalid',
      }));

      const result = await service.rotateSecret('app-1');

      expect(result.secret).toBe('new_app_secret_xyz');
      expect(result.warning).toBe('Old secret is now invalid');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/applications/app-1/rotate-secret`);
      expect(options.method).toBe('POST');
    });
  });

  describe('branding', () => {
    it('should get application branding', async () => {
      const mockBranding = {
        id: 'brand-1',
        application_id: 'app-1',
        primary_color: '#FF0000',
        company_name: 'Acme Corp',
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockBranding));

      const result = await service.getBranding('app-1');

      expect(result.primary_color).toBe('#FF0000');

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/applications/app-1/branding`);
    });

    it('should update application branding', async () => {
      const updatedBranding = {
        id: 'brand-1',
        application_id: 'app-1',
        primary_color: '#00FF00',
        company_name: 'New Corp',
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(updatedBranding));

      const result = await service.updateBranding('app-1', {
        primary_color: '#00FF00',
        company_name: 'New Corp',
      });

      expect(result.primary_color).toBe('#00FF00');
    });
  });

  describe('user management within applications', () => {
    it('should list users of an application', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({
        profiles: [{ id: 'profile-1', user_id: 'user-1', application_id: 'app-1' }],
        total: 1,
        page: 1,
        page_size: 20,
        total_pages: 1,
      }));

      const result = await service.listUsers('app-1');

      expect(result.profiles).toHaveLength(1);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toContain('/api/admin/applications/app-1/users');
    });

    it('should ban a user from an application', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ message: 'User banned' }));

      const result = await service.banUser('app-1', 'user-1', { reason: 'Spamming' });

      expect(result.message).toBe('User banned');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/applications/app-1/users/user-1/ban`);
      expect(JSON.parse(options.body)).toEqual({ reason: 'Spamming' });
    });

    it('should unban a user from an application', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ message: 'User unbanned' }));

      const result = await service.unbanUser('app-1', 'user-1');

      expect(result.message).toBe('User unbanned');

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/applications/app-1/users/user-1/unban`);
    });
  });

  describe('auth config', () => {
    it('should get auth configuration for an application', async () => {
      const mockConfig = {
        application_id: 'app-1',
        name: 'test-app',
        display_name: 'Test Application',
        allowed_auth_methods: ['password', 'otp_email'],
        oauth_providers: ['google'],
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockConfig));

      const result = await service.getAuthConfig('app-1');

      expect(result.application_id).toBe('app-1');
      expect(result.allowed_auth_methods).toContain('password');

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/applications/app-1/auth-config`);
    });
  });

  describe('email templates', () => {
    it('should list email templates for an application', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({
        templates: [{ id: 'tpl-1', type: 'welcome', name: 'Welcome', subject: 'Welcome!', html_body: '<h1>Hi</h1>', variables: [], is_active: true }],
        total: 1,
        page: 1,
        page_size: 20,
        total_pages: 1,
      }));

      const result = await service.listTemplates('app-1');

      expect(result.templates).toHaveLength(1);
      expect(result.templates[0]!.type).toBe('welcome');
    });

    it('should create an email template', async () => {
      const mockTemplate = {
        id: 'tpl-new',
        type: 'custom',
        name: 'Custom',
        subject: 'Custom Email',
        html_body: '<p>Body</p>',
        variables: ['name'],
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockTemplate));

      const result = await service.createTemplate('app-1', {
        type: 'custom',
        name: 'Custom',
        subject: 'Custom Email',
        html_body: '<p>Body</p>',
      });

      expect(result.id).toBe('tpl-new');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/applications/app-1/email-templates`);
      expect(options.method).toBe('POST');
    });

    it('should delete an email template', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ message: 'Template deleted' }));

      const result = await service.deleteTemplate('app-1', 'tpl-1');

      expect(result.message).toBe('Template deleted');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/applications/app-1/email-templates/tpl-1`);
      expect(options.method).toBe('DELETE');
    });
  });
});

// ---------------------------------------------------------------------------
// AdminRBACService Tests
// ---------------------------------------------------------------------------

describe('AdminRBACService', () => {
  let http: HttpClient;
  let service: AdminRBACService;
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    // @ts-ignore
    global.fetch = fetchMock;
    http = createHttpClient();
    service = new AdminRBACService(http);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  // ---- Permissions ----

  describe('listPermissions', () => {
    it('should return a list of permissions', async () => {
      const mockPerms = {
        permissions: [
          createMockPermission({ id: 'perm-1', name: 'articles:read' }),
          createMockPermission({ id: 'perm-2', name: 'articles:write', action: 'write' }),
        ],
        total: 2,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockPerms));

      const result = await service.listPermissions();

      expect(result).toHaveLength(2);
      expect(result[0]!.name).toBe('articles:read');

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/rbac/permissions`);
    });
  });

  describe('getPermission', () => {
    it('should get a permission by ID', async () => {
      const mockPerm = createMockPermission({ id: 'perm-42' });

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockPerm));

      const result = await service.getPermission('perm-42');

      expect(result.id).toBe('perm-42');
    });
  });

  describe('createPermission', () => {
    it('should create a new permission', async () => {
      const newPerm = createMockPermission({ id: 'perm-new', name: 'users:delete' });

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(newPerm));

      const result = await service.createPermission({
        name: 'users:delete',
        resource: 'users',
        action: 'delete',
        description: 'Delete users',
      });

      expect(result.name).toBe('users:delete');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/rbac/permissions`);
      expect(options.method).toBe('POST');
    });
  });

  describe('updatePermission', () => {
    it('should update a permission', async () => {
      const updatedPerm = createMockPermission({ description: 'Updated desc' });

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(updatedPerm));

      const result = await service.updatePermission('perm-1', { description: 'Updated desc' });

      expect(result.description).toBe('Updated desc');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/rbac/permissions/perm-1`);
      expect(options.method).toBe('PUT');
    });
  });

  describe('deletePermission', () => {
    it('should delete a permission', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ message: 'Permission deleted' }));

      const result = await service.deletePermission('perm-1');

      expect(result.message).toBe('Permission deleted');

      const [, options] = fetchMock.mock.calls[0]!;
      expect(options.method).toBe('DELETE');
    });
  });

  describe('getPermissionsByResource', () => {
    it('should filter permissions by resource', async () => {
      const mockPerms = {
        permissions: [
          createMockPermission({ id: 'p1', resource: 'articles', action: 'read' }),
          createMockPermission({ id: 'p2', resource: 'articles', action: 'write' }),
          createMockPermission({ id: 'p3', resource: 'users', action: 'read' }),
        ],
        total: 3,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockPerms));

      const result = await service.getPermissionsByResource('articles');

      expect(result).toHaveLength(2);
      expect(result.every(p => p.resource === 'articles')).toBe(true);
    });
  });

  // ---- Roles ----

  describe('listRoles', () => {
    it('should return a list of roles', async () => {
      const mockRoles = {
        roles: [
          createMockRole({ id: 'role-1', name: 'admin' }),
          createMockRole({ id: 'role-2', name: 'user' }),
        ],
        total: 2,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockRoles));

      const result = await service.listRoles();

      expect(result).toHaveLength(2);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/rbac/roles`);
    });
  });

  describe('getRole', () => {
    it('should get a role by ID', async () => {
      const mockRole = createMockRole({
        id: 'role-editor',
        permissions: [createMockPermission()],
      });

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockRole));

      const result = await service.getRole('role-editor');

      expect(result.id).toBe('role-editor');
      expect(result.permissions).toHaveLength(1);
    });
  });

  describe('createRole', () => {
    it('should create a new role', async () => {
      const newRole = createMockRole({ id: 'role-new', name: 'moderator' });

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(newRole));

      const result = await service.createRole({
        name: 'moderator',
        display_name: 'Moderator',
        permissions: ['perm-1', 'perm-2'],
      });

      expect(result.name).toBe('moderator');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/rbac/roles`);
      expect(options.method).toBe('POST');
    });
  });

  describe('updateRole', () => {
    it('should update a role', async () => {
      const updatedRole = createMockRole({ display_name: 'Senior Editor' });

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(updatedRole));

      const result = await service.updateRole('role-1', { display_name: 'Senior Editor' });

      expect(result.display_name).toBe('Senior Editor');
    });
  });

  describe('deleteRole', () => {
    it('should delete a role', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ message: 'Role deleted' }));

      const result = await service.deleteRole('role-1');

      expect(result.message).toBe('Role deleted');
    });
  });

  describe('addPermissionsToRole', () => {
    it('should merge new permissions with existing ones', async () => {
      const existingRole = createMockRole({
        id: 'role-1',
        permissions: [
          createMockPermission({ id: 'perm-a' }),
          createMockPermission({ id: 'perm-b' }),
        ],
      });

      // First call: getRole
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(existingRole));
      // Second call: updateRole
      const updatedRole = createMockRole({
        id: 'role-1',
        permissions: [
          createMockPermission({ id: 'perm-a' }),
          createMockPermission({ id: 'perm-b' }),
          createMockPermission({ id: 'perm-c' }),
        ],
      });
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(updatedRole));

      const result = await service.addPermissionsToRole('role-1', ['perm-b', 'perm-c']);

      expect(result.permissions).toHaveLength(3);

      // Verify the update request has merged permissions (no duplicates)
      const [, updateOptions] = fetchMock.mock.calls[1]!;
      const updateBody = JSON.parse(updateOptions.body);
      expect(updateBody.permissions).toEqual(['perm-a', 'perm-b', 'perm-c']);
    });
  });

  describe('removePermissionsFromRole', () => {
    it('should remove specified permissions from the role', async () => {
      const existingRole = createMockRole({
        id: 'role-1',
        permissions: [
          createMockPermission({ id: 'perm-a' }),
          createMockPermission({ id: 'perm-b' }),
          createMockPermission({ id: 'perm-c' }),
        ],
      });

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(existingRole));
      const updatedRole = createMockRole({
        id: 'role-1',
        permissions: [createMockPermission({ id: 'perm-a' })],
      });
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(updatedRole));

      const result = await service.removePermissionsFromRole('role-1', ['perm-b', 'perm-c']);

      expect(result.permissions).toHaveLength(1);

      const [, updateOptions] = fetchMock.mock.calls[1]!;
      const updateBody = JSON.parse(updateOptions.body);
      expect(updateBody.permissions).toEqual(['perm-a']);
    });
  });

  // ---- Permission Matrix ----

  describe('getPermissionMatrix', () => {
    it('should return the permission matrix', async () => {
      const mockMatrix: PermissionMatrix = {
        roles: ['admin', 'user'],
        resources: ['articles'],
        actions: ['read', 'write'],
        matrix: [
          { role: 'admin', permissions: { 'articles:read': true, 'articles:write': true } },
          { role: 'user', permissions: { 'articles:read': true, 'articles:write': false } },
        ],
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockMatrix));

      const result = await service.getPermissionMatrix();

      expect(result.roles).toEqual(['admin', 'user']);
      expect(result.matrix).toHaveLength(2);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/rbac/permission-matrix`);
    });
  });

  describe('roleHasPermission', () => {
    it('should return true when role has the permission', async () => {
      const mockMatrix: PermissionMatrix = {
        roles: ['admin'],
        resources: ['articles'],
        actions: ['read'],
        matrix: [
          { role: 'admin', permissions: { 'articles:read': true } },
        ],
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockMatrix));

      const result = await service.roleHasPermission('admin', 'articles:read');

      expect(result).toBe(true);
    });

    it('should return false when role does not have the permission', async () => {
      const mockMatrix: PermissionMatrix = {
        roles: ['user'],
        resources: ['articles'],
        actions: ['write'],
        matrix: [
          { role: 'user', permissions: { 'articles:write': false } },
        ],
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockMatrix));

      const result = await service.roleHasPermission('user', 'articles:write');

      expect(result).toBe(false);
    });

    it('should return false when role is not in the matrix', async () => {
      const mockMatrix: PermissionMatrix = {
        roles: ['admin'],
        resources: [],
        actions: [],
        matrix: [{ role: 'admin', permissions: {} }],
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockMatrix));

      const result = await service.roleHasPermission('nonexistent', 'anything');

      expect(result).toBe(false);
    });
  });

  describe('getCustomRoles / getSystemRoles', () => {
    it('should filter custom (non-system) roles', async () => {
      const mockRoles = {
        roles: [
          createMockRole({ id: 'r1', name: 'admin', is_system_role: true }),
          createMockRole({ id: 'r2', name: 'user', is_system_role: true }),
          createMockRole({ id: 'r3', name: 'editor', is_system_role: false }),
        ],
        total: 3,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockRoles));

      const customRoles = await service.getCustomRoles();

      expect(customRoles).toHaveLength(1);
      expect(customRoles[0]!.name).toBe('editor');
    });

    it('should filter system roles', async () => {
      const mockRoles = {
        roles: [
          createMockRole({ id: 'r1', name: 'admin', is_system_role: true }),
          createMockRole({ id: 'r2', name: 'user', is_system_role: true }),
          createMockRole({ id: 'r3', name: 'editor', is_system_role: false }),
        ],
        total: 3,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockRoles));

      const systemRoles = await service.getSystemRoles();

      expect(systemRoles).toHaveLength(2);
      expect(systemRoles.every(r => r.is_system_role)).toBe(true);
    });
  });
});

// ---------------------------------------------------------------------------
// AdminSystemService Tests
// ---------------------------------------------------------------------------

describe('AdminSystemService', () => {
  let http: HttpClient;
  let service: AdminSystemService;
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    // @ts-ignore
    global.fetch = fetchMock;
    http = createHttpClient();
    service = new AdminSystemService(http);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('getHealth', () => {
    it('should return system health status', async () => {
      const mockHealth: SystemHealthResponse = {
        status: 'healthy',
        services: { database: 'healthy', redis: 'healthy' },
        uptime: 86400,
        version: '1.0.0',
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockHealth));

      const result = await service.getHealth();

      expect(result.status).toBe('healthy');
      expect(result.uptime).toBe(86400);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/system/health`);
    });
  });

  describe('isHealthy', () => {
    it('should return true when system is healthy', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({
        status: 'healthy',
        services: {},
        uptime: 0,
        version: '1.0.0',
      }));

      const result = await service.isHealthy();

      expect(result).toBe(true);
    });

    it('should return false when system is degraded', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({
        status: 'degraded',
        services: { redis: 'unhealthy' },
        uptime: 0,
        version: '1.0.0',
      }));

      const result = await service.isHealthy();

      expect(result).toBe(false);
    });

    it('should return false when health check fails with error', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ message: 'Server error' }, 500));

      const result = await service.isHealthy();

      expect(result).toBe(false);
    });
  });

  describe('getMaintenanceMode / setMaintenanceMode', () => {
    it('should get maintenance mode status', async () => {
      const mockMaintenance: MaintenanceModeResponse = {
        enabled: false,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockMaintenance));

      const result = await service.getMaintenanceMode();

      expect(result.enabled).toBe(false);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/system/maintenance`);
    });

    it('should set maintenance mode', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({
        enabled: true,
        message: 'Scheduled maintenance',
      }));

      const result = await service.setMaintenanceMode({
        enabled: true,
        message: 'Scheduled maintenance',
      });

      expect(result.enabled).toBe(true);
      expect(result.message).toBe('Scheduled maintenance');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/system/maintenance`);
      expect(options.method).toBe('PUT');
    });
  });

  describe('enableMaintenanceMode / disableMaintenanceMode', () => {
    it('should enable maintenance mode with message', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({
        enabled: true,
        message: 'Down for updates',
      }));

      const result = await service.enableMaintenanceMode('Down for updates');

      expect(result.enabled).toBe(true);

      const [, options] = fetchMock.mock.calls[0]!;
      expect(JSON.parse(options.body)).toEqual({ enabled: true, message: 'Down for updates' });
    });

    it('should disable maintenance mode', async () => {
      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse({ enabled: false }));

      const result = await service.disableMaintenanceMode();

      expect(result.enabled).toBe(false);

      const [, options] = fetchMock.mock.calls[0]!;
      expect(JSON.parse(options.body)).toEqual({ enabled: false });
    });
  });

  describe('getGeoDistribution', () => {
    it('should return geo distribution data', async () => {
      const mockGeo: GeoDistributionResponse = {
        locations: [
          { country_code: 'US', country_name: 'United States', city: 'New York', login_count: 100, latitude: 40.7, longitude: -74.0 },
          { country_code: 'GB', country_name: 'United Kingdom', city: 'London', login_count: 50, latitude: 51.5, longitude: -0.1 },
        ],
        total: 150,
        countries: 2,
        cities: 2,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockGeo));

      const result = await service.getGeoDistribution(7);

      expect(result.locations).toHaveLength(2);
      expect(result.total).toBe(150);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toContain('/api/admin/analytics/geo-distribution');
      expect(url).toContain('days=7');
    });
  });

  describe('getTopCountries', () => {
    it('should aggregate and sort countries by login count', async () => {
      const mockGeo: GeoDistributionResponse = {
        locations: [
          { country_code: 'US', country_name: 'United States', city: 'New York', login_count: 80, latitude: 0, longitude: 0 },
          { country_code: 'US', country_name: 'United States', city: 'LA', login_count: 60, latitude: 0, longitude: 0 },
          { country_code: 'GB', country_name: 'United Kingdom', city: 'London', login_count: 50, latitude: 0, longitude: 0 },
          { country_code: 'DE', country_name: 'Germany', city: 'Berlin', login_count: 30, latitude: 0, longitude: 0 },
        ],
        total: 220,
        countries: 3,
        cities: 4,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockGeo));

      const result = await service.getTopCountries(2, 30);

      expect(result).toHaveLength(2);
      expect(result[0]!.country_code).toBe('US');
      expect(result[0]!.login_count).toBe(140); // 80 + 60
      expect(result[1]!.country_code).toBe('GB');
      expect(result[1]!.login_count).toBe(50);
    });
  });

  describe('password policy', () => {
    it('should get password policy', async () => {
      const mockPolicy = {
        min_length: 8,
        require_uppercase: true,
        require_lowercase: true,
        require_numbers: true,
        require_special: false,
      };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(mockPolicy));

      const result = await service.getPasswordPolicy();

      expect(result.min_length).toBe(8);
      expect(result.require_uppercase).toBe(true);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/system/password-policy`);
    });

    it('should update password policy', async () => {
      const updatedPolicy = { min_length: 12, require_special: true };

      fetchMock.mockResolvedValueOnce(mockFetchJsonResponse(updatedPolicy));

      const result = await service.updatePasswordPolicy({ min_length: 12, require_special: true });

      expect(result.min_length).toBe(12);

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/admin/system/password-policy`);
      expect(options.method).toBe('PUT');
    });
  });
});

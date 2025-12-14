
import { User, UserRole, ApiKey, AuditLog, DashboardStats, OAuthAccount, OAuthProviderConfig, EmailTemplate, RoleDefinition, Permission, UserSession, IpRule, WebhookEndpoint, ServiceAccount, BrandingConfig, SmsConfig, PasswordPolicy } from '../types';

// Local type for SystemStatus (not in SDK)
export interface SystemStatus {
  status: 'healthy' | 'degraded' | 'unhealthy';
  database: 'connected' | 'disconnected';
  redis: 'connected' | 'disconnected';
  uptime: number;
  version: string;
  maintenanceMode: boolean;
  maintenanceMessage?: string;
}

// Helpers
const randomId = () => Math.random().toString(36).substring(2, 11);
const randomDate = (start: Date, end: Date) => new Date(start.getTime() + Math.random() * (end.getTime() - start.getTime())).toISOString();

// Mock Users
export const generateUsers = (count: number): User[] => {
  const roleDefinitions = mockRoles;

  return Array.from({ length: count }).map((_, i) => {
    const numRoles = Math.random() > 0.7 ? 2 : 1;
    const shuffled = [...roleDefinitions].sort(() => 0.5 - Math.random());
    const userRoles = shuffled.slice(0, numRoles).map(r => ({
      id: r.id,
      name: r.name,
      display_name: r.display_name
    }));

    return {
      id: randomId(),
      email: `user${i}@example.com`,
      username: `user_${i}`,
      full_name: `User Name ${i}`,
      roles: userRoles,
      account_type: 'human' as const,
      is_active: Math.random() > 0.1,
      email_verified: Math.random() > 0.2,
      phone_verified: Math.random() > 0.6,
      totp_enabled: Math.random() > 0.7,
      phone: Math.random() > 0.5 ? `+1 (555) 000-${1000 + i}` : undefined,
      created_at: randomDate(new Date(2023, 0, 1), new Date()),
      updated_at: randomDate(new Date(2023, 0, 1), new Date()),
      profile_picture_url: `https://picsum.photos/seed/${i}/200/200`
    };
  });
};

// Mock API Keys
export const generateApiKeys = (count: number, users: User[]): ApiKey[] => {
  return Array.from({ length: count }).map((_, i) => {
    const user = users[Math.floor(Math.random() * users.length)];
    const createdAt = randomDate(new Date(2023, 0, 1), new Date());
    return {
      id: randomId(),
      name: `Key for ${user.username} - ${i}`,
      key_prefix: `agw_${randomId().substring(0, 4)}`,
      user_id: user.id,
      scopes: ['users:read', 'profile:read'],
      is_active: Math.random() > 0.2,
      last_used_at: randomDate(new Date(2023, 6, 1), new Date()),
      created_at: createdAt,
      updated_at: createdAt,
    };
  });
};

// Mock Audit Logs
export const generateAuditLogs = (count: number, users?: User[]): AuditLog[] => {
  const actions = ['signin', 'signup', 'api_key_create', 'password_reset', 'oauth_link'];
  const statuses: ('success' | 'failure')[] = ['success', 'success', 'success', 'failure', 'failure'];
  const userAgents = [
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
    'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36',
    'Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15'
  ];

  return Array.from({ length: count }).map((_, i) => {
    const user = users ? users[Math.floor(Math.random() * users.length)] : undefined;
    return {
      id: randomId(),
      action: actions[Math.floor(Math.random() * actions.length)],
      user_id: user?.id || `user_${Math.floor(Math.random() * 10)}`,
      resource: 'auth',
      ip_address: `192.168.1.${Math.floor(Math.random() * 255)}`,
      user_agent: userAgents[Math.floor(Math.random() * userAgents.length)],
      status: statuses[Math.floor(Math.random() * statuses.length)],
      created_at: randomDate(new Date(Date.now() - 7 * 24 * 60 * 60 * 1000), new Date()),
    };
  });
};

export const getMockStats = (): DashboardStats => {
  return {
    totalUsers: 1234,
    activeUsers: 956,
    newUsersToday: 23,
    totalApiKeys: 156,
    activeApiKeys: 142,
    loginAttemptsToday: 487,
    failedLoginAttemptsToday: 12
  };
};

// Mock OAuth Providers
export const mockOAuthProviders: OAuthProviderConfig[] = [
  {
    id: '1',
    provider: 'google',
    client_id: '782934234-random-string.apps.googleusercontent.com',
    client_secret: 'GOCSPX-random-secret-string',
    redirect_uris: ['https://auth.example.com/api/v1/auth/google/callback'],
    is_enabled: true,
    created_at: new Date('2023-01-15').toISOString(),
    updated_at: new Date('2023-01-15').toISOString()
  },
  {
    id: '2',
    provider: 'github',
    client_id: 'Iv1.8a9c8b7d6e5f4g3h',
    client_secret: '8a9c8b7d6e5f4g3h2i1j0k9l8m7n6o5p',
    redirect_uris: ['https://auth.example.com/api/v1/auth/github/callback'],
    is_enabled: true,
    created_at: new Date('2023-02-20').toISOString(),
    updated_at: new Date('2023-02-20').toISOString()
  },
  {
    id: '3',
    provider: 'telegram',
    client_id: '123456789:AAH-random-token',
    client_secret: '',
    redirect_uris: ['https://auth.example.com/api/v1/auth/telegram/callback'],
    is_enabled: false,
    created_at: new Date('2023-03-10').toISOString(),
    updated_at: new Date('2023-03-10').toISOString()
  }
];

// Mock Email Templates
export const mockEmailTemplates: EmailTemplate[] = [
  {
    id: 'tpl_verify',
    type: 'verification',
    name: 'Email Verification',
    subject: 'Verify your email address',
    html_body: `<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f9fafb;">
  <div style="max-width: 600px; margin: 40px auto; padding: 40px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
    <h2 style="color: #0ea5e9; margin-top: 0;">Welcome to Auth Gateway!</h2>
    <p>Hi {{name}},</p>
    <p>Thank you for registering. Please verify your email address to get started by clicking the button below:</p>
    <div style="text-align: center; margin: 30px 0;">
      <a href="{{action_url}}" style="display: inline-block; background-color: #0ea5e9; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: bold;">Verify Email</a>
    </div>
    <p style="font-size: 0.9em; color: #666;">If the button doesn't work, copy and paste this link into your browser:<br/>
    <a href="{{action_url}}" style="color: #0ea5e9;">{{action_url}}</a></p>
    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;" />
    <p style="font-size: 0.8em; color: #888;">If you didn't request this, you can safely ignore this email.</p>
  </div>
</body>
</html>`,
    variables: ['{{name}}', '{{action_url}}', '{{email}}'],
    is_active: true,
    created_at: new Date('2023-09-10').toISOString(),
    updated_at: new Date('2023-09-10').toISOString()
  },
  {
    id: 'tpl_reset',
    type: 'password_reset',
    name: 'Password Reset',
    subject: 'Reset your password',
    html_body: `<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f9fafb;">
  <div style="max-width: 600px; margin: 40px auto; padding: 40px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
    <h2 style="color: #dc2626; margin-top: 0;">Password Reset Request</h2>
    <p>Hi {{name}},</p>
    <p>We received a request to reset your password. Click the link below to choose a new one:</p>
    <div style="margin: 24px 0;">
      <a href="{{action_url}}" style="color: #0ea5e9; font-weight: bold;">Reset my password</a>
    </div>
    <p>This link expires in 1 hour.</p>
    <p style="font-size: 0.8em; color: #888;">Request came from IP: {{ip_address}}</p>
  </div>
</body>
</html>`,
    variables: ['{{name}}', '{{action_url}}', '{{ip_address}}'],
    is_active: true,
    created_at: new Date('2023-09-15').toISOString(),
    updated_at: new Date('2023-09-15').toISOString()
  },
  {
    id: 'tpl_welcome',
    type: 'welcome',
    name: 'Welcome Email',
    subject: 'Welcome to the platform!',
    html_body: `<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0;">
  <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
    <h2>Welcome aboard, {{name}}!</h2>
    <p>We are thrilled to have you with us.</p>
    <p>Your username is: <strong>{{username}}</strong></p>
    <p>If you have any questions, feel free to reply to this email.</p>
    <p>Best regards,<br/>The Team</p>
  </div>
</body>
</html>`,
    variables: ['{{name}}', '{{username}}'],
    is_active: true,
    created_at: new Date('2023-10-01').toISOString(),
    updated_at: new Date('2023-10-01').toISOString()
  }
];

// Mock RBAC Data
const baseTimestamp = new Date('2023-01-01').toISOString();
export const mockPermissions: Permission[] = [
  // User Management
  { id: 'users:read', resource: 'users', action: 'read', name: 'Read Users', description: 'View user list and details', created_at: baseTimestamp, updated_at: baseTimestamp },
  { id: 'users:write', resource: 'users', action: 'write', name: 'Create/Edit Users', description: 'Create and modify users', created_at: baseTimestamp, updated_at: baseTimestamp },
  { id: 'users:delete', resource: 'users', action: 'delete', name: 'Delete Users', description: 'Delete users from system', created_at: baseTimestamp, updated_at: baseTimestamp },

  // API Keys
  { id: 'api_keys:read', resource: 'api_keys', action: 'read', name: 'View API Keys', description: 'View all API keys', created_at: baseTimestamp, updated_at: baseTimestamp },
  { id: 'api_keys:revoke', resource: 'api_keys', action: 'revoke', name: 'Revoke API Keys', description: 'Revoke any API key', created_at: baseTimestamp, updated_at: baseTimestamp },

  // System Settings
  { id: 'system:read', resource: 'system', action: 'read', name: 'View Settings', description: 'View system configuration', created_at: baseTimestamp, updated_at: baseTimestamp },
  { id: 'system:write', resource: 'system', action: 'write', name: 'Edit Settings', description: 'Modify system configuration', created_at: baseTimestamp, updated_at: baseTimestamp },

  // Audit Logs
  { id: 'logs:read', resource: 'audit_logs', action: 'read', name: 'View Logs', description: 'Access audit trail', created_at: baseTimestamp, updated_at: baseTimestamp },
];

export const mockRoles: RoleDefinition[] = [
  {
    id: 'admin',
    name: 'admin',
    display_name: 'Administrator',
    description: 'Full system access',
    is_system_role: true,
    permissions: mockPermissions,
    created_at: new Date('2023-01-01').toISOString(),
    updated_at: new Date('2023-01-01').toISOString()
  },
  {
    id: 'moderator',
    name: 'moderator',
    display_name: 'Moderator',
    description: 'Can manage users but not system settings',
    is_system_role: true,
    permissions: mockPermissions.filter(p => ['users:read', 'users:write', 'logs:read', 'api_keys:read'].includes(p.id)),
    created_at: new Date('2023-02-15').toISOString(),
    updated_at: new Date('2023-02-15').toISOString()
  },
  {
    id: 'user',
    name: 'user',
    display_name: 'User',
    description: 'Standard user access',
    is_system_role: true,
    permissions: [],
    created_at: new Date('2023-03-20').toISOString(),
    updated_at: new Date('2023-03-20').toISOString()
  },
  {
    id: 'support',
    name: 'support',
    display_name: 'Support Agent',
    description: 'Read-only access to users and logs',
    is_system_role: false,
    permissions: mockPermissions.filter(p => ['users:read', 'logs:read'].includes(p.id)),
    created_at: new Date('2023-06-10').toISOString(),
    updated_at: new Date('2023-06-10').toISOString()
  }
];

// Mock IP Rules
export const mockIpRules: IpRule[] = [
  {
    id: 'ip_1',
    type: 'blacklist',
    ip_address: '192.168.1.55',
    description: 'Malicious bot activity',
    created_at: new Date('2023-10-15').toISOString(),
    updated_at: new Date('2023-10-15').toISOString()
  },
  {
    id: 'ip_2',
    type: 'whitelist',
    ip_address: '10.0.0.0/8',
    description: 'Internal corporate network',
    created_at: new Date('2023-01-01').toISOString(),
    updated_at: new Date('2023-01-01').toISOString()
  }
];

// Mock Webhooks
export const mockWebhooks: WebhookEndpoint[] = [
  {
    id: 'wh_1',
    name: 'Main Application Sync',
    url: 'https://api.myapp.com/webhooks/auth',
    events: ['user.created', 'user.deleted', 'user.updated'],
    secret_key: 'whsec_test_1234567890',
    is_active: true,
    failure_count: 0,
    last_triggered_at: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
    created_at: new Date('2023-08-01').toISOString(),
    updated_at: new Date('2023-08-01').toISOString()
  },
  {
    id: 'wh_2',
    name: 'Analytics Tracking',
    url: 'https://analytics.example.com/ingest',
    events: ['auth.login.success', 'auth.login.failed'],
    secret_key: 'whsec_test_abcdefghij',
    is_active: false,
    failure_count: 12,
    last_triggered_at: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(),
    created_at: new Date('2023-09-15').toISOString(),
    updated_at: new Date('2023-09-15').toISOString()
  }
];

// Mock Service Accounts
export const mockServiceAccounts: ServiceAccount[] = [
  {
    id: 'sa_1',
    name: 'Payment Service',
    description: 'Backend service for processing payments',
    client_id: 'svc_payment_8a7d9f2',
    is_active: true,
    created_at: new Date('2023-05-10').toISOString(),
    last_used_at: new Date().toISOString()
  },
  {
    id: 'sa_2',
    name: 'Notification Worker',
    description: 'Async worker for sending emails',
    client_id: 'svc_notify_3k2j1h4',
    is_active: true,
    created_at: new Date('2023-06-22').toISOString(),
    last_used_at: new Date(Date.now() - 1000 * 60 * 60).toISOString()
  }
];

// Mock Branding
export let mockBranding: BrandingConfig = {
  id: 'branding_1',
  company_name: 'Auth Gateway',
  logo_url: '',
  favicon_url: '',
  theme: {
    primary_color: '#2563EB', // blue-600
    secondary_color: '#1E40AF', // blue-800
    background_color: '#F3F4F6' // gray-100
  },
  support_email: 'support@example.com',
  terms_url: '/terms',
  privacy_url: '/privacy',
  created_at: new Date('2023-01-01').toISOString(),
  updated_at: new Date('2023-01-01').toISOString()
};

// Mock SMS Config
export let mockSmsConfig: SmsConfig = {
  provider: 'mock',
  awsRegion: 'us-east-1',
};

// Mock System Status
export let mockSystemStatus: SystemStatus = {
  status: 'healthy',
  database: 'connected',
  redis: 'connected',
  uptime: 3600 * 24 * 5, // 5 days
  version: 'v1.0.0',
  maintenanceMode: false,
  maintenanceMessage: 'System is under maintenance. Please try again later.'
};

// Mock Password Policy
export let mockPasswordPolicy: PasswordPolicy = {
  minLength: 8,
  requireUppercase: true,
  requireLowercase: true,
  requireNumbers: true,
  requireSpecial: false,
  historyCount: 3,
  expiryDays: 90,
  jwtTtlMinutes: 15,
  refreshTtlDays: 7
};

// Generate initial data
export const mockUsers = generateUsers(50);
export const mockApiKeys = generateApiKeys(20, mockUsers);
export const mockLogs = generateAuditLogs(50, mockUsers);

// Generate Mock Sessions
export const generateSessions = (users: User[]): UserSession[] => {
  const sessions: UserSession[] = [];
  users.forEach(user => {
    // Generate 1-3 sessions per user randomly
    if (Math.random() > 0.3) {
      const count = Math.floor(Math.random() * 3) + 1;
      for(let i=0; i<count; i++) {
          const createdDate = randomDate(new Date(Date.now() - 7 * 24 * 60 * 60 * 1000), new Date(Date.now() - 24 * 60 * 60 * 1000));
          const lastActive = randomDate(new Date(Date.now() - 24 * 60 * 60 * 1000), new Date());
          const deviceType = Math.random() > 0.6 ? 'desktop' : 'mobile';
          const os = Math.random() > 0.5 ? (Math.random() > 0.5 ? 'Windows 11' : 'macOS 14') : (Math.random() > 0.5 ? 'iOS 17' : 'Android 14');
          const browser = Math.random() > 0.3 ? 'Chrome 120.0' : 'Safari 17.0';

          sessions.push({
              id: randomId(),
              ip_address: `192.168.${Math.floor(Math.random()*255)}.${Math.floor(Math.random()*255)}`,
              user_agent: `Mozilla/5.0 (${deviceType}; ${os}) AppleWebKit/537.36 (KHTML, like Gecko) ${browser}`,
              created_at: createdDate,
              last_activity: lastActive,
              is_current: i === 0 && Math.random() > 0.5
          });
      }
    }
  });
  return sessions;
}

export const mockSessions = generateSessions(mockUsers);


// Data Access Helpers
export const getUser = (id: string): User | undefined => {
  return mockUsers.find(u => u.id === id);
};

export const updateUser = (id: string, data: Partial<User>): User | undefined => {
  const index = mockUsers.findIndex(u => u.id === id);
  if (index !== -1) {
    mockUsers[index] = { ...mockUsers[index], ...data };
    return mockUsers[index];
  }
  return undefined;
};

export const createUser = (data: Partial<User>): User => {
  const defaultRole = mockRoles.find(r => r.id === 'user');
  const now = new Date().toISOString();
  const newUser: User = {
    id: randomId(),
    email: data.email || '',
    username: data.username || '',
    full_name: data.full_name || '',
    roles: data.roles || (defaultRole ? [{
      id: defaultRole.id,
      name: defaultRole.name,
      display_name: defaultRole.display_name
    }] : []),
    account_type: data.account_type || 'human',
    is_active: data.is_active ?? true,
    email_verified: data.email_verified ?? false,
    phone_verified: data.phone_verified ?? false,
    totp_enabled: data.totp_enabled ?? false,
    phone: data.phone,
    created_at: now,
    updated_at: now,
    profile_picture_url: data.profile_picture_url || `https://picsum.photos/seed/${Math.random()}/200/200`
  };
  mockUsers.unshift(newUser);
  return newUser;
};

// User Actions
export const resetUserTwoFA = (userId: string): boolean => {
  const user = updateUser(userId, { totp_enabled: false });
  return !!user;
}

export const sendPasswordResetEmail = (userId: string): boolean => {
  // Mock sending logic
  console.log(`Sending password reset to user ${userId}`);
  return true;
}

export const getUserApiKeys = (userId: string): ApiKey[] => {
  return mockApiKeys.filter(k => k.user_id === userId);
};

export const getUserLogs = (userId: string): AuditLog[] => {
  return mockLogs.filter(l => l.user_id === userId).sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
};

export const getUserOAuthAccounts = (userId: string): OAuthAccount[] => {
  const providers: OAuthAccount['provider'][] = ['google', 'github', 'yandex', 'telegram'];
  const count = Math.floor(Math.random() * 3);

  return Array.from({ length: count }).map((_, i) => {
    const createdDate = randomDate(new Date(2023, 0, 1), new Date());
    return {
      id: `oauth_${userId}_${i}`,
      provider: providers[i % providers.length],
      user_id: userId,
      provider_user_id: `provider_user_${randomId()}`,
      created_at: createdDate,
      updated_at: createdDate
    };
  });
};

export const getUserSessions = (userId: string): UserSession[] => {
  return mockSessions.sort((a, b) => new Date(b.last_activity).getTime() - new Date(a.last_activity).getTime());
};

export const revokeUserSession = (sessionId: string): boolean => {
  const index = mockSessions.findIndex(s => s.id === sessionId);
  if (index !== -1) {
    mockSessions.splice(index, 1);
    return true;
  }
  return false;
};

// OAuth Provider Access
export const getOAuthProviders = (): OAuthProviderConfig[] => {
  return mockOAuthProviders;
};

export const getOAuthProvider = (id: string): OAuthProviderConfig | undefined => {
  return mockOAuthProviders.find(p => p.id === id);
};

export const updateOAuthProvider = (id: string, data: Partial<OAuthProviderConfig>): OAuthProviderConfig | undefined => {
  const index = mockOAuthProviders.findIndex(p => p.id === id);
  if (index !== -1) {
    mockOAuthProviders[index] = { ...mockOAuthProviders[index], ...data };
    return mockOAuthProviders[index];
  }
  return undefined;
};

export const createOAuthProvider = (data: Omit<OAuthProviderConfig, 'id' | 'created_at' | 'updated_at'>): OAuthProviderConfig => {
  const now = new Date().toISOString();
  const newProvider: OAuthProviderConfig = {
    ...data,
    id: randomId(),
    created_at: now,
    updated_at: now
  };
  mockOAuthProviders.push(newProvider);
  return newProvider;
};

export const deleteOAuthProvider = (id: string): boolean => {
  const index = mockOAuthProviders.findIndex(p => p.id === id);
  if (index !== -1) {
    mockOAuthProviders.splice(index, 1);
    return true;
  }
  return false;
};

// Email Templates Access
export const getEmailTemplates = (): EmailTemplate[] => {
  return mockEmailTemplates;
};

export const getEmailTemplate = (id: string): EmailTemplate | undefined => {
  return mockEmailTemplates.find(t => t.id === id);
};

export const updateEmailTemplate = (id: string, data: Partial<EmailTemplate>): EmailTemplate | undefined => {
  const index = mockEmailTemplates.findIndex(t => t.id === id);
  if (index !== -1) {
    mockEmailTemplates[index] = { ...mockEmailTemplates[index], ...data, updated_at: new Date().toISOString() };
    return mockEmailTemplates[index];
  }
  return undefined;
};

// RBAC Access
export const getRoles = (): RoleDefinition[] => {
  return mockRoles;
};

export const getRoleUserCount = (roleId: string): number => {
  return mockUsers.filter(user => user.roles.some(r => r.id === roleId)).length;
};

export const getRole = (id: string): RoleDefinition | undefined => {
  return mockRoles.find(r => r.id === id);
};

export const getPermissions = (): Permission[] => {
  return mockPermissions;
};

export const getPermission = (id: string): Permission | undefined => {
  return mockPermissions.find(p => p.id === id);
};

export const createPermission = (data: Partial<Permission>): Permission => {
  const resource = data.resource?.toLowerCase() || 'unknown';
  const action = data.action?.toLowerCase() || 'unknown';
  const id = `${resource}:${action}`;
  const now = new Date().toISOString();

  const newPerm: Permission = {
    id: id,
    name: data.name || `${resource} ${action}`,
    resource: resource,
    action: action,
    description: data.description || '',
    created_at: now,
    updated_at: now
  };
  mockPermissions.push(newPerm);
  return newPerm;
};

export const updatePermission = (id: string, data: Partial<Permission>): Permission | undefined => {
  const index = mockPermissions.findIndex(p => p.id === id);
  if (index !== -1) {
    mockPermissions[index] = { ...mockPermissions[index], ...data };
    return mockPermissions[index];
  }
  return undefined;
};

export const deletePermission = (id: string): boolean => {
  const index = mockPermissions.findIndex(p => p.id === id);
  if (index !== -1) {
    mockPermissions.splice(index, 1);
    return true;
  }
  return false;
};

export const updateRole = (id: string, data: Partial<RoleDefinition>): RoleDefinition | undefined => {
  const index = mockRoles.findIndex(r => r.id === id);
  if (index !== -1) {
    mockRoles[index] = { ...mockRoles[index], ...data };
    return mockRoles[index];
  }
  return undefined;
};

export const createRole = (data: Partial<RoleDefinition>): RoleDefinition => {
  const now = new Date().toISOString();
  const roleName = data.name || data.display_name || 'New Role';
  const newRole: RoleDefinition = {
    id: roleName.toLowerCase().replace(/\s+/g, '_') || randomId(),
    name: roleName.toLowerCase().replace(/\s+/g, '_'),
    display_name: data.display_name || roleName,
    description: data.description || '',
    is_system_role: false,
    permissions: data.permissions || [],
    created_at: now,
    updated_at: now
  };
  mockRoles.push(newRole);
  return newRole;
};

export const deleteRole = (id: string): boolean => {
  const index = mockRoles.findIndex(r => r.id === id);
  if (index !== -1 && !mockRoles[index].is_system_role) {
    mockRoles.splice(index, 1);
    return true;
  }
  return false;
};

// IP Rules Access
export const getIpRules = (type?: 'whitelist' | 'blacklist'): IpRule[] => {
  if (type) {
    return mockIpRules.filter(r => r.type === type);
  }
  return mockIpRules;
};

export const createIpRule = (data: Partial<IpRule>): IpRule => {
  const now = new Date().toISOString();
  const newRule: IpRule = {
    id: randomId(),
    type: data.type || 'blacklist',
    ip_address: data.ip_address || '',
    description: data.description,
    created_at: now,
    updated_at: now
  };
  mockIpRules.unshift(newRule);
  return newRule;
};

export const deleteIpRule = (id: string): boolean => {
  const index = mockIpRules.findIndex(r => r.id === id);
  if (index !== -1) {
    mockIpRules.splice(index, 1);
    return true;
  }
  return false;
}

// Webhooks Access
export const getWebhooks = (): WebhookEndpoint[] => {
  return mockWebhooks;
};

export const getWebhook = (id: string): WebhookEndpoint | undefined => {
  return mockWebhooks.find(w => w.id === id);
};

export const createWebhook = (data: Partial<WebhookEndpoint>): WebhookEndpoint => {
  const now = new Date().toISOString();
  const newWebhook: WebhookEndpoint = {
    id: randomId(),
    name: data.name || 'New Webhook',
    url: data.url || '',
    events: data.events || [],
    secret_key: `whsec_${randomId()}${randomId()}`,
    is_active: data.is_active ?? true,
    failure_count: 0,
    created_at: now,
    updated_at: now
  };
  mockWebhooks.unshift(newWebhook);
  return newWebhook;
};

export const updateWebhook = (id: string, data: Partial<WebhookEndpoint>): WebhookEndpoint | undefined => {
  const index = mockWebhooks.findIndex(w => w.id === id);
  if (index !== -1) {
    mockWebhooks[index] = { ...mockWebhooks[index], ...data };
    return mockWebhooks[index];
  }
  return undefined;
};

export const deleteWebhook = (id: string): boolean => {
  const index = mockWebhooks.findIndex(w => w.id === id);
  if (index !== -1) {
    mockWebhooks.splice(index, 1);
    return true;
  }
  return false;
};

// Service Accounts Access
export const getServiceAccounts = (): ServiceAccount[] => {
  return mockServiceAccounts;
};

export const getServiceAccount = (id: string): ServiceAccount | undefined => {
  return mockServiceAccounts.find(sa => sa.id === id);
};

export const createServiceAccount = (data: Partial<ServiceAccount>): { account: ServiceAccount, clientSecret: string } => {
  const clientSecret = `svc_sec_${randomId()}${randomId()}`;
  const newAccount: ServiceAccount = {
    id: randomId(),
    name: data.name || 'New Service',
    description: data.description || '',
    client_id: `svc_${randomId()}`,
    is_active: data.is_active ?? true,
    created_at: new Date().toISOString()
  };
  mockServiceAccounts.unshift(newAccount);
  return { account: newAccount, clientSecret };
};

export const updateServiceAccount = (id: string, data: Partial<ServiceAccount>): ServiceAccount | undefined => {
  const index = mockServiceAccounts.findIndex(sa => sa.id === id);
  if (index !== -1) {
    mockServiceAccounts[index] = { ...mockServiceAccounts[index], ...data };
    return mockServiceAccounts[index];
  }
  return undefined;
};

export const deleteServiceAccount = (id: string): boolean => {
  const index = mockServiceAccounts.findIndex(sa => sa.id === id);
  if (index !== -1) {
    mockServiceAccounts.splice(index, 1);
    return true;
  }
  return false;
};

// Branding Access
export const getBranding = (): BrandingConfig => {
  return mockBranding;
};

export const updateBranding = (data: Partial<BrandingConfig>): BrandingConfig => {
  mockBranding = { ...mockBranding, ...data };
  return mockBranding;
};

// SMS Config Access
export const getSmsConfig = (): SmsConfig => {
  return mockSmsConfig;
};

export const updateSmsConfig = (data: Partial<SmsConfig>): SmsConfig => {
  mockSmsConfig = { ...mockSmsConfig, ...data };
  return mockSmsConfig;
};

// System Status Access
export const getSystemStatus = (): SystemStatus => {
  return mockSystemStatus;
};

export const updateSystemStatus = (data: Partial<SystemStatus>): SystemStatus => {
  mockSystemStatus = { ...mockSystemStatus, ...data };
  return mockSystemStatus;
};

// Password Policy Access
export const getPasswordPolicy = (): PasswordPolicy => {
  return mockPasswordPolicy;
};

export const updatePasswordPolicy = (data: Partial<PasswordPolicy>): PasswordPolicy => {
  mockPasswordPolicy = { ...mockPasswordPolicy, ...data };
  return mockPasswordPolicy;
};
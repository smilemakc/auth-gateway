
import { User, UserRole, ApiKey, AuditLog, DashboardStats, OAuthAccount, OAuthProviderConfig, EmailTemplate, RoleDefinition, Permission, UserSession, IpRule, WebhookEndpoint, ServiceAccount, BrandingConfig, SmsConfig, SystemStatus, PasswordPolicy } from '../types';

// Helpers
const randomId = () => Math.random().toString(36).substring(2, 11);
const randomDate = (start: Date, end: Date) => new Date(start.getTime() + Math.random() * (end.getTime() - start.getTime())).toISOString();

// Mock Users
export const generateUsers = (count: number): User[] => {
  const roles = [UserRole.ADMIN, UserRole.MODERATOR, UserRole.USER];
  return Array.from({ length: count }).map((_, i) => ({
    id: randomId(),
    email: `user${i}@example.com`,
    username: `user_${i}`,
    fullName: `User Name ${i}`,
    role: roles[Math.floor(Math.random() * roles.length)],
    isActive: Math.random() > 0.1,
    isEmailVerified: Math.random() > 0.2,
    is2FAEnabled: Math.random() > 0.7,
    phone: Math.random() > 0.5 ? `+1 (555) 000-${1000 + i}` : undefined,
    createdAt: randomDate(new Date(2023, 0, 1), new Date()),
    lastLogin: randomDate(new Date(2023, 0, 1), new Date()),
    avatarUrl: `https://picsum.photos/seed/${i}/200/200`
  }));
};

// Mock API Keys
export const generateApiKeys = (count: number, users: User[]): ApiKey[] => {
  return Array.from({ length: count }).map((_, i) => {
    const user = users[Math.floor(Math.random() * users.length)];
    return {
      id: randomId(),
      name: `Key for ${user.username} - ${i}`,
      prefix: `agw_${randomId().substring(0, 4)}`,
      ownerId: user.id,
      ownerName: user.username,
      scopes: ['users:read', 'profile:read'],
      status: Math.random() > 0.2 ? 'active' : 'revoked',
      lastUsed: randomDate(new Date(2023, 6, 1), new Date()),
      createdAt: randomDate(new Date(2023, 0, 1), new Date()),
    };
  });
};

// Mock Audit Logs
export const generateAuditLogs = (count: number, users?: User[]): AuditLog[] => {
  const actions = ['signin', 'signup', 'api_key_create', 'password_reset', 'oauth_link'];
  const statuses: ('success' | 'failed' | 'blocked')[] = ['success', 'success', 'success', 'failed', 'blocked'];
  
  return Array.from({ length: count }).map((_, i) => {
    const user = users ? users[Math.floor(Math.random() * users.length)] : undefined;
    return {
      id: randomId(),
      action: actions[Math.floor(Math.random() * actions.length)],
      userId: user?.id || `user_${Math.floor(Math.random() * 10)}`,
      userEmail: user?.email || `user${Math.floor(Math.random() * 10)}@example.com`,
      resource: 'auth',
      ip: `192.168.1.${Math.floor(Math.random() * 255)}`,
      status: statuses[Math.floor(Math.random() * statuses.length)],
      timestamp: randomDate(new Date(Date.now() - 7 * 24 * 60 * 60 * 1000), new Date()),
    };
  });
};

export const getMockStats = (): DashboardStats => {
  // Generate last 30 days data
  const registrations = [];
  const activity = [];
  for (let i = 29; i >= 0; i--) {
    const d = new Date();
    d.setDate(d.getDate() - i);
    const dateStr = d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    registrations.push({ date: dateStr, count: Math.floor(Math.random() * 50) + 10 });
    activity.push({ 
      date: dateStr, 
      success: Math.floor(Math.random() * 200) + 50, 
      failed: Math.floor(Math.random() * 20) 
    });
  }

  return {
    totalUsers: 1234,
    activeUsers: 956,
    usersWith2FA: 789,
    totalApiKeys: 156,
    recentRegistrations: registrations,
    loginActivity: activity
  };
};

// Mock OAuth Providers
export const mockOAuthProviders: OAuthProviderConfig[] = [
  {
    id: '1',
    provider: 'google',
    clientId: '782934234-random-string.apps.googleusercontent.com',
    clientSecret: 'GOCSPX-random-secret-string',
    redirectUris: ['https://auth.example.com/api/v1/auth/google/callback'],
    isEnabled: true,
    createdAt: new Date('2023-01-15').toISOString()
  },
  {
    id: '2',
    provider: 'github',
    clientId: 'Iv1.8a9c8b7d6e5f4g3h',
    clientSecret: '8a9c8b7d6e5f4g3h2i1j0k9l8m7n6o5p',
    redirectUris: ['https://auth.example.com/api/v1/auth/github/callback'],
    isEnabled: true,
    createdAt: new Date('2023-02-20').toISOString()
  },
  {
    id: '3',
    provider: 'telegram',
    clientId: '123456789:AAH-random-token',
    clientSecret: '',
    redirectUris: ['https://auth.example.com/api/v1/auth/telegram/callback'],
    isEnabled: false,
    createdAt: new Date('2023-03-10').toISOString()
  }
];

// Mock Email Templates
export const mockEmailTemplates: EmailTemplate[] = [
  {
    id: 'tpl_verify',
    type: 'verification',
    name: 'Email Verification',
    subject: 'Verify your email address',
    bodyHtml: `<!DOCTYPE html>
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
    lastUpdated: new Date('2023-09-10').toISOString()
  },
  {
    id: 'tpl_reset',
    type: 'reset_password',
    name: 'Password Reset',
    subject: 'Reset your password',
    bodyHtml: `<!DOCTYPE html>
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
    lastUpdated: new Date('2023-09-15').toISOString()
  },
  {
    id: 'tpl_welcome',
    type: 'welcome',
    name: 'Welcome Email',
    subject: 'Welcome to the platform!',
    bodyHtml: `<!DOCTYPE html>
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
    lastUpdated: new Date('2023-10-01').toISOString()
  }
];

// Mock RBAC Data
export const mockPermissions: Permission[] = [
  // User Management
  { id: 'users:read', resource: 'users', action: 'read', name: 'Read Users', description: 'View user list and details' },
  { id: 'users:write', resource: 'users', action: 'write', name: 'Create/Edit Users', description: 'Create and modify users' },
  { id: 'users:delete', resource: 'users', action: 'delete', name: 'Delete Users', description: 'Delete users from system' },
  
  // API Keys
  { id: 'api_keys:read', resource: 'api_keys', action: 'read', name: 'View API Keys', description: 'View all API keys' },
  { id: 'api_keys:revoke', resource: 'api_keys', action: 'revoke', name: 'Revoke API Keys', description: 'Revoke any API key' },
  
  // System Settings
  { id: 'system:read', resource: 'system', action: 'read', name: 'View Settings', description: 'View system configuration' },
  { id: 'system:write', resource: 'system', action: 'write', name: 'Edit Settings', description: 'Modify system configuration' },
  
  // Audit Logs
  { id: 'logs:read', resource: 'audit_logs', action: 'read', name: 'View Logs', description: 'Access audit trail' },
];

export const mockRoles: RoleDefinition[] = [
  {
    id: 'admin',
    name: 'Administrator',
    description: 'Full system access',
    isSystem: true,
    permissions: mockPermissions.map(p => p.id),
    userCount: 5,
    createdAt: new Date('2023-01-01').toISOString()
  },
  {
    id: 'moderator',
    name: 'Moderator',
    description: 'Can manage users but not system settings',
    isSystem: true,
    permissions: ['users:read', 'users:write', 'logs:read', 'api_keys:read'],
    userCount: 12,
    createdAt: new Date('2023-02-15').toISOString()
  },
  {
    id: 'user',
    name: 'User',
    description: 'Standard user access',
    isSystem: true,
    permissions: [],
    userCount: 1217,
    createdAt: new Date('2023-03-20').toISOString()
  },
  {
    id: 'support',
    name: 'Support Agent',
    description: 'Read-only access to users and logs',
    isSystem: false,
    permissions: ['users:read', 'logs:read'],
    userCount: 3,
    createdAt: new Date('2023-06-10').toISOString()
  }
];

// Mock IP Rules
export const mockIpRules: IpRule[] = [
  {
    id: 'ip_1',
    type: 'blacklist',
    ipAddress: '192.168.1.55',
    description: 'Malicious bot activity',
    createdAt: new Date('2023-10-15').toISOString(),
    createdBy: 'System'
  },
  {
    id: 'ip_2',
    type: 'whitelist',
    ipAddress: '10.0.0.0/8',
    description: 'Internal corporate network',
    createdAt: new Date('2023-01-01').toISOString(),
    createdBy: 'Admin'
  }
];

// Mock Webhooks
export const mockWebhooks: WebhookEndpoint[] = [
  {
    id: 'wh_1',
    url: 'https://api.myapp.com/webhooks/auth',
    description: 'Main application sync',
    events: ['user.created', 'user.deleted', 'user.updated'],
    secret: 'whsec_test_1234567890',
    isActive: true,
    failureCount: 0,
    lastTriggeredAt: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
    createdAt: new Date('2023-08-01').toISOString()
  },
  {
    id: 'wh_2',
    url: 'https://analytics.example.com/ingest',
    description: 'Analytics tracking',
    events: ['auth.login.success', 'auth.login.failed'],
    secret: 'whsec_test_abcdefghij',
    isActive: false,
    failureCount: 12,
    lastTriggeredAt: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(),
    createdAt: new Date('2023-09-15').toISOString()
  }
];

// Mock Service Accounts
export const mockServiceAccounts: ServiceAccount[] = [
  {
    id: 'sa_1',
    name: 'Payment Service',
    description: 'Backend service for processing payments',
    clientId: 'svc_payment_8a7d9f2',
    isActive: true,
    createdAt: new Date('2023-05-10').toISOString(),
    lastUsedAt: new Date().toISOString()
  },
  {
    id: 'sa_2',
    name: 'Notification Worker',
    description: 'Async worker for sending emails',
    clientId: 'svc_notify_3k2j1h4',
    isActive: true,
    createdAt: new Date('2023-06-22').toISOString(),
    lastUsedAt: new Date(Date.now() - 1000 * 60 * 60).toISOString()
  }
];

// Mock Branding
export let mockBranding: BrandingConfig = {
  companyName: 'Auth Gateway',
  logoUrl: '',
  faviconUrl: '',
  primaryColor: '#2563EB', // blue-600
  accentColor: '#1E40AF', // blue-800
  backgroundColor: '#F3F4F6', // gray-100
  loginPageTitle: 'Sign in to your account',
  loginPageSubtitle: 'Welcome back! Please enter your details.',
  showSocialLogins: true
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
          sessions.push({
              id: randomId(),
              userId: user.id,
              deviceType: Math.random() > 0.6 ? 'desktop' : 'mobile',
              os: Math.random() > 0.5 ? (Math.random() > 0.5 ? 'Windows 11' : 'macOS 14') : (Math.random() > 0.5 ? 'iOS 17' : 'Android 14'),
              browser: Math.random() > 0.3 ? 'Chrome 120.0' : 'Safari 17.0',
              ipAddress: `192.168.${Math.floor(Math.random()*255)}.${Math.floor(Math.random()*255)}`,
              lastActive: randomDate(new Date(Date.now() - 24 * 60 * 60 * 1000), new Date()), // recent
              isCurrent: i === 0 && Math.random() > 0.5,
              location: ['New York, US', 'London, UK', 'Berlin, DE', 'Tokyo, JP'][Math.floor(Math.random() * 4)]
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
  const newUser: User = {
    id: randomId(),
    email: data.email || '',
    username: data.username || '',
    fullName: data.fullName || '',
    role: data.role || UserRole.USER,
    isActive: data.isActive ?? true,
    isEmailVerified: data.isEmailVerified ?? false,
    is2FAEnabled: data.is2FAEnabled ?? false,
    phone: data.phone,
    createdAt: new Date().toISOString(),
    avatarUrl: `https://picsum.photos/seed/${Math.random()}/200/200`
  };
  mockUsers.unshift(newUser);
  return newUser;
};

// User Actions
export const resetUserTwoFA = (userId: string): boolean => {
  const user = updateUser(userId, { is2FAEnabled: false });
  return !!user;
}

export const sendPasswordResetEmail = (userId: string): boolean => {
  // Mock sending logic
  console.log(`Sending password reset to user ${userId}`);
  return true;
}

export const getUserApiKeys = (userId: string): ApiKey[] => {
  return mockApiKeys.filter(k => k.ownerId === userId);
};

export const getUserLogs = (userId: string): AuditLog[] => {
  return mockLogs.filter(l => l.userId === userId).sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
};

export const getUserOAuthAccounts = (userId: string): OAuthAccount[] => {
  const providers: OAuthAccount['provider'][] = ['google', 'github', 'yandex', 'telegram'];
  const count = Math.floor(Math.random() * 3);
  
  return Array.from({ length: count }).map((_, i) => ({
    id: `oauth_${userId}_${i}`,
    provider: providers[i % providers.length],
    userId: userId,
    userName: `user_oauth_${i}`,
    connectedAt: randomDate(new Date(2023, 0, 1), new Date())
  }));
};

export const getUserSessions = (userId: string): UserSession[] => {
  return mockSessions.filter(s => s.userId === userId).sort((a, b) => new Date(b.lastActive).getTime() - new Date(a.lastActive).getTime());
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

export const createOAuthProvider = (data: Omit<OAuthProviderConfig, 'id' | 'createdAt'>): OAuthProviderConfig => {
  const newProvider: OAuthProviderConfig = {
    ...data,
    id: randomId(),
    createdAt: new Date().toISOString()
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
    mockEmailTemplates[index] = { ...mockEmailTemplates[index], ...data, lastUpdated: new Date().toISOString() };
    return mockEmailTemplates[index];
  }
  return undefined;
};

// RBAC Access
export const getRoles = (): RoleDefinition[] => {
  return mockRoles;
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
  
  const newPerm: Permission = {
    id: id,
    name: data.name || `${resource} ${action}`,
    resource: resource,
    action: action,
    description: data.description || ''
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
  const newRole: RoleDefinition = {
    id: data.name?.toLowerCase().replace(/\s+/g, '_') || randomId(),
    name: data.name || 'New Role',
    description: data.description || '',
    isSystem: false,
    permissions: data.permissions || [],
    userCount: 0,
    createdAt: new Date().toISOString()
  };
  mockRoles.push(newRole);
  return newRole;
};

export const deleteRole = (id: string): boolean => {
  const index = mockRoles.findIndex(r => r.id === id);
  if (index !== -1 && !mockRoles[index].isSystem) {
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
  const newRule: IpRule = {
    id: randomId(),
    type: data.type || 'blacklist',
    ipAddress: data.ipAddress || '',
    description: data.description,
    createdAt: new Date().toISOString(),
    createdBy: 'Admin'
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
  const newWebhook: WebhookEndpoint = {
    id: randomId(),
    url: data.url || '',
    description: data.description,
    events: data.events || [],
    secret: `whsec_${randomId()}${randomId()}`,
    isActive: data.isActive ?? true,
    failureCount: 0,
    createdAt: new Date().toISOString()
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
    clientId: `svc_${randomId()}`,
    isActive: data.isActive ?? true,
    createdAt: new Date().toISOString()
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
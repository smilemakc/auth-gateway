# Frontend Component Decomposition — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Decompose 71 frontend components following SOLID & DRY principles, reducing average component size from ~271 LOC to 60-120 LOC.

**Architecture:** Bottom-up approach — first create shared UI components and hooks as foundation, then restructure i18n, then decompose feature components from largest to smallest.

**Tech Stack:** React 19, TypeScript, Tailwind CSS v4, React Query (TanStack), lucide-react icons, Vite

**Important context:**
- No frontend test framework exists — verify via `npm run build` and manual dev server check (`npm run dev`)
- All components use `React.FC` or `export default function` pattern
- Translations accessed via `useLanguage()` hook → `t('namespace.key')`
- API calls via React Query hooks in `frontend/hooks/`
- Imperative services: `confirm()` and `toast` from `frontend/services/`
- Path alias: `@` = `/frontend/`
- All Tailwind classes use design tokens: `bg-card`, `text-foreground`, `border-border`, etc.

---

## Phase 1: Shared UI Components

### Task 1.1: Create FormField component

**Files:**
- Create: `frontend/components/ui/FormField.tsx`
- Create: `frontend/components/ui/index.ts`

**Step 1: Create the FormField component**

```tsx
// frontend/components/ui/FormField.tsx
import React from 'react';

interface FormFieldProps {
  label: string;
  required?: boolean;
  error?: string;
  children: React.ReactNode;
  className?: string;
}

const FormField: React.FC<FormFieldProps> = ({ label, required, error, children, className }) => {
  return (
    <div className={className}>
      <label className="block text-sm font-medium text-foreground mb-1">
        {label} {required && '*'}
      </label>
      {children}
      {error && (
        <p className="text-xs text-destructive mt-1">{error}</p>
      )}
    </div>
  );
};

export default FormField;
```

**Step 2: Create the TextInput component**

```tsx
// frontend/components/ui/TextInput.tsx
import React from 'react';

interface TextInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'className'> {
  variant?: 'default' | 'compact';
  datalistId?: string;
  datalistOptions?: string[];
}

const TextInput: React.FC<TextInputProps> = ({
  variant = 'default',
  datalistId,
  datalistOptions,
  ...props
}) => {
  const padding = variant === 'compact' ? 'px-3 py-2' : 'px-4 py-2';

  return (
    <>
      <input
        {...props}
        list={datalistId}
        className={`w-full ${padding} border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none text-sm`}
      />
      {datalistId && datalistOptions && (
        <datalist id={datalistId}>
          {datalistOptions.map((opt) => (
            <option key={opt} value={opt} />
          ))}
        </datalist>
      )}
    </>
  );
};

export default TextInput;
```

**Step 3: Create barrel export**

```tsx
// frontend/components/ui/index.ts
export { default as FormField } from './FormField';
export { default as TextInput } from './TextInput';
```

**Step 4: Verify build**

Run: `cd /Users/balashov/PycharmProjects/auth-gateway/frontend && npm run build`
Expected: Build succeeds with no errors

**Step 5: Commit**

```bash
git add frontend/components/ui/
git commit -m "feat(ui): add FormField and TextInput shared components"
```

---

### Task 1.2: Create StatCard component

**Files:**
- Create: `frontend/components/ui/StatCard.tsx`
- Modify: `frontend/components/ui/index.ts`

**Step 1: Create StatCard**

Extract the repeated stat card pattern from `AccessControl.tsx:89-123`.

```tsx
// frontend/components/ui/StatCard.tsx
import React from 'react';

interface StatCardProps {
  icon: React.ReactNode;
  iconBgClass?: string;
  value: number | string;
  label: string;
}

const StatCard: React.FC<StatCardProps> = ({
  icon,
  iconBgClass = 'bg-primary/10',
  value,
  label,
}) => {
  return (
    <div className="bg-card border border-border rounded-xl p-4">
      <div className="flex items-center gap-3">
        <div className={`p-2 ${iconBgClass} rounded-lg`}>
          {icon}
        </div>
        <div>
          <p className="text-2xl font-bold text-foreground">{value}</p>
          <p className="text-sm text-muted-foreground">{label}</p>
        </div>
      </div>
    </div>
  );
};

export default StatCard;
```

**Step 2: Add to barrel export**

Add to `frontend/components/ui/index.ts`:
```tsx
export { default as StatCard } from './StatCard';
```

**Step 3: Verify build**

Run: `cd /Users/balashov/PycharmProjects/auth-gateway/frontend && npm run build`

**Step 4: Commit**

```bash
git add frontend/components/ui/StatCard.tsx frontend/components/ui/index.ts
git commit -m "feat(ui): add StatCard shared component"
```

---

### Task 1.3: Create SearchInput component

**Files:**
- Create: `frontend/components/ui/SearchInput.tsx`
- Modify: `frontend/components/ui/index.ts`

**Step 1: Create SearchInput**

Extract the search pattern from `AccessControl.tsx:127-136`.

```tsx
// frontend/components/ui/SearchInput.tsx
import React from 'react';
import { Search } from 'lucide-react';

interface SearchInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

const SearchInput: React.FC<SearchInputProps> = ({
  value,
  onChange,
  placeholder = '',
  className = '',
}) => {
  return (
    <div className={`relative flex-1 ${className}`}>
      <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
      <input
        type="text"
        placeholder={placeholder}
        value={value}
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
        className="w-full pl-10 pr-4 py-2.5 border border-input rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-ring bg-card"
      />
    </div>
  );
};

export default SearchInput;
```

**Step 2: Add to barrel export and verify**

**Step 3: Commit**

```bash
git add frontend/components/ui/SearchInput.tsx frontend/components/ui/index.ts
git commit -m "feat(ui): add SearchInput shared component"
```

---

### Task 1.4: Create EmptyState and LoadingSpinner components

**Files:**
- Create: `frontend/components/ui/EmptyState.tsx`
- Create: `frontend/components/ui/LoadingSpinner.tsx`
- Modify: `frontend/components/ui/index.ts`

**Step 1: Create EmptyState**

```tsx
// frontend/components/ui/EmptyState.tsx
import React from 'react';

interface EmptyStateProps {
  icon: React.ReactNode;
  message: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

const EmptyState: React.FC<EmptyStateProps> = ({ icon, message, action }) => {
  return (
    <div className="text-center py-12 bg-card rounded-xl border border-border">
      <div className="mx-auto mb-4 text-muted-foreground opacity-50">
        {icon}
      </div>
      <p className="text-muted-foreground">{message}</p>
      {action && (
        <button
          onClick={action.onClick}
          className="mt-4 text-primary hover:underline text-sm font-medium"
        >
          {action.label}
        </button>
      )}
    </div>
  );
};

export default EmptyState;
```

**Step 2: Create LoadingSpinner**

```tsx
// frontend/components/ui/LoadingSpinner.tsx
import React from 'react';
import { Loader2 } from 'lucide-react';

interface LoadingSpinnerProps {
  className?: string;
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ className = 'h-64' }) => {
  return (
    <div className={`flex items-center justify-center ${className}`}>
      <Loader2 className="w-8 h-8 animate-spin text-primary" />
    </div>
  );
};

export default LoadingSpinner;
```

**Step 3: Update barrel export, verify build, commit**

```bash
git add frontend/components/ui/
git commit -m "feat(ui): add EmptyState and LoadingSpinner components"
```

---

### Task 1.5: Create PageHeader component

**Files:**
- Create: `frontend/components/ui/PageHeader.tsx`
- Modify: `frontend/components/ui/index.ts`

**Step 1: Create PageHeader**

Extract the repeated header pattern (back button + title + subtitle + action).

```tsx
// frontend/components/ui/PageHeader.tsx
import React from 'react';
import { ArrowLeft } from 'lucide-react';

interface PageHeaderProps {
  title: string;
  subtitle?: string;
  onBack?: () => void;
  action?: React.ReactNode;
}

const PageHeader: React.FC<PageHeaderProps> = ({ title, subtitle, onBack, action }) => {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div className="flex items-center gap-4">
        {onBack && (
          <button
            onClick={onBack}
            className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
          >
            <ArrowLeft size={24} />
          </button>
        )}
        <div>
          <h1 className="text-2xl font-bold text-foreground">{title}</h1>
          {subtitle && (
            <p className="text-sm text-muted-foreground mt-1">{subtitle}</p>
          )}
        </div>
      </div>
      {action && <div>{action}</div>}
    </div>
  );
};

export default PageHeader;
```

**Step 2: Update barrel export, verify build, commit**

```bash
git add frontend/components/ui/
git commit -m "feat(ui): add PageHeader shared component"
```

---

### Task 1.6: Create ActionButton component

**Files:**
- Create: `frontend/components/ui/ActionButton.tsx`
- Modify: `frontend/components/ui/index.ts`

**Step 1: Create ActionButton**

Extract the repeated button pattern (primary, danger, ghost variants + loading state).

```tsx
// frontend/components/ui/ActionButton.tsx
import React from 'react';
import { Loader2 } from 'lucide-react';

type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'ghost';

interface ActionButtonProps extends Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, 'className'> {
  variant?: ButtonVariant;
  isLoading?: boolean;
  icon?: React.ReactNode;
  children: React.ReactNode;
}

const variantClasses: Record<ButtonVariant, string> = {
  primary: 'bg-primary text-primary-foreground hover:bg-primary/90',
  secondary: 'bg-card text-foreground border border-input hover:bg-accent',
  danger: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
  ghost: 'text-muted-foreground hover:text-foreground hover:bg-accent',
};

const ActionButton: React.FC<ActionButtonProps> = ({
  variant = 'primary',
  isLoading = false,
  icon,
  children,
  disabled,
  ...props
}) => {
  return (
    <button
      {...props}
      disabled={disabled || isLoading}
      className={`flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors disabled:opacity-50 whitespace-nowrap ${variantClasses[variant]}`}
    >
      {isLoading ? <Loader2 size={16} className="animate-spin" /> : icon}
      {children}
    </button>
  );
};

export default ActionButton;
```

**Step 2: Update barrel export, verify build, commit**

```bash
git add frontend/components/ui/
git commit -m "feat(ui): add ActionButton shared component"
```

---

## Phase 2: Shared Hooks

### Task 2.1: Split useRBAC into domain-specific hooks

**Files:**
- Create: `frontend/hooks/rbac/useRoles.ts`
- Create: `frontend/hooks/rbac/usePermissions.ts`
- Create: `frontend/hooks/rbac/useRolePermissions.ts`
- Create: `frontend/hooks/rbac/useUserRoles.ts`
- Create: `frontend/hooks/rbac/index.ts`
- Delete: `frontend/hooks/useRBAC.ts` (after updating all imports)
- Modify: all files that import from `useRBAC` (see step for finding them)

**Step 1: Create useRoles.ts**

Extract lines 1-57 from `hooks/useRBAC.ts`:

```tsx
// frontend/hooks/rbac/useRoles.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../services/apiClient';
import { queryKeys } from '../../services/queryClient';
import { useCurrentAppId } from '../useAppAwareQuery';

export function useRoles(page: number = 1, pageSize: number = 50) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.rbac.roles.list(page, pageSize), appId],
    queryFn: () => apiClient.admin.rbac.listRoles(),
  });
}

export function useRoleDetail(roleId: string) {
  return useQuery({
    queryKey: queryKeys.rbac.roles.detail(roleId),
    queryFn: () => apiClient.admin.rbac.getRole(roleId),
    enabled: !!roleId,
  });
}

export function useCreateRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => apiClient.admin.rbac.createRole(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.all });
    },
  });
}

export function useUpdateRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      apiClient.admin.rbac.updateRole(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.detail(variables.id) });
    },
  });
}

export function useDeleteRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (roleId: string) => apiClient.admin.rbac.deleteRole(roleId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.all });
    },
  });
}
```

**Step 2: Create usePermissions.ts**

Extract lines 59-111 from `hooks/useRBAC.ts`:

```tsx
// frontend/hooks/rbac/usePermissions.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../services/apiClient';
import { queryKeys } from '../../services/queryClient';
import { useCurrentAppId } from '../useAppAwareQuery';

export function usePermissions(page: number = 1, pageSize: number = 100) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.rbac.permissions.list(page, pageSize), appId],
    queryFn: () => apiClient.admin.rbac.listPermissions(),
  });
}

export function usePermissionDetail(permissionId: string) {
  return useQuery({
    queryKey: queryKeys.rbac.permissions.detail(permissionId),
    queryFn: () => apiClient.admin.rbac.getPermission(permissionId),
    enabled: !!permissionId,
  });
}

export function useCreatePermission() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; resource: string; action: string; description?: string }) =>
      apiClient.admin.rbac.createPermission(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.permissions.all });
    },
  });
}

export function useUpdatePermission() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: { name?: string; description?: string } }) =>
      apiClient.admin.rbac.updatePermission(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.permissions.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.permissions.detail(variables.id) });
    },
  });
}

export function useDeletePermission() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (permissionId: string) => apiClient.admin.rbac.deletePermission(permissionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.permissions.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.all });
    },
  });
}
```

**Step 3: Create useRolePermissions.ts**

Extract lines 113-136 from `hooks/useRBAC.ts`:

```tsx
// frontend/hooks/rbac/useRolePermissions.ts
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../services/apiClient';
import { queryKeys } from '../../services/queryClient';

export function useAssignPermissionToRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ roleId, permissionId }: { roleId: string; permissionId: string }) =>
      apiClient.admin.rbac.addPermissionsToRole(roleId, [permissionId]),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.detail(variables.roleId) });
    },
  });
}

export function useRevokePermissionFromRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ roleId, permissionId }: { roleId: string; permissionId: string }) =>
      apiClient.admin.rbac.removePermissionsFromRole(roleId, [permissionId]),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.detail(variables.roleId) });
    },
  });
}
```

**Step 4: Create useUserRoles.ts**

Extract lines 138-176 from `hooks/useRBAC.ts`:

```tsx
// frontend/hooks/rbac/useUserRoles.ts
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../services/apiClient';
import { queryKeys } from '../../services/queryClient';

export function useAssignUserRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ userId, roleId }: { userId: string; roleId: string }) => {
      return await apiClient.admin.users.assignRole(userId, roleId);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.detail(variables.userId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
    },
  });
}

export function useRemoveUserRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ userId, roleId }: { userId: string; roleId: string }) => {
      return await apiClient.admin.users.removeRole(userId, roleId);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.detail(variables.userId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
    },
  });
}

export function useSetUserRoles() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ userId, roleIds }: { userId: string; roleIds: string[] }) => {
      return await apiClient.admin.users.setRoles(userId, roleIds);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.detail(variables.userId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
    },
  });
}
```

**Step 5: Create barrel export**

```tsx
// frontend/hooks/rbac/index.ts
export { useRoles, useRoleDetail, useCreateRole, useUpdateRole, useDeleteRole } from './useRoles';
export { usePermissions, usePermissionDetail, useCreatePermission, useUpdatePermission, useDeletePermission } from './usePermissions';
export { useAssignPermissionToRole, useRevokePermissionFromRole } from './useRolePermissions';
export { useAssignUserRole, useRemoveUserRole, useSetUserRoles } from './useUserRoles';
```

**Step 6: Find all files importing from useRBAC and update imports**

Run: `grep -rl "from.*hooks/useRBAC" frontend/` to find all files.

For each file, change:
```tsx
// OLD:
import { useRoles, useCreateRole } from '../hooks/useRBAC';
// NEW:
import { useRoles, useCreateRole } from '../hooks/rbac';
```

Known consumers (verify with grep):
- `frontend/components/AccessControl.tsx` — uses `useRoles`, `usePermissions`
- `frontend/components/AccessControlCreateRoleForm.tsx` — uses `useCreateRole`
- `frontend/components/AccessControlRoleCard.tsx` — uses `useUpdateRole`, `useDeleteRole`
- `frontend/components/AccessControlPermissionsSection.tsx` — uses `useCreatePermission`, `useDeletePermission`
- `frontend/components/Roles.tsx` — uses role hooks
- `frontend/components/RoleEditor.tsx` — uses role + permission hooks
- `frontend/components/Permissions.tsx` — uses permission hooks
- `frontend/components/PermissionEdit.tsx` — uses permission hooks
- `frontend/components/BulkAssignRoles.tsx` — uses `useRoles`, `useSetUserRoles`
- `frontend/components/UserDetails.tsx` — may use user role hooks
- `frontend/components/UserEdit.tsx` — may use user role hooks

**Step 7: Delete old file**

Delete `frontend/hooks/useRBAC.ts` after all imports updated.

**Step 8: Verify build**

Run: `cd /Users/balashov/PycharmProjects/auth-gateway/frontend && npm run build`
Expected: Build succeeds with no errors

**Step 9: Commit**

```bash
git add frontend/hooks/rbac/ frontend/hooks/useRBAC.ts frontend/components/
git commit -m "refactor(hooks): split useRBAC into domain-specific hooks

Separated 12 exports in useRBAC.ts into 4 focused hook files:
- useRoles.ts (5 hooks)
- usePermissions.ts (5 hooks)
- useRolePermissions.ts (2 hooks)
- useUserRoles.ts (3 hooks)"
```

---

## Phase 3: i18n Restructuring

### Task 3.1: Create i18n module structure with Russian translations

**Files:**
- Create: `frontend/services/i18n/types.ts`
- Create: `frontend/services/i18n/translations/ru/common.ts`
- Create: `frontend/services/i18n/translations/ru/auth.ts`
- Create: `frontend/services/i18n/translations/ru/nav.ts`
- Create: `frontend/services/i18n/translations/ru/users.ts`
- Create: `frontend/services/i18n/translations/ru/applications.ts`
- Create: `frontend/services/i18n/translations/ru/access-control.ts`
- Create: `frontend/services/i18n/translations/ru/oauth.ts`
- Create: `frontend/services/i18n/translations/ru/email.ts`
- Create: `frontend/services/i18n/translations/ru/security.ts`
- Create: `frontend/services/i18n/translations/ru/integrations.ts`
- Create: `frontend/services/i18n/translations/ru/dashboard.ts`
- Create: `frontend/services/i18n/translations/ru/index.ts`

**Step 1: Create types**

```tsx
// frontend/services/i18n/types.ts
export type Language = 'ru' | 'en';

export interface LanguageContextType {
  language: Language;
  setLanguage: (lang: Language) => void;
  t: (key: string) => string;
}
```

**Step 2: Split Russian translations by namespace**

Extract from `frontend/services/i18n.tsx` lines 13-1162. Group by namespace prefix:

```tsx
// frontend/services/i18n/translations/ru/common.ts
// Keys: common.*, confirm.*, breadcrumb.*
const common: Record<string, string> = {
  'common.save': 'Сохранить',
  'common.saving': 'Сохранение...',
  // ... copy ALL common.* keys from i18n.tsx ru block
  // ... copy ALL confirm.* keys
  // ... copy ALL breadcrumb.* keys
};
export default common;
```

Namespace-to-file mapping:

| File | Namespaces | Approx keys |
|------|-----------|-------------|
| `common.ts` | `common.*`, `confirm.*`, `breadcrumb.*` | ~60 |
| `auth.ts` | `auth.*` | ~7 |
| `nav.ts` | `nav.*` | ~31 |
| `users.ts` | `user.*`, `users.*`, `sessions.*`, `groups.*`, `group_details.*`, `group_edit.*`, `bulk.*`, `bulk_update.*`, `bulk_assign.*` | ~200 |
| `applications.ts` | `apps.*`, `app_oauth.*`, `brand.*` | ~160 |
| `access-control.ts` | `access_control.*`, `roles.*`, `perms.*`, `perm_edit.*` | ~50 |
| `oauth.ts` | `oauth.*`, `oauth_edit.*`, `oauth_clients.*` | ~75 |
| `email.ts` | `email.*`, `email_tpl.*` | ~95 |
| `security.ts` | `ip.*`, `audit.*`, `keys.*` | ~40 |
| `integrations.ts` | `ldap.*`, `ldap_edit.*`, `ldap_sync.*`, `saml.*`, `saml_edit.*`, `scim.*`, `tg.*`, `hooks.*`, `webhooks.*`, `sms.*`, `sys.*` | ~190 |
| `dashboard.ts` | `dash.*`, `inspector.*`, `settings.*` | ~60 |

**Step 3: Create Russian index**

```tsx
// frontend/services/i18n/translations/ru/index.ts
import common from './common';
import auth from './auth';
import nav from './nav';
import users from './users';
import applications from './applications';
import accessControl from './access-control';
import oauth from './oauth';
import email from './email';
import security from './security';
import integrations from './integrations';
import dashboard from './dashboard';

const ru: Record<string, string> = {
  ...common,
  ...auth,
  ...nav,
  ...users,
  ...applications,
  ...accessControl,
  ...oauth,
  ...email,
  ...security,
  ...integrations,
  ...dashboard,
};

export default ru;
```

**Step 4: Verify build, commit**

```bash
git add frontend/services/i18n/
git commit -m "refactor(i18n): split Russian translations into domain modules"
```

---

### Task 3.2: Create English translations and context/provider

**Files:**
- Create: `frontend/services/i18n/translations/en/` (mirror ru structure)
- Create: `frontend/services/i18n/translations/index.ts`
- Create: `frontend/services/i18n/context.tsx`
- Create: `frontend/services/i18n/useLanguage.ts`
- Create: `frontend/services/i18n/index.ts`

**Step 1: Split English translations** (same structure as Russian)

Copy from `frontend/services/i18n.tsx` lines 1163-2313.

**Step 2: Create translations index**

```tsx
// frontend/services/i18n/translations/index.ts
import type { Language } from '../types';
import ru from './ru';
import en from './en';

const translations: Record<Language, Record<string, string>> = { ru, en };
export default translations;
```

**Step 3: Create context**

```tsx
// frontend/services/i18n/context.tsx
import React, { createContext, useState, useEffect, type ReactNode } from 'react';
import type { Language, LanguageContextType } from './types';
import translations from './translations';

export const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

export const LanguageProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [language, setLanguageState] = useState<Language>('ru');

  useEffect(() => {
    const savedLang = localStorage.getItem('app_language') as Language;
    if (savedLang && (savedLang === 'en' || savedLang === 'ru')) {
      setLanguageState(savedLang);
    }
  }, []);

  const setLanguage = (lang: Language) => {
    setLanguageState(lang);
    localStorage.setItem('app_language', lang);
  };

  const t = (key: string): string => {
    const value = translations[language][key];
    if (!value && import.meta.env.DEV) {
      console.warn(`[i18n] Missing translation key: "${key}" for language: "${language}"`);
    }
    return value || key;
  };

  return (
    <LanguageContext.Provider value={{ language, setLanguage, t }}>
      {children}
    </LanguageContext.Provider>
  );
};
```

**Step 4: Create useLanguage hook**

```tsx
// frontend/services/i18n/useLanguage.ts
import { useContext } from 'react';
import { LanguageContext } from './context';
import type { LanguageContextType } from './types';

export const useLanguage = (): LanguageContextType => {
  const context = useContext(LanguageContext);
  if (!context) {
    throw new Error('useLanguage must be used within a LanguageProvider');
  }
  return context;
};
```

**Step 5: Create public API**

```tsx
// frontend/services/i18n/index.ts
export { LanguageProvider } from './context';
export { useLanguage } from './useLanguage';
export type { Language, LanguageContextType } from './types';
```

**Step 6: Delete old i18n.tsx**

Delete `frontend/services/i18n.tsx`.

All existing imports use `from '../services/i18n'` — since we're replacing a file with a directory containing `index.ts`, the import paths remain identical. No consumer changes needed.

**Step 7: Verify build**

Run: `cd /Users/balashov/PycharmProjects/auth-gateway/frontend && npm run build`
Expected: Build succeeds with no errors. Check that no translation keys are missing in dev console.

**Step 8: Commit**

```bash
git add frontend/services/i18n/ frontend/services/i18n.tsx
git commit -m "refactor(i18n): complete restructuring into modular architecture

Split 2354-line monolith into domain modules:
- 11 translation files per language
- Separate context, hook, and types
- Import paths unchanged (file → directory/index.ts)"
```

---

## Phase 4: AccessControl Decomposition

### Task 4.1: Decompose AccessControl using shared UI components

**Files:**
- Create: `frontend/components/access-control/AccessControlStats.tsx`
- Create: `frontend/components/access-control/AccessControlRoleList.tsx`
- Modify: `frontend/components/AccessControl.tsx` → move to `frontend/components/access-control/AccessControl.tsx`
- Move: `frontend/components/AccessControlCreateRoleForm.tsx` → `frontend/components/access-control/AccessControlCreateRoleForm.tsx`
- Create: `frontend/components/access-control/index.ts`
- Modify: `frontend/App.tsx` — update import path

**Step 1: Create AccessControlStats**

Extract stats section from `AccessControl.tsx:89-123`:

```tsx
// frontend/components/access-control/AccessControlStats.tsx
import React from 'react';
import { Lock, Shield, Users } from 'lucide-react';
import { StatCard } from '../ui';
import { useLanguage } from '../../services/i18n';

interface AccessControlStatsProps {
  rolesCount: number;
  permissionsCount: number;
  resourcesCount: number;
}

const AccessControlStats: React.FC<AccessControlStatsProps> = ({
  rolesCount,
  permissionsCount,
  resourcesCount,
}) => {
  const { t } = useLanguage();

  return (
    <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <StatCard
        icon={<Shield className="h-5 w-5 text-primary" />}
        iconBgClass="bg-primary/10"
        value={rolesCount}
        label={t('roles.title')}
      />
      <StatCard
        icon={<Lock className="h-5 w-5 text-accent-foreground" />}
        iconBgClass="bg-accent"
        value={permissionsCount}
        label={t('perms.title')}
      />
      <StatCard
        icon={<Users className="h-5 w-5 text-muted-foreground" />}
        iconBgClass="bg-muted"
        value={resourcesCount}
        label={t('access_control.resources')}
      />
    </div>
  );
};

export default AccessControlStats;
```

**Step 2: Create AccessControlRoleList**

Extract role list + search + empty state from `AccessControl.tsx:126-173`:

```tsx
// frontend/components/access-control/AccessControlRoleList.tsx
import React from 'react';
import { Plus, Shield } from 'lucide-react';
import { SearchInput, EmptyState, ActionButton } from '../ui';
import { useLanguage } from '../../services/i18n';
import AccessControlRoleCard from './AccessControlRoleCard';
import type { Permission, RoleDefinition } from '../../types';

interface AccessControlRoleListProps {
  roles: RoleDefinition[];
  groupedPermissions: Record<string, Permission[]>;
  searchTerm: string;
  onSearchChange: (value: string) => void;
  expandedRoles: Set<string>;
  onToggleExpand: (roleId: string) => void;
  onCreateRole: () => void;
}

const AccessControlRoleList: React.FC<AccessControlRoleListProps> = ({
  roles,
  groupedPermissions,
  searchTerm,
  onSearchChange,
  expandedRoles,
  onToggleExpand,
  onCreateRole,
}) => {
  const { t } = useLanguage();

  return (
    <>
      <div className="flex flex-col sm:flex-row gap-4">
        <SearchInput
          value={searchTerm}
          onChange={onSearchChange}
          placeholder={t('access_control.search_roles')}
        />
        <ActionButton icon={<Plus size={18} />} onClick={onCreateRole}>
          {t('access_control.create_role')}
        </ActionButton>
      </div>

      <div className="space-y-4">
        {roles.map((role) => (
          <AccessControlRoleCard
            key={role.id}
            role={role}
            groupedPermissions={groupedPermissions}
            isExpanded={expandedRoles.has(role.id)}
            onToggleExpand={() => onToggleExpand(role.id)}
          />
        ))}

        {roles.length === 0 && (
          <EmptyState
            icon={<Shield size={48} />}
            message={t('access_control.no_roles')}
            action={{ label: t('access_control.create_first_role'), onClick: onCreateRole }}
          />
        )}
      </div>
    </>
  );
};

export default AccessControlRoleList;
```

**Step 3: Rewrite AccessControl as thin orchestrator**

```tsx
// frontend/components/access-control/AccessControl.tsx
import React, { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLanguage } from '../../services/i18n';
import { useRoles, usePermissions } from '../../hooks/rbac';
import { LoadingSpinner, PageHeader } from '../ui';
import type { Permission } from '../../types';
import AccessControlStats from './AccessControlStats';
import AccessControlRoleList from './AccessControlRoleList';
import AccessControlCreateRoleForm from './AccessControlCreateRoleForm';
import AccessControlPermissionsSection from './AccessControlPermissionsSection';

const AccessControl: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();

  const [searchTerm, setSearchTerm] = useState('');
  const [expandedRoles, setExpandedRoles] = useState<Set<string>>(new Set());
  const [showCreateRole, setShowCreateRole] = useState(false);

  const { data: roles = [], isLoading: rolesLoading } = useRoles();
  const { data: permissions = [], isLoading: permissionsLoading } = usePermissions();

  const groupedPermissions = useMemo(() => {
    return permissions.reduce((acc, perm) => {
      if (!acc[perm.resource]) acc[perm.resource] = [];
      acc[perm.resource].push(perm);
      return acc;
    }, {} as Record<string, Permission[]>);
  }, [permissions]);

  const filteredRoles = useMemo(() =>
    roles.filter(r =>
      (r.display_name || r.name).toLowerCase().includes(searchTerm.toLowerCase()) ||
      r.description?.toLowerCase().includes(searchTerm.toLowerCase())
    ), [roles, searchTerm]);

  const toggleRoleExpand = (roleId: string) => {
    setExpandedRoles(prev => {
      const newSet = new Set(prev);
      if (newSet.has(roleId)) newSet.delete(roleId);
      else newSet.add(roleId);
      return newSet;
    });
  };

  if (rolesLoading || permissionsLoading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('nav.access_settings')}
        subtitle={t('access_control.subtitle')}
        onBack={() => navigate('/settings')}
      />

      <AccessControlStats
        rolesCount={roles.length}
        permissionsCount={permissions.length}
        resourcesCount={Object.keys(groupedPermissions).length}
      />

      {showCreateRole && (
        <AccessControlCreateRoleForm onClose={() => setShowCreateRole(false)} />
      )}

      <AccessControlRoleList
        roles={filteredRoles}
        groupedPermissions={groupedPermissions}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        expandedRoles={expandedRoles}
        onToggleExpand={toggleRoleExpand}
        onCreateRole={() => setShowCreateRole(true)}
      />

      <AccessControlPermissionsSection
        permissions={permissions}
        groupedPermissions={groupedPermissions}
      />
    </div>
  );
};

export default AccessControl;
```

**Step 4: Move existing AccessControl sub-components into the directory**

Move files:
- `frontend/components/AccessControlCreateRoleForm.tsx` → `frontend/components/access-control/AccessControlCreateRoleForm.tsx`
- `frontend/components/AccessControlRoleCard.tsx` → `frontend/components/access-control/AccessControlRoleCard.tsx`
- `frontend/components/AccessControlPermissionsSection.tsx` → `frontend/components/access-control/AccessControlPermissionsSection.tsx`

Update their internal imports (change `../hooks/useRBAC` → `../../hooks/rbac`, `../services/i18n` → `../../services/i18n`, `../types` → `../../types`, etc.).

**Step 5: Create barrel export**

```tsx
// frontend/components/access-control/index.ts
export { default as AccessControl } from './AccessControl';
```

**Step 6: Update App.tsx import**

Change:
```tsx
// OLD:
import AccessControl from './components/AccessControl';
// NEW:
import { AccessControl } from './components/access-control';
```

**Step 7: Delete old files from components/**

Delete the 4 old AccessControl* files from `frontend/components/`.

**Step 8: Verify build**

Run: `cd /Users/balashov/PycharmProjects/auth-gateway/frontend && npm run build`

**Step 9: Commit**

```bash
git add frontend/components/access-control/ frontend/components/AccessControl*.tsx frontend/App.tsx
git commit -m "refactor(access-control): decompose into directory with shared UI

- AccessControl: thin orchestrator using PageHeader, LoadingSpinner
- AccessControlStats: extracted stats using StatCard
- AccessControlRoleList: extracted role list using SearchInput, EmptyState, ActionButton"
```

---

### Task 4.2: Decompose AccessControlRoleCard

**Files:**
- Create: `frontend/components/access-control/AccessControlPermissionGrid.tsx`
- Create: `frontend/components/access-control/AccessControlResourceGroup.tsx`
- Modify: `frontend/components/access-control/AccessControlRoleCard.tsx`

**Step 1: Create AccessControlResourceGroup**

Extract the resource group rendering from `AccessControlRoleCard.tsx:183-241`:

```tsx
// frontend/components/access-control/AccessControlResourceGroup.tsx
import React from 'react';
import { Check } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { Permission } from '../../types';

interface AccessControlResourceGroupProps {
  resource: string;
  permissions: Permission[];
  selectedPermIds: string[];
  pendingUpdate: boolean;
  onTogglePermission: (permissionId: string) => void;
  onSelectAll: (resource: string, select: boolean) => void;
}

const AccessControlResourceGroup: React.FC<AccessControlResourceGroupProps> = ({
  resource,
  permissions,
  selectedPermIds,
  pendingUpdate,
  onTogglePermission,
  onSelectAll,
}) => {
  const { t } = useLanguage();
  const selectedCount = permissions.filter(p => selectedPermIds.includes(p.id)).length;
  const allSelected = selectedCount === permissions.length;

  return (
    <div className="bg-card rounded-lg border border-border p-4">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <h5 className="font-medium text-foreground capitalize">
            {resource.replace(/_/g, ' ')}
          </h5>
          <span className="text-xs text-muted-foreground">
            ({selectedCount}/{permissions.length})
          </span>
        </div>
        <button
          onClick={() => onSelectAll(resource, !allSelected)}
          disabled={pendingUpdate}
          className={`text-xs font-medium px-2 py-1 rounded transition-colors ${
            allSelected
              ? 'bg-primary/10 text-primary hover:bg-primary/20'
              : 'bg-muted text-muted-foreground hover:bg-accent'
          }`}
        >
          {allSelected ? t('access_control.deselect_all') : t('access_control.select_all')}
        </button>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-2">
        {permissions.map(perm => {
          const isChecked = selectedPermIds.includes(perm.id);
          return (
            <button
              key={perm.id}
              onClick={() => onTogglePermission(perm.id)}
              disabled={pendingUpdate}
              className={`
                flex items-center gap-2 px-3 py-2 rounded-lg text-left transition-all text-sm
                ${isChecked
                  ? 'bg-primary text-primary-foreground shadow-sm'
                  : 'bg-muted/50 text-foreground hover:bg-muted border border-transparent hover:border-border'
                }
                ${pendingUpdate ? 'opacity-50 cursor-wait' : 'cursor-pointer'}
              `}
              title={perm.description || `${perm.resource}:${perm.action}`}
            >
              <div className={`
                w-4 h-4 rounded flex items-center justify-center shrink-0
                ${isChecked ? 'bg-primary-foreground/20' : 'bg-card border border-input'}
              `}>
                {isChecked && <Check size={12} />}
              </div>
              <span className="truncate font-medium capitalize">{perm.action}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
};

export default AccessControlResourceGroup;
```

**Step 2: Create AccessControlPermissionGrid**

Extract the expanded permissions section from `AccessControlRoleCard.tsx:158-246`:

```tsx
// frontend/components/access-control/AccessControlPermissionGrid.tsx
import React from 'react';
import { AlertCircle, Lock } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { Permission } from '../../types';
import AccessControlResourceGroup from './AccessControlResourceGroup';

interface AccessControlPermissionGridProps {
  groupedPermissions: Record<string, Permission[]>;
  selectedPermIds: string[];
  pendingUpdate: boolean;
  onTogglePermission: (permissionId: string) => void;
  onSelectAllResource: (resource: string, select: boolean) => void;
}

const AccessControlPermissionGrid: React.FC<AccessControlPermissionGridProps> = ({
  groupedPermissions,
  selectedPermIds,
  pendingUpdate,
  onTogglePermission,
  onSelectAllResource,
}) => {
  const { t } = useLanguage();

  return (
    <div className="border-t border-border bg-muted/30">
      <div className="p-4">
        <div className="flex items-center justify-between mb-4">
          <h4 className="font-medium text-foreground flex items-center gap-2">
            <Lock size={16} className="text-primary" />
            {t('roles.permissions')}
          </h4>
          <p className="text-xs text-muted-foreground">
            {t('access_control.toggle_hint')}
          </p>
        </div>

        {Object.keys(groupedPermissions).length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <AlertCircle size={32} className="mx-auto mb-2 opacity-50" />
            <p>{t('access_control.no_permissions_hint')}</p>
          </div>
        ) : (
          <div className="space-y-6">
            {(Object.entries(groupedPermissions) as [string, Permission[]][]).map(([resource, perms]) => (
              <AccessControlResourceGroup
                key={resource}
                resource={resource}
                permissions={perms}
                selectedPermIds={selectedPermIds}
                pendingUpdate={pendingUpdate}
                onTogglePermission={onTogglePermission}
                onSelectAll={onSelectAllResource}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default AccessControlPermissionGrid;
```

**Step 3: Simplify AccessControlRoleCard**

Rewrite to use the extracted sub-components. The card now only has the header + delegates expansion to `AccessControlPermissionGrid`.

The new `AccessControlRoleCard.tsx` should be ~100 LOC: role header (display name, badge, perm count, delete button, expand/collapse) + the permission toggle/select-all handlers + render `AccessControlPermissionGrid` when expanded.

**Step 4: Verify build, commit**

```bash
git add frontend/components/access-control/
git commit -m "refactor(access-control): decompose RoleCard into ResourceGroup and PermissionGrid

- AccessControlResourceGroup: per-resource permission selection
- AccessControlPermissionGrid: expanded permissions container
- AccessControlRoleCard: now ~100 LOC (was 251)"
```

---

### Task 4.3: Decompose AccessControlPermissionsSection

**Files:**
- Create: `frontend/components/access-control/AccessControlPermissionForm.tsx`
- Modify: `frontend/components/access-control/AccessControlPermissionsSection.tsx`

**Step 1: Create AccessControlPermissionForm**

Extract the permission creation form from `AccessControlPermissionsSection.tsx:103-198`:

```tsx
// frontend/components/access-control/AccessControlPermissionForm.tsx
import React, { useCallback, useState } from 'react';
import { Loader2 } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { useCreatePermission } from '../../hooks/rbac';
import { FormField, TextInput, ActionButton } from '../ui';
import { logger } from '@/lib/logger';

const COMMON_ACTIONS = ['create', 'read', 'update', 'delete', 'list', 'manage', 'export', 'import'];

interface AccessControlPermissionFormProps {
  existingResources: string[];
  onClose: () => void;
}

const AccessControlPermissionForm: React.FC<AccessControlPermissionFormProps> = ({
  existingResources,
  onClose,
}) => {
  const { t } = useLanguage();
  const createPermissionMutation = useCreatePermission();

  const [resource, setResource] = useState('');
  const [action, setAction] = useState('');
  const [description, setDescription] = useState('');

  const handleCreate = useCallback(async () => {
    if (!resource.trim() || !action.trim()) return;
    try {
      await createPermissionMutation.mutateAsync({
        name: `${resource}:${action}`,
        resource: resource.toLowerCase().replace(/\s+/g, '_'),
        action: action.toLowerCase().replace(/\s+/g, '_'),
        description,
      });
      setResource('');
      setAction('');
      setDescription('');
      onClose();
    } catch (err) {
      logger.error('Failed to create permission:', err);
    }
  }, [resource, action, description, createPermissionMutation, onClose]);

  return (
    <div className="p-4 bg-muted/30 border-b border-border">
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
        <FormField label={t('access_control.resource_label')} required>
          <TextInput
            value={resource}
            onChange={(e) => setResource(e.target.value)}
            placeholder={t('perm_edit.resource_placeholder')}
            variant="compact"
            datalistId="resources-list"
            datalistOptions={existingResources}
          />
          {existingResources.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-2">
              {existingResources.slice(0, 5).map((r) => (
                <button
                  key={r}
                  type="button"
                  onClick={() => setResource(r)}
                  className="px-2 py-0.5 text-xs bg-muted hover:bg-accent rounded transition-colors"
                >
                  {r}
                </button>
              ))}
            </div>
          )}
        </FormField>

        <FormField label={t('access_control.action_label')} required>
          <TextInput
            value={action}
            onChange={(e) => setAction(e.target.value)}
            placeholder={t('perm_edit.action_placeholder')}
            variant="compact"
            datalistId="actions-list"
            datalistOptions={COMMON_ACTIONS}
          />
          <div className="flex flex-wrap gap-1 mt-2">
            {COMMON_ACTIONS.slice(0, 5).map((a) => (
              <button
                key={a}
                type="button"
                onClick={() => setAction(a)}
                className="px-2 py-0.5 text-xs bg-muted hover:bg-accent rounded transition-colors"
              >
                {a}
              </button>
            ))}
          </div>
        </FormField>

        <FormField label={t('common.description')}>
          <TextInput
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder={t('common.description')}
            variant="compact"
          />
        </FormField>
      </div>

      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {t('access_control.perm_name_will_be')}: <code className="bg-muted px-1.5 py-0.5 rounded">
            {resource || 'resource'}:{action || 'action'}
          </code>
        </p>
        <ActionButton
          onClick={handleCreate}
          disabled={!resource.trim() || !action.trim()}
          isLoading={createPermissionMutation.isPending}
        >
          {t('access_control.create_permission')}
        </ActionButton>
      </div>
    </div>
  );
};

export default AccessControlPermissionForm;
```

**Step 2: Simplify AccessControlPermissionsSection**

Rewrite to delegate form rendering to `AccessControlPermissionForm`. The section now only handles:
- Header with toggle button
- Delegating form to `AccessControlPermissionForm`
- Rendering grouped permissions list with delete buttons

Target: ~120 LOC (was 263).

**Step 3: Verify build, commit**

```bash
git add frontend/components/access-control/
git commit -m "refactor(access-control): extract PermissionForm from PermissionsSection

- AccessControlPermissionForm: self-contained form using FormField, TextInput
- AccessControlPermissionsSection: now ~120 LOC (was 263)"
```

---

### Task 4.4: Update AccessControlCreateRoleForm to use shared UI

**Files:**
- Modify: `frontend/components/access-control/AccessControlCreateRoleForm.tsx`

**Step 1: Refactor to use FormField, TextInput, ActionButton**

Replace inline label+input patterns with `FormField` + `TextInput`, and button with `ActionButton`.

Target: ~60 LOC (was 84). Minor change but establishes the pattern.

**Step 2: Verify build, commit**

```bash
git add frontend/components/access-control/AccessControlCreateRoleForm.tsx
git commit -m "refactor(access-control): update CreateRoleForm to use shared UI components"
```

---

## Phase 5: Large Component Decomposition

### Task 5.1: Decompose UserDetails (559 LOC → 6 components)

**Files:**
- Create: `frontend/components/users/UserProfileCard.tsx`
- Create: `frontend/components/users/UserSecuritySection.tsx`
- Create: `frontend/components/users/UserSessionsSection.tsx`
- Create: `frontend/components/users/UserOAuthSection.tsx`
- Create: `frontend/components/users/UserAuditSection.tsx`
- Create: `frontend/components/users/UserAppProfileSection.tsx`
- Modify: `frontend/components/UserDetails.tsx` → move to `frontend/components/users/UserDetails.tsx`
- Create: `frontend/components/users/index.ts`
- Modify: `frontend/App.tsx` — update import path

**Step 1: Read the full UserDetails.tsx**

Read the complete file to identify section boundaries.

**Step 2: Extract UserProfileCard**

Extract the user info section (avatar, email, phone, created date, status badges) into `UserProfileCard.tsx`.

Props: `user`, `onEdit` callback.

**Step 3: Extract UserSecuritySection**

Extract the 2FA reset, password reset, and danger zone into `UserSecuritySection.tsx`.

Props: `userId`, `user`.

**Step 4: Extract UserSessionsSection**

Extract the sessions list with revoke buttons into `UserSessionsSection.tsx`.

Props: `userId`.

**Step 5: Extract UserOAuthSection**

Extract the OAuth accounts display into `UserOAuthSection.tsx`.

Props: `userId`.

**Step 6: Extract UserAuditSection**

Extract the recent activity/audit log into `UserAuditSection.tsx`.

Props: `userId`, `limit`.

**Step 7: Extract UserAppProfileSection**

Extract the application profile display into `UserAppProfileSection.tsx`.

Props: `userId`, `applicationId`.

**Step 8: Rewrite UserDetails as orchestrator**

UserDetails becomes a thin orchestrator (~80 LOC) that:
- Fetches user data
- Shows loading/error states
- Renders header (PageHeader)
- Composes the 6 sub-components

**Step 9: Create barrel export, update App.tsx**

**Step 10: Verify build, commit**

```bash
git add frontend/components/users/ frontend/components/UserDetails.tsx frontend/App.tsx
git commit -m "refactor(users): decompose UserDetails into 6 focused components

- UserProfileCard: user info display
- UserSecuritySection: 2FA, password reset, danger zone
- UserSessionsSection: active sessions with revoke
- UserOAuthSection: linked OAuth accounts
- UserAuditSection: recent activity log
- UserAppProfileSection: per-app profile
- UserDetails: thin orchestrator (~80 LOC, was 559)"
```

---

### Task 5.2: Decompose Layout (423 LOC → 3 components)

**Files:**
- Create: `frontend/components/layout/Sidebar.tsx`
- Create: `frontend/components/layout/TopBar.tsx`
- Modify: `frontend/components/Layout.tsx` → move to `frontend/components/layout/Layout.tsx`
- Create: `frontend/components/layout/index.ts`
- Modify: `frontend/App.tsx`

**Step 1: Read full Layout.tsx**

**Step 2: Extract Sidebar**

Extract navigation groups, sidebar state, active item highlighting into `Sidebar.tsx`.

Props: `isOpen`, `onClose`, `expandedGroups`, `onToggleGroup`.

**Step 3: Extract TopBar**

Extract header bar (hamburger menu, breadcrumb, theme switcher, language switcher, app selector, notifications bell) into `TopBar.tsx`.

Props: `onToggleSidebar`.

**Step 4: Rewrite Layout as shell**

Layout becomes a thin shell (~80 LOC) orchestrating Sidebar + TopBar + content area.

**Step 5: Verify build, commit**

```bash
git add frontend/components/layout/ frontend/components/Layout.tsx frontend/App.tsx
git commit -m "refactor(layout): decompose Layout into Sidebar and TopBar

- Sidebar: navigation groups with expand/collapse
- TopBar: header with theme, language, app selector
- Layout: thin shell (~80 LOC, was 423)"
```

---

### Task 5.3: Decompose OAuthClientEdit (542 LOC → 4 components)

**Files:**
- Create: `frontend/components/oauth/OAuthClientBasicFields.tsx`
- Create: `frontend/components/oauth/OAuthClientScopeSelector.tsx`
- Create: `frontend/components/oauth/OAuthClientSecretSection.tsx`
- Modify: `frontend/components/OAuthClientEdit.tsx` → move to `frontend/components/oauth/OAuthClientEdit.tsx`
- Create: `frontend/components/oauth/index.ts`

Follow the same pattern: read file, identify sections, extract, simplify orchestrator, verify, commit.

---

### Task 5.4: Decompose LDAPConfigEdit (519 LOC → 4 components)

Same pattern as Task 5.3.

**Files:**
- Create: `frontend/components/ldap/LDAPConnectionFields.tsx`
- Create: `frontend/components/ldap/LDAPSearchFields.tsx`
- Create: `frontend/components/ldap/LDAPMappingFields.tsx`
- Modify: `frontend/components/LDAPConfigEdit.tsx` → `frontend/components/ldap/LDAPConfigEdit.tsx`

---

### Task 5.5: Decompose EmailProviderEdit (514 LOC → 3 components)

Same pattern.

**Files:**
- Create: `frontend/components/email/EmailProviderSMTPFields.tsx`
- Create: `frontend/components/email/EmailProviderTestSection.tsx`
- Modify: `frontend/components/EmailProviderEdit.tsx` → `frontend/components/email/EmailProviderEdit.tsx`

---

### Task 5.6: Decompose remaining large components

Apply the same orchestrator + sections pattern to all remaining components >250 LOC. Priority order (by size):

1. `UserEdit.tsx` (424 LOC) → `users/UserEditBasicFields.tsx` + `users/UserEditAuthFields.tsx`
2. `BulkCreateUsers.tsx` (410 LOC) → `bulk/BulkCreateUsers.tsx` + `bulk/BulkCreateCSVParser.tsx` + `bulk/BulkCreateManualEntry.tsx` + `bulk/BulkCreateResults.tsx`
3. `ApplicationOAuthProviderEdit.tsx` (374 LOC) → split by form sections
4. `ApplicationUsersTab.tsx` (372 LOC) → split table + import modal
5. `Settings.tsx` (354 LOC) → split by settings groups
6. `ApiKeys.tsx` (354 LOC) → split list + create form
7. `Branding.tsx` (333 LOC) → split preview + form
8. `UsersImportModal.tsx` (326 LOC) → split steps
9. `EmailProviders.tsx` (323 LOC) → split list + management
10. Continue for remaining components >250 LOC

For each: read full file → identify sections → extract → simplify orchestrator → verify build → commit.

---

## Phase 6: Migrate Existing Components to Shared UI

### Task 6.1: Audit and replace inline patterns

After all decomposition is done, do a final pass to replace remaining inline patterns across ALL components:

1. Search for `<label className="block text-sm font-medium` → replace with `FormField`
2. Search for `<Loader2 className="w-8 h-8 animate-spin` → replace with `LoadingSpinner`
3. Search for `text-center py-12 bg-card` → replace with `EmptyState`
4. Search for `bg-card border border-border rounded-xl p-4` + stat pattern → replace with `StatCard`
5. Search for `ArrowLeft` + header pattern → replace with `PageHeader`

Run: `cd /Users/balashov/PycharmProjects/auth-gateway/frontend && npm run build`

Commit after each batch of replacements.

---

## Verification Checklist

After completing all phases:

1. `npm run build` — succeeds with no errors
2. `npm run dev` — app loads, navigate through all pages
3. All AccessControl features work (create role, toggle permissions, delete, search)
4. All translation strings render correctly in both RU and EN
5. No console warnings for missing translation keys
6. No broken imports or module resolution errors

---

## Summary

| Phase | Tasks | Est. Files Changed | Key Outcome |
|-------|-------|-------------------|-------------|
| 1. Shared UI | 6 tasks | 8 new files | Reusable FormField, StatCard, SearchInput, EmptyState, LoadingSpinner, PageHeader, ActionButton |
| 2. Shared Hooks | 1 task | 5 new, 10+ modified | useRBAC → 4 focused hooks |
| 3. i18n | 2 tasks | 25+ new files | 2354 LOC → 11 modules per language |
| 4. AccessControl | 4 tasks | 10 new/modified | 782 LOC → 10 focused components |
| 5. Large Components | 6 tasks | 30+ new/modified | Top-10 largest decomposed |
| 6. Migration | 1 task | 50+ modified | All components use shared UI |

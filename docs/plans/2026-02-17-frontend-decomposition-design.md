# Frontend Component Decomposition Design

**Date:** 2026-02-17
**Approach:** Bottom-Up (shared foundation first, then component refactoring)
**Principles:** SOLID & DRY

## Problem

- 71 components, average size ~271 LOC (target: 60-120 LOC)
- 45% of components exceed 250 lines
- ~15-20% code duplication across form patterns, buttons, validation
- i18n.tsx is 2,353 LOC in a single file
- useRBAC.ts has 20 operations in one hook

## Phase 1: Shared UI Components

Extract reusable UI patterns into `frontend/components/ui/`:

| Component | Purpose | Usage |
|-----------|---------|-------|
| `FormField` | label + input + error + required marker | 50+ places |
| `StatCard` | icon + value + label | Dashboard, AccessControl, UserDetails |
| `SearchInput` | search with icon and debounce | Users, Roles, Permissions, Applications |
| `DataTable` | table with sorting, pagination | 15+ list components |
| `PageHeader` | title + breadcrumb + action buttons | every page |
| `EmptyState` | empty list state | all lists |

## Phase 2: Shared Hooks

Extract reusable business logic into `frontend/hooks/`:

| Hook | Purpose |
|------|---------|
| `useFormValidation` | Form state: values, errors, touched, validate, reset |
| `useConfirmAction` | Dangerous action confirmation dialog |
| `usePaginatedQuery` | Pagination + sorting + filtering for lists |
| `useDebounce` | Debounced values for search |

### useRBAC Split (no backward compatibility needed)

```
frontend/hooks/rbac/
  ├── useRoles.ts           — Role CRUD (queries + mutations)
  ├── usePermissions.ts     — Permission CRUD
  ├── useRolePermissions.ts — assign/revoke permissions to roles
  ├── useUserRoles.ts       — assign/remove roles to users
  └── index.ts
```

## Phase 3: i18n Restructuring

```
frontend/services/i18n/
  ├── context.tsx           — LanguageContext + LanguageProvider (~40 LOC)
  ├── useLanguage.ts        — useLanguage hook (~10 LOC)
  ├── types.ts              — TranslationKeys, Language types
  ├── translations/
  │   ├── ru/
  │   │   ├── common.ts     — shared strings (buttons, statuses)
  │   │   ├── auth.ts       — authentication
  │   │   ├── users.ts      — users
  │   │   ├── roles.ts      — roles & permissions
  │   │   ├── applications.ts — applications
  │   │   ├── settings.ts   — settings
  │   │   └── index.ts      — aggregates all modules
  │   ├── en/
  │   │   └── ... (mirror structure)
  │   └── index.ts          — export by language
  └── index.ts              — public API
```

## Phase 4: AccessControl Decomposition

```
frontend/components/access-control/
  ├── AccessControl.tsx              — orchestrator (stats + layout) ~80 LOC
  ├── AccessControlStats.tsx         — statistics (uses StatCard) ~30 LOC
  ├── AccessControlRoleList.tsx      — role list with search ~60 LOC
  ├── AccessControlRoleCard.tsx      — role card (header + delete) ~80 LOC
  ├── AccessControlPermissionGrid.tsx — permission grid for role ~70 LOC
  ├── AccessControlResourceGroup.tsx  — permissions grouped by resource ~50 LOC
  ├── AccessControlPermissionsSection.tsx — permission management section ~60 LOC
  ├── AccessControlPermissionForm.tsx — permission creation form ~60 LOC
  ├── AccessControlCreateRoleForm.tsx — role creation form ~50 LOC
  └── index.ts
```

## Phase 5: Large Component Decomposition

### UserDetails (559 LOC → 6 components)
```
components/users/
  ├── UserDetails.tsx           — orchestrator + layout (~80)
  ├── UserProfileCard.tsx       — basic info + avatar (~80)
  ├── UserSecuritySection.tsx   — 2FA, password, IP filters (~80)
  ├── UserSessionsSection.tsx   — sessions list + revoke (~80)
  ├── UserOAuthSection.tsx      — OAuth connections (~60)
  ├── UserAuditSection.tsx      — audit log (~60)
  └── UserAppProfileSection.tsx — app profile (~60)
```

### OAuthClientEdit (542 LOC → 4 components)
```
components/oauth/
  ├── OAuthClientEdit.tsx       — form orchestrator (~100)
  ├── OAuthClientBasicFields.tsx — name, type, redirect URIs (~120)
  ├── OAuthClientScopeSelector.tsx — scope selection (~100)
  └── OAuthClientSecretSection.tsx — secret display/rotate (~80)
```

### LDAPConfigEdit (519 LOC → 4 components)
```
components/ldap/
  ├── LDAPConfigEdit.tsx          — orchestrator (~80)
  ├── LDAPConnectionFields.tsx    — host, port, bind DN (~120)
  ├── LDAPSearchFields.tsx        — base DN, filter, attributes (~120)
  └── LDAPMappingFields.tsx       — attribute mapping (~100)
```

### EmailProviderEdit (514 LOC → 3 components)
```
components/email/
  ├── EmailProviderEdit.tsx       — orchestrator + type selector (~100)
  ├── EmailProviderSMTPFields.tsx — SMTP config fields (~150)
  └── EmailProviderTestSection.tsx — test send (~80)
```

### Layout (423 LOC → 3 components)
```
components/layout/
  ├── Layout.tsx        — shell + routing (~100)
  ├── Sidebar.tsx       — navigation (~150)
  └── TopBar.tsx        — header + user menu (~80)
```

### Remaining components >250 LOC

Same strategy: **orchestrator + sections**. Each component splits into:
- Orchestrator (data + layout) — up to 100 LOC
- Sections by responsibility — 60-120 LOC each

## Target Metrics

| Metric | Current | Target |
|--------|---------|--------|
| Avg component size | ~271 LOC | 60-120 LOC |
| Components >250 LOC | 32 (45%) | 0 (0%) |
| Code duplication | ~15-20% | <5% |
| Max hook operations | 20 (useRBAC) | 5-7 per hook |
| i18n file size | 2,353 LOC | <100 LOC per file |

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';

// Types
export interface EmailProvider {
  id: string;
  name: string;
  type: 'smtp' | 'sendgrid' | 'ses' | 'mailgun';
  is_active: boolean;
  created_at: string;
  updated_at: string;
  smtp_host?: string;
  smtp_port?: number;
  smtp_username?: string;
  smtp_use_tls?: boolean;
  ses_region?: string;
  ses_access_key_id?: string;
  mailgun_domain?: string;
  has_smtp_password?: boolean;
  has_sendgrid_api_key?: boolean;
  has_ses_secret_access_key?: boolean;
  has_mailgun_api_key?: boolean;
}

export interface CreateEmailProviderRequest {
  name: string;
  type: 'smtp' | 'sendgrid' | 'ses' | 'mailgun';
  is_active: boolean;
  smtp_host?: string;
  smtp_port?: number;
  smtp_username?: string;
  smtp_password?: string;
  smtp_use_tls?: boolean;
  sendgrid_api_key?: string;
  ses_region?: string;
  ses_access_key_id?: string;
  ses_secret_access_key?: string;
  mailgun_domain?: string;
  mailgun_api_key?: string;
}

export interface UpdateEmailProviderRequest {
  name?: string;
  is_active?: boolean;
  smtp_host?: string;
  smtp_port?: number;
  smtp_username?: string;
  smtp_password?: string;
  smtp_use_tls?: boolean;
  sendgrid_api_key?: string;
  ses_region?: string;
  ses_access_key_id?: string;
  ses_secret_access_key?: string;
  mailgun_domain?: string;
  mailgun_api_key?: string;
}

export interface EmailProfile {
  id: string;
  name: string;
  provider_id: string;
  from_email: string;
  from_name: string;
  reply_to?: string;
  is_default: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  provider?: EmailProvider;
}

export interface CreateEmailProfileRequest {
  name: string;
  provider_id: string;
  from_email: string;
  from_name: string;
  reply_to?: string;
  is_default?: boolean;
  is_active?: boolean;
}

export interface UpdateEmailProfileRequest {
  name?: string;
  provider_id?: string;
  from_email?: string;
  from_name?: string;
  reply_to?: string;
  is_default?: boolean;
  is_active?: boolean;
}

// Query keys
export const emailQueryKeys = {
  providers: ['email-providers'] as const,
  provider: (id: string) => ['email-providers', id] as const,
  profiles: ['email-profiles'] as const,
  profile: (id: string) => ['email-profiles', id] as const,
  profileStats: (id: string) => ['email-profiles', id, 'stats'] as const,
};

// Email Providers Hooks
export function useEmailProviders() {
  return useQuery<EmailProvider[]>({
    queryKey: emailQueryKeys.providers,
    queryFn: () => apiClient.admin.emailProviders.list(),
  });
}

export function useEmailProvider(id: string) {
  return useQuery<EmailProvider>({
    queryKey: emailQueryKeys.provider(id),
    queryFn: () => apiClient.admin.emailProviders.get(id),
    enabled: !!id && id !== 'new',
  });
}

export function useCreateEmailProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateEmailProviderRequest) =>
      apiClient.admin.emailProviders.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: emailQueryKeys.providers });
    },
  });
}

export function useUpdateEmailProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateEmailProviderRequest }) =>
      apiClient.admin.emailProviders.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: emailQueryKeys.providers });
      queryClient.invalidateQueries({ queryKey: emailQueryKeys.provider(variables.id) });
    },
  });
}

export function useDeleteEmailProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) =>
      apiClient.admin.emailProviders.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: emailQueryKeys.providers });
    },
  });
}

export function useTestEmailProvider() {
  return useMutation({
    mutationFn: ({ id, email }: { id: string; email: string }) =>
      apiClient.admin.emailProviders.test(id, email),
  });
}

// Email Profiles Hooks
export function useEmailProfiles() {
  return useQuery<EmailProfile[]>({
    queryKey: emailQueryKeys.profiles,
    queryFn: () => apiClient.admin.emailProfiles.list(),
  });
}

export function useEmailProfile(id: string) {
  return useQuery<EmailProfile>({
    queryKey: emailQueryKeys.profile(id),
    queryFn: () => apiClient.admin.emailProfiles.get(id),
    enabled: !!id && id !== 'new',
  });
}

export function useCreateEmailProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateEmailProfileRequest) =>
      apiClient.admin.emailProfiles.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: emailQueryKeys.profiles });
    },
  });
}

export function useUpdateEmailProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateEmailProfileRequest }) =>
      apiClient.admin.emailProfiles.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: emailQueryKeys.profiles });
      queryClient.invalidateQueries({ queryKey: emailQueryKeys.profile(variables.id) });
    },
  });
}

export function useDeleteEmailProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) =>
      apiClient.admin.emailProfiles.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: emailQueryKeys.profiles });
    },
  });
}

export function useSetDefaultEmailProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) =>
      apiClient.admin.emailProfiles.setDefault(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: emailQueryKeys.profiles });
    },
  });
}

export function useTestEmailProfile() {
  return useMutation({
    mutationFn: ({ id, email }: { id: string; email: string }) =>
      apiClient.admin.emailProfiles.test(id, email),
  });
}

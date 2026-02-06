import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '../services/queryClient';
import type {
  Application,
  ApplicationBranding,
  UserApplicationProfile,
  CreateApplicationRequest,
  UpdateApplicationRequest,
  UpdateApplicationBrandingRequest,
  ListApplicationsResponse,
  ListApplicationUsersResponse,
} from '../types';

const API_BASE = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api`;

async function fetchWithAuth(url: string, options: RequestInit = {}) {
  const token = localStorage.getItem('auth_gateway_access_token');
  const appId = localStorage.getItem('auth_gateway_current_app_id');
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(appId ? { 'X-Application-ID': appId } : {}),
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(error.message || error.error || 'Request failed');
  }

  return response.json();
}

// Applications CRUD
export function useApplications(page: number = 1, pageSize: number = 20) {
  return useQuery<ListApplicationsResponse>({
    queryKey: queryKeys.applications.list(page, pageSize),
    queryFn: () => fetchWithAuth(`${API_BASE}/admin/applications?page=${page}&per_page=${pageSize}`),
  });
}

export function useApplicationDetail(id: string) {
  return useQuery<Application>({
    queryKey: queryKeys.applications.detail(id),
    queryFn: () => fetchWithAuth(`${API_BASE}/admin/applications/${id}`),
    enabled: !!id && id !== 'new',
  });
}

export function useCreateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateApplicationRequest) =>
      fetchWithAuth(`${API_BASE}/admin/applications`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.all });
    },
  });
}

export function useUpdateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateApplicationRequest }) =>
      fetchWithAuth(`${API_BASE}/admin/applications/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.detail(variables.id) });
    },
  });
}

export function useDeleteApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) =>
      fetchWithAuth(`${API_BASE}/admin/applications/${id}`, {
        method: 'DELETE',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.all });
    },
  });
}

// Application Branding
export function useApplicationBranding(applicationId: string) {
  return useQuery<ApplicationBranding>({
    queryKey: queryKeys.applications.branding(applicationId),
    queryFn: () => fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/branding`),
    enabled: !!applicationId && applicationId !== 'new',
  });
}

export function useUpdateApplicationBranding() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ applicationId, data }: { applicationId: string; data: UpdateApplicationBrandingRequest }) =>
      fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/branding`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.branding(variables.applicationId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.detail(variables.applicationId) });
    },
  });
}

// Application Users
export function useApplicationUsers(applicationId: string, page: number = 1, pageSize: number = 20) {
  return useQuery<ListApplicationUsersResponse>({
    queryKey: queryKeys.applications.users(applicationId, page, pageSize),
    queryFn: () => fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/users?page=${page}&per_page=${pageSize}`),
    enabled: !!applicationId && applicationId !== 'new',
  });
}

export function useBanUserFromApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ applicationId, userId, reason }: { applicationId: string; userId: string; reason: string }) =>
      fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/users/${userId}/ban`, {
        method: 'POST',
        body: JSON.stringify({ reason }),
      }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.users(variables.applicationId, 1, 20) });
    },
  });
}

export function useUnbanUserFromApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ applicationId, userId }: { applicationId: string; userId: string }) =>
      fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/users/${userId}/unban`, {
        method: 'POST',
      }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.users(variables.applicationId, 1, 20) });
    },
  });
}

// Public branding (no auth required)
export function usePublicApplicationBranding(applicationId: string) {
  return useQuery<ApplicationBranding>({
    queryKey: ['public', 'applications', applicationId, 'branding'],
    queryFn: async () => {
      const response = await fetch(`${API_BASE}/applications/${applicationId}/branding`);
      if (!response.ok) {
        throw new Error('Failed to fetch branding');
      }
      return response.json();
    },
    enabled: !!applicationId,
  });
}

// User's own profile in application
export function useMyApplicationProfile(applicationId?: string) {
  return useQuery<UserApplicationProfile>({
    queryKey: ['user', 'application-profile', applicationId],
    queryFn: () => {
      const url = applicationId
        ? `${API_BASE}/user/application-profile?application_id=${applicationId}`
        : `${API_BASE}/user/application-profile`;
      return fetchWithAuth(url);
    },
    enabled: true,
  });
}

export function useUpdateMyApplicationProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ applicationId, data }: { applicationId?: string; data: Partial<UserApplicationProfile> }) => {
      const url = applicationId
        ? `${API_BASE}/user/application-profile?application_id=${applicationId}`
        : `${API_BASE}/user/application-profile`;
      return fetchWithAuth(url, {
        method: 'PUT',
        body: JSON.stringify(data),
      });
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['user', 'application-profile', variables.applicationId] });
    },
  });
}

// Admin: Get a specific user's application profile
export function useUserApplicationProfile(userId: string, applicationId: string) {
  return useQuery<UserApplicationProfile | null>({
    queryKey: ['admin', 'user-application-profile', userId, applicationId],
    queryFn: async () => {
      // Fetch the user from the application users list and find the specific user
      const response = await fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/users?page=1&per_page=100`);
      const profile = response.profiles?.find((p: UserApplicationProfile) => p.user_id === userId);
      return profile || null;
    },
    enabled: !!userId && !!applicationId,
  });
}

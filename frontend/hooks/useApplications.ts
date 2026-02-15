import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '../services/queryClient';
import { apiClient } from '../services/apiClient';
import type {
  Application,
  ApplicationBranding,
  UserApplicationProfile,
  CreateApplicationRequest,
  UpdateApplicationRequest,
  UpdateApplicationBrandingRequest,
  ListApplicationsResponse,
  ListApplicationUsersResponse,
  ImportUsersRequest,
  ImportUsersResponse,
} from '../types';

// Applications CRUD
export function useApplications(page: number = 1, pageSize: number = 20) {
  return useQuery<ListApplicationsResponse>({
    queryKey: queryKeys.applications.list(page, pageSize),
    queryFn: () => apiClient.admin.applications.list(page, pageSize),
  });
}

export function useApplicationDetail(id: string) {
  return useQuery<Application>({
    queryKey: queryKeys.applications.detail(id),
    queryFn: () => apiClient.admin.applications.getById(id),
    enabled: !!id && id !== 'new',
  });
}

export function useCreateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateApplicationRequest) =>
      apiClient.admin.applications.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.all });
    },
  });
}

export function useUpdateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateApplicationRequest }) =>
      apiClient.admin.applications.update(id, data),
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
      apiClient.admin.applications.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.all });
    },
  });
}

// Application Branding
export function useApplicationBranding(applicationId: string) {
  return useQuery<ApplicationBranding>({
    queryKey: queryKeys.applications.branding(applicationId),
    queryFn: () => apiClient.admin.applications.getBranding(applicationId),
    enabled: !!applicationId && applicationId !== 'new',
  });
}

export function useUpdateApplicationBranding() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ applicationId, data }: { applicationId: string; data: UpdateApplicationBrandingRequest }) =>
      apiClient.admin.applications.updateBranding(applicationId, data),
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
    queryFn: () => apiClient.admin.applications.listUsers(applicationId, page, pageSize),
    enabled: !!applicationId && applicationId !== 'new',
  });
}

export function useBanUserFromApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ applicationId, userId, reason }: { applicationId: string; userId: string; reason: string }) =>
      apiClient.admin.applications.banUser(applicationId, userId, { reason }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.users(variables.applicationId, 1, 20) });
    },
  });
}

export function useUnbanUserFromApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ applicationId, userId }: { applicationId: string; userId: string }) =>
      apiClient.admin.applications.unbanUser(applicationId, userId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.users(variables.applicationId, 1, 20) });
    },
  });
}

export function useImportUsers() {
  const queryClient = useQueryClient();

  return useMutation<ImportUsersResponse, Error, { appId: string; data: ImportUsersRequest }>({
    mutationFn: async ({ data }: { appId: string; data: ImportUsersRequest }) => {
      return apiClient.admin.applications.importUsers(data);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.users(variables.appId, 1, 20) });
    },
  });
}

// Application Secret Rotation
export function useRotateApplicationSecret() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (appId: string) => {
      return apiClient.admin.applications.rotateSecret(appId);
    },
    onSuccess: (_, appId) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.detail(appId) });
    },
  });
}

// Public branding (no auth required)
export function usePublicApplicationBranding(applicationId: string) {
  return useQuery<ApplicationBranding>({
    queryKey: ['public', 'applications', applicationId, 'branding'],
    queryFn: () => apiClient.admin.applications.getPublicBranding(applicationId),
    enabled: !!applicationId,
  });
}

// User's own profile in application
export function useMyApplicationProfile(applicationId?: string) {
  return useQuery<UserApplicationProfile>({
    queryKey: ['user', 'application-profile', applicationId],
    queryFn: () => apiClient.admin.applications.getMyApplicationProfile(applicationId),
    enabled: true,
  });
}

export function useUpdateMyApplicationProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ applicationId, data }: { applicationId?: string; data: Partial<UserApplicationProfile> }) =>
      apiClient.admin.applications.updateMyApplicationProfile(data, applicationId),
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
      const response = await apiClient.admin.applications.listUsers(applicationId, 1, 100);
      const profile = response.profiles?.find((p: UserApplicationProfile) => p.user_id === userId);
      return profile || null;
    },
    enabled: !!userId && !!applicationId,
  });
}

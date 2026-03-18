import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';
import { useCurrentAppId } from './useAppAwareQuery';

export function useSessions(page: number = 1, pageSize: number = 50) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.sessions.list(page, pageSize), appId],
    queryFn: () => apiClient.admin.sessions.list(page, pageSize),
  });
}

export function useUserSessions(userId: string, page: number = 1, pageSize: number = 50) {
  return useQuery({
    queryKey: [...queryKeys.sessions.byUser(userId, page, pageSize)],
    queryFn: () => apiClient.admin.sessions.listUserSessions(userId, page, pageSize),
    enabled: !!userId,
  });
}

export function useSessionDetail(id: string) {
  return useQuery({
    queryKey: queryKeys.sessions.detail(id),
    queryFn: () => apiClient.admin.sessions.get(id),
    enabled: !!id,
  });
}

export function useRevokeSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (sessionId: string) => apiClient.admin.sessions.revoke(sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions.all });
    },
  });
}

export function useRevokeUserSessions() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: string) => apiClient.admin.sessions.revokeAllForUser(userId),
    onSuccess: (_, userId) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.sessions.byUser(userId) });
    },
  });
}

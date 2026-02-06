import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';
import type { IPFilterType } from '../../packages/client-sdk/src/types/common';
import { useCurrentAppId } from './useAppAwareQuery';

export function useIpFilters(type?: IPFilterType, page: number = 1, pageSize: number = 100) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.ipFilters.list(page, pageSize), appId],
    queryFn: () => apiClient.admin.ipFilters.list(type, page, pageSize),
  });
}

export function useWhitelistFilters(page: number = 1, pageSize: number = 100) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: ['ipFilters', 'whitelist', page, pageSize, appId],
    queryFn: () => apiClient.admin.ipFilters.list('whitelist', page, pageSize),
  });
}

export function useBlacklistFilters(page: number = 1, pageSize: number = 100) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: ['ipFilters', 'blacklist', page, pageSize, appId],
    queryFn: () => apiClient.admin.ipFilters.list('blacklist', page, pageSize),
  });
}

export function useCreateIpFilter() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      ip_address: string;
      type: 'whitelist' | 'blacklist';
      description?: string;
    }) => apiClient.admin.ipFilters.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ipFilters.all });
    },
  });
}

export function useDeleteIpFilter() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (filterId: string) => apiClient.admin.ipFilters.delete(filterId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ipFilters.all });
    },
  });
}

export function useAddToWhitelist() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ ipAddress, description }: { ipAddress: string; description?: string }) =>
      apiClient.admin.ipFilters.whitelist(ipAddress, description),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ipFilters.all });
      queryClient.invalidateQueries({ queryKey: ['ipFilters', 'whitelist'] });
    },
  });
}

export function useAddToBlacklist() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ ipAddress, description }: { ipAddress: string; description?: string }) =>
      apiClient.admin.ipFilters.blacklist(ipAddress, description),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ipFilters.all });
      queryClient.invalidateQueries({ queryKey: ['ipFilters', 'blacklist'] });
    },
  });
}

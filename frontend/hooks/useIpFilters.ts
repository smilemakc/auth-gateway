import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useIpFilters(page: number = 1, pageSize: number = 100) {
  return useQuery({
    queryKey: queryKeys.ipFilters.list(page, pageSize),
    queryFn: () => apiClient.admin.ipFilter.list(page, pageSize),
  });
}

export function useIpFilterDetail(filterId: string) {
  return useQuery({
    queryKey: queryKeys.ipFilters.detail(filterId),
    queryFn: () => apiClient.admin.ipFilter.get(filterId),
    enabled: !!filterId,
  });
}

export function useCreateIpFilter() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      ipAddress: string;
      type: 'whitelist' | 'blacklist';
      description?: string;
      enabled?: boolean;
    }) => apiClient.admin.ipFilter.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ipFilters.all });
    },
  });
}

export function useUpdateIpFilter() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      apiClient.admin.ipFilter.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ipFilters.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.ipFilters.detail(variables.id) });
    },
  });
}

export function useDeleteIpFilter() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (filterId: string) => apiClient.admin.ipFilter.delete(filterId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ipFilters.all });
    },
  });
}

export function useToggleIpFilter() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      apiClient.admin.ipFilter.update(id, { enabled }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ipFilters.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.ipFilters.detail(variables.id) });
    },
  });
}

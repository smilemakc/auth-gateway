import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useSCIMConfig() {
  return useQuery({
    queryKey: queryKeys.scim.config,
    queryFn: () => apiClient.admin.scim.getConfig(),
  });
}

export function useSCIMMetadata() {
  return useQuery({
    queryKey: queryKeys.scim.metadata,
    queryFn: () => apiClient.admin.scim.getMetadata(),
  });
}


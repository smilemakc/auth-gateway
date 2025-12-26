import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';
import type { CreateSAMLSPRequest, UpdateSAMLSPRequest } from '@auth-gateway/client-sdk';

export function useSAMLSPs(page: number = 1, pageSize: number = 20) {
  return useQuery({
    queryKey: queryKeys.saml.list(page, pageSize),
    queryFn: () => apiClient.admin.saml.listSPs(page, pageSize),
  });
}

export function useSAMLSP(spId: string) {
  return useQuery({
    queryKey: queryKeys.saml.detail(spId),
    queryFn: () => apiClient.admin.saml.getSP(spId),
    enabled: !!spId,
  });
}

export function useCreateSAMLSP() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateSAMLSPRequest) => apiClient.admin.saml.createSP(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.saml.all });
    },
  });
}

export function useUpdateSAMLSP() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateSAMLSPRequest }) =>
      apiClient.admin.saml.updateSP(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.saml.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.saml.detail(variables.id) });
    },
  });
}

export function useDeleteSAMLSP() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (spId: string) => apiClient.admin.saml.deleteSP(spId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.saml.all });
    },
  });
}

export function useSAMLMetadata() {
  return useQuery({
    queryKey: queryKeys.saml.metadata,
    queryFn: () => apiClient.admin.saml.getMetadata(),
  });
}


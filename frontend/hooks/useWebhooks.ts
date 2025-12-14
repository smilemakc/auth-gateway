import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';
import type {
  CreateWebhookRequest,
  UpdateWebhookRequest,
  TestWebhookRequest,
} from '@auth-gateway/client-sdk';

export function useWebhooks(page: number = 1, pageSize: number = 50) {
  return useQuery({
    queryKey: queryKeys.webhooks.list(page, pageSize),
    queryFn: () => apiClient.admin.webhooks.list(page, pageSize),
  });
}

export function useWebhookDetail(webhookId: string) {
  return useQuery({
    queryKey: queryKeys.webhooks.detail(webhookId),
    queryFn: () => apiClient.admin.webhooks.get(webhookId),
    enabled: !!webhookId,
  });
}

export function useCreateWebhook() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateWebhookRequest) => apiClient.admin.webhooks.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.webhooks.all });
    },
  });
}

export function useUpdateWebhook() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateWebhookRequest }) =>
      apiClient.admin.webhooks.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.webhooks.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.webhooks.detail(variables.id) });
    },
  });
}

export function useDeleteWebhook() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (webhookId: string) => apiClient.admin.webhooks.delete(webhookId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.webhooks.all });
    },
  });
}

export function useToggleWebhook() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      apiClient.admin.webhooks.update(id, { is_active: enabled }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.webhooks.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.webhooks.detail(variables.id) });
    },
  });
}

export function useTestWebhook() {
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: TestWebhookRequest }) =>
      apiClient.admin.webhooks.test(id, data),
  });
}

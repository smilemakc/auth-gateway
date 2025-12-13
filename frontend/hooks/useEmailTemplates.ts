import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useEmailTemplates(page: number = 1, pageSize: number = 50) {
  return useQuery({
    queryKey: queryKeys.emailTemplates.list(page, pageSize),
    queryFn: () => apiClient.admin.templates.list(page, pageSize),
  });
}

export function useEmailTemplateDetail(templateId: string) {
  return useQuery({
    queryKey: queryKeys.emailTemplates.detail(templateId),
    queryFn: () => apiClient.admin.templates.get(templateId),
    enabled: !!templateId,
  });
}

export function useUpdateEmailTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      apiClient.admin.templates.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.emailTemplates.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.emailTemplates.detail(variables.id) });
    },
  });
}

export function useResetEmailTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (templateId: string) => apiClient.admin.templates.reset(templateId),
    onSuccess: (_, templateId) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.emailTemplates.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.emailTemplates.detail(templateId) });
    },
  });
}

export function usePreviewEmailTemplate() {
  return useMutation({
    mutationFn: ({ templateId, data }: { templateId: string; data: any }) =>
      apiClient.admin.templates.preview(templateId, data),
  });
}

export function useSendTestEmail() {
  return useMutation({
    mutationFn: ({ templateId, email }: { templateId: string; email: string }) =>
      apiClient.admin.templates.sendTest(templateId, email),
  });
}

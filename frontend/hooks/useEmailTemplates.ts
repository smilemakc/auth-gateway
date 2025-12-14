import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type {
  EmailTemplate,
  EmailTemplateType,
  CreateEmailTemplateRequest,
  UpdateEmailTemplateRequest,
  PreviewEmailTemplateRequest,
} from '@auth-gateway/client-sdk';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useEmailTemplates() {
  return useQuery({
    queryKey: queryKeys.emailTemplates.all,
    queryFn: () => apiClient.admin.templates.list(),
  });
}

export function useEmailTemplateDetail(templateId: string) {
  return useQuery({
    queryKey: queryKeys.emailTemplates.detail(templateId),
    queryFn: () => apiClient.admin.templates.get(templateId),
    enabled: !!templateId,
  });
}

export function useCreateEmailTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateEmailTemplateRequest) =>
      apiClient.admin.templates.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.emailTemplates.all });
    },
  });
}

export function useUpdateEmailTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateEmailTemplateRequest }) =>
      apiClient.admin.templates.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.emailTemplates.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.emailTemplates.detail(variables.id) });
    },
  });
}

export function useDeleteEmailTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.admin.templates.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.emailTemplates.all });
    },
  });
}

export function usePreviewEmailTemplate() {
  return useMutation({
    mutationFn: (data: PreviewEmailTemplateRequest) =>
      apiClient.admin.templates.preview(data),
  });
}

export function useEmailTemplateTypes() {
  return useQuery({
    queryKey: queryKeys.emailTemplates.types,
    queryFn: () => apiClient.admin.templates.getTypes(),
  });
}

export function useEmailTemplateVariables(type: EmailTemplateType | null) {
  return useQuery({
    queryKey: queryKeys.emailTemplates.variables(type as EmailTemplateType),
    queryFn: () => apiClient.admin.templates.getVariables(type as EmailTemplateType),
    enabled: !!type,
  });
}

export function useEmailTemplatesByType(type: EmailTemplateType | null) {
  return useQuery({
    queryKey: queryKeys.emailTemplates.byType(type as EmailTemplateType),
    queryFn: () => apiClient.admin.templates.getByType(type as EmailTemplateType),
    enabled: !!type,
  });
}

export function useEnableEmailTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.admin.templates.enable(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.emailTemplates.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.emailTemplates.detail(id) });
    },
  });
}

export function useDisableEmailTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.admin.templates.disable(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.emailTemplates.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.emailTemplates.detail(id) });
    },
  });
}

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '../services/queryClient';
import { apiClient } from '../services/apiClient';
import type {
  EmailTemplate,
  EmailTemplateType,
  CreateEmailTemplateRequest,
  UpdateEmailTemplateRequest,
} from '@auth-gateway/client-sdk';

interface EmailTemplateListResponse {
  templates: EmailTemplate[];
}

export function useApplicationTemplates(applicationId: string) {
  return useQuery<EmailTemplateListResponse>({
    queryKey: queryKeys.applications.templates(applicationId),
    queryFn: () => apiClient.admin.applications.listTemplates(applicationId),
    enabled: !!applicationId && applicationId !== 'new',
  });
}

export function useApplicationTemplateDetail(applicationId: string, templateId: string) {
  return useQuery<EmailTemplate>({
    queryKey: queryKeys.applications.templateDetail(applicationId, templateId),
    queryFn: () => apiClient.admin.applications.getTemplate(applicationId, templateId),
    enabled: !!applicationId && applicationId !== 'new' && !!templateId && templateId !== 'new',
  });
}

export function useCreateApplicationTemplate(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateEmailTemplateRequest) =>
      apiClient.admin.applications.createTemplate(applicationId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templates(applicationId) });
    },
  });
}

export function useUpdateApplicationTemplate(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateEmailTemplateRequest }) =>
      apiClient.admin.applications.updateTemplate(applicationId, id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templates(applicationId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templateDetail(applicationId, variables.id) });
    },
  });
}

export function useDeleteApplicationTemplate(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) =>
      apiClient.admin.applications.deleteTemplate(applicationId, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templates(applicationId) });
    },
  });
}

export function useInitializeApplicationTemplates(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () =>
      apiClient.admin.applications.initializeTemplates(applicationId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templates(applicationId) });
    },
  });
}

export function useEnableApplicationTemplate(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) =>
      apiClient.admin.applications.enableTemplate(applicationId, id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templates(applicationId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templateDetail(applicationId, id) });
    },
  });
}

export function useDisableApplicationTemplate(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) =>
      apiClient.admin.applications.disableTemplate(applicationId, id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templates(applicationId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templateDetail(applicationId, id) });
    },
  });
}

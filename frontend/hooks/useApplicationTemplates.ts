import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { queryKeys } from '../services/queryClient';
import type {
  EmailTemplate,
  EmailTemplateType,
  CreateEmailTemplateRequest,
  UpdateEmailTemplateRequest,
} from '@auth-gateway/client-sdk';

const API_BASE = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api`;

async function fetchWithAuth(url: string, options: RequestInit = {}) {
  const token = localStorage.getItem('auth_gateway_access_token');
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(error.message || error.error || 'Request failed');
  }

  return response.json();
}

interface EmailTemplateListResponse {
  templates: EmailTemplate[];
}

export function useApplicationTemplates(applicationId: string) {
  return useQuery<EmailTemplateListResponse>({
    queryKey: queryKeys.applications.templates(applicationId),
    queryFn: () => fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/email-templates`),
    enabled: !!applicationId && applicationId !== 'new',
  });
}

export function useApplicationTemplateDetail(applicationId: string, templateId: string) {
  return useQuery<EmailTemplate>({
    queryKey: queryKeys.applications.templateDetail(applicationId, templateId),
    queryFn: () => fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/email-templates/${templateId}`),
    enabled: !!applicationId && applicationId !== 'new' && !!templateId && templateId !== 'new',
  });
}

export function useCreateApplicationTemplate(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateEmailTemplateRequest) =>
      fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/email-templates`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templates(applicationId) });
    },
  });
}

export function useUpdateApplicationTemplate(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateEmailTemplateRequest }) =>
      fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/email-templates/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      }),
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
      fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/email-templates/${id}`, {
        method: 'DELETE',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templates(applicationId) });
    },
  });
}

export function useInitializeApplicationTemplates(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () =>
      fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/email-templates/initialize`, {
        method: 'POST',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templates(applicationId) });
    },
  });
}

export function useEnableApplicationTemplate(applicationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) =>
      fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/email-templates/${id}/enable`, {
        method: 'POST',
      }),
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
      fetchWithAuth(`${API_BASE}/admin/applications/${applicationId}/email-templates/${id}/disable`, {
        method: 'POST',
      }),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templates(applicationId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.templateDetail(applicationId, id) });
    },
  });
}

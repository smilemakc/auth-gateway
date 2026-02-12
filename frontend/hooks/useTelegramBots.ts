import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type {
  TelegramBot,
  CreateTelegramBotRequest,
  UpdateTelegramBotRequest,
} from '@auth-gateway/client-sdk';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useTelegramBots(appId: string) {
  return useQuery({
    queryKey: queryKeys.applications.telegramBots(appId),
    queryFn: () => apiClient.admin.telegramBots.list(appId),
    enabled: !!appId,
  });
}

export function useTelegramBotDetail(appId: string, botId: string) {
  return useQuery({
    queryKey: queryKeys.applications.telegramBot(appId, botId),
    queryFn: () => apiClient.admin.telegramBots.getById(appId, botId),
    enabled: !!appId && !!botId,
  });
}

export function useCreateTelegramBot() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, data }: { appId: string; data: CreateTelegramBotRequest }) =>
      apiClient.admin.telegramBots.create(appId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.telegramBots(variables.appId) });
    },
  });
}

export function useUpdateTelegramBot() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, id, data }: { appId: string; id: string; data: UpdateTelegramBotRequest }) =>
      apiClient.admin.telegramBots.update(appId, id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.telegramBots(variables.appId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.telegramBot(variables.appId, variables.id) });
    },
  });
}

export function useDeleteTelegramBot() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, id }: { appId: string; id: string }) =>
      apiClient.admin.telegramBots.delete(appId, id),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.telegramBots(variables.appId) });
    },
  });
}

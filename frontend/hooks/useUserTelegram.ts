import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useUserTelegramAccounts(userId: string) {
  return useQuery({
    queryKey: queryKeys.userTelegram.accounts(userId),
    queryFn: () => apiClient.admin.userTelegram.getAccounts(userId),
    enabled: !!userId,
  });
}

export function useUserTelegramBotAccess(userId: string, appId?: string) {
  return useQuery({
    queryKey: queryKeys.userTelegram.botAccess(userId, appId),
    queryFn: () => apiClient.admin.userTelegram.getBotAccess(userId, appId),
    enabled: !!userId,
  });
}

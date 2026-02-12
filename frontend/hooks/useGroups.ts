import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';
import type { CreateGroupRequest, UpdateGroupRequest, AddGroupMembersRequest } from '@auth-gateway/client-sdk';
import { useCurrentAppId } from './useAppAwareQuery';

export function useGroups(page: number = 1, pageSize: number = 20) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.groups.list(page, pageSize), appId],
    queryFn: () => apiClient.admin.groups.list(page, pageSize),
  });
}

export function useGroup(groupId: string) {
  return useQuery({
    queryKey: queryKeys.groups.detail(groupId),
    queryFn: () => apiClient.admin.groups.get(groupId),
    enabled: !!groupId,
  });
}

export function useCreateGroup() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateGroupRequest) => apiClient.admin.groups.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.groups.all });
    },
  });
}

export function useUpdateGroup() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateGroupRequest }) =>
      apiClient.admin.groups.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.groups.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.groups.detail(variables.id) });
    },
  });
}

export function useDeleteGroup() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (groupId: string) => apiClient.admin.groups.delete(groupId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.groups.all });
    },
  });
}

export function useGroupMembers(groupId: string, page: number = 1, pageSize: number = 20) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.groups.members(groupId, page, pageSize), appId],
    queryFn: () => apiClient.admin.groups.getMembers(groupId, page, pageSize),
    enabled: !!groupId,
  });
}

export function useAddGroupMembers() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: AddGroupMembersRequest }) =>
      apiClient.admin.groups.addMembers(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.groups.detail(variables.id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.groups.members(variables.id) });
    },
  });
}

export function useRemoveGroupMember() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ groupId, userId }: { groupId: string; userId: string }) =>
      apiClient.admin.groups.removeMember(groupId, userId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.groups.detail(variables.groupId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.groups.members(variables.groupId) });
    },
  });
}


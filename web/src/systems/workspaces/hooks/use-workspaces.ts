import { useMutation, useQuery, useQueryClient, type QueryKey } from "@tanstack/react-query";

import {
  listWorkspaces,
  registerWorkspace,
  renameWorkspace,
  unregisterWorkspace,
  type RegisterWorkspaceParams,
  type RenameWorkspaceParams,
} from "../adapters/workspaces-api";
import { workspaceKeys } from "../lib/query-keys";
import type { Workspace } from "../types";

export function useWorkspaces() {
  return useQuery<Workspace[]>({
    queryKey: workspaceKeys.lists() as QueryKey,
    queryFn: listWorkspaces,
  });
}

export function useRegisterWorkspace() {
  const queryClient = useQueryClient();
  return useMutation<Workspace, Error, RegisterWorkspaceParams>({
    mutationFn: params => registerWorkspace(params),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() as QueryKey });
    },
  });
}

export function useRenameWorkspace() {
  const queryClient = useQueryClient();
  return useMutation<Workspace, Error, RenameWorkspaceParams>({
    mutationFn: params => renameWorkspace(params),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() as QueryKey });
    },
  });
}

export function useUnregisterWorkspace() {
  const queryClient = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: id => unregisterWorkspace(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() as QueryKey });
    },
  });
}

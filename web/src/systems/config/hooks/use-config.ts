import { useMutation, useQuery, useQueryClient, type QueryKey } from "@tanstack/react-query";

import {
  getGlobalConfig,
  getWorkspaceConfig,
  putGlobalConfig,
  putWorkspaceConfig,
} from "../adapters/config-api";
import { configKeys } from "../lib/query-keys";
import type { ConfigDocument } from "../types";

export function useGlobalConfig() {
  return useQuery<ConfigDocument>({
    queryKey: configKeys.global() as QueryKey,
    queryFn: getGlobalConfig,
  });
}

export function useWorkspaceConfig(workspaceId: string) {
  return useQuery<ConfigDocument>({
    queryKey: configKeys.workspace(workspaceId) as QueryKey,
    queryFn: () => getWorkspaceConfig(workspaceId),
    enabled: Boolean(workspaceId),
  });
}

export function useSaveGlobalConfig() {
  const queryClient = useQueryClient();
  return useMutation<ConfigDocument, Error, ConfigDocument>({
    mutationFn: doc => putGlobalConfig(doc),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: configKeys.global() as QueryKey });
    },
  });
}

export function useSaveWorkspaceConfig(workspaceId: string) {
  const queryClient = useQueryClient();
  return useMutation<ConfigDocument, Error, ConfigDocument>({
    mutationFn: doc => putWorkspaceConfig(workspaceId, doc),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: configKeys.workspace(workspaceId) as QueryKey,
      });
    },
  });
}

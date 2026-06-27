import { useMutation, useQuery, type QueryKey } from "@tanstack/react-query";

import {
  getSetupOptions,
  runSetup,
  type RunSetupParams,
  type SetupOptions,
  type SetupResult,
} from "../adapters/setup-api";

export const setupKeys = {
  all: ["setup"] as const,
  options: (workspaceId: string) => [...setupKeys.all, "options", workspaceId] as const,
};

export function useSetupOptions(workspaceId: string) {
  return useQuery<SetupOptions>({
    queryKey: setupKeys.options(workspaceId) as QueryKey,
    queryFn: () => getSetupOptions(workspaceId),
  });
}

export function useRunSetup() {
  return useMutation<SetupResult, Error, RunSetupParams>({
    mutationFn: runSetup,
  });
}

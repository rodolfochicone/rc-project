import { useMutation, useQueryClient, type QueryKey } from "@tanstack/react-query";

import { dashboardKeys } from "@/systems/dashboard";
import { runKeys, type Run } from "@/systems/runs";

import { startExec, type StartExecParams } from "../adapters/exec-api";

export function useStartExec(workspaceId: string) {
  const queryClient = useQueryClient();
  return useMutation<Run, Error, StartExecParams>({
    mutationFn: params => startExec(params),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: runKeys.lists() as QueryKey });
      void queryClient.invalidateQueries({ queryKey: dashboardKeys.byWorkspace(workspaceId) });
    },
  });
}

import { apiErrorMessage, daemonApiClient, requireData } from "@/lib/api-client";
import type { Run } from "@/systems/runs";

import type { ExecRequestBody, ExecRuntimeOverrides } from "../types";

export interface StartExecParams {
  workspacePath: string;
  prompt: string;
  runtimeOverrides?: ExecRuntimeOverrides;
  interactive?: boolean;
}

function hasOverrides(overrides?: ExecRuntimeOverrides): overrides is ExecRuntimeOverrides {
  return Boolean(overrides) && Object.keys(overrides ?? {}).length > 0;
}

export async function startExec(params: StartExecParams): Promise<Run> {
  const body: ExecRequestBody = {
    workspace_path: params.workspacePath,
    prompt: params.prompt,
    // Detach so the daemon owns the run and we observe it through the live
    // run stream, mirroring how the workflow inventory starts task runs.
    presentation_mode: "detach",
  };
  if (hasOverrides(params.runtimeOverrides)) {
    body.runtime_overrides = params.runtimeOverrides;
  }
  // Only opt into interactivity when explicitly requested; an absent flag keeps
  // the daemon's default auto-approve/finalize behavior (ADR-004).
  if (params.interactive) {
    body.interactive = true;
  }

  const { data, error, response } = await daemonApiClient.POST("/api/exec", { body });
  if (!data) {
    throw new Error(apiErrorMessage(error, "Failed to start exec run"));
  }
  const payload = requireData(data, response, "Failed to start exec run", error);
  return payload.run;
}

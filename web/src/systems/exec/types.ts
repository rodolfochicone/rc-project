import type { components } from "@/generated/rc-openapi";

export type ExecRequestBody = components["schemas"]["ExecRequest"];

/**
 * Runtime override keys accepted by the daemon exec endpoint. Mirrors the
 * subset of `runtimeOverrideInput` that is meaningful for an ad-hoc UI run.
 */
export interface ExecRuntimeOverrides {
  agent_name?: string;
  ide?: string;
  model?: string;
  reasoning_effort?: string;
  access_mode?: string;
  [key: string]: string | undefined;
}

import type { components } from "@/generated/rc-openapi";

export type Run = components["schemas"]["Run"];
export type RunSnapshot = components["schemas"]["RunSnapshotPayload"];
export type RunJobState = components["schemas"]["RunJobState"];
export type RunJobSummary = components["schemas"]["RunJobSummary"];
export type RunTranscript = components["schemas"]["RunTranscriptPayload"];
export type RunTranscriptMessage = components["schemas"]["RunTranscriptMessage"];
export type RunUIMessage = components["schemas"]["RunUIMessage"];
export type RunUIMessagePart = components["schemas"]["RunUIMessagePart"];
export type RunShutdownState = components["schemas"]["RunShutdownState"];
export type RunUsage = components["schemas"]["Usage"];
export type TaskRunRequestBody = components["schemas"]["TaskRunRequest"];
export type RunPendingInput = components["schemas"]["RunPendingInput"];
export type RunInputOption = components["schemas"]["RunInputOption"];
export type RunInputRequest = components["schemas"]["RunInputRequest"];

export type RunListStatusFilter = "active" | "completed" | "failed" | "canceled" | "all";
export type RunListModeFilter = "task" | "review" | "exec" | "all";

export interface RunListParams {
  workspaceId: string | null;
  status?: RunListStatusFilter;
  mode?: RunListModeFilter;
  limit?: number;
}

const TERMINAL_RUN_STATUSES = new Set([
  "completed",
  "succeeded",
  "failed",
  "crashed",
  "canceled",
  "cancelled",
]);

/**
 * isTerminalRunStatus reports whether a run status string represents a finished
 * run. Shared by the run detail route (stream gating) and the detail view (input
 * panel gating) so the terminal-status definition lives in one place.
 */
export function isTerminalRunStatus(status: string | undefined): boolean {
  if (!status) {
    return false;
  }
  return TERMINAL_RUN_STATUSES.has(status.toLowerCase());
}

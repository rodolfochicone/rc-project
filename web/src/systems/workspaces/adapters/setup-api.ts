import { apiBaseUrl, type TransportErrorShape } from "@/lib/api-client";
import { CSRF_HEADER_NAME, readCsrfToken } from "@/lib/csrf";
import { ACTIVE_WORKSPACE_HEADER } from "@/systems/app-shell";

// The setup endpoints are intentionally absent from the browser OpenAPI
// contract (like the workspace mutations), so they use raw `fetch` and attach
// the CSRF + active-workspace headers by hand.

export interface SetupAgent {
  name: string;
  display_name: string;
  detected: boolean;
}

export interface SetupSkill {
  name: string;
  description?: string;
}

export interface SetupOptions {
  agents: SetupAgent[];
  skills: SetupSkill[];
  configured: boolean;
}

export interface SetupInstalledItem {
  skill: string;
  agent: string;
  path: string;
}

export interface SetupFailedItem {
  skill: string;
  agent: string;
  error: string;
}

export interface SetupResult {
  installed: SetupInstalledItem[];
  failed: SetupFailedItem[];
}

export async function getSetupOptions(workspaceId: string): Promise<SetupOptions> {
  const response = await fetch(`${apiBaseUrl}/api/setup/options`, {
    headers: { [ACTIVE_WORKSPACE_HEADER]: workspaceId },
  });
  const payload = (await response.json()) as Partial<SetupOptions> & Partial<TransportErrorShape>;
  if (!response.ok) {
    throw new Error((payload as TransportErrorShape).message ?? "Failed to load setup options");
  }
  return {
    agents: payload.agents ?? [],
    skills: payload.skills ?? [],
    configured: payload.configured ?? false,
  };
}

export interface RunSetupParams {
  workspaceId: string;
  agents: string[];
  skills: string[];
}

export async function runSetup(params: RunSetupParams): Promise<SetupResult> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    [ACTIVE_WORKSPACE_HEADER]: params.workspaceId,
  };
  const token = readCsrfToken();
  if (token) {
    headers[CSRF_HEADER_NAME] = token;
  }
  const response = await fetch(`${apiBaseUrl}/api/setup`, {
    method: "POST",
    headers,
    body: JSON.stringify({ agents: params.agents, skills: params.skills }),
  });
  const payload = (await response.json()) as Partial<SetupResult> & Partial<TransportErrorShape>;
  if (!response.ok) {
    throw new Error((payload as TransportErrorShape).message ?? "Failed to install skills");
  }
  return { installed: payload.installed ?? [], failed: payload.failed ?? [] };
}

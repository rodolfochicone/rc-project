import {
  apiBaseUrl,
  daemonApiClient,
  requireData,
  type TransportErrorShape,
} from "@/lib/api-client";
import { CSRF_HEADER_NAME, readCsrfToken } from "@/lib/csrf";

import type { Workspace } from "../types";

// These workspace mutations are intentionally absent from the browser OpenAPI
// contract, so they cannot use the typed `daemonApiClient` (whose middleware
// attaches the CSRF token automatically). The daemon still enforces the CSRF
// double-submit token on every mutation, so we attach it here by hand.
function mutationHeaders(base: Record<string, string> = {}): Record<string, string> {
  const token = readCsrfToken();
  return token ? { ...base, [CSRF_HEADER_NAME]: token } : base;
}

export async function listWorkspaces(): Promise<Workspace[]> {
  const { data, error, response } = await daemonApiClient.GET("/api/workspaces");
  return requireData(data, response, "Failed to load workspaces", error).workspaces ?? [];
}

export interface RegisterWorkspaceParams {
  name: string;
  rootDir: string;
}

export async function registerWorkspace(params: RegisterWorkspaceParams): Promise<Workspace> {
  const response = await fetch(`${apiBaseUrl}/api/workspaces`, {
    method: "POST",
    headers: mutationHeaders({ "Content-Type": "application/json" }),
    // The daemon's register contract takes `path` (not `root_dir`) plus an
    // optional display name — see WorkspaceRegisterRequest.
    body: JSON.stringify({ path: params.rootDir, name: params.name }),
  });
  const payload = (await response.json()) as {
    workspace?: Workspace;
  } & Partial<TransportErrorShape>;
  if (!response.ok) {
    throw new Error((payload as TransportErrorShape).message ?? "Failed to register workspace");
  }
  if (!payload.workspace) {
    throw new Error("Failed to register workspace: empty response");
  }
  return payload.workspace;
}

export interface RenameWorkspaceParams {
  id: string;
  name: string;
}

export async function renameWorkspace(params: RenameWorkspaceParams): Promise<Workspace> {
  const response = await fetch(`${apiBaseUrl}/api/workspaces/${encodeURIComponent(params.id)}`, {
    method: "PATCH",
    headers: mutationHeaders({ "Content-Type": "application/json" }),
    body: JSON.stringify({ name: params.name }),
  });
  const payload = (await response.json()) as {
    workspace?: Workspace;
  } & Partial<TransportErrorShape>;
  if (!response.ok) {
    throw new Error((payload as TransportErrorShape).message ?? "Failed to rename workspace");
  }
  if (!payload.workspace) {
    throw new Error("Failed to rename workspace: empty response");
  }
  return payload.workspace;
}

export async function unregisterWorkspace(id: string): Promise<void> {
  const response = await fetch(`${apiBaseUrl}/api/workspaces/${encodeURIComponent(id)}`, {
    method: "DELETE",
    headers: mutationHeaders(),
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => ({}))) as Partial<TransportErrorShape>;
    throw new Error(payload.message ?? "Failed to unregister workspace");
  }
}

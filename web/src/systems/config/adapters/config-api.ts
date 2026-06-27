import { daemonApiClient, requireData } from "@/lib/api-client";
import { ACTIVE_WORKSPACE_HEADER } from "@/systems/app-shell";

import type { ConfigDocument } from "../types";

export async function getGlobalConfig(): Promise<ConfigDocument> {
  const { data, error, response } = await daemonApiClient.GET("/api/config/global");
  return requireData(data, response, "Failed to load global config", error).config;
}

export async function putGlobalConfig(doc: ConfigDocument): Promise<ConfigDocument> {
  const { data, error, response } = await daemonApiClient.PUT("/api/config/global", {
    body: doc,
  });
  return requireData(data, response, "Failed to save global config", error).config;
}

export async function getWorkspaceConfig(workspaceId: string): Promise<ConfigDocument> {
  const { data, error, response } = await daemonApiClient.GET("/api/config/workspace", {
    params: { header: { [ACTIVE_WORKSPACE_HEADER]: workspaceId } },
  });
  return requireData(data, response, "Failed to load workspace config", error).config;
}

export async function putWorkspaceConfig(
  workspaceId: string,
  doc: ConfigDocument
): Promise<ConfigDocument> {
  const { data, error, response } = await daemonApiClient.PUT("/api/config/workspace", {
    params: { header: { [ACTIVE_WORKSPACE_HEADER]: workspaceId } },
    body: doc,
  });
  return requireData(data, response, "Failed to save workspace config", error).config;
}

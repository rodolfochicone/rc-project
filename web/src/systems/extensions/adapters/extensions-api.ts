import { daemonApiClient, requireData } from "@/lib/api-client";
import { ACTIVE_WORKSPACE_HEADER } from "@/systems/app-shell";

import type { AgentItem, ExtensionItem } from "../types";

export async function listCatalogAgents(workspaceId: string): Promise<AgentItem[]> {
  const { data, error, response } = await daemonApiClient.GET("/api/catalog/agents", {
    params: { header: { [ACTIVE_WORKSPACE_HEADER]: workspaceId } },
  });
  return requireData(data, response, "Failed to load agents", error).agents ?? [];
}

export async function listCatalogExtensions(workspaceId: string): Promise<ExtensionItem[]> {
  const { data, error, response } = await daemonApiClient.GET("/api/catalog/extensions", {
    params: { header: { [ACTIVE_WORKSPACE_HEADER]: workspaceId } },
  });
  return requireData(data, response, "Failed to load extensions", error).extensions ?? [];
}

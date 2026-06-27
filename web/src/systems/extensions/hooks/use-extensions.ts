import { useQuery, type QueryKey } from "@tanstack/react-query";

import { listCatalogAgents, listCatalogExtensions } from "../adapters/extensions-api";
import { extensionKeys } from "../lib/query-keys";
import type { AgentItem, ExtensionItem } from "../types";

export function useCatalogAgents(workspaceId: string) {
  return useQuery<AgentItem[]>({
    queryKey: extensionKeys.agents(workspaceId) as QueryKey,
    queryFn: () => listCatalogAgents(workspaceId),
    enabled: Boolean(workspaceId),
  });
}

export function useCatalogExtensions(workspaceId: string) {
  return useQuery<ExtensionItem[]>({
    queryKey: extensionKeys.extensions(workspaceId) as QueryKey,
    queryFn: () => listCatalogExtensions(workspaceId),
    enabled: Boolean(workspaceId),
  });
}

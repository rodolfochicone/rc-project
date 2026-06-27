export const extensionKeys = {
  all: ["extensions"] as const,
  agents: (workspaceId: string) => [...extensionKeys.all, "agents", workspaceId] as const,
  extensions: (workspaceId: string) => [...extensionKeys.all, "extensions", workspaceId] as const,
};

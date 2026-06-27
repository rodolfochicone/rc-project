export const configKeys = {
  all: ["config"] as const,
  global: () => [...configKeys.all, "global"] as const,
  workspace: (workspaceId: string) => [...configKeys.all, "workspace", workspaceId] as const,
};

export { WorkspacesView } from "./components/workspaces-view";
export { SetupPromptWatcher } from "./components/setup-prompt-watcher";
export {
  useWorkspaces,
  useRegisterWorkspace,
  useRenameWorkspace,
  useUnregisterWorkspace,
} from "./hooks/use-workspaces";
export {
  listWorkspaces,
  registerWorkspace,
  renameWorkspace,
  unregisterWorkspace,
  type RegisterWorkspaceParams,
  type RenameWorkspaceParams,
} from "./adapters/workspaces-api";
export { workspaceKeys } from "./lib/query-keys";
export type { Workspace, WorkspaceSyncResult } from "./types";

import { type ReactElement } from "react";

import { createFileRoute } from "@tanstack/react-router";

import { AppShellLayout, useActiveWorkspaceContext } from "@/systems/app-shell";
import { WorkspacesView } from "@/systems/workspaces";

export const Route = createFileRoute("/_app/workspaces")({
  component: WorkspacesRoute,
});

function WorkspacesRoute(): ReactElement {
  const { activeWorkspace, workspaces, onSwitchWorkspace } = useActiveWorkspaceContext();

  return (
    <AppShellLayout
      activeWorkspace={activeWorkspace}
      onSwitchWorkspace={onSwitchWorkspace}
      workspaces={workspaces}
    >
      <WorkspacesView activeWorkspace={activeWorkspace} />
    </AppShellLayout>
  );
}

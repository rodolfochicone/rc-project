import { type ReactElement } from "react";

import { createFileRoute } from "@tanstack/react-router";

import { AppShellLayout, useActiveWorkspaceContext } from "@/systems/app-shell";
import { ExecView } from "@/systems/exec";

export const Route = createFileRoute("/_app/exec")({
  component: ExecRoute,
});

function ExecRoute(): ReactElement {
  const { activeWorkspace, workspaces, onSwitchWorkspace } = useActiveWorkspaceContext();

  return (
    <AppShellLayout
      activeWorkspace={activeWorkspace}
      onSwitchWorkspace={onSwitchWorkspace}
      workspaces={workspaces}
    >
      <ExecView activeWorkspace={activeWorkspace} />
    </AppShellLayout>
  );
}

import { type ReactElement } from "react";

import { createFileRoute } from "@tanstack/react-router";

import { AppShellLayout, useActiveWorkspaceContext } from "@/systems/app-shell";
import { ConfigView } from "@/systems/config";

export const Route = createFileRoute("/_app/config")({
  component: ConfigRoute,
});

function ConfigRoute(): ReactElement {
  const { activeWorkspace, workspaces, onSwitchWorkspace } = useActiveWorkspaceContext();

  return (
    <AppShellLayout
      activeWorkspace={activeWorkspace}
      onSwitchWorkspace={onSwitchWorkspace}
      workspaces={workspaces}
    >
      <ConfigView workspaceId={activeWorkspace.id} />
    </AppShellLayout>
  );
}

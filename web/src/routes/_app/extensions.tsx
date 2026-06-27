import { type ReactElement } from "react";

import { createFileRoute } from "@tanstack/react-router";

import { AppShellLayout, useActiveWorkspaceContext } from "@/systems/app-shell";
import { ExtensionsView } from "@/systems/extensions";

export const Route = createFileRoute("/_app/extensions")({
  component: ExtensionsRoute,
});

function ExtensionsRoute(): ReactElement {
  const { activeWorkspace, workspaces, onSwitchWorkspace } = useActiveWorkspaceContext();

  return (
    <AppShellLayout
      activeWorkspace={activeWorkspace}
      onSwitchWorkspace={onSwitchWorkspace}
      workspaces={workspaces}
    >
      <ExtensionsView workspaceId={activeWorkspace.id} />
    </AppShellLayout>
  );
}

import { useState, type FormEvent, type ReactElement } from "react";

import {
  Alert,
  Button,
  EmptyState,
  SectionHeading,
  SkeletonRow,
  StatusBadge,
  SurfaceCard,
  SurfaceCardBody,
  SurfaceCardDescription,
  SurfaceCardEyebrow,
  SurfaceCardHeader,
  SurfaceCardTitle,
} from "@rodolfochicone/ui";
import { Boxes, FolderOpen, Plus, Trash2 } from "lucide-react";

import { apiErrorMessage } from "@/lib/api-client";
import { canPickDirectory, getDesktopBridge } from "@/lib/desktop-bridge";
import type { Workspace as ActiveWorkspace } from "@/systems/app-shell";

import {
  useRegisterWorkspace,
  useUnregisterWorkspace,
  useWorkspaces,
} from "../hooks/use-workspaces";
import type { Workspace } from "../types";
import { SetupSkillsCard } from "./setup-skills-card";

const fieldClass =
  "w-full rounded-[var(--radius-md)] border border-border bg-[color:var(--surface-inset)] px-3 py-2 text-sm text-foreground transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/60";
const labelClass = "block text-sm font-medium text-foreground";

function deriveName(path: string): string {
  const segments = path.split("/").filter(Boolean);
  return segments.at(-1) ?? path;
}

export interface WorkspacesViewProps {
  activeWorkspace: ActiveWorkspace;
}

export function WorkspacesView({ activeWorkspace }: WorkspacesViewProps): ReactElement {
  const { data: workspaces, isLoading, isError, error } = useWorkspaces();
  const register = useRegisterWorkspace();
  const unregister = useUnregisterWorkspace();

  const [path, setPath] = useState("");
  const [name, setName] = useState("");

  const trimmedPath = path.trim();
  const canRegister = trimmedPath.length > 0 && !register.isPending;
  const canBrowse = canPickDirectory();

  async function handleBrowse(): Promise<void> {
    const selected = await getDesktopBridge()?.selectDirectory?.();
    if (selected) {
      setPath(selected);
    }
  }

  function handleRegister(event: FormEvent<HTMLFormElement>): void {
    event.preventDefault();
    if (!canRegister) return;
    register.mutate(
      { rootDir: trimmedPath, name: name.trim() || deriveName(trimmedPath) },
      {
        onSuccess: () => {
          setPath("");
          setName("");
        },
      }
    );
  }

  return (
    <div className="space-y-6" data-testid="workspaces-view">
      <SectionHeading
        description="Register the directories the daemon can act on, then pick one per browser tab."
        eyebrow="Workspaces"
        title="Workspaces"
      />

      {register.isError ? (
        <Alert data-testid="workspaces-register-error" variant="error">
          {apiErrorMessage(register.error, "Failed to register workspace")}
        </Alert>
      ) : null}

      <SurfaceCard>
        <SurfaceCardBody>
          <form
            className="space-y-4"
            data-testid="workspace-register-form"
            onSubmit={handleRegister}
          >
            <div className="space-y-1.5">
              <label className={labelClass} htmlFor="workspace-path">
                Workspace path
              </label>
              <div className="flex gap-2">
                <input
                  className={`${fieldClass} font-mono`}
                  data-testid="workspace-register-path"
                  id="workspace-path"
                  onChange={event => setPath(event.target.value)}
                  placeholder="/absolute/path/to/project"
                  value={path}
                />
                {canBrowse ? (
                  <Button
                    data-testid="workspace-browse"
                    icon={<FolderOpen className="size-4" />}
                    onClick={() => {
                      void handleBrowse();
                    }}
                    type="button"
                    variant="secondary"
                  >
                    Browse
                  </Button>
                ) : null}
              </div>
            </div>
            <div className="space-y-1.5">
              <label className={labelClass} htmlFor="workspace-name">
                Display name <span className="text-muted-foreground">(optional)</span>
              </label>
              <input
                className={fieldClass}
                data-testid="workspace-register-name"
                id="workspace-name"
                onChange={event => setName(event.target.value)}
                placeholder="Defaults to the folder name"
                value={name}
              />
            </div>
            <Button
              data-testid="workspace-register-submit"
              disabled={!canRegister}
              icon={<Plus className="size-4" />}
              loading={register.isPending}
              type="submit"
            >
              Register workspace
            </Button>
          </form>
        </SurfaceCardBody>
      </SurfaceCard>

      <SetupSkillsCard workspace={activeWorkspace} />

      {isError ? (
        <Alert data-testid="workspaces-error" variant="error">
          {apiErrorMessage(error, "Failed to load workspaces")}
        </Alert>
      ) : null}

      {isLoading ? (
        <div className="space-y-2" data-testid="workspaces-loading">
          <SkeletonRow />
          <SkeletonRow />
        </div>
      ) : null}

      {!isLoading && (workspaces ?? []).length === 0 ? (
        <EmptyState
          data-testid="workspaces-empty"
          description="Register a workspace above to let the daemon discover its workflows."
          icon={<Boxes className="size-4" aria-hidden />}
          title="No workspaces registered"
        />
      ) : null}

      {(workspaces ?? []).length > 0 ? (
        <ul className="grid gap-3" data-testid="workspaces-list">
          {(workspaces ?? []).map((ws: Workspace) => (
            <li key={ws.id}>
              <SurfaceCard data-testid={`workspace-item-${ws.id}`}>
                <SurfaceCardHeader>
                  <div className="min-w-0">
                    <SurfaceCardEyebrow>Workspace</SurfaceCardEyebrow>
                    <SurfaceCardTitle data-testid={`workspace-name-${ws.id}`}>
                      {ws.name}
                    </SurfaceCardTitle>
                    <SurfaceCardDescription
                      className="truncate font-mono"
                      data-testid={`workspace-root-${ws.id}`}
                      title={ws.root_dir}
                    >
                      {ws.root_dir}
                    </SurfaceCardDescription>
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    {ws.filesystem_state === "missing" ? (
                      <StatusBadge tone="warning">path missing</StatusBadge>
                    ) : null}
                    {ws.read_only ? <StatusBadge tone="neutral">read-only</StatusBadge> : null}
                  </div>
                </SurfaceCardHeader>
                <SurfaceCardBody>
                  <Button
                    data-testid={`workspace-unregister-${ws.id}`}
                    disabled={unregister.isPending}
                    icon={<Trash2 className="size-4" />}
                    loading={unregister.isPending}
                    onClick={() => {
                      unregister.mutate(ws.id);
                    }}
                    size="sm"
                    variant="ghost"
                  >
                    Remove
                  </Button>
                </SurfaceCardBody>
              </SurfaceCard>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}

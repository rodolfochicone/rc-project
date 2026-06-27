import { useEffect, useRef, type ReactElement } from "react";

import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";

import { useActiveWorkspaceContext } from "@/systems/app-shell";

import { useSetupOptions } from "../hooks/use-setup";

/**
 * SetupPromptWatcher prompts the user to install the rc skills whenever the
 * active workspace points at a project that has not been set up yet. It renders
 * nothing; it only reacts to the active workspace's setup status. Each project
 * is prompted at most once per session so switching back and forth is quiet.
 */
export function SetupPromptWatcher(): ReactElement | null {
  const { activeWorkspace } = useActiveWorkspaceContext();
  const navigate = useNavigate();
  const optionsQuery = useSetupOptions(activeWorkspace.id);
  const promptedRef = useRef<Set<string>>(new Set());

  const configured = optionsQuery.data?.configured;

  useEffect(() => {
    if (configured !== false) return;
    if (promptedRef.current.has(activeWorkspace.id)) return;
    promptedRef.current.add(activeWorkspace.id);

    toast.warning(`${activeWorkspace.name} isn't set up`, {
      description: "Install the rc skills so agents can run them in this project.",
      action: {
        label: "Configure",
        onClick: () => {
          void navigate({ to: "/workspaces" });
        },
      },
    });
  }, [activeWorkspace.id, activeWorkspace.name, configured, navigate]);

  return null;
}

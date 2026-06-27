import { useEffect, useState, type ReactElement } from "react";

import {
  Alert,
  Button,
  SkeletonRow,
  SurfaceCard,
  SurfaceCardBody,
  SurfaceCardEyebrow,
  SurfaceCardHeader,
  SurfaceCardTitle,
} from "@rodolfochicone/ui";
import { Download } from "lucide-react";

import { apiErrorMessage } from "@/lib/api-client";
import type { Workspace } from "@/systems/app-shell";

import { useRunSetup, useSetupOptions } from "../hooks/use-setup";

const labelClass = "block text-sm font-medium text-foreground";
const checkboxClass = "mt-0.5 size-4 rounded border-border accent-[color:var(--primary)]";

function toggle(list: string[], name: string): string[] {
  return list.includes(name) ? list.filter(item => item !== name) : [...list, name];
}

export interface SetupSkillsCardProps {
  workspace: Workspace;
}

export function SetupSkillsCard({ workspace }: SetupSkillsCardProps): ReactElement {
  const optionsQuery = useSetupOptions(workspace.id);
  const runSetup = useRunSetup();

  // Null until options load; then seeded with detected agents and all skills.
  const [selectedAgents, setSelectedAgents] = useState<string[] | null>(null);
  const [selectedSkills, setSelectedSkills] = useState<string[] | null>(null);

  const options = optionsQuery.data;
  useEffect(() => {
    if (!options) return;
    setSelectedAgents(
      prev => prev ?? options.agents.filter(agent => agent.detected).map(a => a.name)
    );
    setSelectedSkills(prev => prev ?? options.skills.map(skill => skill.name));
  }, [options]);

  const agents = selectedAgents ?? [];
  const skills = selectedSkills ?? [];
  const isReadOnly = workspace.read_only;
  const canInstall = agents.length > 0 && !isReadOnly && !runSetup.isPending;

  function handleInstall(): void {
    if (!canInstall) return;
    runSetup.mutate({ workspaceId: workspace.id, agents, skills });
  }

  return (
    <SurfaceCard data-testid="setup-skills-card">
      <SurfaceCardHeader>
        <div className="min-w-0">
          <SurfaceCardEyebrow>Setup</SurfaceCardEyebrow>
          <SurfaceCardTitle>Install rc skills</SurfaceCardTitle>
        </div>
      </SurfaceCardHeader>
      <SurfaceCardBody>
        <p className="mb-4 text-sm text-muted-foreground">
          Install the bundled rc skills into <span className="font-medium">{workspace.name}</span>{" "}
          so the selected agents can run them in this project.
        </p>

        {isReadOnly ? (
          <Alert data-testid="setup-readonly" variant="warning">
            This workspace is read-only. Skill installation is disabled.
          </Alert>
        ) : null}

        {optionsQuery.isError ? (
          <Alert data-testid="setup-options-error" variant="error">
            {apiErrorMessage(optionsQuery.error, "Failed to load setup options")}
          </Alert>
        ) : null}

        {optionsQuery.isLoading ? (
          <div className="space-y-2" data-testid="setup-options-loading">
            <SkeletonRow />
            <SkeletonRow />
          </div>
        ) : null}

        {options && !options.configured && !isReadOnly ? (
          <Alert data-testid="setup-not-configured" variant="warning">
            This project isn&apos;t set up yet. Pick the agents and skills below and install them.
          </Alert>
        ) : null}

        {options ? (
          <div className="space-y-5">
            <fieldset className="space-y-2" data-testid="setup-agents">
              <legend className={labelClass}>Agents</legend>
              {options.agents.length === 0 ? (
                <p className="text-xs text-muted-foreground">No supported agents found.</p>
              ) : (
                <div className="grid gap-2 sm:grid-cols-2">
                  {options.agents.map(agent => (
                    <label className="flex items-start gap-2" key={agent.name}>
                      <input
                        checked={agents.includes(agent.name)}
                        className={checkboxClass}
                        data-testid={`setup-agent-${agent.name}`}
                        disabled={isReadOnly}
                        onChange={() => setSelectedAgents(toggle(agents, agent.name))}
                        type="checkbox"
                      />
                      <span className="text-sm text-foreground">
                        {agent.display_name}
                        {agent.detected ? (
                          <span className="ml-1 text-xs text-muted-foreground">(detected)</span>
                        ) : null}
                      </span>
                    </label>
                  ))}
                </div>
              )}
            </fieldset>

            <fieldset className="space-y-2" data-testid="setup-skills">
              <legend className={labelClass}>Skills</legend>
              <div className="grid gap-2 sm:grid-cols-2">
                {options.skills.map(skill => (
                  <label className="flex items-start gap-2" key={skill.name}>
                    <input
                      checked={skills.includes(skill.name)}
                      className={checkboxClass}
                      data-testid={`setup-skill-${skill.name}`}
                      disabled={isReadOnly}
                      onChange={() => setSelectedSkills(toggle(skills, skill.name))}
                      type="checkbox"
                    />
                    <span className="font-mono text-xs text-foreground" title={skill.description}>
                      {skill.name}
                    </span>
                  </label>
                ))}
              </div>
            </fieldset>

            <Button
              data-testid="setup-install"
              disabled={!canInstall}
              icon={<Download className="size-4" />}
              loading={runSetup.isPending}
              onClick={handleInstall}
            >
              Install {skills.length} skill{skills.length === 1 ? "" : "s"}
            </Button>

            {runSetup.isError ? (
              <Alert data-testid="setup-install-error" variant="error">
                {apiErrorMessage(runSetup.error, "Failed to install skills")}
              </Alert>
            ) : null}

            {runSetup.data ? (
              <Alert
                data-testid="setup-install-result"
                variant={runSetup.data.failed.length > 0 ? "warning" : "success"}
              >
                Installed {runSetup.data.installed.length} skill
                {runSetup.data.installed.length === 1 ? "" : "s"}
                {runSetup.data.failed.length > 0 ? `, ${runSetup.data.failed.length} failed.` : "."}
              </Alert>
            ) : null}
          </div>
        ) : null}
      </SurfaceCardBody>
    </SurfaceCard>
  );
}

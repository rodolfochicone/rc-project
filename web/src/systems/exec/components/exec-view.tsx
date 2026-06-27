import { useId, useState, type FormEvent, type ReactElement } from "react";

import { Alert, Button, SectionHeading, SurfaceCard, SurfaceCardBody } from "@rodolfochicone/ui";
import { useNavigate } from "@tanstack/react-router";
import { Play } from "lucide-react";

import { apiErrorMessage } from "@/lib/api-client";
import {
  ACCESS_MODES,
  RC_SKILLS,
  modelsForRuntime,
  REASONING_EFFORTS,
  RUNTIMES,
} from "@/lib/runtime-catalog";
import type { Workspace } from "@/systems/app-shell";
import { useCatalogAgents } from "@/systems/extensions";

import { useStartExec } from "../hooks/use-exec";
import type { ExecRuntimeOverrides } from "../types";

const fieldClass =
  "w-full rounded-[var(--radius-md)] border border-border bg-[color:var(--surface-inset)] px-3 py-2 text-sm text-foreground transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/60";
const labelClass = "block text-sm font-medium text-foreground";

export interface ExecViewProps {
  activeWorkspace: Workspace;
}

export function ExecView({ activeWorkspace }: ExecViewProps): ReactElement {
  const navigate = useNavigate();
  const startExec = useStartExec(activeWorkspace.id);
  const agentsQuery = useCatalogAgents(activeWorkspace.id);
  const modelListId = useId();

  const [prompt, setPrompt] = useState("");
  const [skills, setSkills] = useState<string[]>([]);
  const [agentName, setAgentName] = useState("");
  const [ide, setIde] = useState("");
  const [model, setModel] = useState("");
  const [reasoningEffort, setReasoningEffort] = useState("");
  const [accessMode, setAccessMode] = useState("");
  const [interactive, setInteractive] = useState(false);

  const isReadOnly = activeWorkspace.read_only;
  const trimmedPrompt = prompt.trim();
  // A run needs either a written prompt or at least one selected skill.
  const canSubmit =
    (trimmedPrompt.length > 0 || skills.length > 0) && !isReadOnly && !startExec.isPending;

  function toggleSkill(name: string): void {
    setSkills(current =>
      current.includes(name) ? current.filter(skill => skill !== name) : [...current, name]
    );
  }

  function buildPrompt(): string {
    const prefix = skills.map(skill => `/${skill}`).join(" ");
    if (!prefix) return trimmedPrompt;
    return trimmedPrompt ? `${prefix}\n\n${trimmedPrompt}` : prefix;
  }

  function buildOverrides(): ExecRuntimeOverrides {
    const overrides: ExecRuntimeOverrides = {};
    if (agentName) overrides.agent_name = agentName;
    const trimmedIde = ide.trim();
    if (trimmedIde) overrides.ide = trimmedIde;
    const trimmedModel = model.trim();
    if (trimmedModel) overrides.model = trimmedModel;
    if (reasoningEffort) overrides.reasoning_effort = reasoningEffort;
    if (accessMode) overrides.access_mode = accessMode;
    return overrides;
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>): void {
    event.preventDefault();
    if (!canSubmit) return;
    startExec.mutate(
      {
        workspacePath: activeWorkspace.root_dir,
        prompt: buildPrompt(),
        runtimeOverrides: buildOverrides(),
        interactive,
      },
      {
        onSuccess: run => {
          void navigate({ to: "/runs/$runId", params: { runId: run.run_id } });
        },
      }
    );
  }

  const agents = agentsQuery.data ?? [];
  const modelSuggestions = modelsForRuntime(ide);

  return (
    <div className="space-y-6" data-testid="exec-view">
      <SectionHeading
        description={`Run an ad-hoc prompt or reusable agent in ${activeWorkspace.name} and follow it live.`}
        eyebrow="Run"
        title="Run a skill"
      />

      {isReadOnly ? (
        <Alert data-testid="exec-readonly" variant="warning">
          This workspace is read-only. Exec runs are disabled.
        </Alert>
      ) : null}

      {startExec.isError ? (
        <Alert data-testid="exec-error" variant="error">
          {apiErrorMessage(startExec.error, "Failed to start exec run")}
        </Alert>
      ) : null}

      <SurfaceCard>
        <SurfaceCardBody>
          <form className="space-y-5" data-testid="exec-form" onSubmit={handleSubmit}>
            <fieldset className="space-y-2" data-testid="exec-skills">
              <legend className={labelClass}>
                Skills <span className="text-muted-foreground">(optional)</span>
              </legend>
              <p className="text-xs text-muted-foreground">
                Select bundled rc skills to prepend to the prompt, or invoke them inline (e.g.
                /code-review).
              </p>
              <div className="flex flex-wrap gap-2">
                {RC_SKILLS.map(skill => {
                  const selected = skills.includes(skill.name);
                  return (
                    <button
                      aria-pressed={selected}
                      className={`rounded-[var(--radius-md)] border px-3 py-1.5 text-left text-xs transition-colors ${
                        selected
                          ? "border-[color:var(--primary)] bg-[color:var(--primary)]/10 text-foreground"
                          : "border-border bg-[color:var(--surface-inset)] text-muted-foreground hover:text-foreground"
                      }`}
                      data-testid={`exec-skill-${skill.name}`}
                      disabled={isReadOnly}
                      key={skill.name}
                      onClick={() => toggleSkill(skill.name)}
                      title={skill.description}
                      type="button"
                    >
                      <span className="font-mono">/{skill.name}</span>
                    </button>
                  );
                })}
              </div>
            </fieldset>

            <div className="space-y-1.5">
              <label className={labelClass} htmlFor="exec-prompt">
                Prompt
              </label>
              <textarea
                className={`${fieldClass} min-h-[8rem] resize-y font-mono`}
                data-testid="exec-prompt"
                disabled={isReadOnly}
                id="exec-prompt"
                onChange={event => setPrompt(event.target.value)}
                placeholder="Describe the task, or invoke a skill (e.g. /code-review)…"
                value={prompt}
              />
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-1.5">
                <label className={labelClass} htmlFor="exec-agent">
                  Agent
                </label>
                <select
                  className={fieldClass}
                  data-testid="exec-agent"
                  disabled={isReadOnly}
                  id="exec-agent"
                  onChange={event => setAgentName(event.target.value)}
                  value={agentName}
                >
                  <option value="">No agent (default)</option>
                  {agents.map(agent => (
                    <option key={agent.name} value={agent.name}>
                      {agent.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="space-y-1.5">
                <label className={labelClass} htmlFor="exec-ide">
                  Runtime
                </label>
                <select
                  className={fieldClass}
                  data-testid="exec-ide"
                  disabled={isReadOnly}
                  id="exec-ide"
                  onChange={event => setIde(event.target.value)}
                  value={ide}
                >
                  <option value="">Runtime default</option>
                  {RUNTIMES.map(runtime => (
                    <option key={runtime.id} value={runtime.id}>
                      {runtime.label}
                    </option>
                  ))}
                </select>
              </div>

              <div className="space-y-1.5">
                <label className={labelClass} htmlFor="exec-model">
                  Model
                </label>
                <input
                  className={fieldClass}
                  data-testid="exec-model"
                  disabled={isReadOnly}
                  id="exec-model"
                  list={modelListId}
                  onChange={event => setModel(event.target.value)}
                  placeholder="Runtime default"
                  value={model}
                />
                <datalist id={modelListId}>
                  {modelSuggestions.map(suggestion => (
                    <option key={suggestion} value={suggestion} />
                  ))}
                </datalist>
              </div>

              <div className="space-y-1.5">
                <label className={labelClass} htmlFor="exec-reasoning">
                  Reasoning effort
                </label>
                <select
                  className={fieldClass}
                  data-testid="exec-reasoning"
                  disabled={isReadOnly}
                  id="exec-reasoning"
                  onChange={event => setReasoningEffort(event.target.value)}
                  value={reasoningEffort}
                >
                  <option value="">Runtime default</option>
                  {REASONING_EFFORTS.map(effort => (
                    <option key={effort} value={effort}>
                      {effort}
                    </option>
                  ))}
                </select>
              </div>

              <div className="space-y-1.5">
                <label className={labelClass} htmlFor="exec-access">
                  Access mode
                </label>
                <select
                  className={fieldClass}
                  data-testid="exec-access"
                  disabled={isReadOnly}
                  id="exec-access"
                  onChange={event => setAccessMode(event.target.value)}
                  value={accessMode}
                >
                  <option value="">Runtime default</option>
                  {ACCESS_MODES.map(mode => (
                    <option key={mode} value={mode}>
                      {mode}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="flex items-center gap-2 text-sm font-medium text-foreground">
                <input
                  checked={interactive}
                  className="size-4 rounded border-border"
                  data-testid="exec-interactive"
                  disabled={isReadOnly}
                  onChange={event => setInteractive(event.target.checked)}
                  type="checkbox"
                />
                Interactive
              </label>
              <p className="text-xs text-muted-foreground">
                Pause for permission requests and skill questions so you can answer them in the run
                detail. The run stays open until you cancel it.
              </p>
            </div>

            <div className="flex items-center gap-3">
              <Button
                data-testid="exec-submit"
                disabled={!canSubmit}
                icon={<Play className="size-4" />}
                loading={startExec.isPending}
                type="submit"
              >
                Run
              </Button>
              <span className="text-xs text-muted-foreground" data-testid="exec-workspace-path">
                {activeWorkspace.root_dir}
              </span>
            </div>
          </form>
        </SurfaceCardBody>
      </SurfaceCard>
    </div>
  );
}

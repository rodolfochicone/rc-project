/**
 * Shared runtime catalog for exec and config forms.
 *
 * The daemon validates the runtime (IDE) against a fixed set of identifiers
 * (see internal/core/model/constants.go), so runtimes are a strict dropdown.
 * Per-runtime model lists are curated suggestions — the daemon accepts any
 * model string, so the Model field stays a combobox (pick or type).
 */

export interface RuntimeOption {
  /** Identifier sent to the daemon as `ide`. */
  id: string;
  /** Human label shown in the dropdown. */
  label: string;
  /** Default model the daemon falls back to for this runtime. */
  defaultModel: string;
}

/** Runtime identifiers mirror internal/core/model/constants.go (IDE* constants). */
export const RUNTIMES: readonly RuntimeOption[] = [
  { id: "claude", label: "Claude", defaultModel: "opus" },
  { id: "codex", label: "Codex", defaultModel: "gpt-5.5" },
  { id: "droid", label: "Droid", defaultModel: "gpt-5.5" },
  { id: "cursor-agent", label: "Cursor", defaultModel: "composer-1" },
  { id: "opencode", label: "OpenCode", defaultModel: "anthropic/claude-opus-4-6" },
  { id: "pi", label: "Pi", defaultModel: "anthropic/claude-opus-4-6" },
  { id: "gemini", label: "Gemini", defaultModel: "gemini-2.5-pro" },
  { id: "copilot", label: "Copilot CLI", defaultModel: "claude-sonnet-4.6" },
] as const;

/**
 * Curated model suggestions per runtime. These power combobox hints only; the
 * daemon accepts any model the runtime supports, so the list is not exhaustive.
 */
export const MODELS_BY_RUNTIME: Record<string, readonly string[]> = {
  claude: ["opus", "sonnet", "haiku"],
  codex: ["gpt-5.5", "gpt-5.4", "gpt-5.4-mini", "gpt-5.3-codex", "gpt-5-codex"],
  droid: ["gpt-5.5", "gpt-5-codex", "claude-sonnet-4-5", "claude-opus-4-1", "glm-4.6"],
  "cursor-agent": ["composer-1", "composer-2.5", "auto", "opus-4.8", "gpt-5.5", "gemini-3-pro"],
  opencode: ["anthropic/claude-opus-4-6", "anthropic/claude-sonnet-4-6", "openai/gpt-5.5"],
  pi: ["anthropic/claude-opus-4-6", "anthropic/claude-sonnet-4-6"],
  gemini: ["gemini-2.5-pro", "gemini-3-pro", "gemini-3-flash", "gemini-2.5-flash"],
  copilot: ["claude-sonnet-4.6", "gpt-5.5", "gemini-2.5-pro"],
};

/**
 * Model suggestions for the selected runtime. With no runtime selected, returns
 * the deduplicated union so the combobox still offers sensible hints.
 */
export function modelsForRuntime(ide: string): readonly string[] {
  const trimmed = ide.trim();
  if (trimmed && MODELS_BY_RUNTIME[trimmed]) {
    return MODELS_BY_RUNTIME[trimmed];
  }
  const seen = new Set<string>();
  for (const models of Object.values(MODELS_BY_RUNTIME)) {
    for (const model of models) seen.add(model);
  }
  return [...seen];
}

export const REASONING_EFFORTS = ["low", "medium", "high", "xhigh"] as const;
export const ACCESS_MODES = ["default", "full"] as const;

export interface SkillOption {
  /** Skill slug invoked as `/<name>` in the prompt. */
  name: string;
  /** Short description shown alongside the toggle. */
  description: string;
}

/** Bundled rc workflow skills (see skills/ in the repo). */
export const RC_SKILLS: readonly SkillOption[] = [
  { name: "rc-create-prd", description: "Brainstorm and write a Product Requirements Document." },
  { name: "rc-create-techspec", description: "Translate a PRD into a Technical Specification." },
  { name: "rc-create-tasks", description: "Break a PRD/TechSpec into executable task files." },
  { name: "rc-execute-task", description: "Execute one PRD task end-to-end." },
  { name: "rc-review-round", description: "Run a comprehensive code review round." },
  { name: "rc-fix-reviews", description: "Remediate batched PR review issues." },
  { name: "rc-final-verify", description: "Enforce fresh verification before completion." },
  { name: "rc-workflow-memory", description: "Maintain workflow-scoped task memory." },
  { name: "rc-git", description: "Branch, push, and open a PR with step-by-step confirmations." },
  {
    name: "rc-jira",
    description:
      "Create, read, comment on, and transition Jira issues via the official Atlassian MCP.",
  },
  { name: "rc", description: "Explain rc capabilities, commands, and configuration." },
] as const;

<div align="center">
  <h1>RC</h1>
  <p><strong>Orchestrate AI coding agents from idea to shipped code — one structured pipeline, shipped as a plugin.</strong></p>
  <p>
    <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
    <a href="https://github.com/rodolfochicone/rc-project/releases"><img src="https://img.shields.io/github/v/release/rodolfochicone/rc-project?include_prereleases" alt="Release"></a>
  </p>
</div>

RC is an **agent plugin** — skills, commands, agents, and hooks — that drives the full lifecycle of
AI-assisted development: optional ideation → PRD → TechSpec → tasks → execution → review →
remediation. It runs **inside your agent host** (Claude Code, OpenCode, and other tools); every
artifact is plain markdown under a project's `.rc/`. No binary, no daemon, no lock-in.

## ✨ Highlights

- **Idea to code in a structured pipeline.** Each phase produces plain-markdown artifacts that feed
  the next: idea (optional) → PRD → TechSpec → tasks → execution → review. Start from an idea for full
  research and debate, or jump straight to a PRD if the scope is clear.
- **Runs in your host.** Claude Code, OpenCode, and other agent tools load the same skills, commands,
  agents, and hooks — auto-discovered from the plugin.
- **Cost-tiered specialists.** Bundled leaf-worker agents route work to the right model: recon on a
  cheap/fast tier (`rc-explorer`, `rc-librarian` on haiku), hard reasoning/review on the strong tier
  (`rc-oracle` on opus), bounded implementation mid-tier (`rc-fixer` on sonnet).
- **Codebase-aware.** Tasks aren't generic prompts — RC explores your codebase to discover patterns
  and grounds each task in real project context.
- **Memory & instincts between runs.** Curated per-project memory, cross-task workflow memory, and an
  observe→distill instincts loop keep context fresh without manual bookkeeping.
- **Multi-perspective review.** `rc-review-round` (multi-lens) and `rc-review-workflow` (automated
  review→fix loop) find and remediate issues.
- **Markdown everywhere.** PRDs, specs, tasks, reviews, and ADRs are human-readable, diffable markdown.
- **Resilience hooks.** Guardrails (git/commit/gate/db), formatting, observability, and a repair-guidance
  hook that helps the agent recover from a failed edit or delegation.
- **Local-first.** All state lives in `.rc/` — nothing leaves your machine.

## 📦 Installation

RC installs through your host's plugin/marketplace mechanism; skills, commands, agents, and hooks are
auto-discovered.

**Claude Code:**

```text
/plugin marketplace add rodolfochicone/rc-project
/plugin install rc@rc-project
```

Commands are namespaced `/rc:rc-*`; update with `/plugin marketplace update`. The marketplace add
clones the private repo, so GitHub read access is required (`gh auth login` or `GH_TOKEN`).

**OpenCode and other hosts:** clone this repo and symlink its asset directories — skills into the
cross-tool `~/.agents/skills` path and the OpenCode-specific agents/commands/plugin (under
`opencode/`) into `~/.config/opencode/`; update with `git pull`. Step-by-step in
[docs/claude-code-plugin.md](docs/claude-code-plugin.md).

> **Recommended companion: [Serena MCP](https://github.com/oraios/serena).** When Serena is connected,
> the code-touching skills (analyze, create-tasks/techspec/prd, execute-task, fix-analysis,
> fix-reviews, review-round, code-review) prefer its LSP-backed symbolic tools for more accurate,
> token-efficient navigation and editing, falling back to Grep/Glob when unavailable. Install it with
> `uv` per Serena's docs.

## 🔄 How It Works

```text
idea (optional) → PRD → TechSpec → tasks → execution → review → remediation → ship
      /rc-idea-factory        /rc-create-*        rc-tasks-workflow /       /rc-git
                                                  rc-execute-task
```

Each phase is a skill. Artifacts land under `.rc/tasks/<slug>/` (`_prd.md`, `_techspec.md`, task
files, `adrs/`, `reviews-NNN/`) and are consumed by the next phase. See `skills/rc/SKILL.md` for the
full reference and `skills/rc/references/workflow-guide.md` for a step-by-step walkthrough.

---

## 🧩 Skill catalog

Skills run **inside** your AI agent — no context switching. Every skill is namespaced under the
plugin: the host auto-routes to a skill by its `description`, or you can invoke one explicitly as
`/rc:<skill>` (e.g. `/rc:rc-analyze`). The catalog below is grouped by job; **"Use when"** tells the
agent (and you) exactly when each one fires.

### Meta — start here

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc` | Explains RC itself — pipeline, artifacts, agents, hooks, config. | You want to know what exists or how the workflow fits together (not to run a step). |
| `rc-find-skills` | Discovers and installs agent skills. | "Is there a skill for X?", "find a skill that…", extending capabilities. |
| `rc-enrichment-prompt` | Rewrites a rough request into a structured, execution-ready prompt (Objective / Context / Requirements / Acceptance criteria). | A request is vague/underspecified, or you ask to "enhance/enrich this prompt" before work starts. |
| `rc-to-prompt` | Packages code + context into a detailed prompt for **another** LLM to solve (the *what*, not the *how*). | Handing a problem off to an external model or teammate. |
| `rc-brainstorming` | Explores intent, requirements, and design before implementation. | Before any creative/greenfield work — features, components, new behavior. |
| `rc-council` | Multi-advisor debate (3–5 archetypes) with opening statements, tensions, and synthesis. | High-impact architecture/tech/product trade-offs; stress-testing a PRD or spec. Not for yes/no lookups. |

### Pipeline — plan

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-create-prd` | Idea → Product Requirements Document (interactive brainstorming + codebase/web research + ADRs). | Starting a new feature/product or capturing requirements. Not for tech design or code. |
| `rc-create-techspec` | PRD → Technical Specification via architecture exploration and clarification. | A PRD exists and needs an implementation design. |
| `rc-create-tasks` | PRD + TechSpec → independently implementable task files, enriched from the codebase. | Breaking a spec into executable, context-grounded tasks. |

### Pipeline — execute

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-execute-task` | Implements one task end-to-end: code, validate, track, commit. | A prompt includes a concrete task file to implement. Not for generic coding without a task file. |
| `rc-tasks-workflow` | Runs a slug's task files via the Claude Code `Workflow` tool — one subagent per task, in dependency order, with test evidence. | Executing all of a slug's tasks, Claude-orchestrated (Claude Code only). |
| `rc-fix-analysis` | Applies the "Implementation plan" from a prior `rc-analyze` report, with tests + verification. | An analysis exists and you want its plan implemented. |

### Analysis & understanding

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-analyze` | Deep, evidence-based read-only analysis → thorough report ending in an actionable plan. | Diagnose a bug, trace a flow, assess impact/feasibility. Not to review a diff for defects. |
| `rc-codemap` | Builds/refreshes a per-directory `codemap.md`; only re-maps changed dirs. | Read first in exploration-heavy tasks to cut token cost. |
| `rc-refactoring-analysis` | Fowler-catalog refactoring audit **plus** architectural audit (dead code, duplication, anti-patterns, type confusion) → prioritized report in `docs/_refacs/`. | Auditing code quality/codebase health, finding dead code, or preparing a refactoring sprint. |
| `rc-graphify` | Turns a codebase (or any input) into a persistent knowledge graph with query/path/explain tools. | Questions about architecture, file relationships, or when `graphify-out/` exists. |
| `rc-ubs` | Runs the Ultimate Bug Scanner over code. | Pre-commit quality checks or validating AI-generated code for bugs/security. |

### Review & remediation

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-code-review` | Rigorous, standards-driven review of the change set → categorized, severity-ranked report in `.rc/`. | On-demand quality gate of a diff/branch before merge. |
| `rc-review-round` | Multi-lens review → structured issue files compatible with `rc-fix-reviews`. | Creating a manual review round (no external provider). |
| `rc-review-workflow` | Automated review→fix→re-review loop via the `Workflow` tool until clean or a round cap. | Hands-off multi-lens review-and-remediate on a slug (Claude Code only). |
| `rc-simplify-review` | Single lens — over-engineering only → ranked delete-list with net lines/deps removable. | Opt-in pre-PR bloat pass, or auditing legacy code. |
| `rc-fix-reviews` | Triage, fix, verify, and resolve batched review issues under `reviews-NNN/`. | Remediating an existing review round. |
| `rc-fix-coderabbit-review` | End-to-end CodeRabbit remediation by PR number (export → fix → resolve threads). | Clearing CodeRabbit feedback on a specific PR. |
| `rc-adversarial-review` | Spawns reviewers on the **opposite** model (Claude↔Codex) to challenge the work. | You want a true cross-model critical pass. |
| `rc-impl-peer-review` | Optional cross-model Opus peer review of a diff, packaged for user-directed fixes. | You explicitly ask for an independent Opus review before commit/PR. |

### Quality, testing & discipline

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-final-verify` | Enforces fresh verification evidence before any "done"/commit/PR claim. | About to report success, hand off, or commit. |
| `rc-gan` | Adversarial generator↔evaluator loop that **exercises the running artifact** to drive subjective quality up to a target. | Refining UI/UX, copy, or CLI ergonomics that a pass/fail gate can't capture. |
| `rc-tdd` | Red-green-refactor loop with vertical tracer-bullet slices. | Building features/fixing bugs test-first; integration-style tests. |
| `rc-testing-anti-patterns` | Prevents testing mock behavior, test-only production methods, and blind mocking. | Writing/changing tests or adding mocks. |
| `rc-systematic-debugging` | Structured root-cause process before proposing fixes. | Any bug, test failure, or unexpected behavior. |
| `rc-no-workarounds` | Gate that rejects hacks/symptom patches (type assertions, lint suppressions, error swallowing…). | Debugging, fixing, or reviewing — to force root-cause fixes. |
| `rc-extreme-software-optimization` | Profile-driven perf optimization with behavior proofs. | "slow / bottleneck / p95 / latency / throughput" or algorithmic wins. |
| `rc-qa-execution` | Full-project QA as a real user: discover contract, run build/lint/test/startup, exercise flows E2E, fix regressions, re-gate. | Validating a branch/release/migration/refactor. Pairs with `rc-qa-report`. |
| `rc-qa-report` | Test plans, test cases, regression suites, and bug reports (Figma MCP validation). | Planning/documenting QA before execution. |

### Docs

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-readme` | Rewrites `README.md` from the real codebase (evidence-based), or guides writing/improving one by hand with templates matched to audience/project type. | Generating/refreshing a README, syncing docs after features land, or drafting/reviewing a README manually. |
| `rc-openapi` | Discovers HTTP endpoints/schemas from source → keeps `openapi.yaml` in sync. | After API changes, or to bootstrap a spec. |
| `rc-postman` | Discovers endpoints/contracts → Postman Collection (v2.1.0) + environments. | After routes/request schemas change. |

### Ship & git

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-git` | Ships work as a branch + PR (confirming each outward step and the PR target), and handles rebases / conflict resolution while preserving a clean history. | Shipping local work as a branch + PR, or rebasing feature branches / resolving conflicts. Not for in-place commits or force-push. |

### Memory & context

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-memory` | The single durable, cross-session memory: curated project facts **and** distilled instincts, as markdown under `.rc/memory/`. | Consult before working; record durable facts/learnings after. |
| `rc-workflow-memory` | Task-scoped memory across a slug's task executions under `.rc/tasks/{name}/memory/`. | A task prompt provides workflow-memory paths to read/update/promote. |
| `rc-compact` | Deliberate conversation compaction at logical boundaries, driven by real token usage. | Long multi-phase runs, to control what context survives. |
| `rc-context-budget` | Audits what fills the context window (agents, skills, rules, MCP schemas) and recommends cuts. | Sessions compact too early, or before adding more tooling. |
| `rc-lesson-learned` | Extracts engineering lessons from recent git history. | "What's the lesson here / takeaway?" reflection on a diff. |

### Config, security, integrations & scaffolding

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-audit` | Security-audits the agent config surface (`.claude`, `.mcp.json`, hooks, installed skills) for secrets, over-broad perms, injection vectors. | Reviewing an agent setup before sharing, or a periodic config audit. |
| `rc-jira` | PM-mode Jira via the Atlassian MCP: shape ideas, create/refine cards into PRD/TechSpec/sub-tasks, execute child tickets, GMUDs. | Any Jira work through the official MCP, with confirmation on writes. |
| `rc-new-project` | Scaffolds a new private repo in `rodolfochicone` from the TypeScript template. | Starting a brand-new TypeScript project. |

### Skill authoring & self-improvement

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-skill-best-practices` | Authors professional-grade skills to the agentskills.io spec (validation script + template + checklist). | Creating a new skill directory or optimizing metadata for discovery. |
| `rc-autoresearch` | Autonomously optimizes a skill: run → score against evals → mutate prompt → keep wins. | "Improve/optimize/benchmark this skill." |
| `rc-hookify` | Authors a new fail-open RC hook from a plain-language rule: writes the script, wires `hooks.json`, documents + verifies it. | Turning an every-time guardrail/formatter/observer into a hook. |

### Frontend & design

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-frontend-design` | Distinctive, production-grade interfaces that avoid generic AI aesthetics. | Building web components/pages/apps, or when a design skill needs project context. |
| `rc-interface-design` | Dashboards, admin panels, tools, interactive products (not marketing pages). | App/tool UI with craft and consistency. |
| `rc-minimalist-ui` | Editorial minimalism: warm monochrome, typographic contrast, flat bento grids. | You want that specific clean, document-style aesthetic. |
| `rc-redesign-existing-projects` | Audits and upgrades an existing site/app to premium quality without breaking it. | Modernizing existing UI in any CSS framework. |
| `rc-a11y` | Accessibility (WCAG 2.2 AA): semantic HTML, ARIA, keyboard nav, focus management, contrast, screen readers. | Building or reviewing UI that must be accessible. |
| `rc-shadcn-ui` | Complete shadcn/ui patterns: install, config, forms (RHF + Zod), theming, components. | Building UI with shadcn/ui + Radix + Tailwind. |
| `rc-storybook-stories` | Create/update/refactor Storybook stories to project patterns. | Adding stories for new components or fixing Storybook issues. |
| `rc-tech-logos` | Installs official tech/brand logos from the Elements registry. | Needing brand logos for landing pages, auth UIs, integrations. |
| `rc-drawio` | Diagrams: flowcharts, architecture, ER, sequence, class, wireframes → PNG/SVG/PDF. | Any request to draw/generate a diagram or `.drawio` file. |

### Content & media

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-seo` | Technical, on-page, and programmatic SEO — audit, content optimization, pages at scale. | Auditing or optimizing a site/content for search ranking. Not for paid media or off-page/backlinks. |
| `rc-video` | Video work — local `ffmpeg` processing, content/script planning, optional VideoDB. | Cutting/transcoding/subtitling media, or planning video content (Reels/Shorts/YouTube). |

### Library & framework references

Deep, opinionated guides for a specific library/runtime — auto-fire on the matching tech.

| Skill | Use when |
| --- | --- |
| `rc-react` | React 19+ components, hooks, state, `useEffect` patterns, TS integration. |
| `rc-vercel-react-best-practices` | React/Next.js **performance** patterns + composition patterns (compound components, render props, context) — Vercel Engineering. |
| `rc-tanstack` | TanStack ecosystem — Query/DB, Form, Router overview **plus** per-rule best practices for Query, Router and Start (server functions, middleware, SSR, auth). |
| `rc-app-renderer-systems` | Domain feature systems under a `systems/` directory (service layer + Query hooks + XState). |
| `rc-tailwindcss` | Tailwind CSS v4 patterns, responsive layouts, `tailwind-variants`. |
| `rc-zustand` | Zustand store organization and client-state patterns. |
| `rc-typescript-advanced` | Generics, conditional/mapped types, template literals, utility types. |
| `rc-zod` | Zod schemas, parsing, `safeParse`, `z.infer`, error handling. |
| `rc-vitest` | Vitest: tests, mocking, coverage, filtering, fixtures. |
| `rc-golang-pro` | Idiomatic concurrent Go, gRPC/REST microservices, pprof, generics. |
| `rc-python` | Idiomatic typed Python 3.12+ — PEP 695 generics, asyncio/TaskGroup, pytest, ruff, uv packaging. |
| `rc-ai-sdk` | Vercel AI SDK: `generateText`/`streamText`, tools, agents, providers, streaming. |

### Backend, data & operations

| Skill | Purpose | Use when |
| --- | --- | --- |
| `rc-sql` | Relational DB — query optimization (EXPLAIN, indexes, N+1), schema design; read-only by default. | Writing/reviewing queries or modeling schema. Not for repo-specific migrations or NoSQL. |
| `rc-observability` | Logs, metrics, traces, and incident response — instrumentation, SLOs, postmortems. | Instrumenting a service, defining alerts, or running root-cause analysis. |
| `rc-resilience` | Event-driven resilience — idempotency, retries/backoff, DLQ, poison messages, timeouts, circuit breaker. | Designing/reviewing message producers/consumers (EventBridge, SQS) or cross-service calls. |

### TUI, CLI & multi-agent

| Skill | Use when |
| --- | --- |
| `rc-tui-design` | Framework-agnostic TUI design — layouts, color, keybindings, dashboards. |
| `rc-tui-glamorous` | Charmbracelet ecosystem router — Gum (shell), Lip Gloss, Bubble Tea, Wish/SSH. |
| `rc-bubbletea` | Deep Go Bubble Tea app patterns — Elm architecture, dual-pane, weight-based layout. |
| `rc-smux` | Control tmux panes and message between AI agents (tmux-bridge CLI). |
| `rc-smux-rc-pairing` | Interactive tmux pairing: orchestrator + Codex author + Claude challenger driving a RC run. |

### Research

| Skill | Use when |
| --- | --- |
| `rc-exa-web-search-free` | Free web/code/company search via the Exa MCP (no API key). |

### Project-specific (Escale)

| Skill | Use when |
| --- | --- |
| `rc-automation-guides` | Working in the `rc-automation` repo — routes to the right dev/ops guide (endpoint, migration, event, deploy, docs…). |
| `rc-portal-design` | Any UI work in `rc-portal` — design system, tokens, Tailwind, Storybook, a11y gates. |

### Bundled extension

Optional, under `extensions/rc-idea-factory`: **`rc-idea-factory`** — raw idea → structured idea spec
with market research, business analysis, and a council debate. Sits before `rc-create-prd`.

---

## ⌨️ Commands

Commands are explicit entry points (`/rc:<command>`) that chain skills for a whole phase. Skills
auto-fire from context; commands are what you type to drive a stage on purpose.

| Command | What it does |
| --- | --- |
| `/rc-plan` | Planning pipeline — PRD → TechSpec → task breakdown, in sequence. |
| `/rc-exec` | Executes a feature's implemented tasks via `rc-execute-task`. |
| `/rc-review` | Review-and-fix loop — one simplify pass, then up to 3 rounds of review-round + fix-reviews. |
| `/rc-pipe` | The **full** pipeline end to end — plan → execute → review → docs → PR. |
| `/rc-gan` | Adversarial quality loop for subjective quality (UI/UX, CLI, copy) against a target score. |
| `/rc-git` | Ships current work as a branch + PR (`rc-git`), then distills session learnings (`rc-memory`). |
| `/rc-docs` | Generates/refreshes project docs — README, Postman, OpenAPI. |
| `/rc-commit-msg` | Generates a Conventional Commit message from the **staged** diff (does not commit). |
| `/rc-plano` | Lightweight Explore→Plan flow — investigate first, deliver a plan for approval, no code yet. |

## 🤖 Bundled specialist agents

Leaf-worker subagents (under `agents/`, discovered as `rc:<name>`) you delegate to, each on a
cost-appropriate model tier. They carry no `Task`/`Agent` tool, so they cannot spawn further
subagents (the recursion cap). See `skills/rc/references/delegation-contract.md`.

| Agent | Lane | Model | Use when |
| --- | --- | --- | --- |
| `rc-explorer` | Read-only codebase recon — returns a compressed map, not file dumps. | haiku | Start of any non-trivial task, to discover what exists. |
| `rc-librarian` | External docs / web / library research. | haiku | The task hinges on current, version-specific library or API behavior. |
| `rc-fixer` | Bounded implementation of well-scoped, mechanical work. | sonnet | Objective, files, and constraints are known — execution, not discovery. |
| `rc-oracle` | Strategic advisor & read-only reviewer for high-stakes calls. | opus | Major architectural decisions or hard debugging with an unclear root cause. |

**Council archetypes** (dispatched by the `rc-council` skill for multi-perspective debate, not called
directly): `architect-advisor` (systems & long-horizon), `pragmatic-engineer` (execution reality),
`security-advocate` (assume-breach), `product-mind` (user value & opportunity cost), `devils-advocate`
(informed skeptic), `the-thinker` (reframing).

## 🧠 Memory & instincts

- **Project memory** — `.rc/memory/` (`INDEX.md` + one file per durable fact), curated by `rc-memory`.
- **Workflow memory** — `.rc/tasks/<slug>/memory/` (shared `MEMORY.md` + per-task files), maintained
  by `rc-workflow-memory` so each agent inherits context from previous tasks.
- **Instincts** — the `observe` hook appends tool observations to `.rc/memory/observations.jsonl`;
  `rc-memory` distills recurring patterns into confidence-scored learnings.

## 🪝 Hooks

Harness-only guardrails (no model-context cost), wired in `hooks/hooks.json`:

- **`git-guard`** — blocks destructive/history-rewriting git commands.
- **`commit-guard`** — gates commits behind verification.
- **`db-guard`** — enforces read-only database access by default (blocks write/DDL SQL without approval).
- **`go-mod-guard`** / **`gateguard`** — protect `go.mod` and force investigation before risky edits.
- **`go-fmt`** — formats Go on write. **`observe`** — feeds the instincts loop. **`repair-guidance`** —
  helps the agent recover from a failed edit/delegation. **`notify`** — Stop/Notification signals.
- **`memory-load`** — `SessionStart` warm-start: surfaces a bounded summary of `.rc/memory/` (facts +
  learnings) and nudges distillation when observations pile up. Silent outside RC projects.

## 🖥️ Supported hosts

RC ships for Claude Code, OpenCode, and other agent tools. Skills, commands, agents, and hooks are
auto-discovered when the plugin is installed; OpenCode-specific variants live under `opencode/`.
Output styles under `output-styles/` are versioned here but applied per-host (Claude Code loads them
from `~/.claude/output-styles/`).

## 🤝 Contributing

There is no build step — components are markdown, JSON, and small Node/Bash scripts. Validate a change
with `node scripts/plugin-smoke.mjs` (frontmatter + hook wiring) and, for task files,
`node scripts/validate-tasks.mjs --slug <slug>`. See `AGENTS.md` for conventions and `CONTRIBUTING.md`.

## 📄 License

MIT — see [LICENSE](LICENSE).

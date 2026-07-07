<div align="center">
  <img src="imgs/rc-avatar-dark.png" alt="rc" width="140">

  <h1>rc</h1>
  <p><strong>AI-assisted development workflows as skills, slash commands, hooks, and agents — for Claude Code and OpenCode.</strong></p>
  <p>
    <a href="LICENSE">
      <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT">
    </a>
  </p>
</div>

rc is a bundle of **skills, slash commands, hooks, and agents** that drive the full lifecycle of AI-assisted development — product ideation (PRD), technical specification, codebase-informed task breakdown, task execution, and PR-review remediation — directly inside your coding agent.

Everything here is **plain markdown and shell**. There is no binary to install and nothing to build: the skills run inside Claude Code or OpenCode, the commands are slash commands, and the hooks are shell scripts your agent invokes at lifecycle events.

## 📦 Install

### Claude Code (plugin marketplace)

```text
/plugin marketplace add rodolfochicone/rc-project   # register the marketplace
/plugin install rc@rc-project                       # install the plugin
```

Plugin skills are namespaced under the plugin, so slash commands surface as `/rc:rc-create-prd`, `/rc:rc-create-techspec`, `/rc:rc-create-tasks`, `/rc:rc-review-round`, and so on. Update with `/plugin marketplace update`.

The plugin ships from this repo's layout (`skills/`, `commands/`, `agents/`, `hooks/hooks.json`) — Claude Code discovers them by convention. See [`docs/claude-code-plugin.md`](docs/claude-code-plugin.md) for the full runbook.

> **Repo is private.** `/plugin marketplace add` clones this repository, so the Claude Code environment needs GitHub read access — sign in with `gh auth login` or export `GH_TOKEN` / `GITHUB_TOKEN` before adding the marketplace.

### OpenCode

OpenCode reads agents, commands, and plugins from its config directory. Copy the `opencode/` bundle into place:

```bash
# Project-local (recommended)
cp -R opencode/agent    .opencode/agent
cp -R opencode/commands .opencode/command
cp -R opencode/plugin   .opencode/plugin

# or global
cp -R opencode/agent    ~/.config/opencode/agent
cp -R opencode/commands ~/.config/opencode/command
cp -R opencode/plugin   ~/.config/opencode/plugin
```

The `opencode/plugin/rc-hooks.ts` plugin gives OpenCode Claude-parity hook enforcement by shelling out to the shared scripts under `hooks/scripts/`.

## 🧩 What's inside

| Directory        | Contents                                                                             |
| ---------------- | ------------------------------------------------------------------------------------ |
| `skills/`        | 31 skills (`SKILL.md` + references) — the workflow logic                             |
| `commands/`      | Claude Code slash commands that route to the skills                                  |
| `agents/`        | 12 Claude Code plugin agents — one per pipeline phase plus two read-only helpers     |
| `hooks/`         | `hooks.json` + shell scripts run at agent lifecycle events                           |
| `opencode/`      | OpenCode agents, commands, and the `rc-hooks` plugin                                 |
| `rules/`         | Coding rules injected into agent context (`common.md`, `go.md`)                      |
| `.claude-plugin/`| Plugin + marketplace manifests                                                       |

## 🛠️ Skills

The core pipeline — Idea → PRD → TechSpec → Tasks → Execution → Review — where each phase produces plain markdown artifacts under `.rc/tasks/<name>/` that feed the next:

| Skill                  | Purpose                                                             |
| ---------------------- | ------------------------------------------------------------------ |
| `rc`                   | Orchestrator — routes to the right phase skill                     |
| `rc-create-prd`        | Idea → Product Requirements Document with ADRs                     |
| `rc-create-techspec`   | PRD → Technical Specification with architecture exploration        |
| `rc-create-tasks`      | PRD + TechSpec → independently implementable task files            |
| `rc-execute-task`      | Execute one task end-to-end: implement, validate, track            |
| `rc-review-round`      | Comprehensive code review → structured issue files                 |
| `rc-fix-reviews`       | Triage, fix, verify, and resolve review issues                     |
| `rc-fix-analysis`      | Turn an analysis plan into applied code changes                    |
| `rc-final-verify`      | Enforce verification evidence before any completion claim          |
| `rc-git`               | Branch, push, and open a PR with a confirmation at each step       |

Quality, analysis, and review:

| Skill                | Purpose                                                                       |
| -------------------- | ----------------------------------------------------------------------------- |
| `rc-analyze`         | Deep, evidence-based diagnosis / tracing of existing code                     |
| `rc-code-review`     | Review a change set for correctness, security, and performance defects        |
| `rc-simplify-review` | Single-lens over-engineering pass → ranked delete-list                        |
| `rc-audit`           | Audit the agent config surface (settings, MCP, hooks) for secrets and risks   |
| `rc-doctor`          | Health-check the rc installation itself — hooks, manifests, parity, frontmatter |
| `rc-gan`             | Adversarial generator↔evaluator loop that drives subjective quality up        |

Deep work, execution, and navigation:

| Skill          | Purpose                                                                          |
| -------------- | -------------------------------------------------------------------------------- |
| `rc-deepwork`  | Scheduler discipline for heavy sessions: plan → review → phased, gated execution |
| `rc-loop`      | Generate → verify → retry against an explicit pass/fail success gate             |
| `rc-worktrees` | Git worktrees as isolated lanes for parallel or risky work                       |
| `rc-codemap`   | Hierarchical per-directory `codemap.md` for token-efficient navigation           |

Memory, context, and learning:

| Skill                 | Purpose                                                                    |
| --------------------- | -------------------------------------------------------------------------- |
| `rc-workflow-memory`  | Cross-task context so agents pick up where the last run left off           |
| `rc-project-memory`   | Durable project facts that persist across sessions                         |
| `rc-instincts`        | Distill recurring corrections into atomic, confidence-scored instincts     |
| `rc-reflect`          | Review recent work and recommend the smallest reusable asset to add        |
| `rc-context-budget`   | Audit what consumes the context window and recommend the highest-impact trims |
| `rc-compact`          | Compact the conversation deliberately at logical task boundaries           |

Docs, APIs, and integrations:

| Skill            | Purpose                                                     |
| ---------------- | ----------------------------------------------------------- |
| `rc-readme`      | Generate or refresh a README grounded in the real codebase  |
| `rc-openapi`     | Produce an OpenAPI spec from the codebase                   |
| `rc-postman`     | Produce a Postman collection                                |
| `rc-jira`        | Create, read, comment, and transition Jira issues via MCP   |
| `rc-new-project` | Scaffold a new project agentically                          |

## ⌨️ Slash commands

Claude Code commands (`commands/`) and OpenCode commands (`opencode/commands/`) route to the skills:

`/rc-plan` · `/rc-exec` · `/rc-pipe` · `/rc-review` · `/rc-git` · `/rc-gan` · `/rc-docs` (Claude) · `/rc-prd` · `/rc-techspec` · `/rc-tasks` · `/rc-fix` (OpenCode)

## 🤖 Agents

One agent per pipeline phase plus two read-only support agents, each pinning a model to its role. In Claude Code they live under `agents/` (invoke via the Task tool or `@rc-*`); in OpenCode under `opencode/agent/`. The `rc` agent orchestrates the rest.

| Agent          | Role                                    | Model  |
| -------------- | --------------------------------------- | ------ |
| `rc`           | Orchestrates the full pipeline          | sonnet |
| `rc-prd`       | Idea → PRD                              | opus   |
| `rc-techspec`  | PRD → TechSpec                          | opus   |
| `rc-tasks`     | PRD + TechSpec → task files             | sonnet |
| `rc-exec`      | Implement one hard task end to end      | opus   |
| `rc-exec-bulk` | Implement many simple tasks in parallel | sonnet |
| `rc-review`    | Independent, critical code review       | opus   |
| `rc-fix`       | Resolve review/QA issues at root cause  | opus   |
| `rc-gan`       | Adversarial quality loop (UI/CLI/copy)  | opus   |
| `rc-git`       | Branch, commit messages, PR             | haiku  |
| `rc-explorer`  | Fast read-only codebase navigation      | haiku  |
| `rc-librarian` | Read-only library / docs research       | sonnet |

## 🪝 Hooks

Shell hooks under `hooks/scripts/`, wired in `hooks/hooks.json` (Claude Code) and mirrored by `opencode/plugin/rc-hooks.ts` (OpenCode):

| Hook                     | Event         | Enforces                                                  |
| ------------------------ | ------------- | -------------------------------------------------------- |
| `git-guard.sh`           | PreToolUse    | Blocks destructive git commands without permission        |
| `commit-guard.sh`        | PreToolUse    | Guards commit messages / attribution rules                |
| `go-mod-guard.sh`        | PreToolUse    | Blocks hand-editing `go.mod` in Go projects               |
| `gateguard.sh`           | PreToolUse    | Gate before edits/writes                                  |
| `go-fmt.sh`              | PostToolUse   | Formats Go after edits                                    |
| `observe.sh`             | PostToolUse   | Records tool observations for `rc-instincts`             |
| `session-recall.sh`      | SessionStart  | Recalls project memory into the session                   |
| `phase-reminder.sh`      | SessionStart  | Reminds the active workflow's pipeline phase and next step |
| `precompact-capture.sh`  | PreCompact    | Captures context before compaction                        |
| `notify.sh`              | Stop / notify | Desktop notifications on terminal state                   |

> The Go-specific hooks (`go-mod-guard`, `go-fmt`) only act inside Go projects; they are no-ops elsewhere.

## 🗺️ Usage map

How the pieces fit: **slash commands** are the entry points, each command drives one or more **skills** (the workflow logic), heavy phases delegate to **agents** (each pinned to the right model), **hooks** guard and observe everything in the background, and the **memory skills** make the next run smarter than the last.

### The pipeline at a glance

```mermaid
flowchart LR
    A(["💡 Idea"]) --> B["/rc-plan<br/>PRD → TechSpec → Tasks"]
    B --> C["/rc-exec<br/>implement each task"]
    C --> D{"Subjective<br/>surface?"}
    D -- "UI / CLI / copy" --> E["/rc-gan<br/>quality loop"]
    D -- no --> F
    E --> F["/rc-review<br/>simplify + review ≤3 rounds + fix"]
    F --> G["/rc-docs<br/>README · OpenAPI · Postman"]
    G --> H["/rc-git<br/>branch → push → PR"]
    H --> I(["📚 rc-instincts<br/>distill learnings"])
    P["/rc-pipe — runs the whole line end to end"] -.-> B
```

Every phase writes markdown artifacts under `.rc/tasks/<slug>/`, so you can stop, edit an artifact by hand, and resume at any phase.

### "I want to…" — the decision table

| I want to…                                        | Claude Code                       | OpenCode                          |
| ------------------------------------------------- | --------------------------------- | --------------------------------- |
| Build a feature end to end                        | `/rc-pipe <idea>`                 | `/rc-pipe <idea>`                 |
| Plan only (PRD → TechSpec → Tasks)                | `/rc-plan <idea>`                 | `/rc-plan <idea>`                 |
| Run a single planning phase                       | `/rc:rc-create-prd` · `-techspec` · `-tasks` | `/rc-prd` · `/rc-techspec` · `/rc-tasks` |
| Execute the planned tasks                         | `/rc-exec <slug>`                 | `/rc-exec <slug>`                 |
| Review-and-fix until clean                        | `/rc-review <slug>`               | `/rc-review <slug>` + `/rc-fix`   |
| Raise UI/CLI/copy quality to a score              | `/rc-gan <target>`                | `/rc-gan <target>`                |
| Ship as branch + PR, then distill learnings       | `/rc-git`                         | `/rc-git`                         |
| Refresh docs (README, OpenAPI, Postman)           | `/rc-docs`                        | ask the `rc` agent                |
| Diagnose a bug / trace how code works             | `rc-analyze` → `rc-fix-analysis`  | same skills via the `rc` agent    |
| Ad-hoc quality gate on a diff                     | `rc-code-review`                  | `rc-review` agent                 |
| Hunt over-engineering only                        | `rc-simplify-review`              | same skill                        |
| Ask "where is X?" cheaply                         | `rc-explorer` agent               | `rc-explorer` agent               |
| Research a library before adopting it             | `rc-librarian` agent              | `rc-librarian` agent              |
| Health-check the rc install itself                | `rc-doctor`                       | `rc-doctor` skill                 |
| Audit agent config for secrets/risks              | `rc-audit`                        | `rc-audit`                        |
| See what bloats the context window                | `rc-context-budget`               | `rc-context-budget`               |
| Compact the session at a sane boundary            | `rc-compact`                      | `rc-compact`                      |
| Remember a durable project fact                   | `rc-project-memory`               | `rc-project-memory`               |
| Turn repeated corrections into instincts          | `rc-instincts`                    | `rc-instincts`                    |
| Find the smallest reusable asset to add           | `rc-reflect`                      | `rc-reflect`                      |
| Work in parallel/risky lanes                      | `rc-worktrees`                    | `rc-worktrees`                    |
| Heavy multi-phase session discipline              | `rc-deepwork`                     | `rc-deepwork`                     |
| Loop until a machine-checkable gate passes        | `rc-loop`                         | `rc-loop`                         |
| Learn what rc can do                              | `rc` skill                        | `rc` agent                        |

### A typical day

1. **New feature** — `/rc-plan payment-retries`, answer the PRD/TechSpec questions, review the generated task files.
2. **Build** — `/rc-exec payment-retries`; each task is implemented, verified, and tracked; hooks guard git and gate edits throughout.
3. **Harden** — `/rc-review payment-retries` loops simplify → review → fix until a round comes back clean.
4. **Ship** — `/rc-git` moves it to a branch and opens the PR, confirming each outward-facing step; `rc-instincts` captures what the session taught.
5. **Anytime** — `rc-explorer` for quick lookups, `rc-analyze` when something breaks, `rc-doctor` when rc itself misbehaves.

An interactive HTML version of this map lives at [`docs/usage-map.html`](docs/usage-map.html).

## 🔎 CI

Every push runs [`validate.yml`](.github/workflows/validate.yml): `bash -n` on all hook scripts, JSON validity + version consistency of the manifests, skill/command/agent frontmatter checks (including the marketplace angle-bracket rule), and Claude↔OpenCode guard-hook parity.

## 📄 License

MIT — see [LICENSE](LICENSE).

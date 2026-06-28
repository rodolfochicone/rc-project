---
name: rc-context-budget
description: Audits what consumes the agent's context window — installed agents, skills, rules, MCP tool schemas, and CLAUDE.md/AGENTS.md — estimates token cost, classifies each as always/sometimes/rarely needed, and recommends the highest-impact reductions. Use when sessions compact too early, when MCP/tooling feels heavy, or before adding more agents/skills. Do not use for application performance profiling or runtime memory analysis.
model: sonnet
effort: medium
argument-hint: "[config-root]"
---

# Context Budget

The context window is a budget, not free space. Every always-loaded agent description, skill, rule, MCP tool schema, and `CLAUDE.md` line is spent on *every* turn — a 200k window can effectively start at 70k once enough tooling is enabled, forcing early compaction and degrading reasoning. This skill inventories that fixed overhead, estimates its token cost, and tells you what to trim. Read-only.

## What to inventory

Discover and measure each contributor (skip what does not exist):

- **Always-loaded prose**: `CLAUDE.md`, `AGENTS.md`, and any always-on rules/`.cursor/rules` injected globally.
- **Agents**: `.claude/agents/**`, `.opencode/agent/**`, `~/.rc/agents/**` — every agent's `description` is loaded whenever the orchestrator considers delegating (Task tool). Long descriptions cost on every invocation.
- **Skills**: bundled + installed `*/SKILL.md`. The `description` frontmatter is what loads for routing; the body loads only when the skill activates.
- **MCP servers**: each enabled server's tool schemas load up front. Estimate **~400–600 tokens per tool**; a server exposing 20 tools is ~10k tokens always-on.
- **Slash commands**: their `description`/`argument-hint`.

## Estimating tokens

- Prose/markdown: `tokens ≈ words × 1.3`. Count words with `wc -w`.
- MCP: `tokens ≈ tool_count × 500`. Count tools from the server's manifest/inspect output, or estimate from its docs.
- Report per-item and a total "fixed overhead before the first user turn".

## Classify and recommend

Tag each contributor:

- **always** — needed every session (core conventions, the agents you always route to). Keep, but keep tight.
- **sometimes** — needed for a class of task (a stack's skill, a provider MCP). Candidate for on-demand enabling.
- **rarely** — niche; loaded by default but seldom used. Prime candidate to disable.

Then give the **top 3 reductions ranked by tokens saved**, each with the concrete action and estimated saving. The usual biggest levers, in order:

1. **Disable rarely-used MCP servers** (biggest single lever — hundreds to thousands of tokens each). Replace a heavy always-on MCP with a thin CLI-wrapping command/skill where possible (e.g. a `gh pr` command instead of the GitHub MCP).
2. **Trim long agent descriptions** — they load on every Task consideration. Move detail into the agent body; keep the description a tight routing trigger.
3. **Shorten always-loaded prose** — fold repeated guidance into a skill/rule that loads on demand instead of living in `CLAUDE.md`.

## Output

Print a table (`contributor | type | est. tokens | class`), the total fixed overhead, and the ranked top-3 reductions with estimated savings and the exact action for each. Note that estimates are approximate (±20%); the goal is relative ranking, not exact accounting.

## Critical Rules

- Read-only. Recommend; do not disable servers, edit configs, or delete skills yourself.
- Never recommend trimming a security control (permission `deny`, a guardrail hook) to save tokens.
- Rank by tokens saved × likelihood-unused; do not just list everything.

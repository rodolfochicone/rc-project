# Delegation Contract

How RC skills hand work to specialist subagents: which agent for which lane, at which model
tier, with a self-contained prompt and non-overlapping write ownership. Referenced by the
workflow skills (`rc-create-tasks`, `rc-execute-task`, `rc-tasks-workflow`,
`rc-review-workflow`). The specialist agents live in `agents/` and are auto-discovered by the
plugin as `rc:<name>`; pass the bare name as `subagent_type` when delegating.

## The pantheon — route by lane and cost

| Agent | Lane | Model tier | Delegate when | Don't |
|---|---|---|---|---|
| `rc-explorer` | read-only recon | **haiku** (cheap/fast) | broad or uncertain search; discover what exists before planning; parallel searches | you already know the path and need the literal contents |
| `rc-librarian` | external docs / web | **haiku** | version-specific library behavior, official API examples, web bug investigation | stable API you're confident about; info already in context |
| `rc-oracle` | architecture, review, hard debugging | **opus** (strong/costly) | high-stakes decisions, a 2nd+ failed fix, risky refactor, code review, simplification | routine choices, first-pass fixes |
| `rc-fixer` | bounded implementation | **sonnet** | clearly scoped mechanical work; parallelizable per folder | needs research/design, unclear requirements |

Reserve the expensive tier (oracle/opus) for judgment that warrants it; push recon and docs to
the cheap tier (explorer/librarian/haiku); keep execution mid-tier (fixer/sonnet). This is the
"mix models for cost" lever — it only pays off if skills actually route through these agents
instead of doing everything at the caller's model.

`rc-oracle` and `rc-fixer` also carry the `Skill` tool, so they can run RC skills (e.g. a
review skill, `rc-final-verify`) while still lacking `Task`/`Agent` — they participate in
skill-driven work without being able to fan out further.

## Model & effort routing

The tiers above are aliases the host resolves to the session's configured models — as of mid-2026:
**haiku** → Haiku 4.5 (fastest/cheapest), **sonnet** → Sonnet 5 (balanced), **opus** → Opus 4.8
(strongest). Reserve the most capable tier (Fable 5, if configured) for the hardest long-horizon
reasoning only — it costs more than Opus. Don't pin a model in an agent's frontmatter unless the
tier is load-bearing; otherwise inherit the session model.

`effort` is a second, orthogonal lever (`low`/`medium`/`high`/`xhigh`/`max`; default `high`). It
controls how much the model thinks and how many tool calls it makes — independent of which model
runs. Route it by task, not by tier:

| Task shape | Effort | Why |
|---|---|---|
| Recon, mechanical edits, cheap subagent lanes (explorer/librarian) | `low` | Fewer, more-consolidated tool calls; terser output |
| Most implementation and review work | `high` | The default sweet spot for quality vs. token cost |
| Hard coding / long-horizon agentic runs (fixer on a big change, oracle) | `xhigh` | Best setting for coding and agentic use cases |
| Correctness matters more than cost (security, data-integrity, a 2nd failed fix) | `max` | Spend to be right |

Two rules of thumb: give long-horizon agentic work the **full task spec up front in one turn** and
run it at `high`/`xhigh`; drop cheap subagents to `low`. Pairing a cheap tier with `low` effort is
the real cost win — mixing models without dialing effort leaves most of the savings on the table.

## Where the pantheon fits (and where it does not)

- **Fits cleanly:** ad-hoc delegation from an orchestrating skill; read-only recon (the
  `rc-create-tasks` exploration step routes to `rc-explorer`); read-only review analysis; a
  single bounded edit.
- **Does NOT fit as-is:** phases that are sequential *by design* to protect the shared working
  tree — notably `rc-tasks-workflow`, which runs one task at a time via the `Workflow` tool and
  keeps worktree-isolated parallelism out of scope on purpose. Do not swap those phase agents
  for a restricted pantheon agent; keep the full-tool executor there. The upgrade path is
  worktree-isolated parallel `rc-fixer`s with per-folder ownership (above), not a drop-in
  `agentType` swap.
- **Wiring note:** when routing a `Workflow`-tool `agent()` to a pantheon agent, confirm the
  exact `agentType` identifier on the target host (plugin agents are namespaced, e.g.
  `rc:rc-oracle`) — a wrong identifier silently falls back to the default agent.

## Task Prompt Contract

Every delegated prompt must be self-contained. Include:

- **objective** — the single outcome the task is responsible for.
- **scope** — the files/dirs/subsystem in play (paths, not vague areas).
- **constraints** — patterns to mirror, conventions, what must not change.
- **ownership** — the exact files/folders this agent may write (see below).
- **edits** — whether writes are allowed at all (read-only agents never edit).
- **expected output** — the shape of the result you want back.
- **validation** — what to run and report (build/tests/lint).
- **what NOT to do** — the out-of-scope traps.

A prompt missing objective, scope, or expected-output leaves the agent guessing — the most
common cause of a wasted delegation. (The `repair-guidance` hook nudges a failed delegation,
but prevention is cheaper than a retry.)

## Write ownership — parallel fan-out

- One write-capable agent owns a file/folder at a time. Never run two `rc-fixer`s over
  overlapping paths.
- Scope parallel fixers by folder/module and state each one's ownership boundary in its prompt.
- Read-only agents (`rc-explorer`, `rc-librarian`, `rc-oracle`) may run in parallel with
  anything, including with each other and with one writer.
- For parallel writers, prefer git worktree isolation (`Agent isolation:"worktree"`) so
  concurrent edits cannot corrupt each other; reconcile the branches afterward.

## Recursion

Specialist agents are **leaf workers**: they are not given the `Task`/`Agent` tool, so they
cannot spawn further subagents. Fan-out happens only at the orchestrating-skill level. Keep it
that way — the tool restriction *is* the recursion cap (no depth counter needed).

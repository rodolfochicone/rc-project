# Provider: Linear (official Linear MCP)

All work goes through the official Linear MCP server (`mcp.linear.app`). Never call the Linear GraphQL API directly.

- **Connectivity probe:** `list_teams`
- **Sync file:** `.rc/tasks/<slug>/_linear-sync.md`
- **Task frontmatter key:** `linear_key: <SUB-ID>`
- **Issue URL:** `https://linear.app/<workspace>/issue/<ID>`

## Tooling contract

| Purpose | Tool |
| --- | --- |
| Verify connectivity & list teams | `list_teams` |
| Team details | `get_team` |
| Workflow states of a team | `list_issue_statuses` (detail: `get_issue_status`) |
| List / create labels | `list_issue_labels` / `create_issue_label` |
| Resolve a user | `list_users` / `get_user` |
| Create **and** update issues (sub-issue via `parentId`; state, priority, labels, assignee, delegate) | `save_issue` |
| Read an issue (incl. sub-issues, git branch name) | `get_issue` |
| Search / filter issues (`assignee: "me"` for the user's own; `parentId` for children) | `list_issues` |
| Read / add or edit comments | `list_comments` / `save_comment` |
| Projects | `list_projects` / `get_project` / `save_project` |
| Cycles / milestones | `list_cycles` / `list_milestones` / `save_milestone` |
| Workspace documents | `list_documents` / `get_document` / `save_document` |

**Three rules that apply to every call:**

1. **Always resolve the team first.** `save_issue` requires a team when creating, and workflow states, labels, cycles, and estimates are all **per-team** — discover them for the chosen team (`list_issue_statuses`, `list_issue_labels`); never guess state or label names.
2. **Markdown is native.** Issue descriptions and comments accept markdown directly — use literal newlines, no escape sequences, no format flag.
3. **`save_*` tools are create-or-update.** `save_issue` without `id` creates; with `id` (e.g. `RC-123`) it updates — same pattern for `save_comment` / `save_project`. Never pass `id` when creating. Sub-issues are `save_issue` with `parentId`; state moves are `save_issue` with `state` set to a state discovered via `list_issue_statuses`. State names are customizable per team — only their categories are fixed (Triage, Backlog, Unstarted, Started, Completed, Canceled).

## Connectivity & scope (Phase 0 completion)

1. Call `list_teams`. Success with ≥1 team → capture each team's id and key (the issue prefix, e.g. `RC`). Empty, error, or tool not present → not configured; guide *MCP setup* below.
2. If exactly one team, use it. If several, ask via **AskUserQuestion** which team to use, listing them by name and key.

## The Linear model (quick map)

- **Hierarchy:** Workspace → Initiatives → **Projects** (with Milestones) → **Issues** (with Sub-issues). **Teams** own issues and each team runs its own workflow and **Cycles** (the sprint equivalent). **Triage** is the review queue for new, unscheduled work.
- **Native child issue:** the **Sub-issue** (`save_issue` with `parentId`) — any regular issue can be a parent.
- **Priorities (exact API values):** `0 = None`, `1 = Urgent`, `2 = High`, `3 = Medium`, `4 = Low`.
- **Coming from Jira:** Epic → Project · Story/Task/Bug → Issue · Sub-task → Sub-issue · Sprint → Cycle · Component → Label.
- **Issue-writing (Linear Method):** write issues, not user stories — a short, direct title; a lean description with just enough context to execute, linking out for depth; one concrete task with a defined outcome per issue; a single owner. Reserve a **Project** for real multi-issue initiatives that need coordination — don't wrap a loose pile of issues in one.
- **Agents are first-class:** issues can be **delegated** to agent app-users (the `delegate` field on `save_issue`) while the human stays the responsible assignee. Delegating is an outward-facing write like any other — confirm first.
- **Git tie-in:** `get_issue` returns the issue's **git branch name** — offer it to the `rc-git` flow so branches match Linear's auto-link format.

## MCP setup (when the probe fails)

Tell the user the official Linear MCP isn't connected and walk them through the two steps — **add** the server, then **authorize** it.

**Step 1 — Add the server (run in their terminal):**

```bash
claude mcp add --transport http linear-server https://mcp.linear.app/mcp
```

**Step 2 — Authorize the connection.** Adding the server does not authenticate it; the user must complete a browser-based **OAuth 2.1** flow (dynamic client registration — no manual OAuth app needed). Inside Claude Code:

1. Run the `/mcp` command.
2. Select the `linear-server` server (it shows as needing authentication).
3. Choose **Authenticate** / **Login** — a browser opens for the Linear OAuth flow.
4. Approve access; the browser confirms and the server status flips to connected.

Access respects the user's existing Linear workspace permissions.

- **Prerequisites:** a Linear workspace, and a browser to authorize.
- For other MCP clients, point them to the same endpoint (`https://mcp.linear.app/mcp`; legacy SSE at `https://mcp.linear.app/sse`) and follow that client's authorization flow.
- Linear has **no official CLI** — programmatic access is the MCP, the GraphQL API (`https://api.linear.app/graphql`), or the `@linear/sdk` TypeScript SDK. Never promise a `linear` CLI.

After they add **and** authorize it, re-run the probe.

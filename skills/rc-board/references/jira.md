# Provider: Jira (official Atlassian MCP)

All work goes through the official Atlassian MCP server (Atlassian Remote MCP / Rovo). Never call the Jira REST API directly.

- **Connectivity probe:** `getAccessibleAtlassianResources`
- **Sync file:** `.rc/tasks/<slug>/_jira-sync.md`
- **Task frontmatter key:** `jira_key: <SUBKEY>`
- **Issue URL:** `<site-url>/browse/<KEY>`

## Tooling contract

| Purpose | Tool |
| --- | --- |
| Verify connectivity & list sites | `getAccessibleAtlassianResources` |
| List projects (create-eligible) | `getVisibleJiraProjects` (`action: "create"`) |
| List issue types for a project | `getJiraProjectIssueTypesMetadata` |
| Field metadata (which are required) | `getJiraIssueTypeMetaWithFields` |
| Resolve a user to an account id | `lookupJiraAccountId` |
| Create an issue | `createJiraIssue` |
| Update fields on an issue | `editJiraIssue` |
| Read an issue (+ comments) | `getJiraIssue` |
| Search issues | `searchJiraIssuesUsingJql` |
| Add / update a comment | `addCommentToJiraIssue` |
| List valid status transitions | `getTransitionsForJiraIssue` |
| Apply a status transition | `transitionJiraIssue` |

**Three rules that apply to every call:**

1. **Always pass `cloudId`** — obtained in the connectivity step. It is required by every tool.
2. **Always send rich text as markdown.** Set `contentFormat: "markdown"` on `createJiraIssue` / `addCommentToJiraIssue`, and `responseContentFormat: "markdown"` on reads. Do **not** hand-build Atlassian Document Format (ADF) JSON — the MCP converts markdown for you.
3. **Sub-tasks need a `parent`.** Create a native sub-task with `createJiraIssue` using the project's sub-task issue type plus `parent: <PARENT-KEY>`. The type name varies by project ("Sub-task", "Subtarefa", …), so discover the exact name and its required fields with `getJiraProjectIssueTypesMetadata` / `getJiraIssueTypeMetaWithFields` before creating.

## Connectivity & scope (Phase 0 completion)

1. Call `getAccessibleAtlassianResources`. Success with ≥1 site → capture each site's `id` (the `cloudId`) and `url`. Empty, error, or tool not present → not configured; guide *MCP setup* below.
2. If exactly one site, use it. If several, ask via **AskUserQuestion** which site to use, listing them by `url`. A site URL is also accepted in `cloudId`, but prefer the resolved UUID.

## The Jira model (quick map)

- **Hierarchy:** Epic → Story/Task/Bug → Sub-task. Projects own issues; sprints schedule them.
- **Native child issue:** the **Sub-task** — the parent must be a **Story or Task**. An **Epic** takes child Stories, not sub-tasks: when refining an Epic, stop and offer to pick/create a Story under it instead.
- **Create-call shape:** `createJiraIssue` with `cloudId`, `projectKey`, `issueTypeName`, `summary`, `description`, `contentFormat: "markdown"`; everything without a dedicated parameter (priority, labels, components, fixVersions, custom fields) goes inside `additional_fields`. Examples: `{"priority": {"name": "High"}}`, `{"labels": ["bug"]}`, `{"components": [{"name": "Backend"}]}`, `{"customfield_10001": "value"}`.
- **Required-field discovery is mandatory before creating:** `getJiraIssueTypeMetaWithFields` for the chosen project + type. `summary` is always required; company-managed and team-managed projects differ on the rest, so never skip it.
- **Assignees:** resolve with `lookupJiraAccountId` and use the returned `assignee_account_id`; don't guess account ids.
- **Comments:** `addCommentToJiraIssue` with `commentBody`, `contentFormat: "markdown"`; restricted visibility via `commentVisibility` (`group`/`role`); edit an existing comment by passing its `commentId`.
- **Transitions:** `getTransitionsForJiraIssue` lists only what's available from the current status; apply with `transitionJiraIssue` using the transition `id` (set the resolution if the workflow asks for one).
- **Search:** JQL, e.g. `project = PROJ AND status = "In Progress" AND assignee = currentUser() ORDER BY updated DESC`; children of a parent via `parent = <KEY> ORDER BY key`.
- **Reads:** `getJiraIssue` with `responseContentFormat: "markdown"`; add `"comment"` to `fields` for comments, `"*all"` for custom fields.

## GMUD (change-management record) — Jira-specific flow

A **GMUD** (Gestão de Mudança / change-management record) governs a controlled change to a production service — a common ITIL-style process, not specific to any one company. Route here whenever the user says GMUD, *mudança*, *change*, or talks about an execution window (*janela*), rollback, or CAB/CCM approval. It follows the org's change template (from *Project conventions*) or the default template below; either way the **rollback plan is non-negotiable** — never create or finalize a GMUD without one.

### GMUD template (mandatory)

Render every section in the user's language. For an **emergencial** change you may abbreviate — keep at least *justification*, *risk*, *implementation*, and *rollback* — but never drop rollback.

1. **Identificação e tipo** — title and change type: **Padrão** (pre-approved, low risk), **Normal**, or **Emergencial**.
2. **Justificativa e objetivo** — why the change is needed and what problem it solves.
3. **Descrição da mudança** — concretely what changes (technical detail).
4. **Serviços e itens de configuração afetados** — impacted systems, services, and CIs.
5. **Risco e impacto** — probability × impact assessment, plus expected downtime/unavailability.
6. **Janela de execução** — start and end date/time.
7. **Plano de implementação** — ordered steps, with the responsible person.
8. **Plano de rollback / contingência** — activation criteria (what must fail to revert), reversion steps, and the maximum time to decide. **Required.**
9. **Plano de testes / validação** — post-implementation checks confirming the service is healthy.
10. **Comunicação e envolvidos** — stakeholders, executor, and the communication plan.
11. **Aprovações** — technical approver and CAB/CCM.

### Create a GMUD

1. **Project** — discover/confirm via `getVisibleJiraProjects`.
2. **Issue type** — `getJiraProjectIssueTypesMetadata`; prefer a change-type ("Mudança", "Change", "GMUD"). If none exists, fall back to Task/Story, tell the user, and carry the whole GMUD template in the description.
3. **Required-field discovery (mandatory)** — `getJiraIssueTypeMetaWithFields`. Change types often expose native fields for window dates, risk, impact, or approver; map the matching template sections onto those fields and put the rest in the description.
4. **Gather content** — fill **every** mandatory template section. Refuse to proceed past a missing rollback plan.
5. **Confirm & create** — preview all sections and fields; on explicit yes, `createJiraIssue` with `contentFormat: "markdown"` and extras in `additional_fields`.
6. **Report** — issue key and browse URL.

### Edit a GMUD

1. **Locate & read** — by key, or search then confirm; read it first so edits land in context.
2. **Pick the change** — update template sections (window, rollback, implementation, risk), other fields, status, or add a note.
3. **Apply** — each is an outward-facing write, so confirm first: field/description edits → `editJiraIssue`; status move → `getTransitionsForJiraIssue` then `transitionJiraIssue`; note → `addCommentToJiraIssue`.
4. **Re-validate** — after editing, confirm the rollback plan is still present and the window is coherent.
5. **Report** — what changed and the browse URL.

## MCP setup (when the probe fails)

Tell the user the official Atlassian MCP isn't connected and walk them through the two steps — **add** the server, then **authorize** it.

**Step 1 — Add the server (run in their terminal):**

```bash
claude mcp add --transport http atlassian https://mcp.atlassian.com/v1/mcp
```

**Step 2 — Authorize the connection.** Adding the server does not authenticate it; the user must complete a browser-based **OAuth 2.1** flow (dynamic client registration — no manual OAuth app needed). Inside Claude Code:

1. Run the `/mcp` command.
2. Select the `atlassian` server (it shows as needing authentication).
3. Choose **Authenticate** / **Login** — a browser opens for the Atlassian OAuth flow.
4. Approve access; the browser confirms and the server status flips to connected.

Access respects the user's existing Atlassian permissions.

- **Prerequisites:** an Atlassian Cloud site with Jira, and a browser to authorize.
- For other MCP clients, point them to the same endpoint (`https://mcp.atlassian.com/v1/mcp`) and follow that client's authorization flow.

After they add **and** authorize it, re-run the probe.

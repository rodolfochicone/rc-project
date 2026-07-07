# rc rules

Rules are the **WHAT** layer of rc's enforcement model. They are short, declarative
statements of the patterns and constraints code must follow, scoped to the files
they apply to by a path glob. They are installed into the harnesses that support
path-scoped rules and are otherwise referenced from `CLAUDE.md` / `AGENTS.md`.

## The three layers: WHAT / HOW / FORCE

rc separates three things that are easy to conflate. Keeping them apart keeps the
always-loaded context small and makes enforcement deterministic where it matters.

| Layer      | Question it answers | Form                       | Applied                                       | Enforcement                               |
| ---------- | ------------------- | -------------------------- | --------------------------------------------- | ----------------------------------------- |
| **Rules**  | _What_ must hold    | `rules/*.md` (glob-scoped) | injected when an edited file matches the glob | awareness — the model sees them           |
| **Skills** | _How_ to do a thing | `skills/*/SKILL.md`        | invoked by trigger/description                | the model chooses to follow the procedure |
| **Hooks**  | _Force_ a guarantee | `hooks/scripts/*.sh`       | fired on a tool event                         | deterministic — `exit 2` blocks the call  |

Rule of thumb:

- A recurring **pattern** ("wrap errors with `%w`", "table-driven tests") → a **rule**. Cheap to state, applies broadly, costs context only when relevant.
- A multi-step **procedure** ("create a PRD", "run the review loop") → a **skill**. Loads its body only when activated.
- A guarantee that **must not depend on the model remembering** ("never edit `go.mod` by hand", "never force-push") → a **hook**. The tool call is physically blocked.

The layers feed each other: a pattern repeated across several skills should be
distilled into a rule (so it is stated once and loaded on demand); a rule whose
violation is costly should be promoted to a hook (so it is enforced, not merely
hoped for). `go.mod` hygiene is the worked example — it is a rule (`rules/go.md`),
a skill instruction (`rc-execute-task`'s laziness ladder), **and** a hook
(`go-mod-guard.sh`), because the cost of getting it wrong justifies all three.

## File format

Each rule file is markdown. Files that apply only to certain paths declare a glob
in YAML frontmatter; files with no `paths` are project-wide.

```yaml
---
paths: ["**/*.go", "**/go.mod"]
---
# Go rules
- Wrap errors with `fmt.Errorf("context: %w", err)`; match with `errors.Is/As`.
```

Keep rules **declarative and short** — they state the WHAT. The HOW lives in
skills; the FORCE lives in hooks. Do not paste procedures or long rationale here.

## Distribution

These rules are installed into each harness in the form that harness understands
(e.g. Cursor `.cursor/rules/`, opencode instruction files). Harnesses without a
native path-scoped rules mechanism rely on `CLAUDE.md` / `AGENTS.md`,
which reference this directory as the source of truth. The source lives here once;
only the per-harness shape is adapted at install time.

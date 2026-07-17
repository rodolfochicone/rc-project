---
name: rc-deslop
description: Removes AI-generated slop from the branch diff — unearned comments, defensive try/catch on trusted paths, `any` casts that only silence the type-checker, and nesting that should be an early return. Use before claiming a coding task complete, before a commit or PR, or when the user asks for slop cleanup. Do not use to hunt correctness or security defects (rc-code-review), to report on over-engineering without editing (rc-simplify-review), or to fix existing review issues (rc-fix-reviews).
model: sonnet
effort: medium
metadata:
  author: Pedro Nauck
  github: https://github.com/pedronauck
  repository: https://github.com/pedronauck/skills
  adapted-by: RC — renamed from deslop, scoped against rc-simplify-review/rc-code-review
---

# Remove AI code slop

Check the diff against main and remove AI-generated slop introduced in the branch. This edits;
it is the cheap sweep that runs before every commit. `rc-simplify-review` is the expensive
read-only counterpart that reports on over-engineering — reach for that one when the question is
"did this need to exist?", not "is this written like a human wrote it?". Keep behavior unchanged
unless fixing a clear bug, and prefer minimal, focused edits over broad rewrites.

## Focus Areas

- Extra comments that are unnecessary or inconsistent with local style — this repo's convention is
  **no comments in code**; keep only pre-existing ones that already match the surrounding file.
  Never strip a comment that states a constraint the code cannot show (a safety invariant, a
  workaround with its reason, a `ponytail:` marker naming a deliberate ceiling).
- Defensive checks or try/catch blocks that are abnormal for trusted code paths. Never remove
  input validation at a trust boundary, error handling that prevents data loss, or a security
  control — those are not slop, however defensive they read.
- Casts to `any` used only to bypass type issues
- Deeply nested code that should be simplified with early returns
- Other patterns inconsistent with the file and surrounding codebase

## Guardrails

- Keep the final summary concise (1-3 sentences).

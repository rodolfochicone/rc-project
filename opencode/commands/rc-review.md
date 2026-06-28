---
description: rc — independent code review (Qwen3.7 Max, high)
agent: rc-review
---

Review the implementation independently and critically: $ARGUMENTS

If the change has a subjective-quality surface (UI/UX, CLI ergonomics, user-facing copy), first run the `rc-gan` skill on that surface to drive it up to a threshold; skip — and say so — for backend/library-only work.

Then use `rc-code-review` (diff) or `rc-review-round` (PRD-scoped). Flag findings with file:line and why they matter; do not fix.

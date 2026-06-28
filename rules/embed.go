// Package rules holds the bundled rc rules — the declarative "WHAT" layer of the
// rules/skills/hooks model (see README.md). Rules are path-scoped patterns that
// code must follow. They are the source of truth referenced by CLAUDE.md /
// AGENTS.md for harnesses without a native per-glob rules mechanism, and can be
// installed natively into harnesses that do support one (e.g. Cursor's
// .cursor/rules/).
package rules

import "embed"

// FS holds every bundled rule file (README.md plus the *.md rule files, each of
// which may declare a `paths:` glob in its frontmatter).
//
//go:embed *.md
var FS embed.FS

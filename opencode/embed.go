package opencode

import "embed"

// FS holds the bundled OpenCode agents and commands installed by `rc setup` for
// the opencode agent. Each phase agent pins a model and reasoning effort, and the
// commands route to those agents so skills run on the intended model.
//
//go:embed agent commands
var FS embed.FS

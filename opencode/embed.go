package opencode

import "embed"

// FS holds the bundled OpenCode agents, commands, and the rc-hooks plugin
// installed by `rc setup` for the opencode agent. Each phase agent pins a model
// and reasoning effort, the commands route to those agents so skills run on the
// intended model, and the plugin provides Claude-parity hook enforcement by
// shelling out to the shared hook scripts.
//
//go:embed agent commands plugin
var FS embed.FS

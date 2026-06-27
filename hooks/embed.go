// Package hooks holds the Claude Code hook scripts bundled into the rc binary.
//
// The same scripts are loaded in place by the Claude Code plugin (via
// hooks/hooks.json) and copied into a project's .claude/rc/hooks/scripts
// directory by `rc setup` for users who install through that channel instead of
// the plugin marketplace.
package hooks

import "embed"

// ScriptsFS holds every bundled hook script as scripts/<name>.sh.
//
//go:embed scripts/*.sh
var ScriptsFS embed.FS

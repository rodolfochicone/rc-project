// Package commands holds the Claude Code slash commands bundled into the rc binary.
package commands

import "embed"

// FS holds every bundled Claude Code slash command as a flat <name>.md file.
//
//go:embed *.md
var FS embed.FS

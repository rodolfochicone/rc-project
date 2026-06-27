package setup

import (
	"fmt"
	"os"
	"path/filepath"
)

// commandsProjectInstallDir is the project-scope location Claude Code reads slash
// commands from. Global scope uses <claudeConfigDir>/commands so it tracks wherever
// Claude Code skills install (honoring the ClaudeConfigDir override).
const commandsProjectInstallDir = ".claude/commands"

func commandsInstallRoot(env resolvedEnvironment, global bool) string {
	if global {
		return filepath.Join(env.claudeConfigDir, "commands")
	}
	return filepath.Join(env.cwd, commandsProjectInstallDir)
}

// PreviewCommandInstall resolves the on-disk install plan for the given commands.
func PreviewCommandInstall(cfg CommandInstallConfig) ([]CommandPreviewItem, error) {
	env, err := resolveEnvironment(cfg.ResolverOptions)
	if err != nil {
		return nil, err
	}

	root := commandsInstallRoot(env, cfg.Global)
	items := make([]CommandPreviewItem, 0, len(cfg.Commands))
	for i := range cfg.Commands {
		targetPath := filepath.Join(root, cfg.Commands[i].FileName)
		items = append(items, CommandPreviewItem{
			Command:       cfg.Commands[i],
			TargetPath:    targetPath,
			WillOverwrite: pathExists(targetPath),
		})
	}
	return items, nil
}

// InstallCommands materializes the given commands into the selected setup scope.
func InstallCommands(cfg CommandInstallConfig) ([]CommandSuccessItem, []CommandFailureItem, error) {
	env, err := resolveEnvironment(cfg.ResolverOptions)
	if err != nil {
		return nil, nil, err
	}

	root := commandsInstallRoot(env, cfg.Global)
	successes := make([]CommandSuccessItem, 0, len(cfg.Commands))
	failures := make([]CommandFailureItem, 0)
	for i := range cfg.Commands {
		success, failure := installCommand(root, cfg.Commands[i])
		if failure != nil {
			failures = append(failures, *failure)
			continue
		}
		successes = append(successes, *success)
	}
	return successes, failures, nil
}

func installCommand(root string, command Command) (*CommandSuccessItem, *CommandFailureItem) {
	targetPath := filepath.Join(root, command.FileName)
	if !isPathSafe(root, targetPath) {
		return nil, commandFailure(command, targetPath, fmt.Errorf("resolved target escapes %s", root))
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, commandFailure(command, targetPath, fmt.Errorf("prepare commands install root %s: %w", root, err))
	}

	// Reuse the shared bundle copier (it writes via os.OpenFile, matching how skills
	// and reusable agents are materialized). WalkDir on a single flat file copies just
	// that file to targetPath.
	if err := copyBundleDirectory(command.SourceFS, command.FileName, targetPath, "command"); err != nil {
		return nil, commandFailure(command, targetPath, err)
	}

	return &CommandSuccessItem{Command: command, Path: targetPath}, nil
}

func commandFailure(command Command, path string, err error) *CommandFailureItem {
	return &CommandFailureItem{Command: command, Path: path, Error: err.Error()}
}

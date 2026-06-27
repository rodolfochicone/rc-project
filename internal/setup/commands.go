package setup

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/rodolfochicone/rc-project/commands"
	"github.com/rodolfochicone/rc-project/internal/core/frontmatter"
)

type (
	// Command describes a bundled Claude Code slash command. Unlike skills and
	// reusable agents (one directory each), a command is a flat <name>.md file, so
	// the command name is the file basename rather than a frontmatter field.
	Command struct {
		Name        string
		Description string
		Origin      AssetOrigin
		SourceFS    fs.FS
		SourceDir   string
		FileName    string
	}

	// CommandInstallConfig selects the commands and scope for an install.
	CommandInstallConfig struct {
		ResolverOptions

		Commands []Command
		Global   bool
	}

	// CommandPreviewItem is the resolved on-disk plan for a single command.
	CommandPreviewItem struct {
		Command       Command
		TargetPath    string
		WillOverwrite bool
	}

	// CommandSuccessItem reports a command that was installed.
	CommandSuccessItem struct {
		Command Command
		Path    string
	}

	// CommandFailureItem reports a command that failed to install.
	CommandFailureItem struct {
		Command Command
		Path    string
		Error   string
	}
)

// ListCommands enumerates bundled slash commands from the provided bundle. Commands
// are flat <name>.md files at the bundle root (not directories), so the command name
// is the file basename. A README.md, if present, is ignored.
func ListCommands(bundle fs.FS) ([]Command, error) {
	if bundle == nil {
		return nil, fmt.Errorf("list bundled commands: bundle is nil")
	}

	entries, err := fs.ReadDir(bundle, ".")
	if err != nil {
		return nil, fmt.Errorf("list bundled commands: %w", err)
	}

	cmds := make([]Command, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".md") || name == "README.md" {
			continue
		}

		command, err := parseCommand(bundle, name)
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, command)
	}

	slices.SortFunc(cmds, func(left, right Command) int {
		return strings.Compare(left.Name, right.Name)
	})
	return cmds, nil
}

func parseCommand(bundle fs.FS, fileName string) (Command, error) {
	name := strings.TrimSuffix(fileName, ".md")
	if err := validateReusableAgentName(name); err != nil {
		return Command{}, fmt.Errorf("read bundled command %q: %w", fileName, err)
	}

	content, err := fs.ReadFile(bundle, fileName)
	if err != nil {
		return Command{}, fmt.Errorf("read bundled command %q: %w", fileName, err)
	}

	var metadata struct {
		Description string `yaml:"description"`
	}
	if _, err := frontmatter.Parse(string(content), &metadata); err != nil {
		return Command{}, fmt.Errorf("read bundled command %q: %w", fileName, err)
	}
	if strings.TrimSpace(metadata.Description) == "" {
		return Command{}, fmt.Errorf("read bundled command %q: missing description", fileName)
	}

	return Command{
		Name:        name,
		Description: strings.TrimSpace(metadata.Description),
		Origin:      AssetOriginBundled,
		SourceFS:    bundle,
		SourceDir:   ".",
		FileName:    fileName,
	}, nil
}

// SelectCommands filters a discovered catalog down to the requested command names.
func SelectCommands(all []Command, names []string) ([]Command, error) {
	return selectByName(all, names, selectByNameConfig[Command]{
		subject:      "bundled commands",
		emptyLabel:   "commands",
		invalidLabel: "command(s)",
		getName: func(command Command) string {
			return command.Name
		},
		normalize: strings.TrimSpace,
		less: func(left, right Command) int {
			return strings.Compare(left.Name, right.Name)
		},
	})
}

// ListBundledCommands returns the slash commands bundled into the rc binary.
func ListBundledCommands() ([]Command, error) {
	return ListCommands(commands.FS)
}

// PreviewBundledCommandInstall resolves the on-disk install plan for bundled commands.
func PreviewBundledCommandInstall(cfg CommandInstallConfig) ([]CommandPreviewItem, error) {
	cmds, err := ListBundledCommands()
	if err != nil {
		return nil, err
	}
	cfg.Commands = cmds
	return PreviewCommandInstall(cfg)
}

// InstallBundledCommands installs every bundled command into the selected scope.
func InstallBundledCommands(cfg CommandInstallConfig) ([]CommandSuccessItem, []CommandFailureItem, error) {
	cmds, err := ListBundledCommands()
	if err != nil {
		return nil, nil, err
	}
	cfg.Commands = cmds
	return InstallCommands(cfg)
}

package cli

import (
	"log/slog"
	"strings"
	"sync"

	extcli "github.com/rodolfochicone/rc-project/internal/cli/extension"
	"github.com/rodolfochicone/rc-project/internal/core/agent"

	// Register the extension-backed run-scope factory used by kernel dispatchers.
	_ "github.com/rodolfochicone/rc-project/internal/core/extension"
	"github.com/rodolfochicone/rc-project/internal/core/kernel"
	"github.com/rodolfochicone/rc-project/internal/core/workspace"
	"github.com/rodolfochicone/rc-project/pkg/rc/events"
	"github.com/spf13/cobra"
)

type commandKind string

const (
	commandKindFetchReviews commandKind = "reviews fetch"
	commandKindFixReviews   commandKind = "reviews fix"
	commandKindWatchReviews commandKind = "reviews watch"
	commandKindExec         commandKind = "exec"
	commandKindArchive      commandKind = "archive"
	commandKindTasksRun     commandKind = "tasks run"
	commandKindSync         commandKind = "sync"
)

var validateRootDispatcher = kernel.ValidateDefaultRegistry

func newRootDispatcher() *kernel.Dispatcher {
	deps := kernel.KernelDeps{
		Logger:        slog.Default(),
		EventBus:      events.New[events.Event](0),
		Workspace:     workspace.Context{},
		AgentRegistry: agent.DefaultRegistry(),
	}

	dispatcher := kernel.BuildDefault(deps)
	if err := validateRootDispatcher(dispatcher); err != nil {
		slog.Default().Error("kernel dispatcher validation failed", "error", err)
	}
	return dispatcher
}

func newLazyRootDispatcher() func() *kernel.Dispatcher {
	return sync.OnceValue(newRootDispatcher)
}

// NewRootCommand returns the reusable rc Cobra command.
func NewRootCommand() *cobra.Command {
	return newRootCommandWithDefaults(newLazyRootDispatcher(), defaultCommandStateDefaults())
}

func newRootCommandWithDefaults(dispatcher func() *kernel.Dispatcher, defaults commandStateDefaults) *cobra.Command {
	root := &cobra.Command{
		Use:          "rc",
		Short:        "Run AI review remediation and PRD task workflows",
		SilenceUsage: true,
		Long: `rc manages review rounds and PRD execution workflows.

Defaults can be stored in ~/.rc/config.toml and overridden per workspace in
.rc/config.toml. Explicit CLI flags always override values loaded from config files.

Use explicit workflow subcommands:
  rc setup         Install bundled public skills for supported agents
  rc init          Scaffold a new project from the rodolfochicone TypeScript template
  rc add           Install individual rc skills into selected agents
  rc agents        Discover and inspect reusable agents
  rc upgrade       Update the CLI to the latest release
  rc ext           Manage bundled, user, and workspace extensions
  rc migrate       Convert legacy workflow artifacts to frontmatter
  rc daemon        Manage the home-scoped daemon bootstrap lifecycle
  rc workspaces    Inspect daemon workspace registrations
  rc tasks         Inspect, validate, and run task workflows
  rc reviews       Fetch, inspect, and remediate review workflows
  rc runs          Inspect and clean persisted daemon run artifacts
  rc sync          Reconcile workflow artifacts into global.db
  rc archive       Move fully completed workflows into .rc/tasks/_archived/
  rc exec          Execute one ad hoc prompt through the shared ACP runtime`,
		// Bare `rc` (no subcommand) prints the standard command help and exits
		// cleanly — no banner, splash, or interactive menu.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		// Thread the global --workspace override into the command context so
		// workspace discovery resolves the requested .rc directory without every
		// call site needing the flag. Empty leaves discovery on its default
		// (walk up from the current directory, then down for a single .rc).
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			override, err := cmd.Flags().GetString("workspace")
			if err != nil {
				return err
			}
			if strings.TrimSpace(override) != "" {
				cmd.SetContext(workspace.WithStartDirOverride(cmd.Context(), override))
			}
			return nil
		},
	}

	root.PersistentFlags().String(
		"workspace",
		"",
		"Directory to resolve the .rc workspace from (default: current directory)",
	)

	root.AddCommand(
		newSetupCommand(nil),
		newInstallCommand(nil),
		newInitCommand(),
		newAddCommand(nil),
		newAgentsCommand(),
		newUpgradeCommand(),
		extcli.NewExtCommand(nil),
		newMigrateCommand(dispatcher),
		newDaemonCommand(),
		newWorkspacesCommand(),
		newTasksCommand(nil, defaults),
		newReviewsCommandWithDefaults(defaults),
		newRunsCommandWithDefaults(defaults),
		newSyncCommand(dispatcher),
		newArchiveCommand(dispatcher),
		newExecCommandWithDefaults(defaults),
		newMemoryCommand(),
		newMCPServeCommand(),
	)
	return root
}

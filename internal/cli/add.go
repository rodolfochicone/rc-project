package cli

import (
	"errors"
	"fmt"

	"github.com/rodolfochicone/rc-project/internal/core/kernel"
	"github.com/rodolfochicone/rc-project/internal/setup"
	"github.com/spf13/cobra"
)

// newAddCommand groups commands that install individual rc assets into a
// project (or user scope) that already uses rc setup.
func newAddCommand(_ *kernel.Dispatcher) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "add",
		Short:        "Add individual rc assets to selected agents",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newAddSkillCommand())
	cmd.AddCommand(newAddCommandCommand())
	return cmd
}

func newAddCommandCommand() *cobra.Command {
	state := newSetupCommandState()
	cmd := &cobra.Command{
		Use:   "command <name>...",
		Short: "Install one or more rc Claude Code slash commands",
		Long: `Install specific rc Claude Code slash commands without running the full setup flow.

Commands are written to .claude/commands/ (project) or ~/.claude/commands/ (with --global).`,
		Example: `  rc add command rc-pipe --yes
  rc add command rc-plan rc-review --global`,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			state.commandNames = args
			return state.runAddCommand(cmd)
		},
	}
	cmd.Flags().BoolVarP(&state.global, "global", "g", false, "Install to the user directory instead of the project")
	cmd.Flags().BoolVarP(&state.yes, "yes", "y", false, "Skip confirmation prompts")
	return cmd
}

// runAddCommand installs an explicit set of slash commands (from positional args)
// into Claude Code's commands directory. It reuses the setup command primitives but
// needs no agent selection — the install target is fixed to .claude/commands/.
func (s *setupCommandState) runAddCommand(cmd *cobra.Command) error {
	resolver := s.resolverOptions()

	if err := s.prepareRunMode(); err != nil {
		return err
	}

	all, err := setup.ListBundledCommands()
	if err != nil {
		return err
	}
	selected, err := setup.SelectCommands(all, s.commandNames)
	if err != nil {
		return err
	}

	previews, err := s.previewCommands(setup.CommandInstallConfig{
		ResolverOptions: resolver,
		Commands:        selected,
		Global:          s.global,
	})
	if err != nil {
		return err
	}

	if !s.yes {
		printCommandPreviewSummary(cmd, previews, s.global)
		confirmed, confirmErr := confirmSetup()
		if confirmErr != nil {
			return confirmErr
		}
		if !confirmed {
			return errors.New("add command canceled")
		}
	}

	successful, failed, err := s.installCommands(setup.CommandInstallConfig{
		ResolverOptions: resolver,
		Commands:        selected,
		Global:          s.global,
	})
	result := &setup.Result{
		Global:             s.global,
		CommandsSuccessful: successful,
		CommandsFailed:     failed,
	}
	printInstallResult(cmd, result)
	if err != nil {
		return err
	}
	if len(failed) > 0 {
		return fmt.Errorf("add command completed with %d failure(s)", len(failed))
	}
	return nil
}

func newAddSkillCommand() *cobra.Command {
	state := newSetupCommandState()
	cmd := &cobra.Command{
		Use:   "skill <name>...",
		Short: "Install one or more rc skills into selected agents",
		Long: `Install specific rc skills without running the full setup flow.

Skills may be bundled rc skills or skills shipped by enabled extensions.
Without --agent, an interactive picker selects the target agents; in
non-interactive mode pass --agent together with --yes.`,
		Example: `  rc add skill rc-git --agent claude --yes
  rc add skill rc-create-prd rc-create-techspec --agent codex --agent claude`,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			state.skillNames = args
			return state.runAddSkill(cmd)
		},
	}
	cmd.Flags().StringSliceVarP(&state.agentNames, "agent", "a", nil, "Target agent/editor name (repeatable)")
	cmd.Flags().BoolVarP(&state.global, "global", "g", false, "Install to the user directory instead of the project")
	cmd.Flags().BoolVar(&state.copy, "copy", false, "Copy files instead of symlinking to agent directories")
	cmd.Flags().BoolVarP(&state.yes, "yes", "y", false, "Skip confirmation prompts")
	return cmd
}

// runAddSkill installs an explicit set of skills (from positional args) into
// the selected agents. It reuses the setup install primitives but installs
// only the requested skills — no reusable agents and no rtk prompt.
func (s *setupCommandState) runAddSkill(cmd *cobra.Command) error {
	ctx, stop := signalCommandContext(cmd)
	defer stop()

	resolver := s.resolverOptions()
	catalog, err := s.loadCatalog(ctx, resolver)
	if err != nil {
		return err
	}

	if err := s.prepareRunMode(); err != nil {
		return err
	}

	selectedSkills, err := setup.SelectSkills(catalog.Skills, s.skillNames)
	if err != nil {
		return err
	}

	supportedAgents, detectedAgents, err := s.loadAgents(resolver)
	if err != nil {
		return err
	}
	selectedAgents, err := s.resolveAgentSelection(supportedAgents, detectedAgents)
	if err != nil {
		return err
	}
	globalScope, err := s.resolveScope(cmd, selectedAgents)
	if err != nil {
		return err
	}
	mode, err := s.resolveInstallMode(cmd, supportedAgents, selectedAgents, globalScope)
	if err != nil {
		return err
	}

	previews, err := s.previewSkills(resolver, selectedSkills, selectedAgents, globalScope, mode)
	if err != nil {
		return err
	}

	printSetupWarnings(cmd, catalog.Conflicts)
	if !s.yes {
		printPreviewSummary(cmd, previews, nil, globalScope, mode)
		confirmed, confirmErr := confirmSetup()
		if confirmErr != nil {
			return confirmErr
		}
		if !confirmed {
			return errors.New("add skill canceled")
		}
	}

	successful, failed, err := s.installSkills(
		resolver,
		previewsToSkills(previews),
		selectedAgents,
		globalScope,
		mode,
	)
	result := &setup.Result{
		Global:     globalScope,
		Mode:       mode,
		Successful: successful,
		Failed:     failed,
	}
	printInstallResult(cmd, result)
	if err != nil {
		return err
	}
	if len(failed) > 0 {
		return fmt.Errorf("add skill completed with %d failure(s)", len(failed))
	}
	return nil
}

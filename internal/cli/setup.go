package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/rodolfochicone/rc-project/internal/core/kernel"
	"github.com/rodolfochicone/rc-project/internal/setup"
	"github.com/rodolfochicone/rc-project/internal/version"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type setupCommandState struct {
	agentNames   []string
	skillNames   []string
	commandNames []string
	global       bool
	copy         bool
	list         bool
	yes          bool
	all          bool
	sync         bool

	loadCatalog           func(context.Context, setup.ResolverOptions) (setup.EffectiveCatalog, error)
	listAgents            func(setup.ResolverOptions) ([]setup.Agent, error)
	detectAgents          func(setup.ResolverOptions) ([]setup.Agent, error)
	previewSkills         setupSkillPreviewFunc
	previewReusableAgents func(setup.ReusableAgentInstallConfig) ([]setup.ReusableAgentPreviewItem, error)
	installSkills         setupSkillInstallFunc
	syncSkills            setupSkillSyncFunc
	installReusableAgents setupReusableAgentInstallFunc
	previewCommands       func(setup.CommandInstallConfig) ([]setup.CommandPreviewItem, error)
	installCommands       func(setup.CommandInstallConfig) ([]setup.CommandSuccessItem, []setup.CommandFailureItem, error)
	installOpenCodeAssets setupOpenCodeInstallFunc
	installHooks          setupHookInstallFunc
	cleanupLegacyAssets   func(setup.LegacyAssetCleanupConfig) (setup.LegacyAssetCleanupResult, error)
	isInteractive         func() bool

	detectTool         func(context.Context, string) (setup.ToolStatus, error)
	runToolInstall     func(context.Context, setup.InstallCommand, io.Writer) error
	confirmToolInstall func(label, display string) (bool, error)
	lookPath           func(string) (string, error)
	goos               string
}

type setupSkillPreviewFunc func(
	setup.ResolverOptions,
	[]setup.Skill,
	[]string,
	bool,
	setup.InstallMode,
) ([]setup.PreviewItem, error)

type setupSkillInstallFunc func(
	setup.ResolverOptions,
	[]setup.Skill,
	[]string,
	bool,
	setup.InstallMode,
) ([]setup.SuccessItem, []setup.FailureItem, error)

type setupReusableAgentInstallFunc func(
	setup.ReusableAgentInstallConfig,
) ([]setup.ReusableAgentSuccessItem, []setup.ReusableAgentFailureItem, error)

type setupSkillSyncFunc func(setup.SyncConfig) (setup.SyncResult, error)

type setupOpenCodeInstallFunc func(
	setup.OpenCodeInstallConfig,
) ([]setup.OpenCodeAssetSuccessItem, []setup.OpenCodeAssetFailureItem, error)

type setupHookInstallFunc func(
	setup.HookInstallConfig,
) ([]setup.HookSuccessItem, []setup.HookFailureItem, error)

type setupInstallPlan struct {
	Config               setup.InstallConfig
	Skills               []setup.Skill
	ReusableAgents       []setup.ReusableAgent
	Commands             []setup.Command
	Previews             []setup.PreviewItem
	ReusableAgentPreview []setup.ReusableAgentPreviewItem
	CommandPreview       []setup.CommandPreviewItem
}

func newSetupCommand(_ *kernel.Dispatcher) *cobra.Command {
	state := newSetupCommandState()
	cmd := &cobra.Command{
		Use:          "setup",
		Short:        "Install rc core assets plus setup assets shipped by enabled extensions",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Install rc's core public skills, any additional skills shipped by enabled
extensions, and reusable agents in either the project or user scope selected during setup.

The command can run interactively or entirely from flags.`,
		Example: `  rc setup
  rc setup --list
  rc setup --agent codex --agent claude --skill rc-create-prd --skill rc-create-techspec --yes
  rc setup --all
  rc setup --sync --agent claude-code --yes
  rc setup --agent cursor --global --copy --yes`,
		RunE: state.run,
	}

	cmd.Flags().StringSliceVarP(&state.agentNames, "agent", "a", nil, "Target agent/editor name (repeatable)")
	cmd.Flags().StringSliceVarP(&state.skillNames, "skill", "s", nil, "Setup skill name to install (repeatable)")
	cmd.Flags().
		StringSliceVarP(&state.commandNames, "command", "c", nil, "Slash command name to install (repeatable)")
	cmd.Flags().BoolVarP(&state.global, "global", "g", false, "Install to the user directory instead of the project")
	cmd.Flags().BoolVar(&state.copy, "copy", false, "Copy files instead of symlinking to agent directories")
	cmd.Flags().BoolVarP(&state.list, "list", "l", false, "List setup assets without installing")
	cmd.Flags().BoolVarP(&state.yes, "yes", "y", false, "Skip confirmation prompts")
	cmd.Flags().
		BoolVar(&state.all, "all", false, "Install all available setup skills to all supported agents without prompts")
	cmd.Flags().
		BoolVar(&state.sync, "sync", false, "Update bundled skills the agent already has and add the missing ones")
	return cmd
}

func newSetupCommandState() *setupCommandState {
	return &setupCommandState{
		loadCatalog:           loadEffectiveSetupCatalog,
		listAgents:            setup.SupportedAgents,
		detectAgents:          setup.DetectInstalledAgents,
		previewSkills:         setup.PreviewSelectedSkills,
		previewReusableAgents: setup.PreviewReusableAgentInstall,
		installSkills:         setup.InstallSelectedSkills,
		syncSkills:            setup.SyncBundledSkills,
		installReusableAgents: setup.InstallReusableAgents,
		previewCommands:       setup.PreviewCommandInstall,
		installCommands:       setup.InstallCommands,
		installOpenCodeAssets: setup.InstallBundledOpenCodeAssets,
		installHooks:          setup.InstallBundledHooks,
		cleanupLegacyAssets:   setup.CleanupLegacyTransferredAssets,
		isInteractive:         isInteractiveTerminal,
		detectTool:            setup.DetectTool,
		runToolInstall:        setup.RunInstall,
		confirmToolInstall:    confirmToolInstall,
		lookPath:              exec.LookPath,
		goos:                  runtime.GOOS,
	}
}

func (s *setupCommandState) run(cmd *cobra.Command, _ []string) error {
	ctx, stop := signalCommandContext(cmd)
	defer stop()

	resolver := s.resolverOptions()
	catalog, err := s.loadCatalog(ctx, resolver)
	if err != nil {
		return err
	}
	if s.list {
		printSetupAssets(cmd, catalog.Skills, catalog.ReusableAgents, catalog.Conflicts)
		commandsList, err := setup.ListBundledCommands()
		if err != nil {
			return err
		}
		printCommandAssets(cmd, commandsList)
		return nil
	}

	if err := s.validateSyncFlags(); err != nil {
		return err
	}

	if !s.yes && s.isInteractive() {
		printWelcomeHeader(cmd)
	}

	if err := s.prepareRunMode(); err != nil {
		return err
	}

	if s.sync {
		return s.runSync(cmd, resolver)
	}

	supportedAgents, detectedAgents, err := s.loadAgents(resolver)
	if err != nil {
		return err
	}

	cfg, previews, reusableAgentPreviews, err := s.buildInstallPlan(
		cmd,
		catalog,
		resolver,
		supportedAgents,
		detectedAgents,
	)
	if err != nil {
		return err
	}
	selectedCommands, commandPreviews, err := s.buildCommandPlan(resolver, cfg.Global)
	if err != nil {
		return err
	}

	printSetupWarnings(cmd, catalog.Conflicts)
	if err := s.confirmPlan(cmd, previews, reusableAgentPreviews, commandPreviews, cfg.Global, cfg.Mode); err != nil {
		return err
	}

	if err := s.executeInstall(cmd, setupInstallPlan{
		Config:               cfg,
		Skills:               previewsToSkills(previews),
		ReusableAgents:       append([]setup.ReusableAgent(nil), catalog.ReusableAgents...),
		Commands:             selectedCommands,
		Previews:             previews,
		ReusableAgentPreview: reusableAgentPreviews,
		CommandPreview:       commandPreviews,
	}); err != nil {
		return err
	}

	return s.ensureRTK(ctx, cmd)
}

// validateSyncFlags rejects flag combinations that contradict the sync intent.
// Sync derives its skill set from what each agent already has plus the missing
// bundled skills, so an explicit --skill selection or the install-everything
// --all flag would override that contract.
func (s *setupCommandState) validateSyncFlags() error {
	if !s.sync {
		return nil
	}
	if s.all {
		return errors.New("--sync cannot be combined with --all")
	}
	if len(s.skillNames) > 0 {
		return errors.New("--sync cannot be combined with --skill")
	}
	return nil
}

// runSync reconciles the bundled skills for each selected agent: it updates the
// ones the agent already has and adds the ones it is missing, leaving current
// and non-bundled skills untouched.
func (s *setupCommandState) runSync(cmd *cobra.Command, resolver setup.ResolverOptions) error {
	supportedAgents, detectedAgents, err := s.loadAgents(resolver)
	if err != nil {
		return err
	}
	selectedAgents, err := s.resolveAgentSelection(supportedAgents, detectedAgents)
	if err != nil {
		return err
	}
	global, err := s.resolveScope(cmd, selectedAgents)
	if err != nil {
		return err
	}

	mode := setup.InstallMode("")
	if s.copy {
		mode = setup.InstallModeCopy
	}

	results := make([]setup.SyncResult, 0, len(selectedAgents))
	failureCount := 0
	for _, name := range selectedAgents {
		result, err := s.syncSkills(setup.SyncConfig{
			ResolverOptions: resolver,
			AgentName:       name,
			Global:          global,
			Mode:            mode,
		})
		if err != nil {
			return err
		}
		results = append(results, result)
		failureCount += len(result.Failed)
	}

	printSyncResults(cmd, results)
	if failureCount > 0 {
		return fmt.Errorf("sync completed with %d failure(s)", failureCount)
	}
	return nil
}

func printSyncResults(cmd *cobra.Command, results []setup.SyncResult) {
	if len(results) == 0 {
		return
	}
	styles := newCLIChromeStyles()
	cwd, homeDir := displayRoots()
	w := cmd.OutOrStdout()

	for i := range results {
		result := &results[i]
		lipgloss.Fprintln(w, styles.sectionTitle.Render(fmt.Sprintf(
			"Sync %s (%s scope)",
			result.Agent.DisplayName,
			installScopeLabel(result.Scope),
		)))
		printSyncSuccessGroup(w, styles, "✓ Added", result.Added, cwd, homeDir)
		printSyncSuccessGroup(w, styles, "✓ Updated", result.Updated, cwd, homeDir)
		if len(result.Unchanged) > 0 {
			lipgloss.Fprintf(
				w,
				"  %s  %s\n",
				styles.label.Render("Unchanged"),
				styles.value.Render(fmt.Sprintf("%d already current", len(result.Unchanged))),
			)
		}
		printSyncFailureGroup(w, styles, result.Failed, cwd, homeDir)
		fmt.Fprintln(w)
	}
}

func printSyncSuccessGroup(
	w io.Writer,
	styles cliChromeStyles,
	label string,
	items []setup.SuccessItem,
	cwd, homeDir string,
) {
	if len(items) == 0 {
		return
	}
	lipgloss.Fprintln(w, styles.successHeader.Render(fmt.Sprintf("  %s (%d)", label, len(items))))
	for i := range items {
		item := &items[i]
		icon := styles.successIcon.Render("✓")
		name := styles.skill.Render(item.Skill.Name)
		path := styles.path.Render(shortenPath(item.Path, cwd, homeDir))
		lipgloss.Fprintf(w, "    %s  %s  %s\n", icon, name, path)
	}
}

func printSyncFailureGroup(
	w io.Writer,
	styles cliChromeStyles,
	items []setup.FailureItem,
	cwd, homeDir string,
) {
	if len(items) == 0 {
		return
	}
	lipgloss.Fprintln(w, styles.failureHeader.Render(fmt.Sprintf("  ✗ Failed (%d)", len(items))))
	for i := range items {
		item := &items[i]
		icon := styles.failureIcon.Render("✗")
		name := styles.skill.Render(item.Skill.Name)
		path := styles.path.Render(shortenPath(item.Path, cwd, homeDir))
		lipgloss.Fprintf(w, "    %s  %s  %s\n", icon, name, path)
		lipgloss.Fprintf(w, "       %s\n", styles.errorMessage.Render(item.Error))
	}
}

// ensureRTK detects the rtk CLI and, when missing, surfaces the OS-appropriate
// install command. During setup the installer runs only after an explicit
// interactive confirmation; --yes/--all and non-interactive runs print guidance
// instead of executing a network installer unattended (runUnattended=false).
func (s *setupCommandState) ensureRTK(ctx context.Context, cmd *cobra.Command) error {
	installer := newRTKInstaller()
	installer.yes = s.yes
	installer.runUnattended = false
	installer.goos = s.goos
	installer.isInteractive = s.isInteractive
	installer.detectTool = s.detectTool
	installer.runInstall = s.runToolInstall
	installer.confirm = s.confirmToolInstall
	installer.lookPath = s.lookPath
	return installer.ensure(ctx, cmd.OutOrStdout())
}

func (s *setupCommandState) prepareRunMode() error {
	if s.all {
		s.yes = true
	}
	if !s.yes && !s.isInteractive() {
		return errors.New("rc setup requires an interactive terminal unless --yes is provided")
	}
	return nil
}

func (s *setupCommandState) resolverOptions() setup.ResolverOptions {
	return currentResolverOptions()
}

func currentResolverOptions() setup.ResolverOptions {
	return setup.ResolverOptions{
		CodeXHome:       strings.TrimSpace(os.Getenv("CODEX_HOME")),
		ClaudeConfigDir: strings.TrimSpace(os.Getenv("CLAUDE_CONFIG_DIR")),
		XDGConfigHome:   strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")),
	}
}

func (s *setupCommandState) loadAgents(resolver setup.ResolverOptions) ([]setup.Agent, []setup.Agent, error) {
	supportedAgents, err := s.listAgents(resolver)
	if err != nil {
		return nil, nil, err
	}

	detectedAgents, err := s.detectAgents(resolver)
	if err != nil {
		return nil, nil, err
	}
	return supportedAgents, detectedAgents, nil
}

func (s *setupCommandState) buildInstallPlan(
	cmd *cobra.Command,
	catalog setup.EffectiveCatalog,
	resolver setup.ResolverOptions,
	supportedAgents []setup.Agent,
	detectedAgents []setup.Agent,
) (setup.InstallConfig, []setup.PreviewItem, []setup.ReusableAgentPreviewItem, error) {
	selectedSkillNames, err := s.resolveSkillSelection(catalog.Skills)
	if err != nil {
		return setup.InstallConfig{}, nil, nil, err
	}
	selectedSkills, err := setup.SelectSkills(catalog.Skills, selectedSkillNames)
	if err != nil {
		return setup.InstallConfig{}, nil, nil, err
	}

	selectedAgents, err := s.resolveAgentSelection(supportedAgents, detectedAgents)
	if err != nil {
		return setup.InstallConfig{}, nil, nil, err
	}

	globalScope, err := s.resolveScope(cmd, selectedAgents)
	if err != nil {
		return setup.InstallConfig{}, nil, nil, err
	}

	mode, err := s.resolveInstallMode(cmd, supportedAgents, selectedAgents, globalScope)
	if err != nil {
		return setup.InstallConfig{}, nil, nil, err
	}

	cfg := setup.InstallConfig{
		ResolverOptions: resolver,
		SkillNames:      selectedSkillNames,
		AgentNames:      selectedAgents,
		Global:          globalScope,
		Mode:            mode,
	}
	previews, err := s.previewSkills(resolver, selectedSkills, selectedAgents, globalScope, mode)
	if err != nil {
		return setup.InstallConfig{}, nil, nil, err
	}
	reusableAgentPreviews, err := s.previewReusableAgents(setup.ReusableAgentInstallConfig{
		ResolverOptions: resolver,
		ReusableAgents:  catalog.ReusableAgents,
		Global:          globalScope,
	})
	if err != nil {
		return setup.InstallConfig{}, nil, nil, err
	}
	return cfg, previews, reusableAgentPreviews, nil
}

func (s *setupCommandState) confirmPlan(
	cmd *cobra.Command,
	previews []setup.PreviewItem,
	reusableAgentPreviews []setup.ReusableAgentPreviewItem,
	commandPreviews []setup.CommandPreviewItem,
	global bool,
	mode setup.InstallMode,
) error {
	if s.yes {
		return nil
	}

	printPreviewSummary(cmd, previews, reusableAgentPreviews, global, mode)
	printCommandPreviewSummary(cmd, commandPreviews, global)
	confirmed, err := confirmSetup()
	if err != nil {
		return err
	}
	if !confirmed {
		return errors.New("setup canceled")
	}
	return nil
}

func (s *setupCommandState) executeInstall(cmd *cobra.Command, plan setupInstallPlan) error {
	result, err := s.installPlan(plan)
	printInstallResult(cmd, result)
	if err != nil {
		return err
	}
	failureCount := len(
		result.Failed,
	) + len(
		result.ReusableAgentsFailed,
	) + len(
		result.CommandsFailed,
	) + len(
		result.HooksFailed,
	) + len(
		result.OpenCodeFailed,
	)
	if failureCount > 0 {
		return fmt.Errorf("setup completed with %d failure(s)", failureCount)
	}
	return nil
}

func (s *setupCommandState) installPlan(plan setupInstallPlan) (*setup.Result, error) {
	if s.cleanupLegacyAssets != nil {
		if _, err := s.cleanupLegacyAssets(setup.LegacyAssetCleanupConfig{
			ResolverOptions: plan.Config.ResolverOptions,
			Global:          plan.Config.Global,
		}); err != nil {
			return nil, fmt.Errorf("cleanup legacy setup assets: %w", err)
		}
	}

	successful, failed, err := s.installSkills(
		plan.Config.ResolverOptions,
		plan.Skills,
		plan.Config.AgentNames,
		plan.Config.Global,
		plan.Config.Mode,
	)
	if err != nil {
		return nil, err
	}

	result := &setup.Result{
		Global:     plan.Config.Global,
		Mode:       plan.Config.Mode,
		Successful: successful,
		Failed:     failed,
	}

	successfulReusableAgents, failedReusableAgents, err := s.installReusableAgents(setup.ReusableAgentInstallConfig{
		ResolverOptions: plan.Config.ResolverOptions,
		ReusableAgents:  plan.ReusableAgents,
		Global:          plan.Config.Global,
	})
	if err != nil {
		return result, fmt.Errorf("install reusable agents: %w", err)
	}
	result.ReusableAgentsSuccessful = successfulReusableAgents
	result.ReusableAgentsFailed = failedReusableAgents

	successfulCommands, failedCommands, err := s.installCommands(setup.CommandInstallConfig{
		ResolverOptions: plan.Config.ResolverOptions,
		Commands:        plan.Commands,
		Global:          plan.Config.Global,
	})
	if err != nil {
		return result, fmt.Errorf("install commands: %w", err)
	}
	result.CommandsSuccessful = successfulCommands
	result.CommandsFailed = failedCommands

	// Hooks are Claude-channel only; install them only when Claude is a target.
	if s.installHooks != nil && agentSelected(plan.Config.AgentNames, "claude", "claude-code") {
		hookSuccessful, hookFailed, err := s.installHooks(setup.HookInstallConfig{
			ResolverOptions: plan.Config.ResolverOptions,
			Global:          plan.Config.Global,
		})
		if err != nil {
			return result, fmt.Errorf("install hooks: %w", err)
		}
		result.HooksSuccessful = hookSuccessful
		result.HooksFailed = hookFailed
	}

	if agentSelected(plan.Config.AgentNames, setup.OpenCodeAgentName) {
		ocSuccessful, ocFailed, err := s.installOpenCodeAssets(setup.OpenCodeInstallConfig{
			ResolverOptions: plan.Config.ResolverOptions,
			Global:          plan.Config.Global,
		})
		if err != nil {
			return result, fmt.Errorf("install opencode assets: %w", err)
		}
		result.OpenCodeSuccessful = ocSuccessful
		result.OpenCodeFailed = ocFailed
	}
	return result, nil
}

// agentSelected reports whether any of the target agent names is in names.
func agentSelected(names []string, targets ...string) bool {
	for _, name := range names {
		for _, target := range targets {
			if name == target {
				return true
			}
		}
	}
	return false
}

func (s *setupCommandState) buildCommandPlan(
	resolver setup.ResolverOptions,
	global bool,
) ([]setup.Command, []setup.CommandPreviewItem, error) {
	all, err := setup.ListBundledCommands()
	if err != nil {
		return nil, nil, err
	}
	selected := all
	if len(s.commandNames) > 0 {
		selected, err = setup.SelectCommands(all, s.commandNames)
		if err != nil {
			return nil, nil, err
		}
	}
	previews, err := s.previewCommands(setup.CommandInstallConfig{
		ResolverOptions: resolver,
		Commands:        selected,
		Global:          global,
	})
	if err != nil {
		return nil, nil, err
	}
	return selected, previews, nil
}

func (s *setupCommandState) resolveSkillSelection(skills []setup.Skill) ([]string, error) {
	if len(s.skillNames) > 0 {
		return append([]string(nil), s.skillNames...), nil
	}
	if s.all || s.yes {
		return skillNames(skills), nil
	}

	selected := skillNames(skills)

	maxNameLen := 0
	for i := range skills {
		if len(skills[i].Name) > maxNameLen {
			maxNameLen = len(skills[i].Name)
		}
	}

	// The picker is wrapped in a full-width border (boxedHuhTheme). Size the
	// form to the terminal minus the box frame so the border closes, and
	// truncate each option to one line (the multi-select would otherwise let
	// the terminal soft-wrap long descriptions). The form renders to stderr.
	formWidth, maxLabelWidth := 0, 0
	if width := terminalWidth(os.Stderr); width > 0 {
		formWidth = width - skillFormFrameWidth
		maxLabelWidth = formWidth - skillOptionChromeWidth
	}

	options := make([]huh.Option[string], 0, len(skills))
	for i := range skills {
		label := fmt.Sprintf("%-*s  %s", maxNameLen, skills[i].Name, skills[i].Description)
		options = append(options, huh.NewOption(truncateToWidth(label, maxLabelWidth), skills[i].Name))
	}

	field := huh.NewMultiSelect[string]().
		Key("skills").
		Title("Setup Skills").
		Description("Select the rc skills to install, including enabled extension assets").
		Options(options...).
		Value(&selected).
		Limit(len(skills)).
		Validate(func(values []string) error {
			if len(values) == 0 {
				return errors.New("select at least one skill")
			}
			return nil
		})
	form := huh.NewForm(huh.NewGroup(field)).WithTheme(boxedHuhTheme())
	if formWidth > 0 {
		form = form.WithWidth(formWidth)
	}
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("select setup skills: %w", err)
	}
	return selected, nil
}

func (s *setupCommandState) resolveAgentSelection(
	supported []setup.Agent,
	detected []setup.Agent,
) ([]string, error) {
	if len(s.agentNames) > 0 {
		return append([]string(nil), s.agentNames...), nil
	}
	if s.all {
		return agentNames(supported), nil
	}
	if s.yes {
		if len(detected) == 0 {
			return nil, errors.New("no agents detected; rerun with --agent or use interactive mode")
		}
		return agentNames(detected), nil
	}

	preselected := defaultAgentSelection(supported)
	options := make([]huh.Option[string], 0, len(supported))
	for _, agent := range supported {
		scopeHint := agent.ProjectRootDir
		if agent.Universal {
			scopeHint = ".agents/skills"
		}
		label := fmt.Sprintf("%s [%s]", agent.DisplayName, scopeHint)
		options = append(options, huh.NewOption(label, agent.Name))
	}

	field := huh.NewMultiSelect[string]().
		Key("agents").
		Title("Target Agents").
		Description("Select the editors/agents where rc should install skills").
		Options(options...).
		Value(&preselected).
		Limit(len(supported)).
		Validate(func(values []string) error {
			if len(values) == 0 {
				return errors.New("select at least one agent")
			}
			return nil
		})
	if err := runPromptField(field); err != nil {
		return nil, fmt.Errorf("select target agents: %w", err)
	}
	return preselected, nil
}

func (s *setupCommandState) resolveScope(cmd *cobra.Command, agents []string) (bool, error) {
	if cmd.Flags().Changed("global") || s.yes {
		return s.global, nil
	}
	if len(agents) == 0 {
		return false, errors.New("resolve installation scope: no agents selected")
	}

	selection := "project"
	field := huh.NewSelect[string]().
		Key("scope").
		Title("Installation Scope").
		Description("Choose whether skills and reusable agents are shared per project or available globally").
		Options(
			huh.NewOption("Project (recommended)", "project"),
			huh.NewOption("Global", "global"),
		).
		Value(&selection)
	if err := runPromptField(field); err != nil {
		return false, fmt.Errorf("select installation scope: %w", err)
	}
	return selection == "global", nil
}

func (s *setupCommandState) resolveInstallMode(
	cmd *cobra.Command,
	supportedAgents []setup.Agent,
	selectedAgents []string,
	global bool,
) (setup.InstallMode, error) {
	if s.copy {
		return setup.InstallModeCopy, nil
	}

	roots := make(map[string]struct{}, len(selectedAgents))
	selected, err := setup.SelectAgents(supportedAgents, selectedAgents)
	if err != nil {
		return "", err
	}
	for _, agent := range selected {
		root := agent.ProjectRootDir
		if global {
			root = agent.GlobalRootDir
		}
		if agent.Universal {
			root = ".agents/skills"
		}
		roots[root] = struct{}{}
	}

	if len(roots) <= 1 {
		return setup.InstallModeCopy, nil
	}
	if s.yes || cmd.Flags().Changed("copy") {
		return setup.InstallModeSymlink, nil
	}

	selection := string(setup.InstallModeSymlink)
	field := huh.NewSelect[string]().
		Key("mode").
		Title("Installation Method").
		Description("Symlink keeps one canonical copy; copy duplicates files into each agent directory").
		Options(
			huh.NewOption("Symlink (recommended)", string(setup.InstallModeSymlink)),
			huh.NewOption("Copy", string(setup.InstallModeCopy)),
		).
		Value(&selection)
	if err := runPromptField(field); err != nil {
		return "", fmt.Errorf("select installation method: %w", err)
	}
	return setup.InstallMode(selection), nil
}

// --- Styled output functions ---

func rcVersionLabel() string {
	v := version.Version
	if v == "" || v == "dev" {
		return "rc dev"
	}
	// Release ldflags already embed a leading "v" (e.g. v0.1.1); trim it so the
	// label never doubles up as "rc vv0.1.1".
	return "rc v" + strings.TrimPrefix(v, "v")
}

func rcWelcomeUser() string {
	u, err := user.Current()
	if err == nil && u.Name != "" {
		return "Welcome back, " + u.Name + "!"
	}
	if err == nil && u.Username != "" {
		return "Welcome back, " + u.Username + "!"
	}
	return "Welcome back!"
}

// printWelcomeHeader renders the Gemini-style setup splash: the rc block-art
// wordmark with the Escale triangle, an eyebrow + greeting, getting-started
// tips, and a status footer. The banner always renders (this runs only in
// interactive setup); the footer separator spans the terminal width when known.
func printWelcomeHeader(cmd *cobra.Command) {
	splash := renderRcSplash(terminalWidth(cmd.OutOrStdout()), "rc // SETUP", rcWelcomeUser())
	lipgloss.Fprintln(cmd.OutOrStdout(), splash)
}

func printSetupAssets(
	cmd *cobra.Command,
	skills []setup.Skill,
	reusableAgents []setup.ReusableAgent,
	conflicts []setup.CatalogConflict,
) {
	if len(skills) == 0 && len(reusableAgents) == 0 {
		printSetupWarnings(cmd, conflicts)
		return
	}
	styles := newCLIChromeStyles()

	if len(skills) > 0 {
		maxNameLen := 0
		for i := range skills {
			if len(skills[i].Name) > maxNameLen {
				maxNameLen = len(skills[i].Name)
			}
		}

		lipgloss.Fprintln(cmd.OutOrStdout(), styles.sectionTitle.Render("Setup Skills"))

		for i := range skills {
			skill := &skills[i]
			name := styles.skill.Render(padRight(skill.Name, maxNameLen))
			source := styles.value.Render(
				"[" + setupAssetSourceLabel(skill.Origin, skill.ExtensionSource, skill.ExtensionName) + "]",
			)
			desc := styles.path.Render(skill.Description)
			lipgloss.Fprintf(cmd.OutOrStdout(), "  %s  %s  %s\n", name, source, desc)
		}
	}

	if len(reusableAgents) > 0 {
		if len(skills) > 0 {
			fmt.Fprintln(cmd.OutOrStdout())
		}

		maxNameLen := 0
		for i := range reusableAgents {
			if len(reusableAgents[i].Name) > maxNameLen {
				maxNameLen = len(reusableAgents[i].Name)
			}
		}

		lipgloss.Fprintln(cmd.OutOrStdout(), styles.sectionTitle.Render("Reusable Agents"))
		for i := range reusableAgents {
			reusableAgent := &reusableAgents[i]
			name := styles.agent.Render(padRight(reusableAgent.Name, maxNameLen))
			source := styles.value.Render(
				"[" + setupAssetSourceLabel(
					reusableAgent.Origin,
					reusableAgent.ExtensionSource,
					reusableAgent.ExtensionName,
				) + "]",
			)
			desc := styles.path.Render(reusableAgent.Description)
			lipgloss.Fprintf(cmd.OutOrStdout(), "  %s  %s  %s\n", name, source, desc)
		}
	}

	printSetupWarnings(cmd, conflicts)
}

func printSetupWarnings(cmd *cobra.Command, conflicts []setup.CatalogConflict) {
	if len(conflicts) == 0 {
		return
	}

	styles := newCLIChromeStyles()
	w := cmd.OutOrStdout()
	fmt.Fprintln(w)
	lipgloss.Fprintln(w, styles.sectionTitle.Render("Warnings"))
	for i := range conflicts {
		lipgloss.Fprintf(w, "  %s  %s\n", styles.warn.Render("!"), formatSetupCatalogConflict(conflicts[i]))
	}
}

func printPreviewSummary(
	cmd *cobra.Command,
	previews []setup.PreviewItem,
	reusableAgentPreviews []setup.ReusableAgentPreviewItem,
	global bool,
	mode setup.InstallMode,
) {
	if len(previews) == 0 && len(reusableAgentPreviews) == 0 {
		return
	}
	styles := newCLIChromeStyles()

	cwd, homeDir := displayRoots()

	w := cmd.OutOrStdout()
	lipgloss.Fprintln(w, styles.sectionTitle.Render("Installation Summary"))
	fmt.Fprintln(w)

	lipgloss.Fprintf(w, "  %s  %s\n", styles.label.Render("Scope "), styles.value.Render(scopeLabel(global)))
	lipgloss.Fprintf(w, "  %s  %s\n", styles.label.Render("Method"), styles.value.Render(string(mode)))
	fmt.Fprintln(w)
	lipgloss.Fprintln(w, styles.separator.Render("  "+strings.Repeat("─", 50)))
	fmt.Fprintln(w)

	maxSkillLen := 0
	maxAgentLen := 0
	for i := range previews {
		if len(previews[i].Skill.Name) > maxSkillLen {
			maxSkillLen = len(previews[i].Skill.Name)
		}
		if len(previews[i].Agent.DisplayName) > maxAgentLen {
			maxAgentLen = len(previews[i].Agent.DisplayName)
		}
	}

	for i := range previews {
		preview := &previews[i]
		name := styles.skill.Render(padRight(preview.Skill.Name, maxSkillLen))
		arrow := styles.arrow.Render("->")
		agent := styles.agent.Render(padRight(preview.Agent.DisplayName, maxAgentLen))
		path := styles.path.Render(shortenPath(preview.TargetPath, cwd, homeDir))

		line := fmt.Sprintf("    %s  %s  %s  %s", name, arrow, agent, path)

		if mode == setup.InstallModeSymlink && !sameInstallPath(preview.CanonicalPath, preview.TargetPath) {
			via := styles.path.Render("via " + shortenPath(preview.CanonicalPath, cwd, homeDir))
			line += "  " + via
		}
		if preview.WillOverwrite {
			line += "  " + styles.warn.Render("[overwrite]")
		}
		lipgloss.Fprintln(w, line)
	}

	if len(reusableAgentPreviews) > 0 {
		fmt.Fprintln(w)
		lipgloss.Fprintln(w, styles.sectionTitle.Render(reusableAgentSectionTitle(global)))
		fmt.Fprintln(w)

		maxReusableAgentLen := 0
		for i := range reusableAgentPreviews {
			if len(reusableAgentPreviews[i].ReusableAgent.Name) > maxReusableAgentLen {
				maxReusableAgentLen = len(reusableAgentPreviews[i].ReusableAgent.Name)
			}
		}

		for i := range reusableAgentPreviews {
			preview := &reusableAgentPreviews[i]
			name := styles.agent.Render(padRight(preview.ReusableAgent.Name, maxReusableAgentLen))
			path := styles.path.Render(shortenPath(preview.TargetPath, cwd, homeDir))

			line := fmt.Sprintf("    %s  %s", name, path)
			if preview.WillOverwrite {
				line += "  " + styles.warn.Render("[overwrite]")
			}
			lipgloss.Fprintln(w, line)
		}
	}
	fmt.Fprintln(w)
}

func printInstallResult(cmd *cobra.Command, result *setup.Result) {
	if result == nil {
		return
	}
	styles := newCLIChromeStyles()
	cwd, homeDir := displayRoots()
	w := cmd.OutOrStdout()
	maxSkillLen, maxAgentLen := computeColumnWidths(result.Successful, result.Failed)
	printSkillInstallSuccesses(w, styles, result.Successful, cwd, homeDir, maxSkillLen, maxAgentLen)
	if len(result.Successful) > 0 && len(result.Failed) > 0 {
		fmt.Fprintln(w)
		lipgloss.Fprintln(w, styles.separator.Render("  "+strings.Repeat("─", 50)))
	}
	printSkillInstallFailures(w, styles, result.Failed, cwd, homeDir, maxSkillLen, maxAgentLen)

	printSecondaryInstallResults(w, styles, result, cwd, homeDir)
	fmt.Fprintln(w)
}

// printSecondaryInstallResults renders the reusable-agent, command, and hook
// sections, inserting a separator before each one that follows another section.
func printSecondaryInstallResults(
	w io.Writer,
	styles cliChromeStyles,
	result *setup.Result,
	cwd, homeDir string,
) {
	// Each section prints unconditionally (the printers no-op when empty); a
	// separator precedes a non-empty section only when an earlier section was
	// also non-empty. Tracking that with a running flag keeps the compound
	// boolean conditions out of one function.
	sections := []struct {
		has   bool
		print func()
	}{
		{len(result.ReusableAgentsSuccessful) > 0 || len(result.ReusableAgentsFailed) > 0, func() {
			printReusableAgentInstallResults(w, styles, result, cwd, homeDir)
		}},
		{len(result.CommandsSuccessful) > 0 || len(result.CommandsFailed) > 0, func() {
			printCommandInstallResults(w, styles, result, cwd, homeDir)
		}},
		{len(result.HooksSuccessful) > 0 || len(result.HooksFailed) > 0, func() {
			printHookInstallResults(w, styles, result, cwd, homeDir)
		}},
		{len(result.OpenCodeSuccessful) > 0 || len(result.OpenCodeFailed) > 0, func() {
			printOpenCodeInstallResults(w, styles, result, cwd, homeDir)
		}},
	}

	printedAny := len(result.Successful) > 0 || len(result.Failed) > 0
	for _, section := range sections {
		if section.has && printedAny {
			printSectionSeparator(w, styles)
		}
		section.print()
		if section.has {
			printedAny = true
		}
	}
}

func printSectionSeparator(w io.Writer, styles cliChromeStyles) {
	fmt.Fprintln(w)
	lipgloss.Fprintln(w, styles.separator.Render("  "+strings.Repeat("─", 50)))
	fmt.Fprintln(w)
}

func printSkillInstallSuccesses(
	w io.Writer,
	styles cliChromeStyles,
	successful []setup.SuccessItem,
	cwd, homeDir string,
	maxSkillLen, maxAgentLen int,
) {
	if len(successful) == 0 {
		return
	}

	lipgloss.Fprintln(w, styles.successHeader.Render(
		fmt.Sprintf("  ✓ Installed (%d)", len(successful)),
	))
	fmt.Fprintln(w)

	for i := range successful {
		item := &successful[i]
		icon := styles.successIcon.Render("✓")
		name := styles.skill.Render(padRight(item.Skill.Name, maxSkillLen))
		arrow := styles.arrow.Render("->")
		agent := styles.agent.Render(padRight(item.Agent.DisplayName, maxAgentLen))
		path := styles.path.Render(shortenPath(item.Path, cwd, homeDir))

		line := fmt.Sprintf("    %s  %s  %s  %s  %s", icon, name, arrow, agent, path)
		if item.Mode == setup.InstallModeSymlink && item.SymlinkFailed {
			line += "  " + styles.warn.Render("[copied after symlink failure]")
		}
		lipgloss.Fprintln(w, line)
	}
}

func printSkillInstallFailures(
	w io.Writer,
	styles cliChromeStyles,
	failed []setup.FailureItem,
	cwd, homeDir string,
	maxSkillLen, maxAgentLen int,
) {
	if len(failed) == 0 {
		return
	}

	lipgloss.Fprintln(w, styles.failureHeader.Render(
		fmt.Sprintf("  ✗ Failed (%d)", len(failed)),
	))
	fmt.Fprintln(w)

	for i := range failed {
		item := &failed[i]
		icon := styles.failureIcon.Render("✗")
		name := styles.skill.Render(padRight(item.Skill.Name, maxSkillLen))
		arrow := styles.arrow.Render("->")
		agent := styles.agent.Render(padRight(item.Agent.DisplayName, maxAgentLen))
		path := styles.path.Render(shortenPath(item.Path, cwd, homeDir))

		lipgloss.Fprintf(w, "    %s  %s  %s  %s  %s\n", icon, name, arrow, agent, path)
		lipgloss.Fprintf(w, "       %s\n", styles.errorMessage.Render(item.Error))
	}
}

func printReusableAgentInstallResults(
	w io.Writer,
	styles cliChromeStyles,
	result *setup.Result,
	cwd, homeDir string,
) {
	if result == nil {
		return
	}
	if len(result.ReusableAgentsSuccessful) == 0 && len(result.ReusableAgentsFailed) == 0 {
		return
	}

	maxReusableAgentLen := reusableAgentColumnWidth(
		result.ReusableAgentsSuccessful,
		result.ReusableAgentsFailed,
	)

	if len(result.ReusableAgentsSuccessful) > 0 {
		lipgloss.Fprintln(w, styles.successHeader.Render(
			fmt.Sprintf(
				"  ✓ Installed %s (%d)",
				reusableAgentResultTitle(result.Global),
				len(result.ReusableAgentsSuccessful),
			),
		))
		fmt.Fprintln(w)

		for i := range result.ReusableAgentsSuccessful {
			item := &result.ReusableAgentsSuccessful[i]
			icon := styles.successIcon.Render("✓")
			name := styles.agent.Render(padRight(item.ReusableAgent.Name, maxReusableAgentLen))
			path := styles.path.Render(shortenPath(item.Path, cwd, homeDir))
			lipgloss.Fprintf(w, "    %s  %s  %s\n", icon, name, path)
		}
	}

	if len(result.ReusableAgentsFailed) == 0 {
		return
	}
	if len(result.ReusableAgentsSuccessful) > 0 {
		fmt.Fprintln(w)
	}

	lipgloss.Fprintln(w, styles.failureHeader.Render(
		fmt.Sprintf(
			"  ✗ Failed %s (%d)",
			reusableAgentResultTitle(result.Global),
			len(result.ReusableAgentsFailed),
		),
	))
	fmt.Fprintln(w)

	for i := range result.ReusableAgentsFailed {
		item := &result.ReusableAgentsFailed[i]
		icon := styles.failureIcon.Render("✗")
		name := styles.agent.Render(padRight(item.ReusableAgent.Name, maxReusableAgentLen))
		path := styles.path.Render(shortenPath(item.Path, cwd, homeDir))
		lipgloss.Fprintf(w, "    %s  %s  %s\n", icon, name, path)
		lipgloss.Fprintf(w, "       %s\n", styles.errorMessage.Render(item.Error))
	}
}

func printCommandInstallResults(
	w io.Writer,
	styles cliChromeStyles,
	result *setup.Result,
	cwd, homeDir string,
) {
	if result == nil {
		return
	}
	if len(result.CommandsSuccessful) == 0 && len(result.CommandsFailed) == 0 {
		return
	}

	maxCommandLen := commandColumnWidth(result.CommandsSuccessful, result.CommandsFailed)

	if len(result.CommandsSuccessful) > 0 {
		lipgloss.Fprintln(w, styles.successHeader.Render(
			fmt.Sprintf("  ✓ Installed %s (%d)", commandResultTitle(result.Global), len(result.CommandsSuccessful)),
		))
		fmt.Fprintln(w)

		for i := range result.CommandsSuccessful {
			item := &result.CommandsSuccessful[i]
			icon := styles.successIcon.Render("✓")
			name := styles.agent.Render(padRight(item.Command.Name, maxCommandLen))
			path := styles.path.Render(shortenPath(item.Path, cwd, homeDir))
			lipgloss.Fprintf(w, "    %s  %s  %s\n", icon, name, path)
		}
	}

	if len(result.CommandsFailed) == 0 {
		return
	}
	if len(result.CommandsSuccessful) > 0 {
		fmt.Fprintln(w)
	}

	lipgloss.Fprintln(w, styles.failureHeader.Render(
		fmt.Sprintf("  ✗ Failed %s (%d)", commandResultTitle(result.Global), len(result.CommandsFailed)),
	))
	fmt.Fprintln(w)

	for i := range result.CommandsFailed {
		item := &result.CommandsFailed[i]
		icon := styles.failureIcon.Render("✗")
		name := styles.agent.Render(padRight(item.Command.Name, maxCommandLen))
		path := styles.path.Render(shortenPath(item.Path, cwd, homeDir))
		lipgloss.Fprintf(w, "    %s  %s  %s\n", icon, name, path)
		lipgloss.Fprintf(w, "       %s\n", styles.errorMessage.Render(item.Error))
	}
}

func printOpenCodeInstallResults(
	w io.Writer,
	styles cliChromeStyles,
	result *setup.Result,
	cwd, homeDir string,
) {
	if result == nil {
		return
	}
	if len(result.OpenCodeSuccessful) == 0 && len(result.OpenCodeFailed) == 0 {
		return
	}

	width := openCodeColumnWidth(result.OpenCodeSuccessful, result.OpenCodeFailed)

	if len(result.OpenCodeSuccessful) > 0 {
		lipgloss.Fprintln(w, styles.successHeader.Render(
			fmt.Sprintf("  ✓ Installed OpenCode assets (%d)", len(result.OpenCodeSuccessful)),
		))
		fmt.Fprintln(w)

		for i := range result.OpenCodeSuccessful {
			item := &result.OpenCodeSuccessful[i]
			icon := styles.successIcon.Render("✓")
			name := styles.agent.Render(padRight(item.Kind+" "+item.Name, width))
			path := styles.path.Render(shortenPath(item.Path, cwd, homeDir))
			lipgloss.Fprintf(w, "    %s  %s  %s\n", icon, name, path)
		}
	}

	if len(result.OpenCodeFailed) == 0 {
		return
	}
	if len(result.OpenCodeSuccessful) > 0 {
		fmt.Fprintln(w)
	}

	lipgloss.Fprintln(w, styles.failureHeader.Render(
		fmt.Sprintf("  ✗ Failed OpenCode assets (%d)", len(result.OpenCodeFailed)),
	))
	fmt.Fprintln(w)

	for i := range result.OpenCodeFailed {
		item := &result.OpenCodeFailed[i]
		icon := styles.failureIcon.Render("✗")
		name := styles.agent.Render(padRight(item.Kind+" "+item.Name, width))
		path := styles.path.Render(shortenPath(item.Path, cwd, homeDir))
		lipgloss.Fprintf(w, "    %s  %s  %s\n", icon, name, path)
		lipgloss.Fprintf(w, "       %s\n", styles.errorMessage.Render(item.Error))
	}
}

func openCodeColumnWidth(successful []setup.OpenCodeAssetSuccessItem, failed []setup.OpenCodeAssetFailureItem) int {
	width := 0
	for i := range successful {
		if length := len(successful[i].Kind + " " + successful[i].Name); length > width {
			width = length
		}
	}
	for i := range failed {
		if length := len(failed[i].Kind + " " + failed[i].Name); length > width {
			width = length
		}
	}
	return width
}

func commandColumnWidth(successful []setup.CommandSuccessItem, failed []setup.CommandFailureItem) int {
	maxCommandLen := 0
	for i := range successful {
		if len(successful[i].Command.Name) > maxCommandLen {
			maxCommandLen = len(successful[i].Command.Name)
		}
	}
	for i := range failed {
		if len(failed[i].Command.Name) > maxCommandLen {
			maxCommandLen = len(failed[i].Command.Name)
		}
	}
	return maxCommandLen
}

// printCommandAssets renders the bundled slash commands for `rc setup --list`.
func printCommandAssets(cmd *cobra.Command, commands []setup.Command) {
	if len(commands) == 0 {
		return
	}
	styles := newCLIChromeStyles()
	w := cmd.OutOrStdout()

	maxLen := 0
	for i := range commands {
		if len(commands[i].Name) > maxLen {
			maxLen = len(commands[i].Name)
		}
	}

	fmt.Fprintln(w)
	lipgloss.Fprintln(w, styles.successHeader.Render(fmt.Sprintf("  Slash Commands (%d)", len(commands))))
	fmt.Fprintln(w)
	for i := range commands {
		name := styles.agent.Render(padRight(commands[i].Name, maxLen))
		desc := styles.path.Render(commands[i].Description)
		lipgloss.Fprintf(w, "    %s  %s\n", name, desc)
	}
}

// printCommandPreviewSummary lists the slash commands that will be installed.
func printCommandPreviewSummary(cmd *cobra.Command, previews []setup.CommandPreviewItem, global bool) {
	if len(previews) == 0 {
		return
	}
	styles := newCLIChromeStyles()
	cwd, homeDir := displayRoots()
	w := cmd.OutOrStdout()

	maxLen := 0
	for i := range previews {
		if len(previews[i].Command.Name) > maxLen {
			maxLen = len(previews[i].Command.Name)
		}
	}

	fmt.Fprintln(w)
	lipgloss.Fprintln(w, styles.successHeader.Render(
		fmt.Sprintf("  %s (%d)", commandResultTitle(global), len(previews)),
	))
	fmt.Fprintln(w)
	for i := range previews {
		name := styles.agent.Render(padRight(previews[i].Command.Name, maxLen))
		path := styles.path.Render(shortenPath(previews[i].TargetPath, cwd, homeDir))
		lipgloss.Fprintf(w, "    %s  %s\n", name, path)
	}
}

func commandResultTitle(global bool) string {
	if global {
		return "Global Slash Commands"
	}
	return "Project Slash Commands"
}

func printHookInstallResults(
	w io.Writer,
	styles cliChromeStyles,
	result *setup.Result,
	cwd, homeDir string,
) {
	if result == nil {
		return
	}
	if len(result.HooksSuccessful) == 0 && len(result.HooksFailed) == 0 {
		return
	}

	maxHookLen := hookColumnWidth(result.HooksSuccessful, result.HooksFailed)

	if len(result.HooksSuccessful) > 0 {
		lipgloss.Fprintln(w, styles.successHeader.Render(
			fmt.Sprintf("  ✓ Installed %s (%d)", hookResultTitle(result.Global), len(result.HooksSuccessful)),
		))
		fmt.Fprintln(w)

		for i := range result.HooksSuccessful {
			item := &result.HooksSuccessful[i]
			icon := styles.successIcon.Render("✓")
			name := styles.agent.Render(padRight(item.Name, maxHookLen))
			path := styles.path.Render(shortenPath(item.Path, cwd, homeDir))
			lipgloss.Fprintf(w, "    %s  %s  %s\n", icon, name, path)
		}
	}

	if len(result.HooksFailed) == 0 {
		return
	}
	if len(result.HooksSuccessful) > 0 {
		fmt.Fprintln(w)
	}

	lipgloss.Fprintln(w, styles.failureHeader.Render(
		fmt.Sprintf("  ✗ Failed %s (%d)", hookResultTitle(result.Global), len(result.HooksFailed)),
	))
	fmt.Fprintln(w)

	for i := range result.HooksFailed {
		item := &result.HooksFailed[i]
		icon := styles.failureIcon.Render("✗")
		name := styles.agent.Render(padRight(item.Name, maxHookLen))
		path := styles.path.Render(shortenPath(item.Path, cwd, homeDir))
		lipgloss.Fprintf(w, "    %s  %s  %s\n", icon, name, path)
		lipgloss.Fprintf(w, "       %s\n", styles.errorMessage.Render(item.Error))
	}
}

func hookColumnWidth(successful []setup.HookSuccessItem, failed []setup.HookFailureItem) int {
	maxHookLen := 0
	for i := range successful {
		if len(successful[i].Name) > maxHookLen {
			maxHookLen = len(successful[i].Name)
		}
	}
	for i := range failed {
		if len(failed[i].Name) > maxHookLen {
			maxHookLen = len(failed[i].Name)
		}
	}
	return maxHookLen
}

func hookResultTitle(global bool) string {
	if global {
		return "Global Claude Hooks"
	}
	return "Project Claude Hooks"
}

func reusableAgentColumnWidth(
	successful []setup.ReusableAgentSuccessItem,
	failed []setup.ReusableAgentFailureItem,
) int {
	maxReusableAgentLen := 0
	for i := range successful {
		if len(successful[i].ReusableAgent.Name) > maxReusableAgentLen {
			maxReusableAgentLen = len(successful[i].ReusableAgent.Name)
		}
	}
	for i := range failed {
		if len(failed[i].ReusableAgent.Name) > maxReusableAgentLen {
			maxReusableAgentLen = len(failed[i].ReusableAgent.Name)
		}
	}
	return maxReusableAgentLen
}

func computeColumnWidths(successful []setup.SuccessItem, failed []setup.FailureItem) (int, int) {
	maxSkill, maxAgent := 0, 0
	for i := range successful {
		if len(successful[i].Skill.Name) > maxSkill {
			maxSkill = len(successful[i].Skill.Name)
		}
		if len(successful[i].Agent.DisplayName) > maxAgent {
			maxAgent = len(successful[i].Agent.DisplayName)
		}
	}
	for i := range failed {
		if len(failed[i].Skill.Name) > maxSkill {
			maxSkill = len(failed[i].Skill.Name)
		}
		if len(failed[i].Agent.DisplayName) > maxAgent {
			maxAgent = len(failed[i].Agent.DisplayName)
		}
	}
	return maxSkill, maxAgent
}

// skillFormFrameWidth is the horizontal frame (rounded border + padding) the
// boxed skill picker adds around its content, subtracted from the terminal
// width so the box closes on the right edge instead of overflowing.
const skillFormFrameWidth = 4

// skillOptionChromeWidth approximates the columns the multi-select reserves
// before the option label inside the box (field border/padding + cursor
// selector + checkbox prefix + a small safety margin) so a truncated label
// stays on one line.
const skillOptionChromeWidth = 6

// truncateToWidth shortens s to at most maxWidth display columns, appending an
// ellipsis when it has to cut. maxWidth <= 0 means "no limit". It is rune-based,
// which matches the ASCII skill names and descriptions used here.
func truncateToWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= maxWidth {
		return s
	}
	if maxWidth == 1 {
		return "…"
	}
	return string(runes[:maxWidth-1]) + "…"
}

// terminalWidth reports the column width of the terminal backing w, or 0 when
// w is not a real terminal (piped output, tests). Callers use 0 to fall back
// to content-based sizing.
func terminalWidth(w io.Writer) int {
	type fdWriter interface{ Fd() uintptr }
	file, ok := w.(fdWriter)
	if !ok {
		return 0
	}
	fd := file.Fd()
	if fd > uintptr(math.MaxInt) {
		return 0
	}
	width, _, err := term.GetSize(int(fd))
	if err != nil || width <= 0 {
		return 0
	}
	return width
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// --- Form and utility functions ---

func confirmSetup() (bool, error) {
	confirmed := false
	field := huh.NewConfirm().
		Key("confirm").
		Title("Proceed with installation?").
		Value(&confirmed)
	if err := runPromptField(field); err != nil {
		return false, fmt.Errorf("confirm installation: %w", err)
	}
	return confirmed, nil
}

func runPromptField(field huh.Field) error {
	return huh.NewForm(huh.NewGroup(field)).WithTheme(darkHuhTheme()).Run()
}

func skillNames(skills []setup.Skill) []string {
	names := make([]string, 0, len(skills))
	for i := range skills {
		names = append(names, skills[i].Name)
	}
	return names
}

func agentNames(agents []setup.Agent) []string {
	names := make([]string, 0, len(agents))
	for _, agent := range agents {
		names = append(names, agent.Name)
	}
	return names
}

func previewsToSkills(previews []setup.PreviewItem) []setup.Skill {
	if len(previews) == 0 {
		return nil
	}

	skills := make([]setup.Skill, 0, len(previews))
	seen := make(map[string]struct{}, len(previews))
	for i := range previews {
		preview := &previews[i]
		key := strings.Join([]string{
			preview.Skill.Name,
			string(preview.Skill.Origin),
			preview.Skill.ExtensionSource,
			preview.Skill.ExtensionName,
			preview.Skill.ManifestPath,
			preview.Skill.ResolvedPath,
		}, "\x00")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		skills = append(skills, preview.Skill)
	}
	return skills
}

func formatSetupCatalogConflict(conflict setup.CatalogConflict) string {
	subject := string(conflict.Kind)
	switch conflict.Resolution {
	case setup.CatalogConflictCoreWins:
		return fmt.Sprintf(
			`ignored extension %s %q from %s because the core %s wins`,
			subject,
			conflict.Name,
			setupAssetSourceLabel(
				conflict.Ignored.Origin,
				conflict.Ignored.ExtensionSource,
				conflict.Ignored.ExtensionName,
			),
			subject,
		)
	case setup.CatalogConflictExtensionPrecedence:
		return fmt.Sprintf(
			`ignored extension %s %q from %s because %s wins by precedence`,
			subject,
			conflict.Name,
			setupAssetSourceLabel(
				conflict.Ignored.Origin,
				conflict.Ignored.ExtensionSource,
				conflict.Ignored.ExtensionName,
			),
			setupAssetSourceLabel(
				conflict.Winner.Origin,
				conflict.Winner.ExtensionSource,
				conflict.Winner.ExtensionName,
			),
		)
	default:
		return fmt.Sprintf(`ignored conflicting %s %q`, subject, conflict.Name)
	}
}

func setupAssetSourceLabel(origin setup.AssetOrigin, extensionSource string, extensionName string) string {
	if origin == setup.AssetOriginBundled {
		return "core"
	}

	parts := make([]string, 0, 2)
	if trimmedSource := strings.TrimSpace(extensionSource); trimmedSource != "" {
		parts = append(parts, trimmedSource)
	}
	if trimmedName := strings.TrimSpace(extensionName); trimmedName != "" {
		parts = append(parts, trimmedName)
	}
	if len(parts) == 0 {
		return "extension"
	}
	return strings.Join(parts, ":")
}

// defaultAgentSelection returns the agents pre-checked in the interactive setup
// form. rc targets Claude Code and Codex as its primary agents, so only those two
// start selected (when supported); every other agent starts unchecked and the user
// opts in explicitly.
func defaultAgentSelection(supported []setup.Agent) []string {
	defaults := []string{"claude-code", "codex"}
	selected := make([]string, 0, len(defaults))
	for _, name := range defaults {
		for i := range supported {
			if supported[i].Name == name {
				selected = append(selected, name)
				break
			}
		}
	}
	if len(selected) > 0 {
		return selected
	}
	return nil
}

func scopeLabel(global bool) string {
	if global {
		return string(setup.InstallScopeGlobal)
	}
	return string(setup.InstallScopeProject)
}

func reusableAgentSectionTitle(global bool) string {
	if global {
		return "Global Reusable Agents"
	}
	return "Project Reusable Agents"
}

func reusableAgentResultTitle(global bool) string {
	return reusableAgentSectionTitle(global)
}

func displayRoots() (string, string) {
	var cwd string
	if value, err := os.Getwd(); err == nil {
		cwd = value
	}

	var homeDir string
	if value, err := os.UserHomeDir(); err == nil {
		homeDir = value
	}
	return cwd, homeDir
}

func shortenPath(fullPath, cwd, homeDir string) string {
	if homeDir != "" && (fullPath == homeDir || strings.HasPrefix(fullPath, homeDir+string(os.PathSeparator))) {
		return "~" + strings.TrimPrefix(fullPath, homeDir)
	}
	if cwd != "" && (fullPath == cwd || strings.HasPrefix(fullPath, cwd+string(os.PathSeparator))) {
		return "." + strings.TrimPrefix(fullPath, cwd)
	}
	return filepath.Clean(fullPath)
}

func sameInstallPath(left, right string) bool {
	return filepath.Clean(left) == filepath.Clean(right)
}

func isInteractiveTerminal() bool {
	stdin, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	stdout, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return stdin.Mode()&os.ModeCharDevice != 0 && stdout.Mode()&os.ModeCharDevice != 0
}

package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/rodolfochicone/rc-project/internal/setup"
	"github.com/spf13/cobra"
)

func TestAddSkillHelpListsFlagsAndExamples(t *testing.T) {
	t.Parallel()

	output, err := executeRootCommand("add", "skill", "--help")
	if err != nil {
		t.Fatalf("execute add skill help: %v", err)
	}

	required := []string{"--agent", "--global", "--copy", "--yes", "rc add skill rc-git"}
	for _, snippet := range required {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected add skill help to include %q\noutput:\n%s", snippet, output)
		}
	}
}

func TestAddCommandListsSkillSubcommand(t *testing.T) {
	t.Parallel()

	output, err := executeRootCommand("add", "--help")
	if err != nil {
		t.Fatalf("execute add help: %v", err)
	}
	if !strings.Contains(output, "skill") {
		t.Fatalf("expected add help to list the skill subcommand\noutput:\n%s", output)
	}
}

func TestAddSkillInstallsOnlySelectedSkillWithoutReusableAgents(t *testing.T) {
	t.Parallel()

	claude := setup.Agent{Name: "claude", DisplayName: "Claude Code", ProjectRootDir: ".claude/skills"}

	state := newSetupCommandState()
	state.yes = true
	state.copy = true
	state.skillNames = []string{"rc-git"}
	state.agentNames = []string{"claude"}

	state.loadCatalog = func(context.Context, setup.ResolverOptions) (setup.EffectiveCatalog, error) {
		return setup.EffectiveCatalog{
			Skills: []setup.Skill{
				{Name: "rc-git", Description: "Branch, push, and open a PR"},
				{Name: "rc-final-verify", Description: "Verify before completion"},
			},
		}, nil
	}
	state.listAgents = func(setup.ResolverOptions) ([]setup.Agent, error) {
		return []setup.Agent{claude}, nil
	}
	state.detectAgents = func(setup.ResolverOptions) ([]setup.Agent, error) {
		return []setup.Agent{claude}, nil
	}

	var previewedSkills []setup.Skill
	state.previewSkills = func(
		_ setup.ResolverOptions,
		skills []setup.Skill,
		_ []string,
		_ bool,
		_ setup.InstallMode,
	) ([]setup.PreviewItem, error) {
		previewedSkills = skills
		items := make([]setup.PreviewItem, 0, len(skills))
		for _, sk := range skills {
			items = append(items, setup.PreviewItem{Skill: sk, Agent: claude, TargetPath: ".claude/skills/" + sk.Name})
		}
		return items, nil
	}

	// add skill must never touch reusable agents or legacy cleanup.
	state.previewReusableAgents = func(setup.ReusableAgentInstallConfig) ([]setup.ReusableAgentPreviewItem, error) {
		t.Fatal("add skill must not preview reusable agents")
		return nil, nil
	}
	state.installReusableAgents = func(
		setup.ReusableAgentInstallConfig,
	) ([]setup.ReusableAgentSuccessItem, []setup.ReusableAgentFailureItem, error) {
		t.Fatal("add skill must not install reusable agents")
		return nil, nil, nil
	}
	state.cleanupLegacyAssets = func(setup.LegacyAssetCleanupConfig) (setup.LegacyAssetCleanupResult, error) {
		t.Fatal("add skill must not run legacy cleanup")
		return setup.LegacyAssetCleanupResult{}, nil
	}

	var installedSkills []setup.Skill
	var installedAgents []string
	var installedMode setup.InstallMode
	state.installSkills = func(
		_ setup.ResolverOptions,
		skills []setup.Skill,
		agents []string,
		global bool,
		mode setup.InstallMode,
	) ([]setup.SuccessItem, []setup.FailureItem, error) {
		installedSkills = skills
		installedAgents = agents
		installedMode = mode
		if global {
			t.Fatal("expected project scope for add skill install")
		}
		out := make([]setup.SuccessItem, 0, len(skills))
		for _, sk := range skills {
			out = append(
				out,
				setup.SuccessItem{Skill: sk, Agent: claude, Path: ".claude/skills/" + sk.Name, Mode: mode},
			)
		}
		return out, nil, nil
	}

	cmd := newAddSkillTestCommand()

	if err := state.runAddSkill(cmd); err != nil {
		t.Fatalf("runAddSkill returned error: %v", err)
	}

	if len(previewedSkills) != 1 || previewedSkills[0].Name != "rc-git" {
		t.Fatalf("expected only rc-git previewed, got %#v", previewedSkills)
	}
	if len(installedSkills) != 1 || installedSkills[0].Name != "rc-git" {
		t.Fatalf("expected only rc-git installed, got %#v", installedSkills)
	}
	if len(installedAgents) != 1 || installedAgents[0] != "claude" {
		t.Fatalf("expected claude as the only target agent, got %#v", installedAgents)
	}
	if installedMode != setup.InstallModeCopy {
		t.Fatalf("expected copy mode, got %q", installedMode)
	}
}

func TestAddSkillRejectsUnknownSkill(t *testing.T) {
	t.Parallel()

	state := newSetupCommandState()
	state.yes = true
	state.skillNames = []string{"does-not-exist"}
	state.agentNames = []string{"claude"}
	state.loadCatalog = func(context.Context, setup.ResolverOptions) (setup.EffectiveCatalog, error) {
		return setup.EffectiveCatalog{
			Skills: []setup.Skill{{Name: "rc-git", Description: "Branch, push, and open a PR"}},
		}, nil
	}
	state.installSkills = func(
		_ setup.ResolverOptions,
		_ []setup.Skill,
		_ []string,
		_ bool,
		_ setup.InstallMode,
	) ([]setup.SuccessItem, []setup.FailureItem, error) {
		t.Fatal("install must not run when a skill name is invalid")
		return nil, nil, nil
	}

	err := state.runAddSkill(newAddSkillTestCommand())
	if err == nil {
		t.Fatal("expected error for unknown skill name")
	}
	if !strings.Contains(err.Error(), "invalid skill(s)") {
		t.Fatalf("expected invalid-skill error, got %v", err)
	}
}

func TestAddCommandHelpListsFlagsAndExamples(t *testing.T) {
	t.Parallel()

	output, err := executeRootCommand("add", "command", "--help")
	if err != nil {
		t.Fatalf("execute add command help: %v", err)
	}

	required := []string{"--global", "--yes", "rc add command rc-pipe"}
	for _, snippet := range required {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected add command help to include %q\noutput:\n%s", snippet, output)
		}
	}
}

func TestAddCommandInstallsOnlySelectedCommand(t *testing.T) {
	t.Parallel()

	state := newSetupCommandState()
	state.yes = true
	state.commandNames = []string{"rc-pipe"}

	var previewed []setup.Command
	state.previewCommands = func(cfg setup.CommandInstallConfig) ([]setup.CommandPreviewItem, error) {
		previewed = cfg.Commands
		items := make([]setup.CommandPreviewItem, 0, len(cfg.Commands))
		for i := range cfg.Commands {
			items = append(items, setup.CommandPreviewItem{
				Command:    cfg.Commands[i],
				TargetPath: ".claude/commands/" + cfg.Commands[i].FileName,
			})
		}
		return items, nil
	}

	var installed []setup.Command
	var installedGlobal bool
	state.installCommands = func(
		cfg setup.CommandInstallConfig,
	) ([]setup.CommandSuccessItem, []setup.CommandFailureItem, error) {
		installed = cfg.Commands
		installedGlobal = cfg.Global
		out := make([]setup.CommandSuccessItem, 0, len(cfg.Commands))
		for i := range cfg.Commands {
			out = append(out, setup.CommandSuccessItem{
				Command: cfg.Commands[i],
				Path:    ".claude/commands/" + cfg.Commands[i].FileName,
			})
		}
		return out, nil, nil
	}

	if err := state.runAddCommand(newAddSkillTestCommand()); err != nil {
		t.Fatalf("runAddCommand returned error: %v", err)
	}

	if len(previewed) != 1 || previewed[0].Name != "rc-pipe" {
		t.Fatalf("expected only rc-pipe previewed, got %#v", previewed)
	}
	if len(installed) != 1 || installed[0].Name != "rc-pipe" {
		t.Fatalf("expected only rc-pipe installed, got %#v", installed)
	}
	if installedGlobal {
		t.Fatal("expected project scope for add command install")
	}
}

func TestAddCommandRejectsUnknownCommand(t *testing.T) {
	t.Parallel()

	state := newSetupCommandState()
	state.yes = true
	state.commandNames = []string{"does-not-exist"}
	state.installCommands = func(
		setup.CommandInstallConfig,
	) ([]setup.CommandSuccessItem, []setup.CommandFailureItem, error) {
		t.Fatal("install must not run when a command name is invalid")
		return nil, nil, nil
	}

	if err := state.runAddCommand(newAddSkillTestCommand()); err == nil {
		t.Fatal("expected error for unknown command name")
	}
}

func newAddSkillTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "skill"}
	var sink bytes.Buffer
	cmd.SetOut(&sink)
	cmd.SetErr(&sink)
	cmd.Flags().Bool("global", false, "global")
	cmd.Flags().Bool("copy", false, "copy")
	return cmd
}

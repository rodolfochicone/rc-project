package setup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testCommandFixtures(t *testing.T) []Command {
	t.Helper()

	bundle := newTestBundle(t, map[string]string{
		"rc-plan.md": "---\ndescription: Plan a feature\n---\nplan body\n",
	})
	commands, err := ListCommands(bundle)
	if err != nil {
		t.Fatalf("list commands: %v", err)
	}
	return commands
}

func TestPreviewCommandInstall(t *testing.T) {
	t.Parallel()

	commands := testCommandFixtures(t)
	cwd := t.TempDir()

	previews, err := PreviewCommandInstall(CommandInstallConfig{
		ResolverOptions: ResolverOptions{CWD: cwd, HomeDir: t.TempDir()},
		Commands:        commands,
	})
	if err != nil {
		t.Fatalf("preview command install: %v", err)
	}
	if len(previews) != 1 {
		t.Fatalf("expected 1 preview, got %d", len(previews))
	}

	want := filepath.Join(cwd, ".claude", "commands", "rc-plan.md")
	if previews[0].TargetPath != want {
		t.Fatalf("target path = %q, want %q", previews[0].TargetPath, want)
	}
	if previews[0].WillOverwrite {
		t.Fatal("WillOverwrite should be false for a clean directory")
	}
}

func TestInstallCommandsProjectScopeWritesFile(t *testing.T) {
	t.Parallel()

	const body = "---\ndescription: Plan a feature\n---\nplan body\n"
	bundle := newTestBundle(t, map[string]string{"rc-plan.md": body})
	commands, err := ListCommands(bundle)
	if err != nil {
		t.Fatalf("list commands: %v", err)
	}

	cwd := t.TempDir()
	successes, failures, err := InstallCommands(CommandInstallConfig{
		ResolverOptions: ResolverOptions{CWD: cwd, HomeDir: t.TempDir()},
		Commands:        commands,
	})
	if err != nil {
		t.Fatalf("install commands: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("unexpected failures: %#v", failures)
	}
	if len(successes) != 1 {
		t.Fatalf("expected 1 success, got %d", len(successes))
	}

	target := filepath.Join(cwd, ".claude", "commands", "rc-plan.md")
	if successes[0].Path != target {
		t.Fatalf("install path = %q, want %q", successes[0].Path, target)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read installed command: %v", err)
	}
	if string(data) != body {
		t.Fatalf("installed content = %q, want %q", string(data), body)
	}
}

func TestInstallCommandsGlobalScopeHonorsClaudeConfigDir(t *testing.T) {
	t.Parallel()

	commands := testCommandFixtures(t)
	claudeDir := t.TempDir()

	successes, failures, err := InstallCommands(CommandInstallConfig{
		ResolverOptions: ResolverOptions{
			CWD:             t.TempDir(),
			HomeDir:         t.TempDir(),
			ClaudeConfigDir: claudeDir,
		},
		Commands: commands,
		Global:   true,
	})
	if err != nil {
		t.Fatalf("install commands: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("unexpected failures: %#v", failures)
	}

	target := filepath.Join(claudeDir, "commands", "rc-plan.md")
	if successes[0].Path != target {
		t.Fatalf("install path = %q, want %q", successes[0].Path, target)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected installed command at %q: %v", target, err)
	}
}

func TestInstallCommandsReportsPerItemFailure(t *testing.T) {
	t.Parallel()

	commands := testCommandFixtures(t)
	cwd := t.TempDir()

	// Make .claude a regular file so creating .claude/commands fails.
	if err := os.WriteFile(filepath.Join(cwd, ".claude"), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed conflicting file: %v", err)
	}

	successes, failures, err := InstallCommands(CommandInstallConfig{
		ResolverOptions: ResolverOptions{CWD: cwd, HomeDir: t.TempDir()},
		Commands:        commands,
	})
	if err != nil {
		t.Fatalf("InstallCommands should surface failures per item, not return error: %v", err)
	}
	if len(successes) != 0 {
		t.Fatalf("expected no successes, got %#v", successes)
	}
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(failures))
	}
	if strings.TrimSpace(failures[0].Error) == "" {
		t.Fatal("failure should carry an error message")
	}
}

func TestInstallBundledSetupAssetsSurfacesCommandInstallError(t *testing.T) {
	bundle := newTestBundle(t, map[string]string{
		"rc-create-prd/SKILL.md": "---\nname: rc-create-prd\ndescription: Create a PRD\n---\n",
	})

	previous := installBundledCommands
	installBundledCommands = func(CommandInstallConfig) ([]CommandSuccessItem, []CommandFailureItem, error) {
		return nil, nil, errors.New("commands unavailable")
	}
	t.Cleanup(func() { installBundledCommands = previous })

	result, err := InstallBundledSetupAssets(InstallConfig{
		Bundle:          bundle,
		ResolverOptions: ResolverOptions{CWD: t.TempDir(), HomeDir: t.TempDir()},
		SkillNames:      []string{"rc-create-prd"},
		AgentNames:      []string{"codex"},
		Mode:            InstallModeCopy,
	})
	if err == nil || !strings.Contains(err.Error(), "install bundled commands") {
		t.Fatalf("expected commands install error, got %v", err)
	}
	if result == nil || len(result.Successful) != 1 {
		t.Fatalf("expected partial skill install result, got %#v", result)
	}
}

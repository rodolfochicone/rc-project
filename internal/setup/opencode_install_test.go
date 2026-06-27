package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallBundledOpenCodeAssetsProjectScope(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	successes, failures, err := InstallBundledOpenCodeAssets(OpenCodeInstallConfig{
		ResolverOptions: ResolverOptions{CWD: cwd, HomeDir: t.TempDir()},
		Global:          false,
	})
	if err != nil {
		t.Fatalf("install opencode assets: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("unexpected failures: %+v", failures)
	}

	agentData, err := os.ReadFile(filepath.Join(cwd, ".opencode", "agent", "rc-exec.md"))
	if err != nil {
		t.Fatalf("read installed agent: %v", err)
	}
	if !strings.Contains(string(agentData), "opencode-go/glm-5.2") {
		t.Fatalf("rc-exec must pin its model so skills run on it, got:\n%s", agentData)
	}
	if !strings.Contains(string(agentData), "reasoningEffort: high") {
		t.Fatal("rc-exec must pin its reasoning effort")
	}

	cmdData, err := os.ReadFile(filepath.Join(cwd, ".opencode", "commands", "rc-plan.md"))
	if err != nil {
		t.Fatalf("read installed command: %v", err)
	}
	if !strings.Contains(string(cmdData), "agent: rc") {
		t.Fatal("rc-plan command must route to the orchestrator agent so each phase runs on its own model")
	}

	agents, commands := countOpenCodeKinds(successes)
	if agents == 0 || commands == 0 {
		t.Fatalf("expected both agents and commands installed, got agents=%d commands=%d", agents, commands)
	}
}

func TestInstallBundledOpenCodeAssetsGlobalScope(t *testing.T) {
	t.Parallel()

	xdg := t.TempDir()
	successes, failures, err := InstallBundledOpenCodeAssets(OpenCodeInstallConfig{
		ResolverOptions: ResolverOptions{HomeDir: t.TempDir(), XDGConfigHome: xdg},
		Global:          true,
	})
	if err != nil {
		t.Fatalf("install opencode assets: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("unexpected failures: %+v", failures)
	}
	if len(successes) == 0 {
		t.Fatal("expected installed assets")
	}

	gitData, err := os.ReadFile(filepath.Join(xdg, "opencode", "agent", "rc-git.md"))
	if err != nil {
		t.Fatalf("read global agent: %v", err)
	}
	if !strings.Contains(string(gitData), "opencode-go/deepseek-v4-flash") {
		t.Fatal("rc-git must pin the cheap git model")
	}
	if !strings.Contains(string(gitData), "reasoningEffort: low") {
		t.Fatal("rc-git must use low effort per the model routing table")
	}
}

func countOpenCodeKinds(items []OpenCodeAssetSuccessItem) (agents, commands int) {
	for i := range items {
		switch items[i].Kind {
		case "agent":
			agents++
		case "command":
			commands++
		}
	}
	return agents, commands
}

package setup

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// twoSkillBundle is the canonical bundle used across the sync tests.
func twoSkillBundle(t *testing.T) fs.FS {
	t.Helper()
	return newTestBundle(t, map[string]string{
		"rc-create-prd/SKILL.md":   "---\nname: rc-create-prd\ndescription: Create a PRD\n---\nbody\n",
		"rc-final-verify/SKILL.md": "---\nname: rc-final-verify\ndescription: Verify completion\n---\nbody\n",
	})
}

func skillNamesFromSuccess(items []SuccessItem) []string {
	names := make([]string, 0, len(items))
	for i := range items {
		names = append(names, items[i].Skill.Name)
	}
	slices.Sort(names)
	return names
}

func skillNamesFromSkills(skills []Skill) []string {
	names := make([]string, 0, len(skills))
	for i := range skills {
		names = append(names, skills[i].Name)
	}
	slices.Sort(names)
	return names
}

func TestSyncAddsMissingSkillsToEmptyProject(t *testing.T) {
	t.Parallel()

	bundle := twoSkillBundle(t)
	projectDir := t.TempDir()
	homeDir := t.TempDir()

	result, err := Sync(SyncConfig{
		Bundle:          bundle,
		ResolverOptions: ResolverOptions{CWD: projectDir, HomeDir: homeDir},
		AgentName:       "claude-code",
	})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}

	if result.Scope != InstallScopeProject {
		t.Fatalf("expected project scope, got %q", result.Scope)
	}
	if got := skillNamesFromSuccess(result.Added); !slices.Equal(got, []string{"rc-create-prd", "rc-final-verify"}) {
		t.Fatalf("expected both skills added, got %v", got)
	}
	if len(result.Updated) != 0 {
		t.Fatalf("expected no updates on empty project, got %v", skillNamesFromSuccess(result.Updated))
	}
	if len(result.Unchanged) != 0 {
		t.Fatalf("expected nothing unchanged on empty project, got %v", skillNamesFromSkills(result.Unchanged))
	}
	assertFileExists(t, filepath.Join(projectDir, ".claude", "skills", "rc-create-prd", "SKILL.md"))
	assertFileExists(t, filepath.Join(projectDir, ".claude", "skills", "rc-final-verify", "SKILL.md"))
}

func TestSyncLeavesCurrentSkillsUnchanged(t *testing.T) {
	t.Parallel()

	bundle := twoSkillBundle(t)
	projectDir := t.TempDir()
	homeDir := t.TempDir()
	cfg := SyncConfig{
		Bundle:          bundle,
		ResolverOptions: ResolverOptions{CWD: projectDir, HomeDir: homeDir},
		AgentName:       "claude-code",
	}

	if _, err := Sync(cfg); err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	result, err := Sync(cfg)
	if err != nil {
		t.Fatalf("second sync: %v", err)
	}

	if len(result.Added) != 0 || len(result.Updated) != 0 {
		t.Fatalf("expected no changes on second sync, added=%v updated=%v",
			skillNamesFromSuccess(result.Added), skillNamesFromSuccess(result.Updated))
	}
	if got := skillNamesFromSkills(result.Unchanged); !slices.Equal(got, []string{"rc-create-prd", "rc-final-verify"}) {
		t.Fatalf("expected both skills unchanged, got %v", got)
	}
}

func TestSyncUpdatesDriftedSkillAndRestoresBundleContent(t *testing.T) {
	t.Parallel()

	bundle := twoSkillBundle(t)
	projectDir := t.TempDir()
	homeDir := t.TempDir()
	cfg := SyncConfig{
		Bundle:          bundle,
		ResolverOptions: ResolverOptions{CWD: projectDir, HomeDir: homeDir},
		AgentName:       "claude-code",
	}

	if _, err := Sync(cfg); err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	driftedPath := filepath.Join(projectDir, ".claude", "skills", "rc-create-prd", "SKILL.md")
	if err := os.WriteFile(driftedPath, []byte("locally edited\n"), 0o644); err != nil {
		t.Fatalf("introduce drift: %v", err)
	}

	result, err := Sync(cfg)
	if err != nil {
		t.Fatalf("resync after drift: %v", err)
	}

	if got := skillNamesFromSuccess(result.Updated); !slices.Equal(got, []string{"rc-create-prd"}) {
		t.Fatalf("expected drifted skill updated, got %v", got)
	}
	if got := skillNamesFromSkills(result.Unchanged); !slices.Equal(got, []string{"rc-final-verify"}) {
		t.Fatalf("expected the current skill left unchanged, got %v", got)
	}
	if len(result.Added) != 0 {
		t.Fatalf("expected no additions, got %v", skillNamesFromSuccess(result.Added))
	}

	want, err := fs.ReadFile(bundle, "rc-create-prd/SKILL.md")
	if err != nil {
		t.Fatalf("read bundle skill: %v", err)
	}
	got, err := os.ReadFile(driftedPath)
	if err != nil {
		t.Fatalf("read restored skill: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("expected sync to restore bundle content\nwant: %q\ngot:  %q", want, got)
	}
}

func TestSyncPreservesNonBundledSkills(t *testing.T) {
	t.Parallel()

	bundle := twoSkillBundle(t)
	projectDir := t.TempDir()
	homeDir := t.TempDir()

	customPath := filepath.Join(projectDir, ".claude", "skills", "my-custom", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(customPath), 0o755); err != nil {
		t.Fatalf("mkdir custom skill: %v", err)
	}
	if err := os.WriteFile(customPath, []byte("custom\n"), 0o644); err != nil {
		t.Fatalf("write custom skill: %v", err)
	}

	result, err := Sync(SyncConfig{
		Bundle:          bundle,
		ResolverOptions: ResolverOptions{CWD: projectDir, HomeDir: homeDir},
		AgentName:       "claude-code",
	})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}

	assertFileExists(t, customPath)
	all := append(skillNamesFromSuccess(result.Added), skillNamesFromSuccess(result.Updated)...)
	all = append(all, skillNamesFromSkills(result.Unchanged)...)
	if slices.Contains(all, "my-custom") {
		t.Fatalf("sync should never touch non-bundled skills, but reported my-custom in %v", all)
	}
}

func TestSyncRoutesToAgentDirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		agentName string
		relDir    string
	}{
		{name: "claude code uses .claude/skills", agentName: "claude-code", relDir: ".claude/skills"},
		{name: "codex uses .agents/skills", agentName: "codex", relDir: ".agents/skills"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bundle := twoSkillBundle(t)
			projectDir := t.TempDir()
			homeDir := t.TempDir()

			if _, err := Sync(SyncConfig{
				Bundle:          bundle,
				ResolverOptions: ResolverOptions{CWD: projectDir, HomeDir: homeDir},
				AgentName:       tt.agentName,
			}); err != nil {
				t.Fatalf("sync: %v", err)
			}

			assertFileExists(t, filepath.Join(projectDir, filepath.FromSlash(tt.relDir), "rc-create-prd", "SKILL.md"))
		})
	}
}

func TestSyncGlobalScopeInstallsUnderHome(t *testing.T) {
	t.Parallel()

	bundle := twoSkillBundle(t)
	projectDir := t.TempDir()
	homeDir := t.TempDir()

	result, err := Sync(SyncConfig{
		Bundle:          bundle,
		ResolverOptions: ResolverOptions{CWD: projectDir, HomeDir: homeDir},
		AgentName:       "claude-code",
		Global:          true,
	})
	if err != nil {
		t.Fatalf("sync global: %v", err)
	}

	if result.Scope != InstallScopeGlobal {
		t.Fatalf("expected global scope, got %q", result.Scope)
	}
	assertFileExists(t, filepath.Join(homeDir, ".claude", "skills", "rc-create-prd", "SKILL.md"))
}

func TestSyncRejectsNilBundle(t *testing.T) {
	t.Parallel()

	_, err := Sync(SyncConfig{
		ResolverOptions: ResolverOptions{CWD: t.TempDir(), HomeDir: t.TempDir()},
		AgentName:       "claude-code",
	})
	if err == nil {
		t.Fatal("expected an error when bundle is nil")
	}
}

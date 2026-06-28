package setup

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func testHookScriptsFS() fstest.MapFS {
	fsys := fstest.MapFS{}
	for _, script := range uniqueHookScripts() {
		fsys["scripts/"+script] = &fstest.MapFile{Data: []byte("#!/usr/bin/env bash\nexit 0\n")}
	}
	for _, script := range rcHookSupportScripts {
		fsys["scripts/"+script] = &fstest.MapFile{Data: []byte("#!/usr/bin/env bash\n# support lib\n")}
	}
	return fsys
}

func readSettings(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read settings %s: %v", path, err)
	}
	root := map[string]any{}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("parse settings %s: %v", path, err)
	}
	return root
}

func hookCommandsForEvent(t *testing.T, settings map[string]any, event string) []string {
	t.Helper()
	hooksObj, ok := settings["hooks"].(map[string]any)
	if !ok {
		return nil
	}
	groups, ok := hooksObj[event].([]any)
	if !ok {
		return nil
	}
	var commands []string
	for _, group := range groups {
		groupObj, ok := group.(map[string]any)
		if !ok {
			continue
		}
		handlers, ok := groupObj["hooks"].([]any)
		if !ok {
			continue
		}
		for _, handler := range handlers {
			handlerObj, ok := handler.(map[string]any)
			if !ok {
				continue
			}
			if command, ok := handlerObj["command"].(string); ok {
				commands = append(commands, command)
			}
		}
	}
	return commands
}

func TestInstallBundledHooksFreshProject(t *testing.T) {
	t.Parallel()
	cwd := t.TempDir()
	home := t.TempDir()

	successes, failures, err := InstallBundledHooks(HookInstallConfig{
		ResolverOptions: ResolverOptions{CWD: cwd, HomeDir: home},
		ScriptsFS:       testHookScriptsFS(),
	})
	if err != nil {
		t.Fatalf("install hooks: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("unexpected failures: %+v", failures)
	}
	// Every hook script, every support file, plus settings.json should be reported as installed.
	if want := len(uniqueHookScripts()) + len(rcHookSupportScripts) + 1; len(successes) != want {
		t.Fatalf("success count = %d, want %d (%+v)", len(successes), want, successes)
	}

	scriptsDir := filepath.Join(cwd, ".claude", "rc", "hooks", "scripts")
	for _, script := range append(append([]string{}, uniqueHookScripts()...), rcHookSupportScripts...) {
		info, err := os.Stat(filepath.Join(scriptsDir, script))
		if err != nil {
			t.Fatalf("stat installed script %s: %v", script, err)
		}
		if info.Mode().Perm()&0o100 == 0 {
			t.Errorf("script %s is not executable: mode %v", script, info.Mode().Perm())
		}
	}

	settings := readSettings(t, filepath.Join(cwd, ".claude", "settings.json"))
	pre := hookCommandsForEvent(t, settings, "PreToolUse")
	post := hookCommandsForEvent(t, settings, "PostToolUse")
	if len(pre) != 4 {
		t.Errorf("PreToolUse command count = %d, want 4 (%v)", len(pre), pre)
	}
	if len(post) != 2 {
		t.Errorf("PostToolUse command count = %d, want 2 (%v)", len(post), post)
	}
	for _, command := range append(append([]string{}, pre...), post...) {
		if !strings.HasPrefix(command, "${CLAUDE_PROJECT_DIR}/.claude/rc/hooks/scripts/") {
			t.Errorf("project command path not relative to CLAUDE_PROJECT_DIR: %q", command)
		}
	}
}

func TestInstallBundledHooksPreservesExistingSettings(t *testing.T) {
	t.Parallel()
	cwd := t.TempDir()
	home := t.TempDir()

	claudeDir := filepath.Join(cwd, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	existing := `{
  "permissions": {"allow": ["Bash(go test:*)"]},
  "hooks": {
    "Stop": [{"matcher": "", "hooks": [{"type": "command", "command": "echo done"}]}],
    "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "echo user-hook"}]}]
  }
}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	if _, failures, err := InstallBundledHooks(HookInstallConfig{
		ResolverOptions: ResolverOptions{CWD: cwd, HomeDir: home},
		ScriptsFS:       testHookScriptsFS(),
	}); err != nil || len(failures) != 0 {
		t.Fatalf("install hooks: err=%v failures=%+v", err, failures)
	}

	settings := readSettings(t, settingsPath)
	if _, ok := settings["permissions"]; !ok {
		t.Error("permissions key was dropped during merge")
	}
	stop := hookCommandsForEvent(t, settings, "Stop")
	foundUserStop := false
	rcStopCount := 0
	for _, command := range stop {
		switch {
		case command == "echo done":
			foundUserStop = true
		case strings.Contains(command, rcHookMarker):
			rcStopCount++
		}
	}
	if !foundUserStop {
		t.Errorf("user Stop hook was dropped: %v", stop)
	}
	if rcStopCount != 1 {
		t.Errorf("expected 1 rc Stop hook, got %d (%v)", rcStopCount, stop)
	}
	pre := hookCommandsForEvent(t, settings, "PreToolUse")
	foundUser := false
	rcCount := 0
	for _, command := range pre {
		switch {
		case command == "echo user-hook":
			foundUser = true
		case strings.Contains(command, rcHookMarker):
			rcCount++
		}
	}
	if !foundUser {
		t.Error("user PreToolUse hook was dropped")
	}
	if rcCount != 4 {
		t.Errorf("expected 4 rc PreToolUse hooks, got %d (%v)", rcCount, pre)
	}
}

func TestInstallBundledHooksIsIdempotent(t *testing.T) {
	t.Parallel()
	cwd := t.TempDir()
	home := t.TempDir()
	cfg := HookInstallConfig{
		ResolverOptions: ResolverOptions{CWD: cwd, HomeDir: home},
		ScriptsFS:       testHookScriptsFS(),
	}
	settingsPath := filepath.Join(cwd, ".claude", "settings.json")

	if _, _, err := InstallBundledHooks(cfg); err != nil {
		t.Fatalf("first install: %v", err)
	}
	first, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read first settings: %v", err)
	}
	if _, _, err := InstallBundledHooks(cfg); err != nil {
		t.Fatalf("second install: %v", err)
	}
	second, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read second settings: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Errorf("settings.json changed on re-install (not idempotent)\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestInstallBundledHooksReplacesStaleRcEntries(t *testing.T) {
	t.Parallel()
	cwd := t.TempDir()
	home := t.TempDir()

	claudeDir := filepath.Join(cwd, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	// A stale rc entry pointing at a script that no longer exists.
	stale := `{
  "hooks": {
    "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "${CLAUDE_PROJECT_DIR}/.claude/rc/hooks/scripts/old-removed.sh"}]}]
  }
}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(stale), 0o644); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	if _, _, err := InstallBundledHooks(HookInstallConfig{
		ResolverOptions: ResolverOptions{CWD: cwd, HomeDir: home},
		ScriptsFS:       testHookScriptsFS(),
	}); err != nil {
		t.Fatalf("install hooks: %v", err)
	}

	for _, command := range hookCommandsForEvent(t, readSettings(t, settingsPath), "PreToolUse") {
		if strings.Contains(command, "old-removed.sh") {
			t.Errorf("stale rc hook was not replaced: %q", command)
		}
	}
}

func TestInstallBundledHooksGlobalScope(t *testing.T) {
	t.Parallel()
	cwd := t.TempDir()
	home := t.TempDir()
	claudeConfig := filepath.Join(home, ".claude")

	if _, failures, err := InstallBundledHooks(HookInstallConfig{
		ResolverOptions: ResolverOptions{CWD: cwd, HomeDir: home, ClaudeConfigDir: claudeConfig},
		ScriptsFS:       testHookScriptsFS(),
		Global:          true,
	}); err != nil || len(failures) != 0 {
		t.Fatalf("install hooks: err=%v failures=%+v", err, failures)
	}

	if _, err := os.Stat(filepath.Join(claudeConfig, "rc", "hooks", "scripts", "git-guard.sh")); err != nil {
		t.Fatalf("global script not installed: %v", err)
	}
	settings := readSettings(t, filepath.Join(claudeConfig, "settings.json"))
	commands := hookCommandsForEvent(t, settings, "PreToolUse")
	if len(commands) == 0 {
		t.Fatal("no global PreToolUse hooks written")
	}
	for _, command := range commands {
		if strings.Contains(command, "${CLAUDE_PROJECT_DIR}") {
			t.Errorf("global command path must be absolute, got %q", command)
		}
		if !filepath.IsAbs(command) {
			t.Errorf("global command path is not absolute: %q", command)
		}
	}
}

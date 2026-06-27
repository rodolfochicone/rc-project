package setup

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rodolfochicone/rc-project/hooks"
)

// rcHookMarker identifies rc-managed hook entries inside a settings.json so they
// can be replaced idempotently while user-authored hooks are preserved.
const rcHookMarker = "/rc/hooks/scripts/"

// hookGroupSpec declares one matcher group of rc hook scripts for an event. The
// ordering of this slice is the canonical install order and drives both script
// copying and settings.json generation deterministically.
type hookGroupSpec struct {
	event   string
	matcher string
	scripts []string
}

var rcHookGroups = []hookGroupSpec{
	{event: "PreToolUse", matcher: "Bash", scripts: []string{"git-guard.sh", "commit-guard.sh"}},
	{event: "PreToolUse", matcher: "Edit|Write|MultiEdit", scripts: []string{"go-mod-guard.sh"}},
	{event: "PostToolUse", matcher: "Edit|Write|MultiEdit", scripts: []string{"go-fmt.sh"}},
}

// HookInstallConfig selects the scope for a hooks install. ScriptsFS defaults to
// the embedded bundle when nil (overridable in tests).
type HookInstallConfig struct {
	ResolverOptions

	ScriptsFS fs.FS
	Global    bool
}

// HookSuccessItem reports a hook artifact (a script or settings.json) that was installed.
type HookSuccessItem struct {
	Name string
	Path string
}

// HookFailureItem reports a hook artifact that failed to install.
type HookFailureItem struct {
	Name  string
	Path  string
	Error string
}

// InstallBundledHooks copies the bundled Claude Code hook scripts into the
// selected scope and merges the rc hook entries into that scope's settings.json,
// preserving any user-authored hooks. It is Claude-only by nature, mirroring how
// bundled slash commands install unconditionally into .claude/.
func InstallBundledHooks(cfg HookInstallConfig) ([]HookSuccessItem, []HookFailureItem, error) {
	env, err := resolveEnvironment(cfg.ResolverOptions)
	if err != nil {
		return nil, nil, err
	}

	bundle := cfg.ScriptsFS
	if bundle == nil {
		bundle = hooks.ScriptsFS
	}

	scriptsDir := hooksScriptsRoot(env, cfg.Global)
	settingsFile := hooksSettingsPath(env, cfg.Global)

	successes := make([]HookSuccessItem, 0, len(rcHookGroups)+1)
	failures := make([]HookFailureItem, 0)

	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("prepare hooks scripts dir %s: %w", scriptsDir, err)
	}

	for _, script := range uniqueHookScripts() {
		target := filepath.Join(scriptsDir, script)
		if !isPathSafe(scriptsDir, target) {
			failures = append(failures, HookFailureItem{
				Name:  script,
				Path:  target,
				Error: fmt.Sprintf("resolved target escapes %s", scriptsDir),
			})
			continue
		}
		if err := copyHookScript(bundle, script, target); err != nil {
			failures = append(failures, HookFailureItem{Name: script, Path: target, Error: err.Error()})
			continue
		}
		successes = append(successes, HookSuccessItem{Name: script, Path: target})
	}

	if err := mergeHookSettings(settingsFile, scriptsDir, cfg.Global, env); err != nil {
		failures = append(failures, HookFailureItem{Name: "settings.json", Path: settingsFile, Error: err.Error()})
		return successes, failures, nil
	}
	successes = append(successes, HookSuccessItem{Name: "settings.json", Path: settingsFile})

	return successes, failures, nil
}

func hooksScriptsRoot(env resolvedEnvironment, global bool) string {
	if global {
		return filepath.Join(env.claudeConfigDir, "rc", "hooks", "scripts")
	}
	return filepath.Join(env.cwd, ".claude", "rc", "hooks", "scripts")
}

func hooksSettingsPath(env resolvedEnvironment, global bool) string {
	if global {
		return filepath.Join(env.claudeConfigDir, "settings.json")
	}
	return filepath.Join(env.cwd, ".claude", "settings.json")
}

func uniqueHookScripts() []string {
	seen := make(map[string]bool)
	out := make([]string, 0)
	for i := range rcHookGroups {
		for _, script := range rcHookGroups[i].scripts {
			if seen[script] {
				continue
			}
			seen[script] = true
			out = append(out, script)
		}
	}
	return out
}

func copyHookScript(bundle fs.FS, name, target string) error {
	src, err := bundle.Open("scripts/" + name)
	if err != nil {
		return fmt.Errorf("open bundled hook %s: %w", name, err)
	}
	defer src.Close()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return fmt.Errorf("create hook script %s: %w", target, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return fmt.Errorf("copy hook script %s: %w", target, err)
	}
	// O_CREATE honors umask, so set the executable bit explicitly.
	if err := os.Chmod(target, 0o755); err != nil {
		return fmt.Errorf("chmod hook script %s: %w", target, err)
	}
	return nil
}

// mergeHookSettings rewrites settingsFile so its `hooks` block contains the
// current rc entries, replacing any prior rc entries (identified by rcHookMarker)
// and preserving every other key and user-authored hook.
func mergeHookSettings(settingsFile, scriptsDir string, global bool, env resolvedEnvironment) error {
	root, err := readJSONObject(settingsFile)
	if err != nil {
		return err
	}

	hooksObj := asJSONObject(root["hooks"])
	for _, event := range orderedHookEvents() {
		kept := filterOutRcGroups(asJSONArray(hooksObj[event]))
		hooksObj[event] = append(kept, rcGroupsForEvent(event, scriptsDir, global, env)...)
	}
	root["hooks"] = hooksObj

	return writeJSONObject(settingsFile, root)
}

func orderedHookEvents() []string {
	seen := make(map[string]bool)
	out := make([]string, 0)
	for i := range rcHookGroups {
		if seen[rcHookGroups[i].event] {
			continue
		}
		seen[rcHookGroups[i].event] = true
		out = append(out, rcHookGroups[i].event)
	}
	return out
}

func rcGroupsForEvent(event, scriptsDir string, global bool, env resolvedEnvironment) []any {
	out := make([]any, 0)
	for i := range rcHookGroups {
		if rcHookGroups[i].event != event {
			continue
		}
		handlers := make([]any, 0, len(rcHookGroups[i].scripts))
		for _, script := range rcHookGroups[i].scripts {
			handlers = append(handlers, map[string]any{
				"type":    "command",
				"command": hookCommandPath(scriptsDir, script, global, env),
			})
		}
		out = append(out, map[string]any{
			"matcher": rcHookGroups[i].matcher,
			"hooks":   handlers,
		})
	}
	return out
}

func hookCommandPath(scriptsDir, script string, global bool, env resolvedEnvironment) string {
	abs := filepath.Join(scriptsDir, script)
	if global {
		return abs
	}
	rel, err := filepath.Rel(env.cwd, abs)
	if err != nil {
		return abs
	}
	return "${CLAUDE_PROJECT_DIR}/" + filepath.ToSlash(rel)
}

func filterOutRcGroups(groups []any) []any {
	out := make([]any, 0, len(groups))
	for _, group := range groups {
		if isRcHookGroup(group) {
			continue
		}
		out = append(out, group)
	}
	return out
}

func isRcHookGroup(group any) bool {
	obj, ok := group.(map[string]any)
	if !ok {
		return false
	}
	handlers, ok := obj["hooks"].([]any)
	if !ok {
		return false
	}
	for _, handler := range handlers {
		handlerObj, ok := handler.(map[string]any)
		if !ok {
			continue
		}
		command, ok := handlerObj["command"].(string)
		if !ok {
			continue
		}
		if strings.Contains(command, rcHookMarker) {
			return true
		}
	}
	return false
}

func readJSONObject(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return map[string]any{}, nil
	}
	root := map[string]any{}
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return root, nil
}

func writeJSONObject(path string, root map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("prepare settings dir for %s: %w", path, err)
	}
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings %s: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func asJSONObject(v any) map[string]any {
	if obj, ok := v.(map[string]any); ok {
		return obj
	}
	return map[string]any{}
}

func asJSONArray(v any) []any {
	if arr, ok := v.([]any); ok {
		return arr
	}
	return nil
}

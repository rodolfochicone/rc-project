package setup

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rodolfochicone/rc-project/hooks"
	"github.com/rodolfochicone/rc-project/opencode"
)

// OpenCodeAgentName is the setup agent identifier that opts a run into installing
// the bundled OpenCode agents and commands.
const OpenCodeAgentName = "opencode"

type (
	// OpenCodeInstallConfig selects the scope for an OpenCode asset install.
	OpenCodeInstallConfig struct {
		ResolverOptions

		Global bool
	}

	// OpenCodeAssetSuccessItem reports a bundled OpenCode agent or command that was installed.
	OpenCodeAssetSuccessItem struct {
		Kind string
		Name string
		Path string
	}

	// OpenCodeAssetFailureItem reports a bundled OpenCode asset that failed to install.
	OpenCodeAssetFailureItem struct {
		Kind  string
		Name  string
		Path  string
		Error string
	}
)

// openCodeRoot resolves the OpenCode config root for the requested scope. Global
// scope uses <xdgConfigHome>/opencode (i.e. ~/.config/opencode); project scope uses
// ./.opencode. OpenCode reads agents from <root>/agent and commands from <root>/commands.
func openCodeRoot(env resolvedEnvironment, global bool) string {
	if global {
		return filepath.Join(env.xdgConfigHome, "opencode")
	}
	return filepath.Join(env.cwd, ".opencode")
}

// InstallBundledOpenCodeAssets installs the bundled OpenCode agents and commands
// into the selected scope so opencode runs rc skills on the intended models.
func InstallBundledOpenCodeAssets(
	cfg OpenCodeInstallConfig,
) ([]OpenCodeAssetSuccessItem, []OpenCodeAssetFailureItem, error) {
	env, err := resolveEnvironment(cfg.ResolverOptions)
	if err != nil {
		return nil, nil, err
	}

	root := openCodeRoot(env, cfg.Global)
	successes := make([]OpenCodeAssetSuccessItem, 0)
	failures := make([]OpenCodeAssetFailureItem, 0)

	for _, group := range []struct{ dir, kind string }{{dir: "agent", kind: "agent"}, {dir: "commands", kind: "command"}} {
		entries, err := fs.ReadDir(opencode.FS, group.dir)
		if err != nil {
			return nil, nil, fmt.Errorf("list bundled opencode %ss: %w", group.kind, err)
		}

		destDir := filepath.Join(root, group.dir)
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !strings.HasSuffix(name, ".md") {
				continue
			}

			target := filepath.Join(destDir, name)
			assetName := strings.TrimSuffix(name, ".md")
			if failure := installOpenCodeAsset(
				opencode.FS,
				path.Join(group.dir, name),
				destDir,
				target,
				group.kind,
				assetName,
			); failure != nil {
				failures = append(failures, *failure)
				continue
			}
			successes = append(successes, OpenCodeAssetSuccessItem{Kind: group.kind, Name: assetName, Path: target})
		}
	}

	// Install the rc-hooks plugin into BOTH .opencode/plugin/ and
	// .opencode/plugins/ — opencode has used both directory names across
	// versions, so shipping to both makes the hooks load regardless. The plugin
	// guards against double-registration if a version loads both copies.
	for _, pluginDirName := range []string{"plugin", "plugins"} {
		pluginS, pluginF, err := installOpenCodeFromDir(
			opencode.FS, "plugin", filepath.Join(root, pluginDirName), ".ts", "plugin",
		)
		if err != nil {
			return nil, nil, err
		}
		successes = append(successes, pluginS...)
		failures = append(failures, pluginF...)
	}

	// The hook scripts are the SAME ones the Claude channel uses, so the guard
	// logic has a single source of truth. The plugin invokes them via
	// `bash <script>`, so the executable bit is not required.
	scriptS, scriptF, err := installOpenCodeFromDir(
		hooks.ScriptsFS, "scripts", filepath.Join(root, "rc", "hooks", "scripts"), ".sh", "hook-script",
	)
	if err != nil {
		return nil, nil, err
	}
	successes = append(successes, scriptS...)
	failures = append(failures, scriptF...)

	return successes, failures, nil
}

// installOpenCodeFromDir copies every file with the given suffix from sourceDir
// in the bundle into destDir, reporting each as the given kind. The asset name is
// the file name with the suffix trimmed.
func installOpenCodeFromDir(
	bundle fs.FS, sourceDir, destDir, suffix, kind string,
) ([]OpenCodeAssetSuccessItem, []OpenCodeAssetFailureItem, error) {
	entries, err := fs.ReadDir(bundle, sourceDir)
	if err != nil {
		return nil, nil, fmt.Errorf("list bundled opencode %ss: %w", kind, err)
	}
	successes := make([]OpenCodeAssetSuccessItem, 0)
	failures := make([]OpenCodeAssetFailureItem, 0)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, suffix) {
			continue
		}
		target := filepath.Join(destDir, name)
		assetName := strings.TrimSuffix(name, suffix)
		if failure := installOpenCodeAsset(
			bundle,
			path.Join(sourceDir, name),
			destDir,
			target,
			kind,
			assetName,
		); failure != nil {
			failures = append(failures, *failure)
			continue
		}
		successes = append(successes, OpenCodeAssetSuccessItem{Kind: kind, Name: assetName, Path: target})
	}
	return successes, failures, nil
}

func installOpenCodeAsset(bundle fs.FS, source, destDir, target, kind, name string) *OpenCodeAssetFailureItem {
	if !isPathSafe(destDir, target) {
		return openCodeFailure(kind, name, target, fmt.Errorf("resolved target escapes %s", destDir))
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return openCodeFailure(kind, name, target, fmt.Errorf("prepare opencode install dir %s: %w", destDir, err))
	}
	data, err := fs.ReadFile(bundle, source)
	if err != nil {
		return openCodeFailure(kind, name, target, err)
	}
	if err := os.WriteFile(target, data, 0o600); err != nil {
		return openCodeFailure(kind, name, target, err)
	}
	return nil
}

func openCodeFailure(kind, name, target string, err error) *OpenCodeAssetFailureItem {
	return &OpenCodeAssetFailureItem{Kind: kind, Name: name, Path: target, Error: err.Error()}
}

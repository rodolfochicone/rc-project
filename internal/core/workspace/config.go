package workspace

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	toml "github.com/pelletier/go-toml/v2"
	rcconfig "github.com/rodolfochicone/rc-project/internal/config"
	"github.com/rodolfochicone/rc-project/internal/core/model"
)

var osUserHomeDir = os.UserHomeDir
var discoverWorkspaceRoot = discoverWorkspaceRootFromStart
var discoverCache sync.Map

// maxWorkspaceSearchDepth bounds how far below the start directory the downward
// search for descendant .rc markers descends. It covers common monorepo layouts
// (packages/<pkg>/.rc, apps/<app>/.rc) while preventing a pathological full-tree
// walk when rc runs from a directory with no .rc above it (e.g. $HOME).
const maxWorkspaceSearchDepth = 6

// skipWorkspaceSearchDirs names directories the downward marker search never
// descends into — large dependency trees, VCS metadata, and archive state.
var skipWorkspaceSearchDirs = map[string]struct{}{
	"node_modules": {},
	".git":         {},
	"vendor":       {},
	"_archived":    {},
}

type startDirOverrideKey struct{}

// WithStartDirOverride returns a context that carries a request-scoped workspace
// start directory. Discover honors it only when called with an empty startDir,
// so an explicit startDir argument always wins. The CLI uses this to thread the
// global --workspace flag into discovery without changing every call site.
func WithStartDirOverride(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, startDirOverrideKey{}, strings.TrimSpace(dir))
}

func startDirOverride(ctx context.Context) string {
	value, ok := ctx.Value(startDirOverrideKey{}).(string)
	if !ok {
		return ""
	}
	return value
}

// MultipleWorkspacesError is returned by Discover when no .rc directory exists at
// or above the start directory but several exist beneath it, leaving the target
// workspace ambiguous. The caller must disambiguate by changing into the desired
// subproject or passing --workspace.
type MultipleWorkspacesError struct {
	Start      string
	Candidates []string
}

func (e *MultipleWorkspacesError) Error() string {
	return fmt.Sprintf(
		"multiple .rc workspaces found under %s; cd into the target subproject or pass --workspace <dir>: %s",
		e.Start,
		strings.Join(e.Candidates, ", "),
	)
}

type discoverCacheEntry struct {
	ready chan struct{}
	root  string
	err   error
}

type configPaths struct {
	workspaceRoot string
	globalRoot    string
	workspacePath string
	globalPath    string
	workspaceSeen bool
	globalSeen    bool
}

func Resolve(ctx context.Context, startDir string) (Context, error) {
	root, err := Discover(ctx, startDir)
	if err != nil {
		return Context{}, err
	}

	cfg, paths, err := loadEffectiveConfig(ctx, root)
	if err != nil {
		return Context{}, err
	}

	return Context{
		Root:                root,
		RcDir:               model.RcDir(root),
		ConfigPath:          paths.effectivePath(),
		WorkspaceConfigPath: paths.workspacePath,
		GlobalConfigPath:    paths.globalPath,
		Config:              cfg,
	}, nil
}

func Discover(ctx context.Context, startDir string) (string, error) {
	if err := context.Cause(ctx); err != nil {
		return "", fmt.Errorf("discover workspace: %w", err)
	}

	resolvedStart := strings.TrimSpace(startDir)
	if resolvedStart == "" {
		resolvedStart = startDirOverride(ctx)
	}
	if resolvedStart == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}
		resolvedStart = cwd
	}

	absStart, err := filepath.Abs(resolvedStart)
	if err != nil {
		return "", fmt.Errorf("resolve workspace start dir: %w", err)
	}

	entry := &discoverCacheEntry{ready: make(chan struct{})}
	actual, loaded := discoverCache.LoadOrStore(absStart, entry)
	cachedEntry, ok := actual.(*discoverCacheEntry)
	if !ok || cachedEntry == nil {
		return "", fmt.Errorf("discover workspace: unexpected cache entry %T", actual)
	}
	entry = cachedEntry
	if loaded {
		select {
		case <-entry.ready:
			if entry.err != nil {
				return "", entry.err
			}
			return entry.root, nil
		case <-ctx.Done():
			return "", fmt.Errorf("discover workspace: %w", context.Cause(ctx))
		}
	}

	entry.root, entry.err = discoverWorkspaceRoot(ctx, resolvedStart)
	close(entry.ready)
	if entry.err != nil {
		discoverCache.Delete(absStart)
		return "", entry.err
	}
	return entry.root, nil
}

func discoverWorkspaceRootFromStart(ctx context.Context, startDir string) (string, error) {
	absStart, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolve workspace start dir: %w", err)
	}
	realStart, err := filepath.EvalSymlinks(absStart)
	if err != nil {
		return "", fmt.Errorf("resolve workspace start dir symlinks: %w", err)
	}

	globalMarkerDir, hasGlobalMarker := discoverGlobalWorkspaceMarkerDir()
	current := realStart
	for {
		if err := context.Cause(ctx); err != nil {
			return "", fmt.Errorf("discover workspace: %w", err)
		}

		candidate := filepath.Join(current, model.WorkflowRootDirName)
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			// The home-scoped rc directory stores global runtime/config state.
			// It must not redefine arbitrary paths under HOME as local workspaces.
			if !hasGlobalMarker || !sameWorkspaceMarkerDir(candidate, globalMarkerDir) {
				return current, nil
			}
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("stat workspace marker %s: %w", candidate, err)
		}

		parent := filepath.Dir(current)
		if parent == current {
			return resolveWithoutAncestorMarker(ctx, realStart, globalMarkerDir, hasGlobalMarker)
		}
		current = parent
	}
}

// resolveWithoutAncestorMarker handles the case where no .rc marker exists at or
// above realStart. It searches beneath realStart so that running rc from a
// monorepo root (whose .rc directories live in subprojects) resolves correctly:
// exactly one descendant marker is used automatically, several are reported as a
// MultipleWorkspacesError, and none preserves the legacy fallback to realStart.
func resolveWithoutAncestorMarker(
	ctx context.Context,
	realStart string,
	globalMarkerDir string,
	hasGlobalMarker bool,
) (string, error) {
	candidates, err := findDescendantWorkspaceMarkers(ctx, realStart, globalMarkerDir, hasGlobalMarker)
	if err != nil {
		return "", err
	}
	switch len(candidates) {
	case 0:
		return realStart, nil
	case 1:
		return candidates[0], nil
	default:
		return "", &MultipleWorkspacesError{Start: realStart, Candidates: candidates}
	}
}

// findDescendantWorkspaceMarkers walks beneath startDir collecting the parent
// directory of every .rc marker found, excluding the home-scoped global marker.
// The walk skips dependency/VCS/archive directories, does not descend into found
// markers, and is bounded by maxWorkspaceSearchDepth. Results are sorted for
// deterministic output.
func findDescendantWorkspaceMarkers(
	ctx context.Context,
	startDir string,
	globalMarkerDir string,
	hasGlobalMarker bool,
) ([]string, error) {
	var roots []string
	walkErr := filepath.WalkDir(startDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if cause := context.Cause(ctx); cause != nil {
			return fmt.Errorf("discover workspace: %w", cause)
		}
		if !entry.IsDir() {
			return nil
		}

		if entry.Name() == model.WorkflowRootDirName {
			if !hasGlobalMarker || !sameWorkspaceMarkerDir(path, globalMarkerDir) {
				roots = append(roots, filepath.Dir(path))
			}
			return fs.SkipDir
		}

		if path == startDir {
			return nil
		}
		if _, skip := skipWorkspaceSearchDirs[entry.Name()]; skip {
			return fs.SkipDir
		}
		if workspaceSearchDepth(startDir, path) >= maxWorkspaceSearchDepth {
			return fs.SkipDir
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("search descendant workspace markers: %w", walkErr)
	}
	sort.Strings(roots)
	return roots, nil
}

func workspaceSearchDepth(startDir string, path string) int {
	rel, err := filepath.Rel(startDir, path)
	if err != nil || rel == "." {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator)) + 1
}

func discoverGlobalWorkspaceMarkerDir() (string, bool) {
	homeDir, err := osUserHomeDir()
	if err != nil {
		return "", false
	}
	resolvedHomeDir, err := resolveConfigBaseDir(homeDir)
	if err != nil {
		return "", false
	}

	markerDir := filepath.Join(resolvedHomeDir, model.WorkflowRootDirName)
	resolvedMarkerDir, err := filepath.EvalSymlinks(markerDir)
	if err == nil {
		return filepath.Clean(resolvedMarkerDir), true
	}
	return filepath.Clean(markerDir), true
}

func sameWorkspaceMarkerDir(left string, right string) bool {
	left = canonicalWorkspaceMarkerDir(left)
	right = canonicalWorkspaceMarkerDir(right)
	if left == "" || right == "" {
		return false
	}
	return left == right
}

func canonicalWorkspaceMarkerDir(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return filepath.Clean(resolved)
	}
	return filepath.Clean(path)
}

func LoadConfig(ctx context.Context, workspaceRoot string) (ProjectConfig, string, error) {
	cfg, paths, err := loadEffectiveConfig(ctx, workspaceRoot)
	if err != nil {
		return ProjectConfig{}, "", err
	}
	return cfg, paths.effectivePath(), nil
}

func LoadGlobalConfig(ctx context.Context) (ProjectConfig, string, error) {
	if err := context.Cause(ctx); err != nil {
		return ProjectConfig{}, "", fmt.Errorf("load global config: %w", err)
	}

	paths, err := resolveConfigPaths(".")
	if err != nil {
		return ProjectConfig{}, "", fmt.Errorf("resolve global config paths: %w", err)
	}

	cfg, _, err := loadConfigFile(ctx, paths.globalPath, globalConfigScope, paths.globalRoot)
	if err != nil {
		return ProjectConfig{}, "", err
	}
	return cfg, paths.globalPath, nil
}

// GlobalConfigPath returns the on-disk path for the global (~/.rc/config.toml)
// config file using the same home-directory resolver as LoadGlobalConfig.
// This is the single source of truth for callers that need the path for
// atomic writes — using any other home-dir lookup can diverge from reads.
func GlobalConfigPath() (string, error) {
	paths, err := resolveConfigPaths(".")
	if err != nil {
		return "", fmt.Errorf("resolve global config path: %w", err)
	}
	return paths.globalPath, nil
}

func loadEffectiveConfig(ctx context.Context, workspaceRoot string) (ProjectConfig, configPaths, error) {
	if err := context.Cause(ctx); err != nil {
		return ProjectConfig{}, configPaths{}, fmt.Errorf("load workspace config: %w", err)
	}

	paths, err := resolveConfigPaths(workspaceRoot)
	if err != nil {
		return ProjectConfig{}, configPaths{}, fmt.Errorf("resolve config paths: %w", err)
	}

	globalCfg, globalSeen, err := loadConfigFile(ctx, paths.globalPath, globalConfigScope, paths.globalRoot)
	if err != nil {
		return ProjectConfig{}, configPaths{}, err
	}
	workspaceCfg, workspaceSeen, err := loadConfigFile(
		ctx,
		paths.workspacePath,
		workspaceConfigScope,
		paths.workspaceRoot,
	)
	if err != nil {
		return ProjectConfig{}, configPaths{}, err
	}

	paths.globalSeen = globalSeen
	paths.workspaceSeen = workspaceSeen

	cfg := buildEffectiveProjectConfig(globalCfg, workspaceCfg)
	if err := cfg.validate(effectiveConfigScope); err != nil {
		return ProjectConfig{}, configPaths{}, err
	}
	return cfg, paths, nil
}

func resolveConfigPaths(workspaceRoot string) (configPaths, error) {
	paths := configPaths{
		workspaceRoot: workspaceRoot,
		workspacePath: model.ConfigPathForWorkspace(workspaceRoot),
	}

	homeDir, err := osUserHomeDir()
	if err != nil {
		return configPaths{}, fmt.Errorf("lookup user home directory: %w", err)
	}
	resolvedHomeDir, err := resolveConfigBaseDir(homeDir)
	if err != nil {
		return configPaths{}, fmt.Errorf("resolve global config base dir: %w", err)
	}

	homePaths, err := rcconfig.ResolveHomePathsFrom(filepath.Join(resolvedHomeDir, rcconfig.DirName))
	if err != nil {
		return configPaths{}, fmt.Errorf("resolve global config base dir: %w", err)
	}

	paths.globalRoot = resolvedHomeDir
	paths.globalPath = homePaths.ConfigFile
	return paths, nil
}

func loadConfigFile(
	ctx context.Context,
	configPath string,
	scope string,
	baseDir string,
) (ProjectConfig, bool, error) {
	if err := context.Cause(ctx); err != nil {
		return ProjectConfig{}, false, fmt.Errorf("load %s: %w", scope, err)
	}
	if strings.TrimSpace(configPath) == "" {
		return ProjectConfig{}, false, nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ProjectConfig{}, false, nil
		}
		return ProjectConfig{}, false, fmt.Errorf("read %s: %w", scope, err)
	}
	if err := rejectLegacyConfigSections(content, scope); err != nil {
		return ProjectConfig{}, true, err
	}

	var cfg ProjectConfig
	decoder := toml.NewDecoder(bytes.NewReader(content)).DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return ProjectConfig{}, true, fmt.Errorf("decode %s: %w", scope, err)
	}

	cfg, err = normalizeProjectConfigPaths(cfg, baseDir)
	if err != nil {
		return ProjectConfig{}, true, fmt.Errorf("normalize %s: %w", scope, err)
	}
	if err := cfg.validate(scope); err != nil {
		return ProjectConfig{}, true, err
	}
	return cfg, true, nil
}

func rejectLegacyConfigSections(content []byte, scope string) error {
	var raw map[string]any
	if err := toml.Unmarshal(content, &raw); err != nil {
		return nil
	}
	if _, ok := raw["start"]; ok {
		return fmt.Errorf("%s section [start] was removed; use [tasks.run] instead", scope)
	}
	if _, ok := raw["dev"]; ok {
		return fmt.Errorf("%s section [dev] was removed with the rc dev command; remove it from your config", scope)
	}
	if _, ok := raw["accounts"]; ok {
		return fmt.Errorf(
			"%s section [accounts] was removed with the rc dev command; remove it from your config",
			scope,
		)
	}
	return nil
}

func (p configPaths) effectivePath() string {
	if p.workspaceSeen {
		return p.workspacePath
	}
	if p.globalSeen {
		return p.globalPath
	}
	return p.workspacePath
}

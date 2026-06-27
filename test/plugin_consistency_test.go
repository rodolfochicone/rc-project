package test

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rodolfochicone/rc-project/commands"
	"github.com/rodolfochicone/rc-project/skills"
)

// pluginManifest mirrors the subset of .claude-plugin/plugin.json the
// distribution contract depends on. It lives only in the test package: the
// manifests are static assets, not a runtime type.
type pluginManifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      struct {
		Name string `json:"name"`
	} `json:"author"`
}

// marketplaceManifest mirrors the subset of .claude-plugin/marketplace.json
// that the single-plugin marketplace contract (ADR-001) depends on.
type marketplaceManifest struct {
	Name  string `json:"name"`
	Owner struct {
		Name string `json:"name"`
	} `json:"owner"`
	Plugins []struct {
		Name        string `json:"name"`
		Source      string `json:"source"`
		Version     string `json:"version"`
		Description string `json:"description"`
	} `json:"plugins"`
}

func readPluginManifest(t *testing.T) pluginManifest {
	t.Helper()
	path := filepath.Join(repoRoot(t), ".claude-plugin", "plugin.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}
	var m pluginManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("plugin.json is not valid JSON: %v", err)
	}
	return m
}

func readMarketplaceManifest(t *testing.T) marketplaceManifest {
	t.Helper()
	path := filepath.Join(repoRoot(t), ".claude-plugin", "marketplace.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read marketplace.json: %v", err)
	}
	var m marketplaceManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("marketplace.json is not valid JSON: %v", err)
	}
	return m
}

func TestPluginManifestInvariants(t *testing.T) {
	t.Parallel()

	plugin := readPluginManifest(t)
	market := readMarketplaceManifest(t)

	if len(market.Plugins) != 1 {
		t.Fatalf("marketplace must list exactly one plugin, got %d", len(market.Plugins))
	}
	entry := market.Plugins[0]

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"plugin name", plugin.Name, "rc"},
		{"marketplace name", market.Name, "rc-project"},
		{"marketplace plugin name", entry.Name, "rc"},
		{"marketplace plugin source", entry.Source, "."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}

	nonEmpty := []struct {
		name string
		got  string
	}{
		{"plugin version", plugin.Version},
		{"plugin description", plugin.Description},
		{"plugin author name", plugin.Author.Name},
		{"marketplace owner name", market.Owner.Name},
		{"marketplace plugin version", entry.Version},
	}
	for _, tt := range nonEmpty {
		t.Run(tt.name+" non-empty", func(t *testing.T) {
			t.Parallel()
			if strings.TrimSpace(tt.got) == "" {
				t.Errorf("%s must not be empty", tt.name)
			}
		})
	}
}

// TestManifestVersionsAgree guards the release contract: task_02 rewrites both
// manifests, and the consistency invariant is that they stay in lockstep.
func TestManifestVersionsAgree(t *testing.T) {
	t.Parallel()

	plugin := readPluginManifest(t)
	market := readMarketplaceManifest(t)
	if len(market.Plugins) != 1 {
		t.Fatalf("marketplace must list exactly one plugin, got %d", len(market.Plugins))
	}
	if plugin.Version != market.Plugins[0].Version {
		t.Errorf("manifest versions disagree: plugin.json=%q marketplace.json=%q",
			plugin.Version, market.Plugins[0].Version)
	}
}

// TestPluginSkillsCohereWithEmbeddedAssets asserts the plugin ships every
// bundled skill: each top-level skills.FS directory must expose a SKILL.md so
// Claude Code can discover it under the rc: namespace.
func TestPluginSkillsCohereWithEmbeddedAssets(t *testing.T) {
	t.Parallel()

	entries, err := fs.ReadDir(skills.FS, ".")
	if err != nil {
		t.Fatalf("read embedded skills root: %v", err)
	}

	var dirs int
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirs++
		dir := e.Name()
		t.Run(dir, func(t *testing.T) {
			t.Parallel()
			if !strings.HasPrefix(dir, "rc") {
				t.Errorf("embedded skill %q does not carry the rc prefix", dir)
			}
			if _, err := fs.Stat(skills.FS, dir+"/SKILL.md"); err != nil {
				t.Errorf("embedded skill %q is missing SKILL.md: %v", dir, err)
			}
		})
	}
	if dirs == 0 {
		t.Fatal("no embedded skills found; plugin would ship empty")
	}
}

// TestPluginCommandsCohereWithEmbeddedAssets asserts the plugin ships the
// expected rc-* slash commands and that every embedded command file keeps the
// rc- prefix and .md extension.
func TestPluginCommandsCohereWithEmbeddedAssets(t *testing.T) {
	t.Parallel()

	entries, err := fs.ReadDir(commands.FS, ".")
	if err != nil {
		t.Fatalf("read embedded commands root: %v", err)
	}

	present := make(map[string]bool, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		present[name] = true
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if !strings.HasPrefix(name, "rc-") || !strings.HasSuffix(name, ".md") {
				t.Errorf("embedded command %q must match rc-*.md", name)
			}
		})
	}

	for _, want := range []string{"rc-exec.md", "rc-review.md"} {
		if !present[want] {
			t.Errorf("expected embedded command %q is missing", want)
		}
	}
}

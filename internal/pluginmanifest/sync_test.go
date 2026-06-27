package pluginmanifest

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const (
	validPlugin = `{
  "name": "rc",
  "version": "0.0.0",
  "description": "rc workflow skills and commands",
  "author": { "name": "Escale" }
}
`
	validMarketplace = `{
  "name": "rc-project",
  "owner": { "name": "Escale" },
  "plugins": [
    { "name": "rc", "source": ".", "version": "0.0.0", "description": "rc plugin" }
  ]
}
`
)

// writeManifests seeds a temp .claude-plugin dir and returns its path plus the
// two manifest paths.
func writeManifests(t *testing.T, plugin, marketplace string) (dir, pluginPath, marketPath string) {
	t.Helper()
	dir = t.TempDir()
	pluginPath = filepath.Join(dir, "plugin.json")
	marketPath = filepath.Join(dir, "marketplace.json")
	if plugin != "" {
		if err := os.WriteFile(pluginPath, []byte(plugin), 0o600); err != nil {
			t.Fatalf("seed plugin.json: %v", err)
		}
	}
	if marketplace != "" {
		if err := os.WriteFile(marketPath, []byte(marketplace), 0o600); err != nil {
			t.Fatalf("seed marketplace.json: %v", err)
		}
	}
	return dir, pluginPath, marketPath
}

func readVersions(t *testing.T, pluginPath, marketPath string) (pluginVer string, marketVers []string) {
	t.Helper()
	var p pluginDoc
	raw, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("decode plugin.json: %v", err)
	}
	var m marketplaceDoc
	raw, err = os.ReadFile(marketPath)
	if err != nil {
		t.Fatalf("read marketplace.json: %v", err)
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decode marketplace.json: %v", err)
	}
	vers := make([]string, len(m.Plugins))
	for i, pl := range m.Plugins {
		vers[i] = pl.Version
	}
	return p.Version, vers
}

func TestNormalizeVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain semver", "1.2.3", "1.2.3"},
		{"leading v stripped", "v1.2.3", "1.2.3"},
		{"surrounding whitespace", "  v2.0.0\n", "2.0.0"},
		{"snapshot form", "0.0.0-next-abcdef", "0.0.0-next-abcdef"},
		{"empty", "", ""},
		{"only v", "v", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeVersion(tt.in); got != tt.want {
				t.Errorf("NormalizeVersion(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSyncRewritesBothManifests(t *testing.T) {
	t.Parallel()
	dir, pluginPath, marketPath := writeManifests(t, validPlugin, validMarketplace)

	if err := Sync(dir, "1.2.3"); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	pluginVer, marketVers := readVersions(t, pluginPath, marketPath)
	if pluginVer != "1.2.3" {
		t.Errorf("plugin version = %q, want 1.2.3", pluginVer)
	}
	for _, v := range marketVers {
		if v != "1.2.3" {
			t.Errorf("marketplace plugin version = %q, want 1.2.3", v)
		}
	}
}

func TestSyncNormalizesLeadingV(t *testing.T) {
	t.Parallel()
	dir, pluginPath, marketPath := writeManifests(t, validPlugin, validMarketplace)

	if err := Sync(dir, "v4.5.6"); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	pluginVer, marketVers := readVersions(t, pluginPath, marketPath)
	if pluginVer != "4.5.6" {
		t.Errorf("plugin version = %q, want 4.5.6 (leading v stripped)", pluginVer)
	}
	if len(marketVers) == 0 || marketVers[0] != "4.5.6" {
		t.Errorf("marketplace versions = %v, want [4.5.6]", marketVers)
	}
}

func TestSyncPreservesOtherFields(t *testing.T) {
	t.Parallel()
	dir, pluginPath, marketPath := writeManifests(t, validPlugin, validMarketplace)

	if err := Sync(dir, "9.9.9"); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	var p pluginDoc
	raw, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("decode plugin.json: %v", err)
	}
	if p.Name != "rc" || p.Description == "" || p.Author.Name != "Escale" {
		t.Errorf("Sync mutated non-version fields: %+v", p)
	}

	var m marketplaceDoc
	raw, err = os.ReadFile(marketPath)
	if err != nil {
		t.Fatalf("read marketplace.json: %v", err)
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decode marketplace.json: %v", err)
	}
	if m.Name != "rc-project" || m.Owner.Name != "Escale" || m.Plugins[0].Source != "." {
		t.Errorf("Sync mutated non-version fields: %+v", m)
	}
}

func TestSyncIsIdempotent(t *testing.T) {
	t.Parallel()
	dir, pluginPath, marketPath := writeManifests(t, validPlugin, validMarketplace)

	if err := Sync(dir, "2.0.0"); err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	firstPlugin, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}
	firstMarket, err := os.ReadFile(marketPath)
	if err != nil {
		t.Fatalf("read marketplace.json: %v", err)
	}

	if err := Sync(dir, "2.0.0"); err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	secondPlugin, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}
	secondMarket, err := os.ReadFile(marketPath)
	if err != nil {
		t.Fatalf("read marketplace.json: %v", err)
	}

	if !bytes.Equal(firstPlugin, secondPlugin) {
		t.Errorf("plugin.json not idempotent:\nfirst:\n%s\nsecond:\n%s", firstPlugin, secondPlugin)
	}
	if !bytes.Equal(firstMarket, secondMarket) {
		t.Errorf("marketplace.json not idempotent:\nfirst:\n%s\nsecond:\n%s", firstMarket, secondMarket)
	}
}

func TestSyncErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		plugin      string
		marketplace string
		version     string
		seedPlugin  bool
	}{
		{name: "empty version", plugin: validPlugin, marketplace: validMarketplace, version: "", seedPlugin: true},
		{name: "version only v", plugin: validPlugin, marketplace: validMarketplace, version: "v", seedPlugin: true},
		{
			name:        "missing plugin manifest",
			plugin:      "",
			marketplace: validMarketplace,
			version:     "1.0.0",
			seedPlugin:  false,
		},
		{
			name:        "malformed plugin manifest",
			plugin:      "{ not json",
			marketplace: validMarketplace,
			version:     "1.0.0",
			seedPlugin:  true,
		},
		{
			name:        "malformed marketplace manifest",
			plugin:      validPlugin,
			marketplace: "{ not json",
			version:     "1.0.0",
			seedPlugin:  true,
		},
		{
			name:        "marketplace without plugins",
			plugin:      validPlugin,
			marketplace: `{"name":"rc-project","owner":{"name":"Escale"},"plugins":[]}`,
			version:     "1.0.0",
			seedPlugin:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir, pluginPath, _ := writeManifests(t, tt.plugin, tt.marketplace)

			err := Sync(dir, tt.version)
			if err == nil {
				t.Fatalf("Sync(%q) = nil, want error", tt.version)
			}

			// A malformed marketplace must not leave a partially-written
			// plugin.json: the original bytes must be intact.
			if tt.name == "malformed marketplace manifest" {
				got, readErr := os.ReadFile(pluginPath)
				if readErr != nil {
					t.Fatalf("read plugin.json: %v", readErr)
				}
				if string(got) != validPlugin {
					t.Errorf("plugin.json was partially written on failure:\n%s", got)
				}
			}
		})
	}
}

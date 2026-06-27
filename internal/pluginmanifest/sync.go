// Package pluginmanifest keeps the Claude Code plugin manifests
// (.claude-plugin/plugin.json and marketplace.json) in lockstep with the
// release tag. The OSS GoReleaser flow invokes Sync so a forgotten manual bump
// cannot silently stop users from receiving plugin updates.
package pluginmanifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// author mirrors the plugin.json author object.
type author struct {
	Name string `json:"name"`
}

// pluginDoc models .claude-plugin/plugin.json. Field order matches the
// committed manifest so a rewrite produces a stable, low-noise diff.
type pluginDoc struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      author `json:"author"`
}

// owner mirrors the marketplace.json owner object.
type owner struct {
	Name string `json:"name"`
}

// marketplacePlugin models one entry of marketplace.json's plugins array.
type marketplacePlugin struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// marketplaceDoc models .claude-plugin/marketplace.json.
type marketplaceDoc struct {
	Name    string              `json:"name"`
	Owner   owner               `json:"owner"`
	Plugins []marketplacePlugin `json:"plugins"`
}

// NormalizeVersion trims surrounding whitespace and strips a single leading
// "v" so a release tag such as "v1.2.3" becomes the "1.2.3" form the manifests
// carry.
func NormalizeVersion(raw string) string {
	return strings.TrimPrefix(strings.TrimSpace(raw), "v")
}

// Sync rewrites the version field in both manifests under pluginDir to the
// normalized form of rawVersion, preserving every other field. It parses and
// validates both manifests before writing either, so a malformed manifest
// aborts the release without a partial write. Running it twice with the same
// version is idempotent.
func Sync(pluginDir, rawVersion string) error {
	version := NormalizeVersion(rawVersion)
	if version == "" {
		return fmt.Errorf("target version must not be empty (got %q)", rawVersion)
	}

	pluginPath := filepath.Join(pluginDir, "plugin.json")
	marketPath := filepath.Join(pluginDir, "marketplace.json")

	var plugin pluginDoc
	if err := readManifest(pluginPath, &plugin); err != nil {
		return err
	}
	var market marketplaceDoc
	if err := readManifest(marketPath, &market); err != nil {
		return err
	}
	if len(market.Plugins) == 0 {
		return fmt.Errorf("%s: marketplace lists no plugins", marketPath)
	}

	plugin.Version = version
	for i := range market.Plugins {
		market.Plugins[i].Version = version
	}

	if err := writeManifest(pluginPath, plugin); err != nil {
		return err
	}
	return writeManifest(marketPath, market)
}

// readManifest decodes a manifest file, failing loudly when it is missing or
// not valid JSON.
func readManifest(path string, dst any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read manifest %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("parse manifest %s: %w", path, err)
	}
	return nil
}

// writeManifest serializes a manifest with 2-space indentation and a trailing
// newline, matching the committed formatting.
func writeManifest(path string, doc any) error {
	encoded, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("encode manifest %s: %w", path, err)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		return fmt.Errorf("write manifest %s: %w", path, err)
	}
	return nil
}

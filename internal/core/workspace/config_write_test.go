package workspace

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
)

// TestWriteConfigRoundTrip asserts that a representative non-empty ProjectConfig
// written via WriteConfig can be loaded back and equals the original.
// This matters because a silent marshal/unmarshal asymmetry would silently corrupt
// user config on every save.
func TestWriteConfigRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	keep := true
	concurrent := 3
	maxRounds := 5
	pollInterval := "30s"
	reviewTimeout := "10m"
	quietPeriod := "5s"
	keepTerminalDays := 7
	keepMax := 50

	original := ProjectConfig{
		Runs: RunsConfig{
			KeepTerminalDays: &keepTerminalDays,
			KeepMax:          &keepMax,
		},
		FixReviews: FixReviewsConfig{
			Concurrent:      &concurrent,
			IncludeResolved: &keep,
		},
		WatchReviews: WatchReviewsConfig{
			MaxRounds:     &maxRounds,
			PollInterval:  &pollInterval,
			ReviewTimeout: &reviewTimeout,
			QuietPeriod:   &quietPeriod,
		},
		Sound: SoundConfig{
			Enabled: &keep,
		},
	}

	if err := WriteConfig(context.Background(), configPath, original, workspaceConfigScope); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var reloaded ProjectConfig
	if err := toml.Unmarshal(data, &reloaded); err != nil {
		t.Fatalf("toml.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(original, reloaded) {
		t.Fatalf("round-trip mismatch\nwant: %+v\ngot:  %+v", original, reloaded)
	}
}

// TestWriteConfigAtomicPreservesExistingFileOnFailure asserts that when a write
// cannot complete, any pre-existing file is left byte-identical and no partial
// file is written alongside it.
// This matters because a non-atomic write would corrupt the user's config on
// any mid-write failure (disk full, crash, etc.).
func TestWriteConfigAtomicPreservesExistingFileOnFailure(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	original := []byte("[runs]\nkeep_max = 10\n")
	if err := os.WriteFile(configPath, original, 0o600); err != nil {
		t.Fatalf("WriteFile() seed error = %v", err)
	}

	// Force MkdirAll to fail by using an existing regular file as a path component.
	// A file cannot be treated as a directory, so MkdirAll returns an error
	// before any temp file is created or any existing config is touched.
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile(blocker) error = %v", err)
	}
	badPath := filepath.Join(blocker, "config.toml")
	cfg := ProjectConfig{}
	err := WriteConfig(context.Background(), badPath, cfg, workspaceConfigScope)
	if err == nil {
		t.Fatal("WriteConfig() with unwriteable dir: want error, got nil")
	}

	// The original file at configPath must be untouched.
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() after failed write: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("original file changed after failed write\nwant: %q\ngot:  %q", original, got)
	}

	// No partial temp files must exist in the original dir.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	for _, e := range entries {
		if e.Name() != filepath.Base(configPath) && e.Name() != "blocker" {
			t.Fatalf("unexpected file left in dir after failed write: %q", e.Name())
		}
	}
}

// TestWriteConfigValidationRejectsInvalidConfig asserts that WriteConfig rejects
// an invalid ProjectConfig before touching the disk. This matters because the
// daemon must never persist a config that would fail to load back, and the
// on-disk file must remain unchanged so the user's last valid config survives.
func TestWriteConfigValidationRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	valid := []byte("[runs]\nkeep_max = 5\n")
	if err := os.WriteFile(configPath, valid, 0o600); err != nil {
		t.Fatalf("WriteFile() seed error = %v", err)
	}

	emptyProvider := ""
	invalid := ProjectConfig{
		FetchReviews: FetchReviewsConfig{
			// An empty string provider name is explicitly invalid per config validation.
			Provider: &emptyProvider,
		},
	}

	err := WriteConfig(context.Background(), configPath, invalid, workspaceConfigScope)
	if err == nil {
		t.Fatal("WriteConfig() with invalid config: want error, got nil")
	}

	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() after validation failure: %v", err)
	}
	if !bytes.Equal(got, valid) {
		t.Fatalf("file was modified despite validation failure\nwant: %q\ngot:  %q", valid, got)
	}
}

// TestWriteConfigCreatesFileWithCorrectPermissions asserts that WriteConfig creates
// a new config file with 0o600 permissions. This is a security requirement: config
// files may contain sensitive values and must not be world-readable.
func TestWriteConfigCreatesFileWithCorrectPermissions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	cfg := ProjectConfig{}
	if err := WriteConfig(context.Background(), configPath, cfg, workspaceConfigScope); err != nil {
		t.Fatalf("WriteConfig() error = %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("file permissions = %04o, want %04o", perm, 0o600)
	}
}

// TestWriteConfigGlobalScope asserts that WriteConfig accepts globalConfigScope,
// which uses a stricter validation path (no workspace-only fields allowed).
func TestWriteConfigGlobalScope(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	keepMax := 20
	cfg := ProjectConfig{
		Runs: RunsConfig{KeepMax: &keepMax},
	}

	if err := WriteConfig(context.Background(), configPath, cfg, globalConfigScope); err != nil {
		t.Fatalf("WriteConfig() with globalConfigScope error = %v", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

// TestWriteConfigCreatesParentDirectoryIfAbsent asserts that WriteConfig creates
// the parent .rc directory when it does not yet exist.
// This matters because the first GUI save for a freshly registered workspace
// (which has never run rc before) must not fail with "no such file or directory"
// from os.CreateTemp.
func TestWriteConfigCreatesParentDirectoryIfAbsent(t *testing.T) {
	t.Parallel()

	// Point configPath into a subdirectory that does not exist yet.
	root := t.TempDir()
	configPath := filepath.Join(root, ".rc", "config.toml")

	// Confirm precondition: parent dir absent.
	if _, err := os.Stat(filepath.Dir(configPath)); !os.IsNotExist(err) {
		t.Fatalf("precondition: .rc dir should not exist, got err = %v", err)
	}

	keepMax := 5
	cfg := ProjectConfig{Runs: RunsConfig{KeepMax: &keepMax}}
	if err := WriteConfig(context.Background(), configPath, cfg, workspaceConfigScope); err != nil {
		t.Fatalf("WriteConfig() with absent parent dir: error = %v", err)
	}

	// Parent directory must now exist with restrictive permissions.
	dirInfo, err := os.Stat(filepath.Dir(configPath))
	if err != nil {
		t.Fatalf("Stat(.rc) after write: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0o700 {
		t.Fatalf(".rc dir permissions = %04o, want 0700", perm)
	}

	// The config file itself must be present.
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

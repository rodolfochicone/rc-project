package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// WriteConfig validates cfg, marshals it to TOML, and writes it atomically to
// configPath using a temp-file + fsync + rename pattern. The scope parameter
// controls which validation rules are applied (globalConfigScope vs
// workspaceConfigScope). The original file is left untouched on any failure
// before the final rename.
func WriteConfig(_ context.Context, configPath string, cfg ProjectConfig, scope string) error {
	if err := cfg.validate(scope); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	payload, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	dir := filepath.Dir(configPath)
	base := filepath.Base(configPath)

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, base+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp config file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp config file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync temp config file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp config file: %w", err)
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		return fmt.Errorf("chmod temp config file: %w", err)
	}
	if err := os.Rename(tmpPath, configPath); err != nil {
		return fmt.Errorf("rename temp config file: %w", err)
	}

	return syncConfigDir(dir)
}

func syncConfigDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config directory %q for sync: %w", path, err)
	}
	defer func() {
		_ = dir.Close()
	}()
	if err := dir.Sync(); err != nil {
		return fmt.Errorf("sync config directory %q: %w", path, err)
	}
	return nil
}

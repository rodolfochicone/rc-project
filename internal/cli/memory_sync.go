package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/rodolfochicone/rc-project/internal/core/model"
	"github.com/rodolfochicone/rc-project/internal/core/projectmemory"
	"github.com/spf13/cobra"
)

// projectMemoryMirrorDir resolves the committed text-mirror directory (<.rc base>/memory) for
// the current workspace. The mirror is the shareable source of truth that `export` writes and
// `import` reads; the SQLite database is a local cache rebuilt from it.
func projectMemoryMirrorDir(ctx context.Context) (string, error) {
	root, err := discoverWorkspaceRoot(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Join(model.RcDir(root), projectmemory.MirrorDirName), nil
}

func newMemoryExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "export",
		Short:        "Write every memory to the committed text mirror (.rc/memory/)",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Write one markdown-per-fact file to .rc/memory/ for every stored memory.

The mirror is committed to git and shared across machines; .rc/memory.db stays a local cache.
Export overwrites mirror files but never deletes them. Commit the mirror, then teammates run
` + "`rc memory import`" + ` to load it.`,
		Example: `  rc memory export`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, stop := signalCommandContext(cmd)
			defer stop()

			mirrorDir, err := projectMemoryMirrorDir(ctx)
			if err != nil {
				return withExitCode(2, err)
			}

			st, _, err := openProjectMemory(ctx)
			if err != nil {
				return withExitCode(2, err)
			}
			defer closeProjectMemory(ctx, st)

			memories, err := st.List(ctx, projectmemory.ListFilter{})
			if err != nil {
				return mapMemoryError(err)
			}

			if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
				return withExitCode(2, fmt.Errorf("create mirror directory: %w", err))
			}
			// Refuse to silently overwrite: two distinct memories whose names collide (e.g. keys
			// differing only by case or punctuation) would otherwise drop one record from the mirror.
			writtenBy := make(map[string]string, len(memories))
			for i := range memories {
				name := projectmemory.MirrorFileName(memories[i])
				if prevID, clash := writtenBy[name]; clash {
					return withExitCode(1, fmt.Errorf(
						"mirror file name collision: memories %q and %q both map to %q; disambiguate their keys",
						prevID, memories[i].ID, name,
					))
				}
				writtenBy[name] = memories[i].ID

				data, marshalErr := projectmemory.MarshalMemory(memories[i])
				if marshalErr != nil {
					return withExitCode(2, marshalErr)
				}
				path := filepath.Join(mirrorDir, name)
				if writeErr := os.WriteFile(path, data, 0o600); writeErr != nil {
					return withExitCode(2, fmt.Errorf("write mirror file %q: %w", path, writeErr))
				}
			}
			return writeMemoryText(cmd, fmt.Sprintf("Exported %d memories to %s\n", len(memories), mirrorDir))
		},
	}
	return cmd
}

// skippedMirrorFile records a mirror file that could not be read or parsed, so import can
// report it without aborting the rest of the batch.
type skippedMirrorFile struct {
	path string
	err  error
}

func newMemoryImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "import",
		Short:        "Load memories from the committed text mirror (.rc/memory/) into the database",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Read every markdown file under .rc/memory/ and upsert it into the local database.

Import is most-recent-wins by updated_at and never deletes rows. A malformed file is reported
and skipped while the rest still import; skipped files make the command exit non-zero.`,
		Example: `  rc memory import`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, stop := signalCommandContext(cmd)
			defer stop()

			mirrorDir, err := projectMemoryMirrorDir(ctx)
			if err != nil {
				return withExitCode(2, err)
			}

			records, skipped, err := readMirrorRecords(mirrorDir)
			if err != nil {
				return withExitCode(2, err)
			}

			st, _, err := openProjectMemory(ctx)
			if err != nil {
				return withExitCode(2, err)
			}
			defer closeProjectMemory(ctx, st)

			result, err := st.Import(ctx, records)
			if err != nil {
				return mapMemoryError(err)
			}

			for i := range skipped {
				slog.Default().Warn("skipped malformed mirror file", "file", skipped[i].path, "error", skipped[i].err)
			}

			summary := fmt.Sprintf(
				"Imported from %s: added=%d updated=%d skipped=%d malformed=%d\n",
				mirrorDir, result.Added, result.Updated, result.Skipped, len(skipped),
			)
			if writeErr := writeMemoryText(cmd, summary); writeErr != nil {
				return writeErr
			}
			if len(skipped) > 0 {
				return withExitCode(1, fmt.Errorf("%d malformed mirror file(s) skipped", len(skipped)))
			}
			return nil
		},
	}
	return cmd
}

// readMirrorRecords parses every *.md file in mirrorDir into a memory. Unreadable or malformed
// files are collected as skipped rather than failing the whole read. A missing directory is
// treated as an empty mirror.
func readMirrorRecords(mirrorDir string) ([]projectmemory.Memory, []skippedMirrorFile, error) {
	entries, err := os.ReadDir(mirrorDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read mirror directory: %w", err)
	}

	var (
		records []projectmemory.Memory
		skipped []skippedMirrorFile
	)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(mirrorDir, entry.Name())
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			skipped = append(skipped, skippedMirrorFile{path: path, err: readErr})
			continue
		}
		memory, parseErr := projectmemory.ParseMemory(data)
		if parseErr != nil {
			skipped = append(skipped, skippedMirrorFile{path: path, err: parseErr})
			continue
		}
		records = append(records, memory)
	}
	return records, skipped, nil
}

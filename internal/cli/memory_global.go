package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	rcconfig "github.com/rodolfochicone/rc-project/internal/config"
	"github.com/rodolfochicone/rc-project/internal/core/model"
	"github.com/rodolfochicone/rc-project/internal/core/projectmemory"
	"github.com/spf13/cobra"
)

// defaultMemorySearchLimit mirrors the project-memory store's own default and bounds a merged
// workspace+global search when the caller passes no limit.
const defaultMemorySearchLimit = 8

// globalMemoryDBPath returns ~/.rc/memory.db — the home-scoped store shared across every
// workspace, one scope up from each per-project <root>/.rc/memory.db.
func globalMemoryDBPath() (string, error) {
	home, err := rcconfig.ResolveHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, projectmemory.DBFileName), nil
}

// openMemoryScope opens the global store when global is set, otherwise the workspace store.
func openMemoryScope(ctx context.Context, global bool) (*projectmemory.Store, projectmemory.Retriever, error) {
	if global {
		return openGlobalMemory(ctx, true)
	}
	return openProjectMemory(ctx)
}

// openGlobalMemory opens the home-scoped memory store. With createIfMissing false it returns
// (nil, nil, nil) when the database does not yet exist, so read paths can skip the global tier
// without materializing an empty database.
func openGlobalMemory(
	ctx context.Context,
	createIfMissing bool,
) (*projectmemory.Store, projectmemory.Retriever, error) {
	dbPath, err := globalMemoryDBPath()
	if err != nil {
		return nil, nil, err
	}
	if !createIfMissing {
		if _, statErr := os.Stat(dbPath); errors.Is(statErr, fs.ErrNotExist) {
			return nil, nil, nil
		} else if statErr != nil {
			return nil, nil, fmt.Errorf("stat global memory %q: %w", dbPath, statErr)
		}
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, nil, fmt.Errorf("prepare global memory dir: %w", err)
	}
	return openMemoryAt(ctx, dbPath)
}

// memoryGlobalFlag reports whether the persistent --global flag is set on the memory command.
func memoryGlobalFlag(cmd *cobra.Command) bool {
	global, err := cmd.Flags().GetBool("global")
	return err == nil && global
}

// searchMemories runs a scoped search. With --global it queries only the global store; by
// default it merges workspace results with global ones so promoted, cross-project facts surface
// in every workspace.
func searchMemories(
	ctx context.Context,
	cmd *cobra.Command,
	query projectmemory.SearchQuery,
) ([]projectmemory.SearchHit, error) {
	if memoryGlobalFlag(cmd) {
		st, retriever, err := openGlobalMemory(ctx, true)
		if err != nil {
			return nil, err
		}
		defer closeProjectMemory(ctx, st)
		return retriever.Search(ctx, query)
	}

	workspaceStore, workspaceRetriever, err := openProjectMemory(ctx)
	if err != nil {
		return nil, err
	}
	defer closeProjectMemory(ctx, workspaceStore)
	workspaceHits, err := workspaceRetriever.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	globalStore, globalRetriever, err := openGlobalMemory(ctx, false)
	if err != nil {
		return nil, err
	}
	if globalStore == nil {
		return workspaceHits, nil
	}
	defer closeProjectMemory(ctx, globalStore)
	globalHits, err := globalRetriever.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	return mergeSearchHits(workspaceHits, globalHits, query.Limit), nil
}

// mergeSearchHits combines primary (workspace) and secondary (global) hits, de-duplicating so a
// fact promoted from this workspace is not shown twice: primary wins on a (scope, key) or
// (scope, title) collision. Results are ordered by ascending BM25 score and capped at limit.
func mergeSearchHits(primary, secondary []projectmemory.SearchHit, limit int) []projectmemory.SearchHit {
	if limit <= 0 {
		limit = defaultMemorySearchLimit
	}
	seen := make(map[string]struct{}, len(primary)+len(secondary))
	merged := make([]projectmemory.SearchHit, 0, len(primary)+len(secondary))
	for _, group := range [][]projectmemory.SearchHit{primary, secondary} {
		for i := range group {
			identity := memoryIdentityKey(group[i].Memory)
			if _, ok := seen[identity]; ok {
				continue
			}
			seen[identity] = struct{}{}
			merged = append(merged, group[i])
		}
	}
	sort.SliceStable(merged, func(i, j int) bool { return merged[i].Score < merged[j].Score })
	if len(merged) > limit {
		merged = merged[:limit]
	}
	return merged
}

// memoryIdentityKey identifies "the same fact" across the workspace and global stores: by
// (scope, key) when a stable key is set, else by (scope, title). Used to de-duplicate merged
// search results.
func memoryIdentityKey(m projectmemory.Memory) string {
	scope := strings.ToLower(strings.TrimSpace(m.Scope))
	if key := strings.ToLower(strings.TrimSpace(m.Key)); key != "" {
		return "key\x00" + scope + "\x00" + key
	}
	return "title\x00" + scope + "\x00" + strings.ToLower(strings.TrimSpace(m.Title))
}

// promoteMemory copies the memory identified by id from src into dst. A keyed fact upserts by
// (scope, key), so re-promoting a refreshed fact updates the global copy in place.
func promoteMemory(ctx context.Context, src, dst *projectmemory.Store, id string) (projectmemory.Memory, error) {
	memory, err := src.Get(ctx, id)
	if err != nil {
		return projectmemory.Memory{}, err
	}
	return dst.Add(ctx, projectmemory.AddInput{
		Scope:  memory.Scope,
		Key:    memory.Key,
		Title:  memory.Title,
		Body:   memory.Body,
		Tags:   memory.Tags,
		Source: "promote",
	})
}

// newMemoryPromoteCommand builds `rc memory promote`, lifting a per-project fact into the
// home-scoped global store so it informs work in every workspace.
func newMemoryPromoteCommand() *cobra.Command {
	var outputFormat string
	cmd := &cobra.Command{
		Use:          "promote <id>",
		Short:        "Copy a workspace memory into the global store shared across workspaces",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		Long: `Promote a per-project memory into the home-scoped global store (~/.rc/memory.db).

The fact is upserted by (scope, key) when it has a key, so re-promoting a refreshed fact updates
the global copy in place. Global facts then surface in every workspace's default search.`,
		Example: `  rc memory promote mem-1a2b3c4d`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signalCommandContext(cmd)
			defer stop()

			format, err := normalizeOperatorOutputFormat(outputFormat)
			if err != nil {
				return withExitCode(1, err)
			}

			source, _, err := openProjectMemory(ctx)
			if err != nil {
				return withExitCode(2, err)
			}
			defer closeProjectMemory(ctx, source)

			destination, _, err := openGlobalMemory(ctx, true)
			if err != nil {
				return withExitCode(2, err)
			}
			defer closeProjectMemory(ctx, destination)

			promoted, err := promoteMemory(ctx, source, destination, args[0])
			if err != nil {
				return mapMemoryError(err)
			}
			return writeMemoryRecord(cmd, format, promoted)
		},
	}
	cmd.Flags().StringVar(&outputFormat, "format", operatorOutputFormatText, "Output format: text or json")
	return cmd
}

// runNativeExport writes the scoped store's memories in Claude-native memory format, honoring
// --global and an optional target directory. It backs `rc memory export --native`.
func runNativeExport(ctx context.Context, cmd *cobra.Command, dir string) error {
	global := memoryGlobalFlag(cmd)
	memStore, _, err := openMemoryScope(ctx, global)
	if err != nil {
		return withExitCode(2, err)
	}
	defer closeProjectMemory(ctx, memStore)

	memories, err := memStore.List(ctx, projectmemory.ListFilter{})
	if err != nil {
		return mapMemoryError(err)
	}

	targetDir := strings.TrimSpace(dir)
	if targetDir == "" {
		targetDir, err = defaultNativeDir(ctx, global)
		if err != nil {
			return withExitCode(2, err)
		}
	}
	count, err := writeNativeMemory(targetDir, memories)
	if err != nil {
		return withExitCode(2, err)
	}
	return writeMemoryText(cmd, fmt.Sprintf("Exported %d memories (native format) to %s\n", count, targetDir))
}

// defaultNativeDir is where `rc memory export --native` writes when no --dir is given:
// <.rc base>/memory-native for a workspace, or ~/.rc/memory-native for the global store.
func defaultNativeDir(ctx context.Context, global bool) (string, error) {
	if global {
		home, err := rcconfig.ResolveHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "memory-native"), nil
	}
	root, err := discoverWorkspaceRoot(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Join(model.RcDir(root), "memory-native"), nil
}

// writeNativeMemory renders memories into Claude-native auto-memory shape at dir: a MEMORY.md
// pointer index plus one markdown file per fact (read on demand). It returns the count written
// and refuses to silently drop two facts whose file names collide.
func writeNativeMemory(dir string, memories []projectmemory.Memory) (int, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, fmt.Errorf("create native memory dir %q: %w", dir, err)
	}

	index := &strings.Builder{}
	_, _ = fmt.Fprint(index, "# Project Memory\n\n")

	writtenBy := make(map[string]string, len(memories))
	for i := range memories {
		name := projectmemory.MirrorFileName(memories[i])
		if prevID, clash := writtenBy[name]; clash {
			return 0, fmt.Errorf(
				"native file name collision: memories %q and %q both map to %q",
				prevID,
				memories[i].ID,
				name,
			)
		}
		writtenBy[name] = memories[i].ID

		if err := os.WriteFile(filepath.Join(dir, name), []byte(nativeMemoryFileBody(memories[i])), 0o600); err != nil {
			return 0, fmt.Errorf("write native memory file %q: %w", name, err)
		}
		_, _ = fmt.Fprintf(index, "- [%s](%s) — %s\n", memories[i].Title, name, memories[i].Scope)
	}

	if err := os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte(index.String()), 0o600); err != nil {
		return 0, fmt.Errorf("write native MEMORY.md: %w", err)
	}
	return len(memories), nil
}

// nativeMemoryFileBody renders one fact as a standalone markdown topic file.
func nativeMemoryFileBody(m projectmemory.Memory) string {
	body := &strings.Builder{}
	_, _ = fmt.Fprintf(body, "# %s\n\n%s\n", m.Title, m.Body)
	if len(m.Tags) > 0 {
		_, _ = fmt.Fprintf(body, "\n_tags: %s_\n", strings.Join(m.Tags, ", "))
	}
	return body.String()
}

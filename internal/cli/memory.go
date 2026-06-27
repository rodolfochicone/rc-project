package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rodolfochicone/rc-project/internal/core/model"
	"github.com/rodolfochicone/rc-project/internal/core/projectmemory"
	"github.com/spf13/cobra"
)

// newMemoryCommand builds the `rc memory` command tree. It manages the per-project
// curated memory stored in <workspace>/.rc/memory.db, which rc skills read and write
// to carry decisions, conventions and gotchas across runs.
func newMemoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "memory",
		Short:        "Manage the per-project memory database",
		SilenceUsage: true,
		Long: `Read and write the per-project curated memory stored in .rc/memory.db.

Memories are short, scoped facts (decisions, conventions, gotchas, glossary, context)
that rc skills consult before working and record afterward. Retrieval is keyword-ranked
via SQLite FTS5, so exact symbol and identifier matches are preserved.`,
	}
	cmd.AddCommand(
		newMemoryAddCommand(),
		newMemorySearchCommand(),
		newMemoryGetCommand(),
		newMemoryListCommand(),
		newMemoryUpdateCommand(),
		newMemoryDeleteCommand(),
		newMemoryReindexCommand(),
	)
	return cmd
}

// openProjectMemory resolves the workspace, opens its project-memory database, and
// returns both the store (for CRUD) and the retriever to use for search. When semantic
// embeddings are enabled (RC_MEMORY_EMBEDDINGS=ollama) the retriever is hybrid; otherwise
// it is the lexical store itself.
func openProjectMemory(ctx context.Context) (*projectmemory.Store, projectmemory.Retriever, error) {
	root, err := discoverWorkspaceRoot(ctx)
	if err != nil {
		return nil, nil, err
	}
	dbPath := filepath.Join(model.RcDir(root), projectmemory.DBFileName)

	embedder, err := projectMemoryEmbedderFromEnv()
	if err != nil {
		return nil, nil, err
	}

	var opts []projectmemory.Option
	if embedder != nil {
		opts = append(opts, projectmemory.WithEmbedder(embedder))
	}
	st, err := projectmemory.Open(ctx, dbPath, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("open project memory: %w", err)
	}

	var retriever projectmemory.Retriever = st
	if embedder != nil {
		retriever = projectmemory.NewHybridRetriever(st, embedder)
	}
	return st, retriever, nil
}

// projectMemoryEmbedderFromEnv builds a semantic embedder when RC_MEMORY_EMBEDDINGS is set
// to "ollama". It returns (nil, nil) when embeddings are disabled, keeping retrieval
// lexical. RC_MEMORY_MODEL defaults to embeddinggemma; RC_MEMORY_ENDPOINT to local Ollama.
func projectMemoryEmbedderFromEnv() (projectmemory.Embedder, error) {
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("RC_MEMORY_EMBEDDINGS")), "ollama") {
		return nil, nil
	}
	modelName := strings.TrimSpace(os.Getenv("RC_MEMORY_MODEL"))
	if modelName == "" {
		modelName = "embeddinggemma"
	}
	embedder, err := projectmemory.NewOllamaEmbedder(os.Getenv("RC_MEMORY_ENDPOINT"), modelName)
	if err != nil {
		return nil, fmt.Errorf("configure project memory embeddings: %w", err)
	}
	return embedder, nil
}

// closeProjectMemory closes the store, logging (not failing) on a checkpoint error,
// since committed rows are already durable in the WAL.
func closeProjectMemory(ctx context.Context, st *projectmemory.Store) {
	if err := st.Close(ctx); err != nil {
		slog.Default().Warn("close project memory", "error", err)
	}
}

// mapMemoryError maps store sentinels to operator exit codes: invalid input and
// not-found are user errors (1); everything else is an operational failure (2).
func mapMemoryError(err error) error {
	switch {
	case errors.Is(err, projectmemory.ErrInvalidInput), errors.Is(err, projectmemory.ErrNotFound):
		return withExitCode(1, err)
	default:
		return withExitCode(2, err)
	}
}

// memoryJSON is the stable JSON shape emitted with --format json for skills to parse.
type memoryJSON struct {
	ID        string   `json:"id"`
	Scope     string   `json:"scope"`
	Key       string   `json:"key,omitempty"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Tags      []string `json:"tags"`
	Source    string   `json:"source,omitempty"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	Score     *float64 `json:"score,omitempty"`
}

func toMemoryJSON(m projectmemory.Memory) memoryJSON {
	tags := m.Tags
	if tags == nil {
		tags = []string{}
	}
	return memoryJSON{
		ID:        m.ID,
		Scope:     m.Scope,
		Key:       m.Key,
		Title:     m.Title,
		Body:      m.Body,
		Tags:      tags,
		Source:    m.Source,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

func renderMemory(m projectmemory.Memory) string {
	b := &strings.Builder{}
	key := ""
	if m.Key != "" {
		key = "/" + m.Key
	}
	// strings.Builder never returns a write error, so the counts are discarded.
	_, _ = fmt.Fprintf(b, "[%s] %s%s — %s\n", m.ID, m.Scope, key, m.Title)
	if len(m.Tags) > 0 {
		_, _ = fmt.Fprintf(b, "  tags: %s\n", strings.Join(m.Tags, ", "))
	}
	_, _ = fmt.Fprintf(b, "  %s\n", m.Body)
	_, _ = fmt.Fprintf(b, "  updated %s\n", m.UpdatedAt.Format(time.RFC3339))
	return b.String()
}

func writeMemoryText(cmd *cobra.Command, body string) error {
	if _, err := fmt.Fprint(cmd.OutOrStdout(), body); err != nil {
		return withExitCode(2, fmt.Errorf("write memory output: %w", err))
	}
	return nil
}

func writeMemoryRecord(cmd *cobra.Command, format string, m projectmemory.Memory) error {
	if format == operatorOutputFormatJSON {
		if err := writeOperatorJSON(cmd.OutOrStdout(), toMemoryJSON(m)); err != nil {
			return withExitCode(2, fmt.Errorf("write memory json: %w", err))
		}
		return nil
	}
	return writeMemoryText(cmd, renderMemory(m))
}

func writeMemoryList(cmd *cobra.Command, format string, memories []projectmemory.Memory) error {
	if format == operatorOutputFormatJSON {
		payload := make([]memoryJSON, 0, len(memories))
		for i := range memories {
			payload = append(payload, toMemoryJSON(memories[i]))
		}
		if err := writeOperatorJSON(cmd.OutOrStdout(), payload); err != nil {
			return withExitCode(2, fmt.Errorf("write memory json: %w", err))
		}
		return nil
	}

	if len(memories) == 0 {
		return writeMemoryText(cmd, "No memories found.\n")
	}
	b := &strings.Builder{}
	for i := range memories {
		_, _ = fmt.Fprint(b, renderMemory(memories[i]))
	}
	return writeMemoryText(cmd, b.String())
}

func writeMemoryHits(cmd *cobra.Command, format string, hits []projectmemory.SearchHit) error {
	if format == operatorOutputFormatJSON {
		payload := make([]memoryJSON, 0, len(hits))
		for i := range hits {
			record := toMemoryJSON(hits[i].Memory)
			score := hits[i].Score
			record.Score = &score
			payload = append(payload, record)
		}
		if err := writeOperatorJSON(cmd.OutOrStdout(), payload); err != nil {
			return withExitCode(2, fmt.Errorf("write memory json: %w", err))
		}
		return nil
	}

	if len(hits) == 0 {
		return writeMemoryText(cmd, "No memories found.\n")
	}
	b := &strings.Builder{}
	for i := range hits {
		_, _ = fmt.Fprint(b, renderMemory(hits[i].Memory))
	}
	return writeMemoryText(cmd, b.String())
}

// --- add ---------------------------------------------------------------------

type memoryAddState struct {
	scope        string
	key          string
	title        string
	body         string
	source       string
	tags         []string
	outputFormat string
}

func newMemoryAddCommand() *cobra.Command {
	state := &memoryAddState{}
	cmd := &cobra.Command{
		Use:          "add",
		Short:        "Insert or upsert a project memory",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Example: `  rc memory add --scope convention --title "Use modernc sqlite" --body "Pure-Go, CGO-free."
  rc memory add --scope gotcha --key race --title "Run loop race" --body "Guard with mutex."`,
		RunE: state.run,
	}
	cmd.Flags().StringVar(&state.scope, "scope", "", "Memory scope: decision, convention, gotcha, glossary, context")
	cmd.Flags().StringVar(&state.key, "key", "", "Stable key for upsert within a scope (optional)")
	cmd.Flags().StringVar(&state.title, "title", "", "Short title")
	cmd.Flags().StringVar(&state.body, "body", "", "Memory body")
	cmd.Flags().StringSliceVar(&state.tags, "tags", nil, "Comma-separated tags")
	cmd.Flags().StringVar(&state.source, "source", "", "Origin (skill or command) that produced this memory")
	cmd.Flags().StringVar(&state.outputFormat, "format", operatorOutputFormatText, "Output format: text or json")
	return cmd
}

func (s *memoryAddState) run(cmd *cobra.Command, _ []string) error {
	ctx, stop := signalCommandContext(cmd)
	defer stop()

	format, err := normalizeOperatorOutputFormat(s.outputFormat)
	if err != nil {
		return withExitCode(1, err)
	}

	st, _, err := openProjectMemory(ctx)
	if err != nil {
		return withExitCode(2, err)
	}
	defer closeProjectMemory(ctx, st)

	memory, err := st.Add(ctx, projectmemory.AddInput{
		Scope:  s.scope,
		Key:    s.key,
		Title:  s.title,
		Body:   s.body,
		Tags:   s.tags,
		Source: s.source,
	})
	if err != nil {
		return mapMemoryError(err)
	}
	return writeMemoryRecord(cmd, format, memory)
}

// --- search ------------------------------------------------------------------

type memorySearchState struct {
	scope        string
	limit        int
	outputFormat string
}

func newMemorySearchCommand() *cobra.Command {
	state := &memorySearchState{}
	cmd := &cobra.Command{
		Use:          "search <query>",
		Short:        "Find relevant project memories (BM25-ranked)",
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
		Example: `  rc memory search sqlite driver
  rc memory search "wal checkpoint" --scope gotcha --format json`,
		RunE: state.run,
	}
	cmd.Flags().StringVar(&state.scope, "scope", "", "Restrict results to a single scope")
	cmd.Flags().IntVar(&state.limit, "limit", 0, "Maximum results (default 8)")
	cmd.Flags().StringVar(&state.outputFormat, "format", operatorOutputFormatText, "Output format: text or json")
	return cmd
}

func (s *memorySearchState) run(cmd *cobra.Command, args []string) error {
	ctx, stop := signalCommandContext(cmd)
	defer stop()

	format, err := normalizeOperatorOutputFormat(s.outputFormat)
	if err != nil {
		return withExitCode(1, err)
	}

	st, retriever, err := openProjectMemory(ctx)
	if err != nil {
		return withExitCode(2, err)
	}
	defer closeProjectMemory(ctx, st)

	hits, err := retriever.Search(ctx, projectmemory.SearchQuery{
		Text:  strings.Join(args, " "),
		Scope: s.scope,
		Limit: s.limit,
	})
	if err != nil {
		return mapMemoryError(err)
	}
	return writeMemoryHits(cmd, format, hits)
}

// --- get ---------------------------------------------------------------------

type memoryGetState struct {
	scope        string
	key          string
	outputFormat string
}

func newMemoryGetCommand() *cobra.Command {
	state := &memoryGetState{}
	cmd := &cobra.Command{
		Use:          "get [id]",
		Short:        "Show one memory by id, or by --scope and --key",
		SilenceUsage: true,
		Args:         cobra.MaximumNArgs(1),
		Example: `  rc memory get mem-1a2b3c4d
  rc memory get --scope convention --key db-driver`,
		RunE: state.run,
	}
	cmd.Flags().StringVar(&state.scope, "scope", "", "Scope (with --key) to resolve a memory")
	cmd.Flags().StringVar(&state.key, "key", "", "Key (with --scope) to resolve a memory")
	cmd.Flags().StringVar(&state.outputFormat, "format", operatorOutputFormatText, "Output format: text or json")
	return cmd
}

func (s *memoryGetState) run(cmd *cobra.Command, args []string) error {
	ctx, stop := signalCommandContext(cmd)
	defer stop()

	format, err := normalizeOperatorOutputFormat(s.outputFormat)
	if err != nil {
		return withExitCode(1, err)
	}

	st, _, err := openProjectMemory(ctx)
	if err != nil {
		return withExitCode(2, err)
	}
	defer closeProjectMemory(ctx, st)

	var memory projectmemory.Memory
	switch {
	case len(args) == 1:
		memory, err = st.Get(ctx, args[0])
	case s.scope != "" && s.key != "":
		memory, err = st.GetByKey(ctx, s.scope, s.key)
	default:
		return withExitCode(1, errors.New("get requires an <id> argument or both --scope and --key"))
	}
	if err != nil {
		return mapMemoryError(err)
	}
	return writeMemoryRecord(cmd, format, memory)
}

// --- list --------------------------------------------------------------------

type memoryListState struct {
	scope        string
	tag          string
	limit        int
	outputFormat string
}

func newMemoryListCommand() *cobra.Command {
	state := &memoryListState{}
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List memories, newest first",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Example: `  rc memory list
  rc memory list --scope decision --limit 20 --format json`,
		RunE: state.run,
	}
	cmd.Flags().StringVar(&state.scope, "scope", "", "Restrict to a single scope")
	cmd.Flags().StringVar(&state.tag, "tag", "", "Restrict to memories carrying this tag")
	cmd.Flags().IntVar(&state.limit, "limit", 0, "Maximum results (default: all)")
	cmd.Flags().StringVar(&state.outputFormat, "format", operatorOutputFormatText, "Output format: text or json")
	return cmd
}

func (s *memoryListState) run(cmd *cobra.Command, _ []string) error {
	ctx, stop := signalCommandContext(cmd)
	defer stop()

	format, err := normalizeOperatorOutputFormat(s.outputFormat)
	if err != nil {
		return withExitCode(1, err)
	}

	st, _, err := openProjectMemory(ctx)
	if err != nil {
		return withExitCode(2, err)
	}
	defer closeProjectMemory(ctx, st)

	memories, err := st.List(ctx, projectmemory.ListFilter{
		Scope: s.scope,
		Tag:   s.tag,
		Limit: s.limit,
	})
	if err != nil {
		return mapMemoryError(err)
	}
	return writeMemoryList(cmd, format, memories)
}

// --- update ------------------------------------------------------------------

type memoryUpdateState struct {
	title        string
	body         string
	source       string
	tags         []string
	outputFormat string
}

func newMemoryUpdateCommand() *cobra.Command {
	state := &memoryUpdateState{}
	cmd := &cobra.Command{
		Use:          "update <id>",
		Short:        "Update fields of an existing memory",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		Example: `  rc memory update mem-1a2b3c4d --body "Refreshed guidance."
  rc memory update mem-1a2b3c4d --tags go,sqlite`,
		RunE: state.run,
	}
	cmd.Flags().StringVar(&state.title, "title", "", "New title")
	cmd.Flags().StringVar(&state.body, "body", "", "New body")
	cmd.Flags().StringSliceVar(&state.tags, "tags", nil, "Replacement comma-separated tags")
	cmd.Flags().StringVar(&state.source, "source", "", "New origin (skill or command)")
	cmd.Flags().StringVar(&state.outputFormat, "format", operatorOutputFormatText, "Output format: text or json")
	return cmd
}

func (s *memoryUpdateState) run(cmd *cobra.Command, args []string) error {
	ctx, stop := signalCommandContext(cmd)
	defer stop()

	format, err := normalizeOperatorOutputFormat(s.outputFormat)
	if err != nil {
		return withExitCode(1, err)
	}

	input := projectmemory.UpdateInput{}
	if cmd.Flags().Changed("title") {
		input.Title = &s.title
	}
	if cmd.Flags().Changed("body") {
		input.Body = &s.body
	}
	if cmd.Flags().Changed("source") {
		input.Source = &s.source
	}
	if cmd.Flags().Changed("tags") {
		tags := s.tags
		input.Tags = &tags
	}

	st, _, err := openProjectMemory(ctx)
	if err != nil {
		return withExitCode(2, err)
	}
	defer closeProjectMemory(ctx, st)

	memory, err := st.Update(ctx, args[0], input)
	if err != nil {
		return mapMemoryError(err)
	}
	return writeMemoryRecord(cmd, format, memory)
}

// --- delete ------------------------------------------------------------------

func newMemoryDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete <id>",
		Short:        "Delete a memory by id",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		Example:      `  rc memory delete mem-1a2b3c4d`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signalCommandContext(cmd)
			defer stop()

			st, _, err := openProjectMemory(ctx)
			if err != nil {
				return withExitCode(2, err)
			}
			defer closeProjectMemory(ctx, st)

			if err := st.Delete(ctx, args[0]); err != nil {
				return mapMemoryError(err)
			}
			return writeMemoryText(cmd, fmt.Sprintf("Deleted memory %s\n", args[0]))
		},
	}
	return cmd
}

// --- reindex -----------------------------------------------------------------

func newMemoryReindexCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "reindex",
		Short:        "Regenerate semantic embeddings for all memories",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Regenerate the embedding of every memory using the configured model.

Requires semantic embeddings enabled via RC_MEMORY_EMBEDDINGS=ollama (with optional
RC_MEMORY_MODEL and RC_MEMORY_ENDPOINT). Run after enabling embeddings or switching models
so existing memories become searchable semantically.`,
		Example: `  RC_MEMORY_EMBEDDINGS=ollama RC_MEMORY_MODEL=embeddinggemma rc memory reindex`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, stop := signalCommandContext(cmd)
			defer stop()

			st, _, err := openProjectMemory(ctx)
			if err != nil {
				return withExitCode(2, err)
			}
			defer closeProjectMemory(ctx, st)

			count, err := st.Reindex(ctx)
			if err != nil {
				return mapMemoryError(err)
			}
			return writeMemoryText(cmd, fmt.Sprintf("Reindexed %d memories.\n", count))
		},
	}
	return cmd
}

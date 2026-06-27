package projectmemory

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rodolfochicone/rc-project/internal/store"
)

// vectorCandidate pairs a memory with its stored embedding for cosine scoring.
type vectorCandidate struct {
	Memory
	vector []float32
}

// embedText is the text fed to the embedder for a memory: title, body, and tags.
func embedText(m Memory) string {
	return strings.TrimSpace(m.Title + "\n" + m.Body + "\n" + strings.Join(m.Tags, " "))
}

// maybeEmbed generates and stores the embedding for a memory when an embedder is
// configured. It is best-effort: a failure (e.g. Ollama unreachable) is logged and
// swallowed so it never blocks the write, since committed rows are already durable.
func (s *Store) maybeEmbed(ctx context.Context, m Memory) {
	if s.embedder == nil {
		return
	}
	vec, err := s.embedder.Embed(ctx, embedText(m))
	if err != nil {
		slog.Default().Warn("project memory embed failed", "id", m.ID, "error", err)
		return
	}
	if err := s.saveVector(ctx, m.ID, s.embedder.Model(), vec); err != nil {
		slog.Default().Warn("project memory save vector failed", "id", m.ID, "error", err)
	}
}

// saveVector upserts the embedding for a memory.
func (s *Store) saveVector(ctx context.Context, memID, model string, vec []float32) error {
	const upsert = `INSERT INTO memory_vectors (mem_id, model, dims, vector, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(mem_id) DO UPDATE SET
			model = excluded.model,
			dims = excluded.dims,
			vector = excluded.vector,
			created_at = excluded.created_at`
	if _, err := s.db.ExecContext(
		ctx, upsert,
		memID, model, len(vec), encodeVector(vec), store.FormatTimestamp(s.now()),
	); err != nil {
		return fmt.Errorf("projectmemory: save vector: %w", err)
	}
	return nil
}

// vectorCandidates loads memories that have an embedding from the given model, optionally
// restricted to a scope. The corpus is small (curated memory), so callers score these in
// memory with cosine rather than relying on a vector index.
func (s *Store) vectorCandidates(ctx context.Context, scope, model string) ([]vectorCandidate, error) {
	const query = `SELECT m.id, m.scope, m.mem_key, m.title, m.body, m.tags, m.source,
			m.created_at, m.updated_at, v.vector
		FROM memory_vectors v
		JOIN memories m ON m.id = v.mem_id
		WHERE v.model = ?
		  AND (? = '' OR m.scope = ?)`

	rows, err := s.db.QueryContext(ctx, query, model, scope, scope)
	if err != nil {
		return nil, fmt.Errorf("projectmemory: load vector candidates: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var candidates []vectorCandidate
	for rows.Next() {
		var (
			memory  Memory
			key     sql.NullString
			tags    string
			created string
			updated string
			blob    []byte
		)
		if err := rows.Scan(
			&memory.ID, &memory.Scope, &key, &memory.Title, &memory.Body,
			&tags, &memory.Source, &created, &updated, &blob,
		); err != nil {
			return nil, fmt.Errorf("projectmemory: scan vector candidate: %w", err)
		}
		hydrated, err := hydrateMemory(memory, key, tags, created, updated)
		if err != nil {
			return nil, err
		}
		vec, err := decodeVector(blob)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, vectorCandidate{Memory: hydrated, vector: vec})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("projectmemory: iterate vector candidates: %w", err)
	}
	return candidates, nil
}

// Reindex regenerates embeddings for every memory using the configured embedder. Unlike
// the best-effort embedding on write, this is an explicit operation: it fails hard if the
// embedder is absent or unreachable. It returns the number of memories embedded.
func (s *Store) Reindex(ctx context.Context) (int, error) {
	if s.embedder == nil {
		return 0, fmt.Errorf("projectmemory: reindex requires a configured embedder: %w", ErrInvalidInput)
	}

	memories, err := s.List(ctx, ListFilter{})
	if err != nil {
		return 0, err
	}

	count := 0
	for i := range memories {
		vec, err := s.embedder.Embed(ctx, embedText(memories[i]))
		if err != nil {
			return count, fmt.Errorf("projectmemory: reindex embed %s: %w", memories[i].ID, err)
		}
		if err := s.saveVector(ctx, memories[i].ID, s.embedder.Model(), vec); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

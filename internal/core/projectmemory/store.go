// Package projectmemory implements a per-project SQLite store of curated memories
// (decisions, conventions, gotchas, glossary, context) that rc skills and commands
// read and write through the `rc memory` command. It is local-first and lives in the
// project's resolved .rc base directory. Retrieval is keyword-ranked via SQLite
// FTS5/BM25, which keeps exact symbol and identifier matches that pure vector search
// would lose. The database is the single source of truth; there is no markdown mirror.
package projectmemory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rodolfochicone/rc-project/internal/store"
)

// DBFileName is the project-memory database file, resolved under the .rc base directory.
const DBFileName = "memory.db"

// defaultSearchLimit bounds Search results when the caller does not specify a limit,
// keeping injected context small enough to respect the agent's context budget.
const defaultSearchLimit = 8

const memoryColumns = `id, scope, mem_key, title, body, tags, source, created_at, updated_at`

// Sentinel errors returned by the store.
var (
	// ErrNotFound reports that no memory matched the requested identifier or key.
	ErrNotFound = errors.New("projectmemory: memory not found")
	// ErrInvalidInput reports that a required field was empty or malformed.
	ErrInvalidInput = errors.New("projectmemory: invalid input")
)

// Memory is a single curated project memory record.
type Memory struct {
	ID        string
	Scope     string
	Key       string // optional stable key; unique within a scope when set
	Title     string
	Body      string
	Tags      []string
	Source    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SearchHit pairs a memory with its BM25 relevance score (lower is more relevant).
type SearchHit struct {
	Memory
	Score float64
}

// AddInput describes a memory to insert or upsert. Scope, Title and Body are required.
type AddInput struct {
	Scope  string
	Key    string
	Title  string
	Body   string
	Tags   []string
	Source string
}

// UpdateInput carries the fields to change on an existing memory. Nil fields are left
// unchanged; non-nil fields replace the stored value.
type UpdateInput struct {
	Title  *string
	Body   *string
	Tags   *[]string
	Source *string
}

// ListFilter narrows and bounds a List query.
type ListFilter struct {
	Scope string
	Tag   string
	Limit int
}

// SearchQuery is a full-text query over titles, bodies and tags, optionally scoped.
type SearchQuery struct {
	Text  string
	Scope string
	Limit int
}

// Store is a per-project memory database handle. It is safe for concurrent use by
// multiple goroutines; the underlying *sql.DB manages its own connection pool.
type Store struct {
	db       *sql.DB
	now      func() time.Time
	embedder Embedder
}

// Option configures a Store.
type Option func(*Store)

// WithClock overrides the time source. Intended for deterministic tests.
func WithClock(now func() time.Time) Option {
	return func(s *Store) {
		if now != nil {
			s.now = now
		}
	}
}

// WithEmbedder enables semantic embeddings: every Add and Update generates and stores a
// vector (best-effort), and Reindex regenerates them. Without it, the store is purely
// lexical.
func WithEmbedder(embedder Embedder) Option {
	return func(s *Store) {
		s.embedder = embedder
	}
}

// Open opens (creating if needed) the project-memory database at dbPath and applies
// pending migrations. The caller owns the returned Store and must Close it.
func Open(ctx context.Context, dbPath string, opts ...Option) (*Store, error) {
	s := &Store{now: func() time.Time { return time.Now().UTC() }}
	for _, opt := range opts {
		opt(s)
	}

	db, err := store.OpenSQLiteDatabase(ctx, dbPath, func(ctx context.Context, db *sql.DB) error {
		return migrate(ctx, db, s.now)
	})
	if err != nil {
		return nil, fmt.Errorf("projectmemory: open %q: %w", dbPath, err)
	}
	s.db = db
	return s, nil
}

// Close checkpoints the WAL and closes the underlying database handle.
func (s *Store) Close(ctx context.Context) error {
	return store.CloseSQLiteDatabase(ctx, s.db)
}

// Add inserts a new memory. When Key is set and a memory already exists for the same
// (scope, key), the existing record is updated in place instead of duplicated.
func (s *Store) Add(ctx context.Context, in AddInput) (Memory, error) {
	scope := strings.TrimSpace(in.Scope)
	title := strings.TrimSpace(in.Title)
	body := strings.TrimSpace(in.Body)
	if scope == "" || title == "" || body == "" {
		return Memory{}, fmt.Errorf("projectmemory: add requires scope, title and body: %w", ErrInvalidInput)
	}
	key := strings.TrimSpace(in.Key)
	tags := normalizeTags(in.Tags)
	source := strings.TrimSpace(in.Source)
	stamp := store.FormatTimestamp(s.now())
	id := store.NewID("mem")

	const upsert = `INSERT INTO memories
		(id, scope, mem_key, title, body, tags, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(scope, mem_key) WHERE mem_key IS NOT NULL DO UPDATE SET
			title = excluded.title,
			body = excluded.body,
			tags = excluded.tags,
			source = excluded.source,
			updated_at = excluded.updated_at`

	if _, err := s.db.ExecContext(
		ctx, upsert,
		id, scope, store.NullableString(key), title, body, tags, source, stamp, stamp,
	); err != nil {
		return Memory{}, fmt.Errorf("projectmemory: add: %w", err)
	}

	var (
		stored Memory
		getErr error
	)
	if key != "" {
		stored, getErr = s.GetByKey(ctx, scope, key)
	} else {
		stored, getErr = s.Get(ctx, id)
	}
	if getErr != nil {
		return Memory{}, getErr
	}
	s.maybeEmbed(ctx, stored)
	return stored, nil
}

// Get returns the memory with the given id.
func (s *Store) Get(ctx context.Context, id string) (Memory, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Memory{}, fmt.Errorf("projectmemory: get requires an id: %w", ErrInvalidInput)
	}
	row := s.db.QueryRowContext(ctx, `SELECT `+memoryColumns+` FROM memories WHERE id = ?`, id)
	return scanMemoryRow(row)
}

// GetByKey returns the memory identified by its (scope, key) pair.
func (s *Store) GetByKey(ctx context.Context, scope, key string) (Memory, error) {
	scope = strings.TrimSpace(scope)
	key = strings.TrimSpace(key)
	if scope == "" || key == "" {
		return Memory{}, fmt.Errorf("projectmemory: get-by-key requires scope and key: %w", ErrInvalidInput)
	}
	row := s.db.QueryRowContext(
		ctx,
		`SELECT `+memoryColumns+` FROM memories WHERE scope = ? AND mem_key = ?`,
		scope, key,
	)
	return scanMemoryRow(row)
}

// Update applies the non-nil fields of in to the memory with the given id and returns
// the updated record. Title and Body may not be cleared to empty.
func (s *Store) Update(ctx context.Context, id string, in UpdateInput) (Memory, error) {
	existing, err := s.Get(ctx, id)
	if err != nil {
		return Memory{}, err
	}

	if in.Title != nil {
		existing.Title = strings.TrimSpace(*in.Title)
	}
	if in.Body != nil {
		existing.Body = strings.TrimSpace(*in.Body)
	}
	if in.Tags != nil {
		existing.Tags = splitTags(normalizeTags(*in.Tags))
	}
	if in.Source != nil {
		existing.Source = strings.TrimSpace(*in.Source)
	}
	if existing.Title == "" || existing.Body == "" {
		return Memory{}, fmt.Errorf("projectmemory: update cannot clear title or body: %w", ErrInvalidInput)
	}

	stamp := store.FormatTimestamp(s.now())
	if _, err := s.db.ExecContext(
		ctx,
		`UPDATE memories SET title = ?, body = ?, tags = ?, source = ?, updated_at = ? WHERE id = ?`,
		existing.Title, existing.Body, normalizeTags(existing.Tags), existing.Source, stamp, existing.ID,
	); err != nil {
		return Memory{}, fmt.Errorf("projectmemory: update: %w", err)
	}

	updated, err := s.Get(ctx, existing.ID)
	if err != nil {
		return Memory{}, err
	}
	s.maybeEmbed(ctx, updated)
	return updated, nil
}

// Delete removes the memory with the given id. It returns ErrNotFound when no row matched.
func (s *Store) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("projectmemory: delete requires an id: %w", ErrInvalidInput)
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM memories WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("projectmemory: delete: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("projectmemory: delete rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// listQuery is static: optional filters are driven by parameters (`? = ” OR ...`)
// rather than string concatenation, so the SQL text never mixes with runtime values.
const listQuery = `SELECT ` + memoryColumns + ` FROM memories
	WHERE (? = '' OR scope = ?)
	  AND (? = '' OR (',' || tags || ',') LIKE ?)
	ORDER BY updated_at DESC, id ASC
	LIMIT ?`

// List returns memories matching the filter, newest first. A non-positive Limit
// returns all matches.
func (s *Store) List(ctx context.Context, filter ListFilter) ([]Memory, error) {
	scope := strings.TrimSpace(filter.Scope)
	tag := normalizeTag(filter.Tag)
	tagPattern := ""
	if tag != "" {
		tagPattern = "%," + tag + ",%"
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = -1 // SQLite treats LIMIT -1 as unbounded.
	}

	rows, err := s.db.QueryContext(ctx, listQuery, scope, scope, tag, tagPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("projectmemory: list: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var memories []Memory
	for rows.Next() {
		memory, err := scanMemoryRow(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, memory)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("projectmemory: iterate list: %w", err)
	}
	return memories, nil
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanMemoryRow(sc rowScanner) (Memory, error) {
	var (
		memory  Memory
		key     sql.NullString
		tags    string
		created string
		updated string
	)
	if err := sc.Scan(
		&memory.ID, &memory.Scope, &key, &memory.Title, &memory.Body,
		&tags, &memory.Source, &created, &updated,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Memory{}, ErrNotFound
		}
		return Memory{}, fmt.Errorf("projectmemory: scan memory: %w", err)
	}
	return hydrateMemory(memory, key, tags, created, updated)
}

func hydrateMemory(memory Memory, key sql.NullString, tags, created, updated string) (Memory, error) {
	createdAt, err := store.ParseTimestamp(created)
	if err != nil {
		return Memory{}, fmt.Errorf("projectmemory: parse created_at: %w", err)
	}
	updatedAt, err := store.ParseTimestamp(updated)
	if err != nil {
		return Memory{}, fmt.Errorf("projectmemory: parse updated_at: %w", err)
	}
	memory.Key = strings.TrimSpace(key.String)
	memory.Tags = splitTags(tags)
	memory.CreatedAt = createdAt
	memory.UpdatedAt = updatedAt
	return memory, nil
}

// normalizeTags trims, lowercases, de-duplicates and sorts tags into a stable CSV.
func normalizeTags(tags []string) string {
	seen := make(map[string]struct{}, len(tags))
	cleaned := make([]string, 0, len(tags))
	for _, tag := range tags {
		normalized := normalizeTag(tag)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		cleaned = append(cleaned, normalized)
	}
	sort.Strings(cleaned)
	return strings.Join(cleaned, ",")
}

func normalizeTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}

func splitTags(csv string) []string {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			tags = append(tags, trimmed)
		}
	}
	return tags
}

package projectmemory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rodolfochicone/rc-project/internal/store"
)

// migrationStep is a single versioned schema change applied once and recorded
// in schema_migrations, mirroring the convention used by internal/store/globaldb.
type migrationStep struct {
	version    int
	name       string
	statements []string
}

const migrationsTableStmt = `CREATE TABLE IF NOT EXISTS schema_migrations (
	version    INTEGER PRIMARY KEY,
	name       TEXT NOT NULL,
	applied_at TEXT NOT NULL
);`

// migrations defines the project-memory schema: a single curated memories table,
// supporting indexes, and an external-content FTS5 index kept in sync by triggers.
var migrations = []migrationStep{
	{
		version: 1,
		name:    "project_memory_initial",
		statements: []string{
			`CREATE TABLE IF NOT EXISTS memories (
				id         TEXT PRIMARY KEY,
				scope      TEXT NOT NULL,
				mem_key    TEXT,
				title      TEXT NOT NULL,
				body       TEXT NOT NULL,
				tags       TEXT NOT NULL DEFAULT '',
				source     TEXT NOT NULL DEFAULT '',
				created_at TEXT NOT NULL,
				updated_at TEXT NOT NULL
			);`,
			`CREATE INDEX IF NOT EXISTS idx_memories_scope ON memories(scope);`,
			`CREATE INDEX IF NOT EXISTS idx_memories_updated_at ON memories(updated_at DESC);`,
			`CREATE UNIQUE INDEX IF NOT EXISTS uq_memories_scope_key
				ON memories(scope, mem_key)
				WHERE mem_key IS NOT NULL;`,
			`CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
				title, body, tags,
				content='memories', content_rowid='rowid', tokenize='unicode61'
			);`,
			`CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON memories BEGIN
				INSERT INTO memories_fts(rowid, title, body, tags)
				VALUES (new.rowid, new.title, new.body, new.tags);
			END;`,
			`CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON memories BEGIN
				INSERT INTO memories_fts(memories_fts, rowid, title, body, tags)
				VALUES ('delete', old.rowid, old.title, old.body, old.tags);
			END;`,
			`CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON memories BEGIN
				INSERT INTO memories_fts(memories_fts, rowid, title, body, tags)
				VALUES ('delete', old.rowid, old.title, old.body, old.tags);
				INSERT INTO memories_fts(rowid, title, body, tags)
				VALUES (new.rowid, new.title, new.body, new.tags);
			END;`,
		},
	},
	{
		version: 2,
		name:    "project_memory_trigram_index",
		statements: []string{
			// A second FTS5 index with the trigram tokenizer enables substring matching
			// (e.g. "sqlite" inside "OpenSQLiteDatabase"), which unicode61 cannot do.
			`CREATE VIRTUAL TABLE IF NOT EXISTS memories_trigram USING fts5(
				title, body, tags,
				content='memories', content_rowid='rowid', tokenize='trigram'
			);`,
			// Backfill any rows written under schema v1 before the triggers existed.
			`INSERT INTO memories_trigram(rowid, title, body, tags)
				SELECT rowid, title, body, tags FROM memories;`,
			`CREATE TRIGGER IF NOT EXISTS memories_tri_ai AFTER INSERT ON memories BEGIN
				INSERT INTO memories_trigram(rowid, title, body, tags)
				VALUES (new.rowid, new.title, new.body, new.tags);
			END;`,
			`CREATE TRIGGER IF NOT EXISTS memories_tri_ad AFTER DELETE ON memories BEGIN
				INSERT INTO memories_trigram(memories_trigram, rowid, title, body, tags)
				VALUES ('delete', old.rowid, old.title, old.body, old.tags);
			END;`,
			`CREATE TRIGGER IF NOT EXISTS memories_tri_au AFTER UPDATE ON memories BEGIN
				INSERT INTO memories_trigram(memories_trigram, rowid, title, body, tags)
				VALUES ('delete', old.rowid, old.title, old.body, old.tags);
				INSERT INTO memories_trigram(rowid, title, body, tags)
				VALUES (new.rowid, new.title, new.body, new.tags);
			END;`,
		},
	},
	{
		version: 3,
		name:    "project_memory_vectors",
		statements: []string{
			// Optional dense embeddings for semantic retrieval. One vector per memory,
			// tagged with the producing model so a model switch can be detected and
			// re-indexed. The vector is a derived artifact: deleting a memory drops it.
			`CREATE TABLE IF NOT EXISTS memory_vectors (
				mem_id     TEXT PRIMARY KEY REFERENCES memories(id) ON DELETE CASCADE,
				model      TEXT NOT NULL,
				dims       INTEGER NOT NULL,
				vector     BLOB NOT NULL,
				created_at TEXT NOT NULL
			);`,
			`CREATE INDEX IF NOT EXISTS idx_memory_vectors_model ON memory_vectors(model);`,
		},
	},
}

// migrate ensures the schema_migrations bookkeeping table exists and applies every
// pending migration in order, each inside its own transaction.
func migrate(ctx context.Context, db *sql.DB, now func() time.Time) error {
	if ctx == nil {
		return errors.New("projectmemory: migrate context is required")
	}
	if db == nil {
		return errors.New("projectmemory: migrate database is required")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	if err := store.EnsureSchema(ctx, db, []string{migrationsTableStmt}); err != nil {
		return fmt.Errorf("projectmemory: ensure migrations table: %w", err)
	}

	applied, err := appliedVersions(ctx, db)
	if err != nil {
		return err
	}

	for _, step := range migrations {
		if applied[step.version] {
			continue
		}
		if err := applyStep(ctx, db, step, now); err != nil {
			return err
		}
	}
	return nil
}

func appliedVersions(ctx context.Context, db *sql.DB) (map[int]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("projectmemory: query schema migrations: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	versions := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("projectmemory: scan schema migration: %w", err)
		}
		versions[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("projectmemory: iterate schema migrations: %w", err)
	}
	return versions, nil
}

func applyStep(ctx context.Context, db *sql.DB, step migrationStep, now func() time.Time) (retErr error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("projectmemory: begin migration %d: %w", step.version, err)
	}

	committed := false
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				retErr = errors.Join(
					retErr,
					fmt.Errorf("projectmemory: rollback migration %d: %w", step.version, rollbackErr),
				)
			}
		}
	}()

	for _, stmt := range step.statements {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf(
				"projectmemory: apply migration %d (%s): %w",
				step.version, strings.TrimSpace(step.name), err,
			)
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)`,
		step.version,
		strings.TrimSpace(step.name),
		store.FormatTimestamp(now()),
	); err != nil {
		return fmt.Errorf("projectmemory: record migration %d: %w", step.version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("projectmemory: commit migration %d: %w", step.version, err)
	}
	committed = true
	return nil
}

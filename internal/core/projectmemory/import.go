package projectmemory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/rodolfochicone/rc-project/internal/store"
)

// ImportResult reports the outcome of a batch import: how many records were inserted,
// updated in place, or left unchanged because they were not strictly newer.
type ImportResult struct {
	Added   int
	Updated int
	Skipped int
}

// importOutcome classifies what Import did with a single record.
type importOutcome int

const (
	importSkipped importOutcome = iota
	importAdded
	importUpdated
)

// Import upserts fully-specified records inside one transaction, preserving each record's id,
// created_at and updated_at (never re-stamping with now()) and applying most-recent-wins by
// UpdatedAt. Identity is (Scope, Key) when Key is set, otherwise ID. It never deletes rows.
// On any error the whole batch rolls back, so the database is never left partially imported.
func (s *Store) Import(ctx context.Context, records []Memory) (result ImportResult, retErr error) {
	if len(records) == 0 {
		return ImportResult{}, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ImportResult{}, fmt.Errorf("projectmemory: import begin: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				retErr = errors.Join(retErr, fmt.Errorf("projectmemory: import rollback: %w", rollbackErr))
			}
		}
	}()

	imported := make([]Memory, 0, len(records))
	for i := range records {
		outcome, stored, err := importOne(ctx, tx, records[i])
		if err != nil {
			return ImportResult{}, err
		}
		switch outcome {
		case importAdded:
			result.Added++
			imported = append(imported, stored)
		case importUpdated:
			result.Updated++
			imported = append(imported, stored)
		case importSkipped:
			result.Skipped++
		}
	}

	if err := tx.Commit(); err != nil {
		return ImportResult{}, fmt.Errorf("projectmemory: import commit: %w", err)
	}
	committed = true

	// Embeddings are derived artifacts updated best-effort after the rows are durable.
	for i := range imported {
		s.maybeEmbed(ctx, imported[i])
	}
	return result, nil
}

// importOne inserts, updates, or skips a single record within the import transaction and
// returns the stored record (for added/updated) so the caller can refresh embeddings.
func importOne(ctx context.Context, tx *sql.Tx, record Memory) (importOutcome, Memory, error) {
	id := strings.TrimSpace(record.ID)
	scope := strings.TrimSpace(record.Scope)
	title := strings.TrimSpace(record.Title)
	body := strings.TrimSpace(record.Body)
	if id == "" || scope == "" || title == "" || body == "" {
		return importSkipped, Memory{},
			fmt.Errorf("projectmemory: import requires id, scope, title and body: %w", ErrInvalidInput)
	}
	if record.CreatedAt.IsZero() || record.UpdatedAt.IsZero() {
		// A zero timestamp formats to "" and produces a row that fails to parse on read; reject it
		// here rather than silently writing an unreadable record.
		return importSkipped, Memory{},
			fmt.Errorf("projectmemory: import requires non-zero created_at and updated_at: %w", ErrInvalidInput)
	}
	key := strings.TrimSpace(record.Key)
	tags := normalizeTags(record.Tags)
	source := strings.TrimSpace(record.Source)
	createdAt := store.FormatTimestamp(record.CreatedAt)
	updatedAt := store.FormatTimestamp(record.UpdatedAt)

	existing, found, err := findExistingForImport(ctx, tx, id, scope, key)
	if err != nil {
		return importSkipped, Memory{}, err
	}

	if !found {
		const insert = `INSERT INTO memories
			(id, scope, mem_key, title, body, tags, source, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		if _, err := tx.ExecContext(
			ctx, insert,
			id, scope, store.NullableString(key), title, body, tags, source, createdAt, updatedAt,
		); err != nil {
			return importSkipped, Memory{}, fmt.Errorf("projectmemory: import insert: %w", err)
		}
		return importAdded, storedMemory(id, scope, key, title, body, tags, source, record), nil
	}

	if !record.UpdatedAt.After(existing.UpdatedAt) {
		return importSkipped, Memory{}, nil
	}

	const update = `UPDATE memories
		SET title = ?, body = ?, tags = ?, source = ?, created_at = ?, updated_at = ?
		WHERE id = ?`
	if _, err := tx.ExecContext(
		ctx, update,
		title, body, tags, source, createdAt, updatedAt, existing.ID,
	); err != nil {
		return importSkipped, Memory{}, fmt.Errorf("projectmemory: import update: %w", err)
	}
	return importUpdated, storedMemory(existing.ID, scope, key, title, body, tags, source, record), nil
}

// findExistingForImport resolves the row a record would conflict with: by (scope, key) when a
// key is set, otherwise by id. A missing row is reported as not found, not an error.
func findExistingForImport(ctx context.Context, tx *sql.Tx, id, scope, key string) (Memory, bool, error) {
	var row *sql.Row
	if key != "" {
		row = tx.QueryRowContext(
			ctx, `SELECT `+memoryColumns+` FROM memories WHERE scope = ? AND mem_key = ?`, scope, key,
		)
	} else {
		row = tx.QueryRowContext(ctx, `SELECT `+memoryColumns+` FROM memories WHERE id = ?`, id)
	}
	memory, err := scanMemoryRow(row)
	if errors.Is(err, ErrNotFound) {
		return Memory{}, false, nil
	}
	if err != nil {
		return Memory{}, false, err
	}
	return memory, true, nil
}

// storedMemory builds the in-memory view of a just-written record for embedding refresh.
func storedMemory(id, scope, key, title, body, tags, source string, record Memory) Memory {
	return Memory{
		ID:        id,
		Scope:     scope,
		Key:       key,
		Title:     title,
		Body:      body,
		Tags:      splitTags(tags),
		Source:    source,
		CreatedAt: record.CreatedAt.UTC(),
		UpdatedAt: record.UpdatedAt.UTC(),
	}
}

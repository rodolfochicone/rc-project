package cli

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/rodolfochicone/rc-project/internal/core/projectmemory"
)

func openTempMemory(t *testing.T) *projectmemory.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), projectmemory.DBFileName)
	st, err := projectmemory.Open(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("open temp memory: %v", err)
	}
	t.Cleanup(func() { _ = st.Close(context.Background()) })
	return st
}

func hit(id, scope, key, title string, score float64) projectmemory.SearchHit {
	return projectmemory.SearchHit{
		Memory: projectmemory.Memory{ID: id, Scope: scope, Key: key, Title: title},
		Score:  score,
	}
}

func TestMergeSearchHits(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		primary   []projectmemory.SearchHit
		secondary []projectmemory.SearchHit
		limit     int
		wantIDs   []string
	}{
		{
			name:      "global fact surfaces alongside workspace, ordered by score",
			primary:   []projectmemory.SearchHit{hit("w1", "convention", "", "A", 0.5)},
			secondary: []projectmemory.SearchHit{hit("g1", "gotcha", "", "B", 0.2)},
			limit:     8,
			wantIDs:   []string{"g1", "w1"},
		},
		{
			name:      "keyed collision keeps the workspace copy",
			primary:   []projectmemory.SearchHit{hit("w1", "convention", "db", "Driver", 0.9)},
			secondary: []projectmemory.SearchHit{hit("g1", "convention", "db", "Driver", 0.1)},
			limit:     8,
			wantIDs:   []string{"w1"},
		},
		{
			name:      "title collision without key keeps the workspace copy",
			primary:   []projectmemory.SearchHit{hit("w1", "convention", "", "Same Title", 0.9)},
			secondary: []projectmemory.SearchHit{hit("g1", "convention", "", "same title", 0.1)},
			limit:     8,
			wantIDs:   []string{"w1"},
		},
		{
			name:      "limit keeps the most relevant across both tiers",
			primary:   []projectmemory.SearchHit{hit("w1", "s", "", "A", 0.9), hit("w2", "s", "", "B", 0.8)},
			secondary: []projectmemory.SearchHit{hit("g1", "s", "", "C", 0.1)},
			limit:     2,
			wantIDs:   []string{"g1", "w2"},
		},
		{
			name:      "non-positive limit falls back to the default",
			primary:   []projectmemory.SearchHit{hit("w1", "s", "", "A", 0.5)},
			secondary: []projectmemory.SearchHit{hit("g1", "s", "", "B", 0.4)},
			limit:     0,
			wantIDs:   []string{"g1", "w1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			merged := mergeSearchHits(tt.primary, tt.secondary, tt.limit)
			gotIDs := make([]string, 0, len(merged))
			for _, h := range merged {
				gotIDs = append(gotIDs, h.ID)
			}
			if !slices.Equal(gotIDs, tt.wantIDs) {
				t.Errorf("merged ids = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

func TestWriteNativeMemoryWritesIndexAndFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	memories := []projectmemory.Memory{
		{
			ID:    "mem-1",
			Scope: "convention",
			Key:   "db-driver",
			Title: "Use modernc sqlite",
			Body:  "Pure-Go, CGO-free.",
			Tags:  []string{"go", "sqlite"},
		},
		{ID: "mem-2", Scope: "gotcha", Title: "WAL checkpoint", Body: "Checkpoint before close."},
	}

	count, err := writeNativeMemory(dir, memories)
	if err != nil {
		t.Fatalf("writeNativeMemory: %v", err)
	}
	if count != len(memories) {
		t.Fatalf("count = %d, want %d", count, len(memories))
	}

	index := readFileString(t, filepath.Join(dir, "MEMORY.md"))
	for _, want := range []string{"Use modernc sqlite", "convention__db-driver.md", "WAL checkpoint", "mem-2.md"} {
		if !strings.Contains(index, want) {
			t.Errorf("index missing %q:\n%s", want, index)
		}
	}

	keyed := readFileString(t, filepath.Join(dir, "convention__db-driver.md"))
	for _, want := range []string{"# Use modernc sqlite", "Pure-Go, CGO-free.", "_tags: go, sqlite_"} {
		if !strings.Contains(keyed, want) {
			t.Errorf("keyed file missing %q:\n%s", want, keyed)
		}
	}

	keyless := readFileString(t, filepath.Join(dir, "mem-2.md"))
	if !strings.Contains(keyless, "Checkpoint before close.") {
		t.Errorf("keyless file missing body:\n%s", keyless)
	}
}

func TestWriteNativeMemoryDetectsFilenameCollision(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	memories := []projectmemory.Memory{
		{ID: "mem-1", Scope: "convention", Key: "dup", Title: "First", Body: "a"},
		{ID: "mem-2", Scope: "convention", Key: "dup", Title: "Second", Body: "b"},
	}

	if _, err := writeNativeMemory(dir, memories); err == nil {
		t.Fatal("expected a collision error, got nil")
	} else if !strings.Contains(err.Error(), "collision") {
		t.Fatalf("error = %v, want a collision error", err)
	}
}

func TestPromoteMemoryCopiesAndUpsertsInPlace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	source := openTempMemory(t)
	destination := openTempMemory(t)

	added, err := source.Add(ctx, projectmemory.AddInput{
		Scope: "convention",
		Key:   "db-driver",
		Title: "Use modernc sqlite",
		Body:  "Pure-Go, CGO-free.",
		Tags:  []string{"go"},
	})
	if err != nil {
		t.Fatalf("seed source: %v", err)
	}

	promoted, err := promoteMemory(ctx, source, destination, added.ID)
	if err != nil {
		t.Fatalf("promoteMemory: %v", err)
	}
	if promoted.Title != added.Title || promoted.Body != added.Body {
		t.Errorf("promoted fact = %+v, want title/body from %+v", promoted, added)
	}
	if promoted.Source != "promote" {
		t.Errorf("promoted source = %q, want %q", promoted.Source, "promote")
	}

	inGlobal, err := destination.GetByKey(ctx, "convention", "db-driver")
	if err != nil {
		t.Fatalf("get promoted from destination: %v", err)
	}
	if inGlobal.Body != "Pure-Go, CGO-free." {
		t.Errorf("destination body = %q, want original", inGlobal.Body)
	}

	// Re-promoting a refreshed fact must update the global copy in place, not duplicate it.
	refreshed := "Pure-Go, CGO-free, FTS5 built in."
	if _, err := source.Update(ctx, added.ID, projectmemory.UpdateInput{Body: &refreshed}); err != nil {
		t.Fatalf("update source: %v", err)
	}
	if _, err := promoteMemory(ctx, source, destination, added.ID); err != nil {
		t.Fatalf("re-promote: %v", err)
	}

	all, err := destination.List(ctx, projectmemory.ListFilter{Scope: "convention"})
	if err != nil {
		t.Fatalf("list destination: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("destination has %d convention memories, want 1 (upsert in place)", len(all))
	}
	if all[0].Body != refreshed {
		t.Errorf("destination body = %q, want refreshed %q", all[0].Body, refreshed)
	}
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

package projectmemory

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeEmbedder maps text to a small "concept presence" vector so that synonyms with no
// shared characters (e.g. "telemetry" and "logging") land on the same dimension and thus
// have cosine similarity 1. This lets hybrid retrieval be tested deterministically without
// a real model.
type fakeEmbedder struct {
	model string
	fail  bool
}

func (f fakeEmbedder) Model() string { return f.model }

func (f fakeEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	if f.fail {
		return nil, errors.New("embedder unavailable")
	}
	groups := [][]string{
		{"logging", "telemetry", "observability", "logs", "slog"},
		{"sqlite", "database", "driver"},
		{"race", "concurrency", "mutex", "goroutine"},
	}
	lower := strings.ToLower(text)
	vec := make([]float32, len(groups))
	for dim, words := range groups {
		for _, word := range words {
			if strings.Contains(lower, word) {
				vec[dim] = 1
				break
			}
		}
	}
	return vec, nil
}

func newTestStoreWithEmbedder(t *testing.T, embedder Embedder) *Store {
	t.Helper()
	ctx := context.Background()

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var tick int64
	clock := func() time.Time {
		tick++
		return base.Add(time.Duration(tick) * time.Second)
	}

	st, err := Open(ctx, filepath.Join(t.TempDir(), DBFileName), WithClock(clock), WithEmbedder(embedder))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := st.Close(ctx); closeErr != nil {
			t.Errorf("close store: %v", closeErr)
		}
	})
	return st
}

// TestHybridRetrieverFindsSemanticMatch is the core payoff: a query whose term never
// appears literally in any memory is still found because it is semantically related.
func TestHybridRetrieverFindsSemanticMatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	embedder := fakeEmbedder{model: "concept"}
	st := newTestStoreWithEmbedder(t, embedder)

	added, err := st.Add(ctx, AddInput{Scope: "convention", Title: "Logs", Body: "Use slog for logging everywhere."})
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	// "telemetry" shares no token with the memory, so the lexical leg finds nothing,
	// but it maps to the same concept dimension as "logging".
	if lexical, err := st.Search(ctx, SearchQuery{Text: "telemetry"}); err != nil || len(lexical) != 0 {
		t.Fatalf("precondition: lexical should miss; hits=%d err=%v", len(lexical), err)
	}

	hybrid := NewHybridRetriever(st, embedder)
	hits, err := hybrid.Search(ctx, SearchQuery{Text: "telemetry"})
	if err != nil {
		t.Fatalf("hybrid search: %v", err)
	}
	if len(hits) != 1 || hits[0].ID != added.ID {
		t.Fatalf("hybrid semantic match failed: %+v", hits)
	}
}

func TestHybridRetrieverDegradesToLexicalWhenEmbedderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t) // no embedder: no vectors written

	_, err := st.Add(ctx, AddInput{Scope: "convention", Title: "Driver", Body: "Use the sqlite driver."})
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	hybrid := NewHybridRetriever(st, fakeEmbedder{model: "concept", fail: true})
	hits, err := hybrid.Search(ctx, SearchQuery{Text: "sqlite"})
	if err != nil {
		t.Fatalf("hybrid search should not fail when embedder errors: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("degraded lexical hits=%d, want 1", len(hits))
	}
}

func TestStoreEmbedsOnAddAndReindex(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStoreWithEmbedder(t, fakeEmbedder{model: "concept"})

	added, err := st.Add(ctx, AddInput{Scope: "convention", Title: "Logs", Body: "Use slog for logging."})
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	candidates, err := st.vectorCandidates(ctx, "", "concept")
	if err != nil {
		t.Fatalf("vector candidates: %v", err)
	}
	if len(candidates) != 1 || candidates[0].ID != added.ID {
		t.Fatalf("expected a stored vector for the added memory, got %d", len(candidates))
	}

	count, err := st.Reindex(ctx)
	if err != nil {
		t.Fatalf("reindex: %v", err)
	}
	if count != 1 {
		t.Fatalf("reindex count=%d, want 1", count)
	}
}

func TestReindexWithoutEmbedderFails(t *testing.T) {
	t.Parallel()
	if _, err := newTestStore(t).Reindex(context.Background()); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("got %v, want ErrInvalidInput", err)
	}
}

func TestDeleteCascadesToVector(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStoreWithEmbedder(t, fakeEmbedder{model: "concept"})

	added, err := st.Add(ctx, AddInput{Scope: "gotcha", Title: "Race", Body: "Guard the registry against a race."})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := st.Delete(ctx, added.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	candidates, err := st.vectorCandidates(ctx, "", "concept")
	if err != nil {
		t.Fatalf("vector candidates: %v", err)
	}
	if len(candidates) != 0 {
		t.Fatalf("vector survived memory delete: %d candidates", len(candidates))
	}
}

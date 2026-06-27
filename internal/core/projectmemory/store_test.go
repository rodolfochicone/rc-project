package projectmemory

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// newTestStore opens an isolated project-memory database backed by a temp directory.
// The clock advances one second per call so updated_at ordering is deterministic.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var tick int64
	clock := func() time.Time {
		tick++
		return base.Add(time.Duration(tick) * time.Second)
	}

	st, err := Open(ctx, filepath.Join(t.TempDir(), DBFileName), WithClock(clock))
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

func TestOpenAppliesMigrationsIdempotently(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), DBFileName)

	first, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	if _, err := first.Add(ctx, AddInput{Scope: "decision", Title: "t", Body: "b"}); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := first.Close(ctx); err != nil {
		t.Fatalf("close first: %v", err)
	}

	// Reopening must re-run migration bookkeeping without error and preserve data.
	second, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	t.Cleanup(func() { _ = second.Close(ctx) })

	got, err := second.List(ctx, ListFilter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("after reopen got %d memories, want 1", len(got))
	}
}

func TestAddAndGetRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	added, err := st.Add(ctx, AddInput{
		Scope:  "convention",
		Key:    "db-driver",
		Title:  "Use modernc.org/sqlite",
		Body:   "Pure-Go, CGO-free, FTS5 built in.",
		Tags:   []string{"DB", "sqlite", "db"}, // mixed case + duplicate
		Source: "rc-execute-task",
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if added.ID == "" {
		t.Fatal("add returned empty id")
	}
	// Tags must be normalized (lowercased, de-duplicated, sorted) so retrieval is stable.
	if want := []string{"db", "sqlite"}; !equalStrings(added.Tags, want) {
		t.Fatalf("tags = %v, want %v", added.Tags, want)
	}

	got, err := st.Get(ctx, added.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != added.Title || got.Body != added.Body || got.Key != "db-driver" {
		t.Fatalf("get mismatch: %+v", got)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Fatal("timestamps must be populated")
	}
}

// TestAddUpsertsByKey encodes the core promise: a stable (scope, key) lets a skill
// refresh a fact without creating a duplicate, while preserving the original created_at.
func TestAddUpsertsByKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	first, err := st.Add(ctx, AddInput{Scope: "gotcha", Key: "race", Title: "Old", Body: "old body"})
	if err != nil {
		t.Fatalf("first add: %v", err)
	}

	second, err := st.Add(ctx, AddInput{Scope: "gotcha", Key: "race", Title: "New", Body: "new body"})
	if err != nil {
		t.Fatalf("second add: %v", err)
	}

	all, err := st.List(ctx, ListFilter{Scope: "gotcha"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("upsert created %d rows, want 1", len(all))
	}
	if second.Title != "New" || second.Body != "new body" {
		t.Fatalf("upsert did not refresh content: %+v", second)
	}
	if !second.CreatedAt.Equal(first.CreatedAt) {
		t.Fatalf("upsert changed created_at: first=%s second=%s", first.CreatedAt, second.CreatedAt)
	}
	if !second.UpdatedAt.After(first.UpdatedAt) {
		t.Fatalf("upsert did not advance updated_at: first=%s second=%s", first.UpdatedAt, second.UpdatedAt)
	}
}

func TestAddWithoutKeyAllowsDistinctRecords(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	for _, body := range []string{"first", "second"} {
		if _, err := st.Add(ctx, AddInput{Scope: "context", Title: "note", Body: body}); err != nil {
			t.Fatalf("add %q: %v", body, err)
		}
	}
	all, err := st.List(ctx, ListFilter{Scope: "context"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("keyless adds collapsed to %d rows, want 2", len(all))
	}
}

func TestAddValidationRejectsEmptyRequiredFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	cases := map[string]AddInput{
		"missing scope": {Title: "t", Body: "b"},
		"missing title": {Scope: "s", Body: "b"},
		"missing body":  {Scope: "s", Title: "t"},
		"blank title":   {Scope: "s", Title: "   ", Body: "b"},
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := st.Add(ctx, in); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("got %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestGetMissingReturnsNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	if _, err := st.Get(ctx, "mem-does-not-exist"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestUpdateAppliesPartialChanges(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	added, err := st.Add(ctx, AddInput{Scope: "decision", Title: "Keep title", Body: "old"})
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	newBody := "updated body"
	updated, err := st.Update(ctx, added.ID, UpdateInput{Body: &newBody})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Title != "Keep title" {
		t.Fatalf("title changed unexpectedly: %q", updated.Title)
	}
	if updated.Body != newBody {
		t.Fatalf("body = %q, want %q", updated.Body, newBody)
	}
	if !updated.UpdatedAt.After(added.UpdatedAt) {
		t.Fatal("update did not advance updated_at")
	}
}

func TestUpdateCannotClearRequiredFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	added, err := st.Add(ctx, AddInput{Scope: "decision", Title: "t", Body: "b"})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	blank := "   "
	if _, err := st.Update(ctx, added.ID, UpdateInput{Body: &blank}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("got %v, want ErrInvalidInput", err)
	}
}

func TestUpdateMissingReturnsNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	title := "x"
	if _, err := st.Update(ctx, "nope", UpdateInput{Title: &title}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestDeleteRemovesAndReportsMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	added, err := st.Add(ctx, AddInput{Scope: "decision", Title: "t", Body: "b"})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := st.Delete(ctx, added.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := st.Get(ctx, added.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("get after delete: got %v, want ErrNotFound", err)
	}
	if err := st.Delete(ctx, added.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second delete: got %v, want ErrNotFound", err)
	}
}

func TestListFiltersByScopeAndTagNewestFirst(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	seed := []AddInput{
		{Scope: "convention", Title: "older", Body: "b", Tags: []string{"go"}},
		{Scope: "convention", Title: "newer", Body: "b", Tags: []string{"go", "sql"}},
		{Scope: "gotcha", Title: "other", Body: "b", Tags: []string{"go"}},
	}
	for _, in := range seed {
		if _, err := st.Add(ctx, in); err != nil {
			t.Fatalf("seed %q: %v", in.Title, err)
		}
	}

	byScope, err := st.List(ctx, ListFilter{Scope: "convention"})
	if err != nil {
		t.Fatalf("list by scope: %v", err)
	}
	if len(byScope) != 2 {
		t.Fatalf("scope filter returned %d, want 2", len(byScope))
	}
	if byScope[0].Title != "newer" {
		t.Fatalf("expected newest first, got %q", byScope[0].Title)
	}

	bySQLTag, err := st.List(ctx, ListFilter{Tag: "sql"})
	if err != nil {
		t.Fatalf("list by tag: %v", err)
	}
	if len(bySQLTag) != 1 || bySQLTag[0].Title != "newer" {
		t.Fatalf("tag filter = %+v, want only 'newer'", bySQLTag)
	}
}

// TestSearchRanksByRelevance verifies BM25 ordering and scope filtering — the property
// that makes the store useful to a skill: the most relevant memory comes first.
func TestSearchRanksByRelevance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	seed := []AddInput{
		{Scope: "convention", Title: "SQLite driver choice", Body: "Use modernc sqlite driver for the database layer."},
		{Scope: "convention", Title: "Logging", Body: "Use slog for structured logging everywhere."},
		{Scope: "gotcha", Title: "sqlite wal", Body: "Checkpoint the sqlite WAL before closing the database."},
	}
	for _, in := range seed {
		if _, err := st.Add(ctx, in); err != nil {
			t.Fatalf("seed %q: %v", in.Title, err)
		}
	}

	hits, err := st.Search(ctx, SearchQuery{Text: "sqlite database"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("search returned no hits")
	}
	if !contains(hits[0].Title, "SQLite", "sqlite") {
		t.Fatalf("top hit %q not about sqlite", hits[0].Title)
	}
	// Lower BM25 is more relevant; results must be non-decreasing.
	for i := 1; i < len(hits); i++ {
		if hits[i].Score < hits[i-1].Score {
			t.Fatalf("results not ordered by score: %f before %f", hits[i-1].Score, hits[i].Score)
		}
	}

	scoped, err := st.Search(ctx, SearchQuery{Text: "sqlite", Scope: "gotcha"})
	if err != nil {
		t.Fatalf("scoped search: %v", err)
	}
	for _, hit := range scoped {
		if hit.Scope != "gotcha" {
			t.Fatalf("scoped search leaked scope %q", hit.Scope)
		}
	}
}

// TestSearchToleratesSpecialCharacters guards against FTS5 syntax errors leaking from
// raw user input: quotes and operator-like characters must not crash the query.
func TestSearchToleratesSpecialCharacters(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	_, err := st.Add(ctx, AddInput{
		Scope: "context",
		Title: "note",
		Body:  "the daemon exits before flush",
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	for _, raw := range []string{`daemon "flush"`, `daemon OR exit`, `(flush)`, `   `} {
		if _, err := st.Search(ctx, SearchQuery{Text: raw}); err != nil && !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("search(%q) errored unexpectedly: %v", raw, err)
		}
	}
}

// TestSearchReflectsUpdatesAndDeletes proves the FTS triggers keep the index in sync.
func TestSearchReflectsUpdatesAndDeletes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	added, err := st.Add(ctx, AddInput{Scope: "context", Title: "alpha", Body: "kangaroo marsupial"})
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	newBody := "penguin antarctic bird"
	if _, err := st.Update(ctx, added.ID, UpdateInput{Body: &newBody}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if hits, err := st.Search(ctx, SearchQuery{Text: "kangaroo"}); err != nil || len(hits) != 0 {
		t.Fatalf("stale term still indexed: hits=%d err=%v", len(hits), err)
	}
	if hits, err := st.Search(ctx, SearchQuery{Text: "penguin"}); err != nil || len(hits) != 1 {
		t.Fatalf("updated term not indexed: hits=%d err=%v", len(hits), err)
	}

	if err := st.Delete(ctx, added.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if hits, err := st.Search(ctx, SearchQuery{Text: "penguin"}); err != nil || len(hits) != 0 {
		t.Fatalf("deleted memory still searchable: hits=%d err=%v", len(hits), err)
	}
}

// TestSearchPrefixMatch verifies the prefix stage: a partial term still finds the memory.
func TestSearchPrefixMatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	_, err := st.Add(ctx, AddInput{Scope: "convention", Title: "Driver", Body: "Use the sqlite driver."})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	hits, err := st.Search(ctx, SearchQuery{Text: "sqlit"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("prefix search hits=%d, want 1", len(hits))
	}
}

// TestSearchOrExpansionFindsPartialMatches verifies that when no single memory contains
// all terms (so the exact AND stage is empty), the OR-prefix stage still surfaces the
// memories that match some terms.
func TestSearchOrExpansionFindsPartialMatches(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	for _, in := range []AddInput{
		{Scope: "convention", Title: "Driver", Body: "Use the sqlite driver."},
		{Scope: "convention", Title: "Logs", Body: "Use slog for logging."},
	} {
		if _, err := st.Add(ctx, in); err != nil {
			t.Fatalf("add %q: %v", in.Title, err)
		}
	}

	hits, err := st.Search(ctx, SearchQuery{Text: "sqlite logging"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 2 {
		t.Fatalf("or-expansion hits=%d, want 2", len(hits))
	}
}

// TestSearchTrigramMatchesSubstringInIdentifier verifies the trigram stage: a term that
// appears only as a substring inside a camelCase identifier is still found, which the
// unicode61 exact/prefix stages cannot do.
func TestSearchTrigramMatchesSubstringInIdentifier(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	if _, err := st.Add(ctx, AddInput{
		Scope: "gotcha",
		Title: "Pooling",
		Body:  "OpenSQLiteDatabase configures the connection pool.",
	}); err != nil {
		t.Fatalf("add: %v", err)
	}

	hits, err := st.Search(ctx, SearchQuery{Text: "sqlite"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("trigram substring hits=%d, want 1", len(hits))
	}
}

// TestSearchRanksExactBeforeBroadMatches verifies the cascade ordering: a memory that
// matches all terms exactly ranks ahead of one that only matches via OR-expansion.
func TestSearchRanksExactBeforeBroadMatches(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	exact, err := st.Add(ctx, AddInput{Scope: "context", Title: "Both", Body: "alpha and beta together."})
	if err != nil {
		t.Fatalf("add exact: %v", err)
	}
	if _, err := st.Add(ctx, AddInput{Scope: "context", Title: "One", Body: "only alpha here."}); err != nil {
		t.Fatalf("add partial: %v", err)
	}

	hits, err := st.Search(ctx, SearchQuery{Text: "alpha beta"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 2 {
		t.Fatalf("hits=%d, want 2", len(hits))
	}
	if hits[0].ID != exact.ID {
		t.Fatalf("expected exact match first, got %q", hits[0].Title)
	}
}

func TestNormalizeTags(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   []string
		want string
	}{
		{"empty", nil, ""},
		{"trim lower dedupe sort", []string{" Go ", "sql", "go", "SQL"}, "go,sql"},
		{"drops blanks", []string{"", "  ", "x"}, "x"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeTags(tc.in); got != tc.want {
				t.Fatalf("normalizeTags(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestTokenizeQuery(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "   ", nil},
		{"splits on punctuation", "log/slog driver", []string{"log", "slog", "driver"}},
		{"dedupes", "sqlite sqlite db", []string{"sqlite", "db"}},
		{"keeps operator words as plain tokens", "a OR b", []string{"a", "OR", "b"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tokenizeQuery(tc.in); !equalStrings(got, tc.want) {
				t.Fatalf("tokenizeQuery(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestMatchExprBuilders(t *testing.T) {
	t.Parallel()
	tokens := []string{"sqlite", "db", "go"}

	if got, want := exactExpr(tokens), `"sqlite" "db" "go"`; got != want {
		t.Fatalf("exactExpr = %q, want %q", got, want)
	}
	if got, want := prefixExpr(tokens), `"sqlite"* "db"* "go"*`; got != want {
		t.Fatalf("prefixExpr = %q, want %q", got, want)
	}
	if got, want := orPrefixExpr(tokens), `"sqlite"* OR "db"* OR "go"*`; got != want {
		t.Fatalf("orPrefixExpr = %q, want %q", got, want)
	}
	// Only terms with at least three runes survive the trigram stage.
	if got, want := trigramExpr(tokens), `"sqlite"`; got != want {
		t.Fatalf("trigramExpr = %q, want %q", got, want)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func contains(s string, subs ...string) bool {
	for _, sub := range subs {
		if sub != "" && strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

package projectmemory

import (
	"context"
	"errors"
	"testing"
	"time"
)

func importStamp(offset int) time.Time {
	return time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC).Add(time.Duration(offset) * time.Hour)
}

func TestImport_InsertsAbsentRecordPreservingIdentityAndTimestamps(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	created := importStamp(0)
	updated := importStamp(1)
	record := Memory{
		ID:        "mem-import-001",
		Scope:     "decision",
		Key:       "shared",
		Title:     "Title",
		Body:      "Body",
		Tags:      []string{"b", "a"},
		Source:    "rc-import",
		CreatedAt: created,
		UpdatedAt: updated,
	}

	result, err := st.Import(ctx, []Memory{record})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if result.Added != 1 || result.Updated != 0 || result.Skipped != 0 {
		t.Fatalf("result = %+v, want Added=1", result)
	}

	got, err := st.Get(ctx, "mem-import-001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.CreatedAt.Equal(created) || !got.UpdatedAt.Equal(updated) {
		t.Fatalf("timestamps not preserved: created=%s updated=%s", got.CreatedAt, got.UpdatedAt)
	}
	if got.Key != "shared" || got.Source != "rc-import" {
		t.Fatalf("fields not preserved: %+v", got)
	}
}

func TestImport_MostRecentWinsByUpdatedAt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name            string
		incomingUpdated time.Time
		wantOutcome     string
		wantBody        string
	}{
		{name: "strictly newer updates", incomingUpdated: importStamp(5), wantOutcome: "updated", wantBody: "new body"},
		{name: "equal is skipped", incomingUpdated: importStamp(1), wantOutcome: "skipped", wantBody: "old body"},
		{name: "older is skipped", incomingUpdated: importStamp(0), wantOutcome: "skipped", wantBody: "old body"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			st := newTestStore(t)

			seed := Memory{
				ID: "mem-existing", Scope: "decision", Key: "k",
				Title: "old", Body: "old body",
				CreatedAt: importStamp(0), UpdatedAt: importStamp(1),
			}
			if _, err := st.Import(ctx, []Memory{seed}); err != nil {
				t.Fatalf("seed import: %v", err)
			}

			incoming := Memory{
				ID: "mem-different-id", Scope: "decision", Key: "k",
				Title: "new", Body: "new body",
				CreatedAt: importStamp(0), UpdatedAt: tc.incomingUpdated,
			}
			result, err := st.Import(ctx, []Memory{incoming})
			if err != nil {
				t.Fatalf("Import: %v", err)
			}

			switch tc.wantOutcome {
			case "updated":
				if result.Updated != 1 {
					t.Fatalf("result = %+v, want Updated=1", result)
				}
			case "skipped":
				if result.Skipped != 1 {
					t.Fatalf("result = %+v, want Skipped=1", result)
				}
			}

			got, err := st.GetByKey(ctx, "decision", "k")
			if err != nil {
				t.Fatalf("GetByKey: %v", err)
			}
			if got.Body != tc.wantBody {
				t.Fatalf("body = %q, want %q", got.Body, tc.wantBody)
			}
			if got.ID != "mem-existing" {
				t.Fatalf("keyed update must keep existing id, got %q", got.ID)
			}
		})
	}
}

func TestImport_KeylessMatchedById(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	first := Memory{
		ID:        "mem-keyless",
		Scope:     "gotcha",
		Title:     "t",
		Body:      "first",
		CreatedAt: importStamp(0),
		UpdatedAt: importStamp(1),
	}
	if _, err := st.Import(ctx, []Memory{first}); err != nil {
		t.Fatalf("first import: %v", err)
	}

	second := Memory{
		ID:        "mem-keyless",
		Scope:     "gotcha",
		Title:     "t",
		Body:      "second",
		CreatedAt: importStamp(0),
		UpdatedAt: importStamp(2),
	}
	result, err := st.Import(ctx, []Memory{second})
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if result.Updated != 1 {
		t.Fatalf("result = %+v, want Updated=1", result)
	}

	got, err := st.Get(ctx, "mem-keyless")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Body != "second" {
		t.Fatalf("body = %q, want %q", got.Body, "second")
	}
}

func TestImport_NeverDeletesExistingRows(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	keep := Memory{
		ID:        "mem-keep",
		Scope:     "decision",
		Key:       "keep",
		Title:     "t",
		Body:      "keep",
		CreatedAt: importStamp(0),
		UpdatedAt: importStamp(1),
	}
	if _, err := st.Import(ctx, []Memory{keep}); err != nil {
		t.Fatalf("seed import: %v", err)
	}

	other := Memory{
		ID:        "mem-other",
		Scope:     "decision",
		Key:       "other",
		Title:     "t",
		Body:      "other",
		CreatedAt: importStamp(0),
		UpdatedAt: importStamp(1),
	}
	if _, err := st.Import(ctx, []Memory{other}); err != nil {
		t.Fatalf("second import: %v", err)
	}

	if _, err := st.GetByKey(ctx, "decision", "keep"); err != nil {
		t.Fatalf("pre-existing row was removed: %v", err)
	}
}

func TestImport_RollsBackEntireBatchOnError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	valid := Memory{
		ID:        "mem-valid",
		Scope:     "decision",
		Key:       "valid",
		Title:     "t",
		Body:      "b",
		CreatedAt: importStamp(0),
		UpdatedAt: importStamp(1),
	}
	invalid := Memory{
		ID:        "",
		Scope:     "decision",
		Title:     "t",
		Body:      "b",
		CreatedAt: importStamp(0),
		UpdatedAt: importStamp(1),
	}

	_, err := st.Import(ctx, []Memory{valid, invalid})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Import error = %v, want ErrInvalidInput", err)
	}

	if _, err := st.Get(ctx, "mem-valid"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("valid record must have rolled back, Get err = %v", err)
	}
}

func TestImport_KeepsSearchIndexConsistent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	record := Memory{
		ID:        "mem-search",
		Scope:     "decision",
		Key:       "idx",
		Title:     "unicornkeyword",
		Body:      "b",
		CreatedAt: importStamp(0),
		UpdatedAt: importStamp(1),
	}
	if _, err := st.Import(ctx, []Memory{record}); err != nil {
		t.Fatalf("Import: %v", err)
	}

	hits, err := st.Search(ctx, SearchQuery{Text: "unicornkeyword"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 1 || hits[0].ID != "mem-search" {
		t.Fatalf("search after import = %+v, want one hit mem-search", hits)
	}
}

func TestImport_IsIdempotent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	records := []Memory{
		{
			ID:        "mem-a",
			Scope:     "decision",
			Key:       "a",
			Title:     "t",
			Body:      "a",
			CreatedAt: importStamp(0),
			UpdatedAt: importStamp(1),
		},
		{ID: "mem-b", Scope: "gotcha", Title: "t", Body: "b", CreatedAt: importStamp(0), UpdatedAt: importStamp(1)},
	}
	if _, err := st.Import(ctx, records); err != nil {
		t.Fatalf("first import: %v", err)
	}

	result, err := st.Import(ctx, records)
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if result.Added != 0 || result.Updated != 0 || result.Skipped != 2 {
		t.Fatalf("re-import result = %+v, want all skipped", result)
	}
}

func TestImport_RejectsZeroTimestamps(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name   string
		record Memory
	}{
		{
			name:   "zero created_at",
			record: Memory{ID: "mem-z1", Scope: "decision", Title: "t", Body: "b", UpdatedAt: importStamp(1)},
		},
		{
			name:   "zero updated_at",
			record: Memory{ID: "mem-z2", Scope: "decision", Title: "t", Body: "b", CreatedAt: importStamp(1)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			st := newTestStore(t)
			if _, err := st.Import(ctx, []Memory{tc.record}); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Import error = %v, want ErrInvalidInput", err)
			}
			if _, err := st.Get(ctx, tc.record.ID); !errors.Is(err, ErrNotFound) {
				t.Fatalf("record must not have been written, Get err = %v", err)
			}
		})
	}
}

func TestImport_EmptyBatchIsNoOp(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := newTestStore(t)

	result, err := st.Import(ctx, nil)
	if err != nil {
		t.Fatalf("Import(nil): %v", err)
	}
	if result != (ImportResult{}) {
		t.Fatalf("result = %+v, want zero", result)
	}
}

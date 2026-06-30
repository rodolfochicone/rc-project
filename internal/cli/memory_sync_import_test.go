package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rodolfochicone/rc-project/internal/core/projectmemory"
)

func writeMirrorFile(t *testing.T, ws string, memory projectmemory.Memory) {
	t.Helper()
	data, err := projectmemory.MarshalMemory(memory)
	if err != nil {
		t.Fatalf("MarshalMemory: %v", err)
	}
	if err := os.MkdirAll(mirrorDir(ws), 0o755); err != nil {
		t.Fatalf("mkdir mirror: %v", err)
	}
	path := filepath.Join(mirrorDir(ws), projectmemory.MirrorFileName(memory))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write mirror file: %v", err)
	}
}

func listMemories(t *testing.T, ws string) map[string]memoryJSON {
	t.Helper()
	out, err := runMemory(t, ws, "list", "--format", "json")
	if err != nil {
		t.Fatalf("list: %v\n%s", err, out)
	}
	var records []memoryJSON
	if err := json.Unmarshal([]byte(out), &records); err != nil {
		t.Fatalf("unmarshal list: %v\n%s", err, out)
	}
	byID := make(map[string]memoryJSON, len(records))
	for i := range records {
		byID[records[i].ID] = records[i]
	}
	return byID
}

func TestMemoryImportRoundTripAcrossWorkspaces(t *testing.T) {
	source := newMemoryWorkspace(t)
	addMemory(t, source, "--scope", "decision", "--key", "sharing", "--title", "Share", "--body", "Commit a mirror.")
	addMemory(t, source, "--scope", "convention", "--key", "driver", "--title", "Driver", "--body", "Use modernc.")
	addMemory(t, source, "--scope", "gotcha", "--title", "WAL", "--body", "Checkpoint on close.")

	if out, err := runMemory(t, source, "export"); err != nil {
		t.Fatalf("export: %v\n%s", err, out)
	}

	// Simulate a fresh clone: copy the committed mirror into a new workspace and import.
	dest := newMemoryWorkspace(t)
	if err := os.MkdirAll(mirrorDir(dest), 0o755); err != nil {
		t.Fatalf("mkdir dest mirror: %v", err)
	}
	for _, name := range mirrorFiles(t, source) {
		data, err := os.ReadFile(filepath.Join(mirrorDir(source), name))
		if err != nil {
			t.Fatalf("read source mirror %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(mirrorDir(dest), name), data, 0o600); err != nil {
			t.Fatalf("write dest mirror %s: %v", name, err)
		}
	}

	if out, err := runMemory(t, dest, "import"); err != nil {
		t.Fatalf("import: %v\n%s", err, out)
	}

	want := listMemories(t, source)
	got := listMemories(t, dest)
	if len(got) != len(want) {
		t.Fatalf("imported %d memories, want %d", len(got), len(want))
	}
	for id, wantRecord := range want {
		gotRecord, ok := got[id]
		if !ok {
			t.Fatalf("imported set missing id %q", id)
		}
		if gotRecord.Body != wantRecord.Body || gotRecord.Scope != wantRecord.Scope || gotRecord.Key != wantRecord.Key {
			t.Fatalf("round-trip mismatch for %q: got %+v want %+v", id, gotRecord, wantRecord)
		}
	}
}

func TestMemoryImportMostRecentWinsAndAddsNew(t *testing.T) {
	ws := newMemoryWorkspace(t)

	// Local record carries a fresh (now) timestamp.
	local := addMemory(t, ws, "--scope", "decision", "--key", "k", "--title", "local", "--body", "local-new")

	// A mirror file for the same (scope, key) but older — must lose to the local version.
	old := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	writeMirrorFile(t, ws, projectmemory.Memory{
		ID: "mem-remote", Scope: "decision", Key: "k",
		Title: "remote", Body: "remote-old",
		CreatedAt: old, UpdatedAt: old,
	})
	// A brand-new keyless fact from the mirror — must be added.
	writeMirrorFile(t, ws, projectmemory.Memory{
		ID: "mem-newfact", Scope: "context", Title: "new", Body: "added-from-mirror",
		CreatedAt: old, UpdatedAt: old,
	})

	if out, err := runMemory(t, ws, "import"); err != nil {
		t.Fatalf("import: %v\n%s", err, out)
	}

	records := listMemories(t, ws)
	if got := records[local.ID]; got.Body != "local-new" {
		t.Fatalf("local record body = %q, want local-new (local is newer)", got.Body)
	}
	if got, ok := records["mem-newfact"]; !ok || got.Body != "added-from-mirror" {
		t.Fatalf("new keyless fact not imported: %+v", got)
	}
}

func TestMemoryImportSkipsMalformedFileAndExitsNonZero(t *testing.T) {
	ws := newMemoryWorkspace(t)

	valid := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	writeMirrorFile(
		t,
		ws,
		projectmemory.Memory{
			ID:        "mem-1",
			Scope:     "decision",
			Key:       "a",
			Title:     "t",
			Body:      "b1",
			CreatedAt: valid,
			UpdatedAt: valid,
		},
	)
	writeMirrorFile(
		t,
		ws,
		projectmemory.Memory{
			ID:        "mem-2",
			Scope:     "decision",
			Key:       "b",
			Title:     "t",
			Body:      "b2",
			CreatedAt: valid,
			UpdatedAt: valid,
		},
	)
	if err := os.WriteFile(
		filepath.Join(mirrorDir(ws), "broken.md"),
		[]byte("not valid frontmatter"),
		0o600,
	); err != nil {
		t.Fatalf("write broken file: %v", err)
	}

	out, err := runMemory(t, ws, "import")
	if err == nil {
		t.Fatalf("expected non-zero exit for malformed file, got nil\n%s", out)
	}

	records := listMemories(t, ws)
	if len(records) != 2 {
		t.Fatalf("imported %d valid records, want 2 (malformed skipped)", len(records))
	}
}

func TestMemoryImportEmptyMirrorIsNoOp(t *testing.T) {
	ws := newMemoryWorkspace(t)

	out, err := runMemory(t, ws, "import")
	if err != nil {
		t.Fatalf("import: %v\n%s", err, out)
	}
	if records := listMemories(t, ws); len(records) != 0 {
		t.Fatalf("expected empty database, got %d records", len(records))
	}
}

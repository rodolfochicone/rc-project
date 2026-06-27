package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newMemoryWorkspace creates a temp directory containing a .rc folder so workspace
// discovery resolves it as the project root for the memory database.
func newMemoryWorkspace(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ws, ".rc"), 0o755); err != nil {
		t.Fatalf("mkdir .rc: %v", err)
	}
	return ws
}

// runMemory executes `rc memory <args>` in-process against the given workspace and
// returns the combined output and the command error.
func runMemory(t *testing.T, workspace string, args ...string) (string, error) {
	t.Helper()
	full := append([]string{"memory"}, args...)
	full = append(full, "--workspace", workspace)

	cmd := NewRootCommand()
	cmd.SetArgs(full)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	return out.String(), err
}

func TestMemoryAddSearchGetRoundTrip(t *testing.T) {
	ws := newMemoryWorkspace(t)

	out, err := runMemory(t, ws,
		"add",
		"--scope", "convention",
		"--title", "Use modernc sqlite",
		"--body", "Pure-Go, CGO-free sqlite driver with FTS5.",
		"--tags", "go,sqlite",
		"--format", "json",
	)
	if err != nil {
		t.Fatalf("add: %v\n%s", err, out)
	}
	var added memoryJSON
	if err := json.Unmarshal([]byte(out), &added); err != nil {
		t.Fatalf("unmarshal add output: %v\n%s", err, out)
	}
	if added.ID == "" {
		t.Fatal("add returned empty id")
	}
	if want := []string{"go", "sqlite"}; strings.Join(added.Tags, ",") != strings.Join(want, ",") {
		t.Fatalf("tags = %v, want %v", added.Tags, want)
	}

	searchOut, err := runMemory(t, ws, "search", "sqlite", "driver", "--format", "json")
	if err != nil {
		t.Fatalf("search: %v\n%s", err, searchOut)
	}
	var hits []memoryJSON
	if err := json.Unmarshal([]byte(searchOut), &hits); err != nil {
		t.Fatalf("unmarshal search output: %v\n%s", err, searchOut)
	}
	if len(hits) == 0 || hits[0].ID != added.ID {
		t.Fatalf("search did not surface the added memory: %+v", hits)
	}
	if hits[0].Score == nil {
		t.Fatal("search hit missing bm25 score")
	}

	getOut, err := runMemory(t, ws, "get", added.ID, "--format", "json")
	if err != nil {
		t.Fatalf("get: %v\n%s", err, getOut)
	}
	var got memoryJSON
	if err := json.Unmarshal([]byte(getOut), &got); err != nil {
		t.Fatalf("unmarshal get output: %v\n%s", err, getOut)
	}
	if got.Title != added.Title {
		t.Fatalf("get title = %q, want %q", got.Title, added.Title)
	}
}

func TestMemoryUpsertByKeyDoesNotDuplicate(t *testing.T) {
	ws := newMemoryWorkspace(t)

	for _, body := range []string{"old guidance", "new guidance"} {
		if out, err := runMemory(t, ws,
			"add", "--scope", "gotcha", "--key", "race", "--title", "Run loop race", "--body", body,
		); err != nil {
			t.Fatalf("add %q: %v\n%s", body, err, out)
		}
	}

	listOut, err := runMemory(t, ws, "list", "--scope", "gotcha", "--format", "json")
	if err != nil {
		t.Fatalf("list: %v\n%s", err, listOut)
	}
	var memories []memoryJSON
	if err := json.Unmarshal([]byte(listOut), &memories); err != nil {
		t.Fatalf("unmarshal list output: %v\n%s", err, listOut)
	}
	if len(memories) != 1 {
		t.Fatalf("upsert by key produced %d rows, want 1", len(memories))
	}
	if memories[0].Body != "new guidance" {
		t.Fatalf("upsert did not refresh body: %q", memories[0].Body)
	}
}

func TestMemoryUpdateAndDelete(t *testing.T) {
	ws := newMemoryWorkspace(t)

	addOut, err := runMemory(t, ws,
		"add", "--scope", "decision", "--title", "Keep title", "--body", "original", "--format", "json",
	)
	if err != nil {
		t.Fatalf("add: %v\n%s", err, addOut)
	}
	var added memoryJSON
	if err := json.Unmarshal([]byte(addOut), &added); err != nil {
		t.Fatalf("unmarshal add: %v", err)
	}

	updOut, err := runMemory(t, ws, "update", added.ID, "--body", "updated", "--format", "json")
	if err != nil {
		t.Fatalf("update: %v\n%s", err, updOut)
	}
	var updated memoryJSON
	if err := json.Unmarshal([]byte(updOut), &updated); err != nil {
		t.Fatalf("unmarshal update: %v", err)
	}
	if updated.Body != "updated" || updated.Title != "Keep title" {
		t.Fatalf("update result unexpected: %+v", updated)
	}

	if out, err := runMemory(t, ws, "delete", added.ID); err != nil {
		t.Fatalf("delete: %v\n%s", err, out)
	}
	if _, err := runMemory(t, ws, "get", added.ID); err == nil {
		t.Fatal("get after delete: expected error, got nil")
	}
}

func TestMemoryGetMissingReturnsError(t *testing.T) {
	ws := newMemoryWorkspace(t)

	out, err := runMemory(t, ws, "get", "mem-missing")
	if err == nil {
		t.Fatalf("expected error for missing id, got nil; output:\n%s", out)
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %v, want it to mention 'not found'", err)
	}
}

func TestMemoryAddValidationRejectsMissingFields(t *testing.T) {
	ws := newMemoryWorkspace(t)

	if _, err := runMemory(t, ws, "add", "--scope", "decision", "--title", "no body"); err == nil {
		t.Fatal("expected validation error when body is missing")
	}
}

func TestMemoryReindexRequiresEmbeddings(t *testing.T) {
	t.Setenv("RC_MEMORY_EMBEDDINGS", "")
	ws := newMemoryWorkspace(t)
	// With embeddings disabled, reindex has no embedder and must fail.
	if _, err := runMemory(t, ws, "reindex"); err == nil {
		t.Fatal("expected reindex to fail when embeddings are disabled")
	}
}

// memoryEmbedStub stands in for a local Ollama daemon, returning a fixed embedding so the
// hybrid path can be exercised end-to-end without a real model.
func memoryEmbedStub(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": [][]float32{{1, 0, 0}}})
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestMemoryHybridSearchWithEmbeddingsEnabled(t *testing.T) {
	t.Setenv("RC_MEMORY_EMBEDDINGS", "ollama")
	t.Setenv("RC_MEMORY_MODEL", "stub-model")
	t.Setenv("RC_MEMORY_ENDPOINT", memoryEmbedStub(t))

	ws := newMemoryWorkspace(t)
	_, err := runMemory(t, ws, "add", "--scope", "convention", "--title", "T", "--body", "semantic body")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if out, reindexErr := runMemory(t, ws, "reindex"); reindexErr != nil {
		t.Fatalf("reindex: %v\n%s", reindexErr, out)
	}

	out, err := runMemory(t, ws, "search", "anything", "--format", "json")
	if err != nil {
		t.Fatalf("search: %v\n%s", err, out)
	}
	var hits []memoryJSON
	if err := json.Unmarshal([]byte(out), &hits); err != nil {
		t.Fatalf("unmarshal search: %v\n%s", err, out)
	}
	if len(hits) != 1 {
		t.Fatalf("hybrid search hits=%d, want 1", len(hits))
	}
}

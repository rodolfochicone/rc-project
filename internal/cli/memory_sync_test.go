package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rodolfochicone/rc-project/internal/core/projectmemory"
)

func mirrorDir(ws string) string {
	return filepath.Join(ws, ".rc", "memory")
}

func addMemory(t *testing.T, ws string, args ...string) memoryJSON {
	t.Helper()
	out, err := runMemory(t, ws, append([]string{"add", "--format", "json"}, args...)...)
	if err != nil {
		t.Fatalf("add: %v\n%s", err, out)
	}
	var added memoryJSON
	if err := json.Unmarshal([]byte(out), &added); err != nil {
		t.Fatalf("unmarshal add output: %v\n%s", err, out)
	}
	return added
}

func mirrorFiles(t *testing.T, ws string) []string {
	t.Helper()
	entries, err := os.ReadDir(mirrorDir(ws))
	if err != nil {
		t.Fatalf("read mirror dir: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func TestMemoryExportWritesOneFilePerRecord(t *testing.T) {
	ws := newMemoryWorkspace(t)

	addMemory(t, ws, "--scope", "decision", "--key", "sharing", "--title", "Share", "--body", "Commit a mirror.")
	addMemory(t, ws, "--scope", "convention", "--key", "db-driver", "--title", "Driver", "--body", "Use modernc.")
	keyless := addMemory(t, ws, "--scope", "gotcha", "--title", "WAL", "--body", "Checkpoint on close.")

	out, err := runMemory(t, ws, "export")
	if err != nil {
		t.Fatalf("export: %v\n%s", err, out)
	}

	got := mirrorFiles(t, ws)
	want := map[string]bool{
		"decision__sharing.md":     true,
		"convention__db-driver.md": true,
		keyless.ID + ".md":         true,
	}
	if len(got) != len(want) {
		t.Fatalf("mirror files = %v, want %d files", got, len(want))
	}
	for _, name := range got {
		if !want[name] {
			t.Fatalf("unexpected mirror file %q (have %v)", name, got)
		}
	}
}

func TestMemoryExportFileParsesBackToRecord(t *testing.T) {
	ws := newMemoryWorkspace(t)
	added := addMemory(t, ws,
		"--scope", "decision", "--key", "sharing",
		"--title", "Share project memory",
		"--body", "Commit one markdown file per fact.",
		"--tags", "memory,sharing",
		"--source", "rc-analyze",
	)

	if out, err := runMemory(t, ws, "export"); err != nil {
		t.Fatalf("export: %v\n%s", err, out)
	}

	data, err := os.ReadFile(filepath.Join(mirrorDir(ws), "decision__sharing.md"))
	if err != nil {
		t.Fatalf("read mirror file: %v", err)
	}
	parsed, err := projectmemory.ParseMemory(data)
	if err != nil {
		t.Fatalf("ParseMemory: %v", err)
	}
	if parsed.ID != added.ID || parsed.Scope != "decision" || parsed.Key != "sharing" {
		t.Fatalf("parsed identity mismatch: %+v", parsed)
	}
	if parsed.Title != added.Title || parsed.Body != added.Body || parsed.Source != "rc-analyze" {
		t.Fatalf("parsed content mismatch: %+v", parsed)
	}
	if strings.Join(parsed.Tags, ",") != "memory,sharing" {
		t.Fatalf("parsed tags = %v", parsed.Tags)
	}
}

func TestMemoryExportCreatesMirrorDir(t *testing.T) {
	ws := newMemoryWorkspace(t)
	addMemory(t, ws, "--scope", "decision", "--key", "k", "--title", "t", "--body", "b")

	if _, err := os.Stat(mirrorDir(ws)); !os.IsNotExist(err) {
		t.Fatalf("mirror dir should not exist before export, stat err = %v", err)
	}
	if out, err := runMemory(t, ws, "export"); err != nil {
		t.Fatalf("export: %v\n%s", err, out)
	}
	if info, err := os.Stat(mirrorDir(ws)); err != nil || !info.IsDir() {
		t.Fatalf("mirror dir not created: info=%v err=%v", info, err)
	}
}

func TestMemoryExportEmptyDatabaseWritesNoFiles(t *testing.T) {
	ws := newMemoryWorkspace(t)

	out, err := runMemory(t, ws, "export")
	if err != nil {
		t.Fatalf("export: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Exported 0 memories") {
		t.Fatalf("export output = %q, want count 0", out)
	}
	if got := mirrorFiles(t, ws); len(got) != 0 {
		t.Fatalf("mirror files = %v, want none", got)
	}
}

func TestMemoryExportFailsOnFileNameCollision(t *testing.T) {
	ws := newMemoryWorkspace(t)
	// Two distinct memories whose keys sanitize to the same mirror file name.
	addMemory(t, ws, "--scope", "decision", "--key", "db-driver", "--title", "a", "--body", "a")
	addMemory(t, ws, "--scope", "decision", "--key", "DB Driver", "--title", "b", "--body", "b")

	out, err := runMemory(t, ws, "export")
	if err == nil {
		t.Fatalf("expected collision error, got nil\n%s", out)
	}
	if !strings.Contains(out, "collision") {
		t.Fatalf("error output = %q, want a collision message", out)
	}
}

func TestMemoryExportIsIdempotent(t *testing.T) {
	ws := newMemoryWorkspace(t)
	addMemory(t, ws, "--scope", "decision", "--key", "k", "--title", "t", "--body", "b")

	if out, err := runMemory(t, ws, "export"); err != nil {
		t.Fatalf("first export: %v\n%s", err, out)
	}
	first, err := os.ReadFile(filepath.Join(mirrorDir(ws), "decision__k.md"))
	if err != nil {
		t.Fatalf("read after first export: %v", err)
	}

	if out, err := runMemory(t, ws, "export"); err != nil {
		t.Fatalf("second export: %v\n%s", err, out)
	}
	second, err := os.ReadFile(filepath.Join(mirrorDir(ws), "decision__k.md"))
	if err != nil {
		t.Fatalf("read after second export: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("re-export not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

package projectmemory

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func sampleTime(t *testing.T) time.Time {
	t.Helper()
	return time.Date(2026, time.June, 30, 16, 24, 25, 0, time.UTC)
}

func TestMarshalParseMemory_RoundTrip(t *testing.T) {
	t.Parallel()

	stamp := sampleTime(t)
	tests := []struct {
		name   string
		memory Memory
	}{
		{
			name: "keyed with tags and source",
			memory: Memory{
				ID:        "mem-335245db2db2e8e6",
				Scope:     "decision",
				Key:       "project-memory-sharing",
				Title:     "Share project memory via a text mirror",
				Body:      "Commit one markdown file per fact under .rc/memory/.",
				Tags:      []string{"architecture", "memory", "sharing"},
				Source:    "rc-analyze",
				CreatedAt: stamp,
				UpdatedAt: stamp,
			},
		},
		{
			name: "keyless without tags or source",
			memory: Memory{
				ID:        "mem-abc123def456",
				Scope:     "gotcha",
				Title:     "WAL checkpoint on close",
				Body:      "Close checkpoints the WAL; committed rows are durable.",
				CreatedAt: stamp,
				UpdatedAt: stamp.Add(time.Hour),
			},
		},
		{
			name: "multi-line body",
			memory: Memory{
				ID:        "mem-multiline01",
				Scope:     "context",
				Title:     "Build order",
				Body:      "First the format.\nThen the import.\nThen the commands.",
				CreatedAt: stamp,
				UpdatedAt: stamp,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			data, err := MarshalMemory(tc.memory)
			if err != nil {
				t.Fatalf("MarshalMemory: %v", err)
			}

			got, err := ParseMemory(data)
			if err != nil {
				t.Fatalf("ParseMemory: %v", err)
			}

			assertMemoryEqual(t, tc.memory, got)
		})
	}
}

func TestMirrorFileName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		memory Memory
		want   string
	}{
		{
			name:   "keyed maps to scope and key",
			memory: Memory{ID: "mem-1", Scope: "decision", Key: "project-memory-sharing"},
			want:   "decision__project-memory-sharing.md",
		},
		{
			name:   "key with spaces uppercase and slashes is sanitized",
			memory: Memory{ID: "mem-2", Scope: "Convention", Key: "DB Driver/v2"},
			want:   "convention__db-driver-v2.md",
		},
		{
			name:   "keyless maps to id",
			memory: Memory{ID: "mem-abc123", Scope: "gotcha"},
			want:   "mem-abc123.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := MirrorFileName(tc.memory); got != tc.want {
				t.Fatalf("MirrorFileName = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMirrorFileName_SanitizationIsIdempotent(t *testing.T) {
	t.Parallel()

	memory := Memory{ID: "mem-3", Scope: "Convention", Key: "DB Driver/v2"}
	first := MirrorFileName(memory)

	stripped := strings.TrimSuffix(first, ".md")
	parts := strings.SplitN(stripped, "__", 2)
	if len(parts) != 2 {
		t.Fatalf("expected scope__key form, got %q", first)
	}
	reSanitized := Memory{ID: memory.ID, Scope: parts[0], Key: parts[1]}
	if second := MirrorFileName(reSanitized); second != first {
		t.Fatalf("sanitization not idempotent: first %q, second %q", first, second)
	}
}

func TestParseMemory_RejectsMissingRequiredFields(t *testing.T) {
	t.Parallel()

	stamp := marshalledTimestamp(t)
	tests := []struct {
		name string
		data string
	}{
		{
			name: "missing title",
			data: "---\nid: mem-1\nscope: decision\ncreated_at: " + stamp + "\nupdated_at: " + stamp + "\n---\n\nbody text\n",
		},
		{
			name: "missing scope",
			data: "---\nid: mem-1\ntitle: A title\ncreated_at: " + stamp + "\nupdated_at: " + stamp + "\n---\n\nbody text\n",
		},
		{
			name: "missing id",
			data: "---\nscope: decision\ntitle: A title\ncreated_at: " + stamp + "\nupdated_at: " + stamp + "\n---\n\nbody text\n",
		},
		{
			name: "missing body",
			data: "---\nid: mem-1\nscope: decision\ntitle: A title\ncreated_at: " + stamp + "\nupdated_at: " + stamp + "\n---\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := ParseMemory([]byte(tc.data)); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("ParseMemory error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestParseMemory_RejectsUnparseableTimestamp(t *testing.T) {
	t.Parallel()

	data := "---\nid: mem-1\nscope: decision\ntitle: A title\ncreated_at: not-a-time\nupdated_at: not-a-time\n---\n\nbody text\n"
	if _, err := ParseMemory([]byte(data)); err == nil {
		t.Fatal("expected error for unparseable timestamp, got nil")
	}
}

func marshalledTimestamp(t *testing.T) string {
	t.Helper()
	out, err := MarshalMemory(Memory{
		ID:        "x",
		Scope:     "x",
		Title:     "x",
		Body:      "x",
		CreatedAt: sampleTime(t),
		UpdatedAt: sampleTime(t),
	})
	if err != nil {
		t.Fatalf("MarshalMemory: %v", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "created_at:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "created_at:"))
		}
	}
	t.Fatal("created_at not found in marshaled output")
	return ""
}

func assertMemoryEqual(t *testing.T, want, got Memory) {
	t.Helper()
	if got.ID != want.ID || got.Scope != want.Scope || got.Key != want.Key {
		t.Fatalf("identity mismatch: got %+v, want %+v", got, want)
	}
	if got.Title != want.Title || got.Body != want.Body || got.Source != want.Source {
		t.Fatalf("content mismatch: got %+v, want %+v", got, want)
	}
	if !got.CreatedAt.Equal(want.CreatedAt) || !got.UpdatedAt.Equal(want.UpdatedAt) {
		t.Fatalf("timestamp mismatch: got created=%s updated=%s, want created=%s updated=%s",
			got.CreatedAt, got.UpdatedAt, want.CreatedAt, want.UpdatedAt)
	}
	if strings.Join(got.Tags, ",") != strings.Join(want.Tags, ",") {
		t.Fatalf("tags mismatch: got %v, want %v", got.Tags, want.Tags)
	}
}

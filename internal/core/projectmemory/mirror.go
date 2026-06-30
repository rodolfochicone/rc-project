package projectmemory

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rodolfochicone/rc-project/internal/core/frontmatter"
	"github.com/rodolfochicone/rc-project/internal/store"
)

// MirrorDirName is the project-memory text mirror directory, resolved under the .rc base
// directory. It holds one markdown-per-fact file and is the committed, shareable source of
// truth; the SQLite database is a local cache rebuilt from it via `rc memory import`.
const MirrorDirName = "memory"

// mirrorFrontmatter is the YAML front matter of a mirror file. Timestamps are kept as the
// store's canonical text layout so a marshal/parse round-trip is stable.
type mirrorFrontmatter struct {
	ID        string   `yaml:"id"`
	Scope     string   `yaml:"scope"`
	Key       string   `yaml:"key,omitempty"`
	Title     string   `yaml:"title"`
	Tags      []string `yaml:"tags,omitempty"`
	Source    string   `yaml:"source,omitempty"`
	CreatedAt string   `yaml:"created_at"`
	UpdatedAt string   `yaml:"updated_at"`
}

// nonAlphanumeric matches runs of characters that are not lowercase letters or digits; it
// drives mirror filename sanitization.
var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// MarshalMemory renders a memory as a markdown file with YAML front matter followed by the
// body. The inverse of ParseMemory.
func MarshalMemory(m Memory) ([]byte, error) {
	fm := mirrorFrontmatter{
		ID:        strings.TrimSpace(m.ID),
		Scope:     strings.TrimSpace(m.Scope),
		Key:       strings.TrimSpace(m.Key),
		Title:     strings.TrimSpace(m.Title),
		Tags:      splitTags(normalizeTags(m.Tags)),
		Source:    strings.TrimSpace(m.Source),
		CreatedAt: store.FormatTimestamp(m.CreatedAt),
		UpdatedAt: store.FormatTimestamp(m.UpdatedAt),
	}
	out, err := frontmatter.Format(fm, m.Body)
	if err != nil {
		return nil, fmt.Errorf("projectmemory: marshal memory %q: %w", m.ID, err)
	}
	return []byte(out), nil
}

// ParseMemory parses a mirror file back into a Memory. It returns ErrInvalidInput when a
// required field (id, scope, title, body, created_at, updated_at) is missing or malformed.
func ParseMemory(data []byte) (Memory, error) {
	var fm mirrorFrontmatter
	body, err := frontmatter.Parse(string(data), &fm)
	if err != nil {
		return Memory{}, fmt.Errorf("projectmemory: parse memory: %w", err)
	}

	id := strings.TrimSpace(fm.ID)
	scope := strings.TrimSpace(fm.Scope)
	title := strings.TrimSpace(fm.Title)
	memoryBody := strings.TrimRight(body, "\n")
	switch {
	case id == "":
		return Memory{}, fmt.Errorf("projectmemory: parse memory missing id: %w", ErrInvalidInput)
	case scope == "":
		return Memory{}, fmt.Errorf("projectmemory: parse memory missing scope: %w", ErrInvalidInput)
	case title == "":
		return Memory{}, fmt.Errorf("projectmemory: parse memory missing title: %w", ErrInvalidInput)
	case strings.TrimSpace(memoryBody) == "":
		return Memory{}, fmt.Errorf("projectmemory: parse memory missing body: %w", ErrInvalidInput)
	}

	createdAt, err := store.ParseTimestamp(fm.CreatedAt)
	if err != nil {
		return Memory{}, fmt.Errorf("projectmemory: parse memory created_at: %w", err)
	}
	updatedAt, err := store.ParseTimestamp(fm.UpdatedAt)
	if err != nil {
		return Memory{}, fmt.Errorf("projectmemory: parse memory updated_at: %w", err)
	}

	return Memory{
		ID:        id,
		Scope:     scope,
		Key:       strings.TrimSpace(fm.Key),
		Title:     title,
		Body:      memoryBody,
		Tags:      splitTags(normalizeTags(fm.Tags)),
		Source:    strings.TrimSpace(fm.Source),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// MirrorFileName is the deterministic file name for a memory: "<scope>__<sanitized-key>.md"
// when Key is set, otherwise "<id>.md". The same (scope, key) maps to the same file on every
// machine, so git merges a shared fact in place.
func MirrorFileName(m Memory) string {
	if key := strings.TrimSpace(m.Key); key != "" {
		return sanitizeMirrorSegment(m.Scope) + "__" + sanitizeMirrorSegment(key) + ".md"
	}
	return strings.TrimSpace(m.ID) + ".md"
}

// sanitizeMirrorSegment lowercases a path segment and replaces every non-alphanumeric run
// with a single hyphen, trimming leading and trailing hyphens. It is idempotent.
func sanitizeMirrorSegment(segment string) string {
	lower := strings.ToLower(strings.TrimSpace(segment))
	return strings.Trim(nonAlphanumeric.ReplaceAllString(lower, "-"), "-")
}

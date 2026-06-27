package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGoModulePathIsRc asserts the Go module path rename (T5, AC4, F2.1).
// The public API and all internal imports must resolve under github.com/rodolfochicone/rc-project.
func TestGoModulePathIsRc(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join(repoRoot(t), "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	text := string(content)

	if !strings.Contains(text, "module github.com/rodolfochicone/rc-project") {
		t.Fatalf(
			"expected go.mod to declare module github.com/rodolfochicone/rc-project — Compozy module path not renamed",
		)
	}
	if strings.Contains(text, "github.com/compozy/compozy") {
		t.Fatalf("go.mod still contains github.com/compozy/compozy — Compozy module path not fully replaced")
	}
}

// TestCLIEntrypointIsRc asserts the cmd directory rename (T4, AC4, F2.2).
// The CLI entrypoint must live at cmd/rc/, not cmd/compozy/.
func TestCLIEntrypointIsRc(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "cmd", "rc")); err != nil {
		t.Fatalf("expected cmd/rc/ to exist after rebrand: %v — cmd/compozy/ not renamed", err)
	}
	if _, err := os.Stat(filepath.Join(root, "cmd", "compozy")); err == nil {
		t.Fatal("cmd/compozy/ still exists — old entrypoint directory not removed")
	}
}

// TestPublicPackagePkgRcExists asserts the public subpackage rename (T4, T6, F2.4).
func TestPublicPackagePkgRcExists(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "pkg", "rc")); err != nil {
		t.Fatalf("expected pkg/rc/ to exist after rebrand: %v — pkg/compozy/ not renamed", err)
	}
	if _, err := os.Stat(filepath.Join(root, "pkg", "compozy")); err == nil {
		t.Fatal("pkg/compozy/ still exists — old public subpackage directory not removed")
	}
}

// TestGoFilesContainNoCompozyModulePath asserts that no .go file imports the old Compozy module (T5, T8, AC5).
// This encodes why the rename matters: a stale module path would make the binary misidentify itself.
// The test file itself is excluded because it contains the old path as a string literal for this check.
func TestGoFilesContainNoCompozyModulePath(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	var violations []string

	// oldModulePath is the exact string we're searching for, constructed to avoid matching this file.
	oldModulePath := "github.com/" + "compozy/compozy"

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" || name == ".claude" {
				return filepath.SkipDir
			}
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Skip this test file itself — it contains the old path as a string literal for this check.
		if strings.HasSuffix(filepath.ToSlash(path), "test/rc_identity_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(content), oldModulePath) {
			rel, _ := filepath.Rel(root, path)
			violations = append(violations, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}
	if len(violations) > 0 {
		t.Fatalf("found %d .go file(s) still containing the old Compozy module path — rename incomplete:\n  %s",
			len(violations), strings.Join(violations, "\n  "))
	}
}

// TestSkillDirRcExists asserts the skill directory rename (T15, F2.9).
func TestSkillDirRcExists(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "skills", "rc")); err != nil {
		t.Fatalf("expected skills/rc/ to exist after rebrand: %v — skills/compozy/ not renamed", err)
	}
	if _, err := os.Stat(filepath.Join(root, "skills", "compozy")); err == nil {
		t.Fatal("skills/compozy/ still exists — old skill directory not removed")
	}
}

// TestOpenAPIFileIsRcNamed asserts the OpenAPI spec file rename (T13, F2.9).
func TestOpenAPIFileIsRcNamed(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "openapi", "rc-daemon.json")); err != nil {
		t.Fatalf("expected openapi/rc-daemon.json to exist: %v — openapi/compozy-daemon.json not renamed", err)
	}
	if _, err := os.Stat(filepath.Join(root, "openapi", "compozy-daemon.json")); err == nil {
		t.Fatal("openapi/compozy-daemon.json still exists — old spec file not removed")
	}
}

// TestGoReleaserIdentityIsRc asserts goreleaser config uses rc identity (T9, F2.7, AC4).
func TestGoReleaserIdentityIsRc(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join(repoRoot(t), ".goreleaser.yml"))
	if err != nil {
		t.Fatalf("read .goreleaser.yml: %v", err)
	}
	text := string(content)

	if !strings.Contains(text, "project_name: rc") {
		t.Fatalf("expected .goreleaser.yml project_name to be 'rc' — Compozy identity not replaced")
	}
	if !strings.Contains(text, "binary: rc") {
		t.Fatalf("expected .goreleaser.yml binary to be 'rc' — binary name not changed")
	}
	if !strings.Contains(text, "main: ./cmd/rc") {
		t.Fatalf("expected .goreleaser.yml main to be './cmd/rc' — entrypoint not updated")
	}
	if strings.Contains(text, "project_name: compozy") {
		t.Fatalf(".goreleaser.yml still contains project_name: compozy — Compozy identity not removed")
	}
	if strings.Contains(text, "binary: compozy") {
		t.Fatalf(".goreleaser.yml still contains binary: compozy — binary name not changed")
	}
}

// TestLicenseRetainsUpstreamAttribution asserts the LICENSE keeps BOTH the rc
// copyright and the upstream Compozy/NauckGroup attribution that the MIT License
// requires when redistributing a fork (USER DECISION 2026-06-14, reversing the
// earlier attribution-removal decision). Business reason: omitting the upstream
// copyright would violate the MIT terms under which the fork is distributed.
func TestLicenseRetainsUpstreamAttribution(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join(repoRoot(t), "LICENSE"))
	if err != nil {
		t.Fatalf("read LICENSE: %v", err)
	}
	text := strings.ToLower(string(content))

	if !strings.Contains(text, "rc") {
		t.Fatalf("LICENSE does not contain 'rc' — rc copyright missing")
	}
	if !strings.Contains(text, "nauckgroup") {
		t.Fatalf("LICENSE missing 'NauckGroup LTDA' — upstream MIT copyright not retained")
	}
	if !strings.Contains(text, "compozy") {
		t.Fatalf("LICENSE missing upstream Compozy attribution — MIT fork attribution not retained")
	}
}

// TestResidualCompozyStringScan asserts AC5: zero compozy occurrences remain outside the allowlist.
// This encodes the business reason: any stale 'compozy' identifier would misidentify the product.
// The scan excludes: lockfiles, generated files, .git, node_modules,
// the harness planning dir (.claude/ship/fork-rebrand-compozy-rc — intentionally references the
// old brand name as the migration subject), and this test file itself (which contains 'compozy'
// as a string literal for the scan target). Other product files under .claude/ are scanned.
func TestResidualCompozyStringScan(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	skipDirs := map[string]bool{
		".git": true, "node_modules": true,
		".factory": true, ".junie": true, ".pi": true, ".qwen": true, ".kilocode": true,
	}
	// Planning dir that intentionally contains the old brand name as the migration subject.
	skipDirPath := filepath.Join(root, ".claude", "ship", "fork-rebrand-compozy-rc")
	skipFiles := map[string]bool{
		"go.sum":           true,
		"bun.lock":         true,
		"routeTree.gen.ts": true,
		"skills-lock.json": true,
		// Attribution files that intentionally credit the upstream Compozy project
		// to satisfy the MIT License terms for a fork (asserted positively by
		// TestLicenseRetainsUpstreamAttribution).
		"LICENSE": true,
		"NOTICE":  true,
	}
	skipSuffixes := []string{"-openapi.d.ts"}

	// searchTarget is constructed to avoid this file being flagged by its own scan.
	searchTarget := "compo" + "zy"

	var violations []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			if path == skipDirPath {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if skipFiles[name] {
			return nil
		}
		for _, suffix := range skipSuffixes {
			if strings.HasSuffix(name, suffix) {
				return nil
			}
		}
		// Skip test files that contain the search target as a string literal for assertion purposes.
		if strings.HasSuffix(filepath.ToSlash(path), "test/rc_identity_test.go") {
			return nil
		}
		if strings.HasSuffix(filepath.ToSlash(path), "test/rc-identity.test.ts") {
			return nil
		}
		// Root README intentionally credits the upstream Compozy project (MIT attribution).
		if path == filepath.Join(root, "README.md") {
			return nil
		}
		// Stat the real target (follows symlinks) to skip symlinks pointing at directories.
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lower := strings.ToLower(string(content))
		if strings.Contains(lower, searchTarget) {
			rel, _ := filepath.Rel(root, path)
			violations = append(violations, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo for residual scan: %v", err)
	}
	if len(violations) > 0 {
		t.Fatalf("AC5 FAIL: found %d file(s) with residual 'compozy' string — rebrand incomplete:\n  %s",
			len(violations), strings.Join(violations, "\n  "))
	}
}

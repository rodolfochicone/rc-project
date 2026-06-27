package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestRenderRcBannerUsesBlockArt(t *testing.T) {
	t.Parallel()

	banner := renderRcBanner()
	if !strings.Contains(banner, "█") {
		t.Fatalf("expected block-art banner to contain █ glyphs, got %q", banner)
	}
	if lines := strings.Count(banner, "\n") + 1; lines != len(rcBannerRows) {
		t.Fatalf("expected banner to render %d rows, got %d", len(rcBannerRows), lines)
	}
}

func TestRenderStartupTipsListsCoreCommands(t *testing.T) {
	t.Parallel()

	tips := renderStartupTips()
	for _, want := range []string{"Tips for getting started:", "rc setup", "rc tasks run", "rc --help"} {
		if !strings.Contains(tips, want) {
			t.Fatalf("expected tips to mention %q, got %q", want, tips)
		}
	}
}

func TestRenderStatusFooterShowsVersionAndSpansWidth(t *testing.T) {
	t.Parallel()

	const width = 100
	footer := renderStatusFooter(width)
	if !strings.Contains(footer, rcVersionLabel()) {
		t.Fatalf("expected footer to show version label %q, got %q", rcVersionLabel(), footer)
	}

	separator, _, _ := strings.Cut(footer, "\n")
	if got := lipgloss.Width(separator); got != width {
		t.Fatalf("expected footer separator to span %d columns, got %d", width, got)
	}
}

func TestGradientHexAtClampsToBrandRamp(t *testing.T) {
	t.Parallel()

	if got := gradientHexAt(-0.5); got != "#F26B21" {
		t.Fatalf("expected gradient start to clamp to brand orange, got %q", got)
	}
	if got := gradientHexAt(1.5); got != "#FDB813" {
		t.Fatalf("expected gradient end to clamp to accent gold, got %q", got)
	}
	mid := gradientHexAt(0.5)
	if len(mid) != 7 || mid[0] != '#' {
		t.Fatalf("expected mid gradient to be a hex color, got %q", mid)
	}
}

func TestCurrentGitBranchReadsHead(t *testing.T) {
	t.Parallel()

	t.Run("branch ref", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeGitHead(t, dir, "ref: refs/heads/feature/login\n")
		if got := currentGitBranch(dir); got != "feature/login" {
			t.Fatalf("expected branch name from HEAD, got %q", got)
		}
	})

	t.Run("detached head", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		writeGitHead(t, dir, "0123456789abcdef0123456789abcdef01234567\n")
		if got := currentGitBranch(dir); got != "0123456" {
			t.Fatalf("expected short sha for detached HEAD, got %q", got)
		}
	})

	t.Run("no repository", func(t *testing.T) {
		t.Parallel()
		if got := currentGitBranch(t.TempDir()); got != "" {
			t.Fatalf("expected empty branch outside a repo, got %q", got)
		}
	})
}

func writeGitHead(t *testing.T, dir, contents string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte(contents), 0o644); err != nil {
		t.Fatalf("write HEAD: %v", err)
	}
}

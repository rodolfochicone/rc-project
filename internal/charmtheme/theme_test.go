package charmtheme

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestRcPaletteDefaults(t *testing.T) {
	t.Parallel()

	if !sameColor(ColorBgBase, lipgloss.Color("#0C0A09")) {
		t.Fatalf("expected base background #0C0A09, got %v", ColorBgBase)
	}
	if !sameColor(ColorBgSurface, lipgloss.Color("#1C1917")) {
		t.Fatalf("expected surface background #1C1917, got %v", ColorBgSurface)
	}
	if !sameColor(ColorBgOverlay, lipgloss.Color("#292524")) {
		t.Fatalf("expected overlay background #292524, got %v", ColorBgOverlay)
	}
	if !sameColor(ColorBrand, lipgloss.Color("#F26B21")) {
		t.Fatalf("expected brand color #F26B21, got %v", ColorBrand)
	}
	if !sameColor(ColorAccent, lipgloss.Color("#FBB034")) {
		t.Fatalf("expected accent color #FBB034, got %v", ColorAccent)
	}
	if !sameColor(ColorBorder, lipgloss.Color("#44403C")) {
		t.Fatalf("expected border color #44403C, got %v", ColorBorder)
	}
	if got, want := ProgressGradientStart, "#F26B21"; got != want {
		t.Fatalf("expected progress gradient start %s, got %s", want, got)
	}
	if got, want := ProgressGradientEnd, "#FBB034"; got != want {
		t.Fatalf("expected progress gradient end %s, got %s", want, got)
	}
}

// TestRcPaletteUsesOrangeAmberBrand asserts the rc rebrand palette (T11, AC6, F3.1).
// Neutral backgrounds and semantic colors are unchanged; only brand/accent tokens swap from
// rc green to rc orange→amber.
func TestRcPaletteUsesOrangeAmberBrand(t *testing.T) {
	t.Parallel()

	if !sameColor(ColorBrand, lipgloss.Color("#F26B21")) {
		t.Fatalf("expected rc brand color #F26B21 (orange), got %v — rc green not replaced", ColorBrand)
	}
	if !sameColor(ColorAccent, lipgloss.Color("#FBB034")) {
		t.Fatalf("expected rc accent color #FBB034 (amber), got %v — rc green not replaced", ColorAccent)
	}
	if !sameColor(ColorAccentAlt, lipgloss.Color("#FDB813")) {
		t.Fatalf("expected rc accent-alt color #FDB813, got %v — rc green not replaced", ColorAccentAlt)
	}
	if !sameColor(ColorAccentDeep, lipgloss.Color("#F37021")) {
		t.Fatalf("expected rc accent-deep color #F37021, got %v — rc green not replaced", ColorAccentDeep)
	}
	if got, want := ProgressGradientStart, "#F26B21"; got != want {
		t.Fatalf("expected rc progress gradient start %s (orange), got %s — rc green not replaced", want, got)
	}
	if got, want := ProgressGradientEnd, "#FBB034"; got != want {
		t.Fatalf("expected rc progress gradient end %s (amber), got %s — rc green not replaced", want, got)
	}
	forbiddenGreenHexes := []string{"#CAEA28", "#A3E635", "#84CC16", "#65A30D"}
	for _, hex := range forbiddenGreenHexes {
		if sameColor(ColorBrand, lipgloss.Color(hex)) ||
			sameColor(ColorAccent, lipgloss.Color(hex)) ||
			sameColor(ColorAccentAlt, lipgloss.Color(hex)) ||
			sameColor(ColorAccentDeep, lipgloss.Color(hex)) {
			t.Fatalf("rc green brand hex %s still present in rc palette — rebrand incomplete", hex)
		}
	}
}

func TestTechBorderUsesSquareChrome(t *testing.T) {
	t.Parallel()

	if got := TechBorder.TopLeft; got != "┌" {
		t.Fatalf("expected top-left border glyph ┌, got %q", got)
	}
	if got := TechBorder.TopRight; got != "┐" {
		t.Fatalf("expected top-right border glyph ┐, got %q", got)
	}
	if got := TechBorder.BottomLeft; got != "└" {
		t.Fatalf("expected bottom-left border glyph └, got %q", got)
	}
	if got := TechBorder.BottomRight; got != "┘" {
		t.Fatalf("expected bottom-right border glyph ┘, got %q", got)
	}
}

func sameColor(left, right color.Color) bool {
	lr, lg, lb, la := left.RGBA()
	rr, rg, rb, ra := right.RGBA()
	return lr == rr && lg == rg && lb == rb && la == ra
}

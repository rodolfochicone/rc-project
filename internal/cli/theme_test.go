package cli

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/rodolfochicone/rc-project/internal/charmtheme"
	"github.com/spf13/cobra"
)

func TestDarkHuhThemeUsesRcTokens(t *testing.T) {
	t.Parallel()

	styles := darkHuhTheme().Theme(true)

	if !sameColor(styles.Focused.Title.GetForeground(), charmtheme.ColorBrand) {
		t.Fatalf("expected focused title to use brand color, got %v", styles.Focused.Title.GetForeground())
	}
	if !sameColor(styles.Focused.SelectSelector.GetForeground(), charmtheme.ColorAccent) {
		t.Fatalf("expected selector to use accent color, got %v", styles.Focused.SelectSelector.GetForeground())
	}
	if !sameColor(styles.Focused.Base.GetBorderLeftForeground(), charmtheme.ColorBorderFocus) {
		t.Fatalf(
			"expected focused border to use focus border color, got %v",
			styles.Focused.Base.GetBorderLeftForeground(),
		)
	}
	if !sameColor(styles.Focused.FocusedButton.GetBackground(), charmtheme.ColorBrand) {
		t.Fatalf(
			"expected focused button background to use brand color, got %v",
			styles.Focused.FocusedButton.GetBackground(),
		)
	}
	if !sameColor(styles.Focused.TextInput.Prompt.GetForeground(), charmtheme.ColorAccent) {
		t.Fatalf(
			"expected text input prompt to use accent color, got %v",
			styles.Focused.TextInput.Prompt.GetForeground(),
		)
	}
	if !sameColor(styles.Help.ShortKey.GetForeground(), charmtheme.ColorAccentDeep) {
		t.Fatalf("expected help key style to use deep accent color, got %v", styles.Help.ShortKey.GetForeground())
	}
	if sameColor(styles.Focused.SelectSelector.GetForeground(), lipgloss.Color("#F780E2")) {
		t.Fatalf("expected selector to avoid stock Charm fuchsia accent")
	}

	border := styles.Focused.Base.GetBorderStyle()
	if border.Left != charmtheme.TechBorder.Left {
		t.Fatalf(
			"expected focused border style to use technical border glyph %q, got %q",
			charmtheme.TechBorder.Left,
			border.Left,
		)
	}
}

func TestRenderFormIntroUsesTechnicalChrome(t *testing.T) {
	t.Parallel()

	rendered := renderFormIntro()
	if !strings.Contains(rendered, "rc // INTERACTIVE INPUT") {
		t.Fatalf("expected interactive input banner, got %q", rendered)
	}
	if !strings.Contains(rendered, "┌") {
		t.Fatalf("expected technical border chrome, got %q", rendered)
	}
	if strings.ContainsAny(rendered, "╭╮╰╯") {
		t.Fatalf("expected to avoid rounded border glyphs, got %q", rendered)
	}
}

// TestSetupWelcomeHeaderUsesRoundedBorder asserts T10/AC6/PRD-AC8: the welcome header must use
// lipgloss.RoundedBorder() (╭╮╰╯ corners) not the technical square border (┌┐└┘).
// The rounded orange box is the specified rc brand layout.
// TestSetupWelcomeHeaderRendersBannerAndTips asserts the Gemini-style splash:
// the setup header keeps its "rc // SETUP" eyebrow, renders the block-art
// wordmark, and shows the getting-started tips. Dropping any of these would
// silently regress the redesigned startup screen.
func TestSetupWelcomeHeaderRendersBannerAndTips(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := &cobra.Command{Use: "setup"}
	cmd.SetOut(&out)

	printWelcomeHeader(cmd)
	rendered := out.String()
	if !strings.Contains(rendered, "rc // SETUP") {
		t.Fatalf("expected setup header eyebrow, got %q", rendered)
	}
	if !strings.Contains(rendered, "█") {
		t.Fatalf("expected setup header to render the block-art rc banner, got %q", rendered)
	}
	if !strings.Contains(rendered, "Tips for getting started:") {
		t.Fatalf("expected setup header to include getting-started tips, got %q", rendered)
	}
}

// TestSetupWelcomeHeaderContainsGreeting asserts T10/AC6: the header shows a "Welcome back" greeting.
// Removing this line would silently break the specified rich header layout.
func TestSetupWelcomeHeaderContainsGreeting(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := &cobra.Command{Use: "setup"}
	cmd.SetOut(&out)

	printWelcomeHeader(cmd)
	rendered := out.String()
	if !strings.Contains(rendered, "Welcome back") {
		t.Fatalf("expected welcome greeting in header, got %q", rendered)
	}
}

// TestSetupWelcomeHeaderContainsVersionLabel asserts T10/AC6: the header shows the rc version label.
// Removing this label would silently drop the product version from the setup screen.
func TestSetupWelcomeHeaderContainsVersionLabel(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := &cobra.Command{Use: "setup"}
	cmd.SetOut(&out)

	printWelcomeHeader(cmd)
	rendered := out.String()
	if !strings.Contains(rendered, "rc") {
		t.Fatalf("expected rc version label in header, got %q", rendered)
	}
}

// TestSetupWelcomeHeaderRendersGreeting asserts the header shows a personalized
// greeting above the redesigned banner. The greeting is a required element of
// the rc startup layout.
func TestSetupWelcomeHeaderRendersGreeting(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := &cobra.Command{Use: "setup"}
	cmd.SetOut(&out)

	printWelcomeHeader(cmd)
	rendered := out.String()
	if !strings.Contains(rendered, rcWelcomeUser()) {
		t.Fatalf("expected personalized greeting in header, got %q", rendered)
	}
}

func TestCLIChromeStylesUseRcPalette(t *testing.T) {
	t.Parallel()

	styles := newCLIChromeStyles()
	if !sameColor(styles.title.GetForeground(), charmtheme.ColorBrand) {
		t.Fatalf("expected title style to use brand color, got %v", styles.title.GetForeground())
	}
	if !sameColor(styles.box.GetBackground(), charmtheme.ColorBgSurface) {
		t.Fatalf("expected box background to use surface color, got %v", styles.box.GetBackground())
	}
	if !sameColor(styles.box.GetBorderLeftForeground(), charmtheme.ColorAccentDeep) {
		t.Fatalf("expected box border to use deep accent color, got %v", styles.box.GetBorderLeftForeground())
	}
	if !sameColor(styles.warn.GetForeground(), charmtheme.ColorWarning) {
		t.Fatalf("expected warning style to use warning color, got %v", styles.warn.GetForeground())
	}
}

// TestRenderFormIntroUsesRcLabel asserts the interactive input banner rename (T10/T12, AC6, F3.4).
// The brand label must read "rc // INTERACTIVE INPUT" — not the old brand label.
func TestRenderFormIntroUsesRcLabel(t *testing.T) {
	t.Parallel()

	rendered := renderFormIntro()
	if !strings.Contains(rendered, "rc // INTERACTIVE INPUT") {
		t.Fatalf("expected rc // INTERACTIVE INPUT label in interactive banner, got %q", rendered)
	}
}

// TestSetupWelcomeHeaderUsesRcLabel asserts the setup header rename (T10/T12, AC6, F3.4).
// The setup header must show the rc brand label.
func TestSetupWelcomeHeaderUsesRcLabel(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := &cobra.Command{Use: "setup"}
	cmd.SetOut(&out)

	printWelcomeHeader(cmd)
	rendered := out.String()
	if !strings.Contains(rendered, "rc") {
		t.Fatalf("expected rc label in welcome header, got %q", rendered)
	}
}

func sameColor(left, right color.Color) bool {
	lr, lg, lb, la := left.RGBA()
	rr, rg, rb, ra := right.RGBA()
	return lr == rr && lg == rg && lb == rb && la == ra
}

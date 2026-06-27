package cli

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/rodolfochicone/rc-project/internal/charmtheme"
)

// defaultSplashWidth sizes the footer separator when the output is not a real
// terminal (piped output, tests) and no column width is available.
const defaultSplashWidth = 80

// rcPeakRows is the Escale "peak" logo (the rc mountain) drawn in the same ANSI
// Shadow style as the wordmark: a solid block face with the line outline (╗ ═ ╚
// ╝) tracing the right-hand staircase and the base toward the bottom-right.
var rcPeakRows = []string{
	"     ██╗",
	"    ████╗",
	"   ██████╗",
	"  ████████╗",
	" ██████████╗",
	"╚══════════╝",
}

// rcWordRows is the "rc" wordmark in the figlet "ANSI Shadow" style: solid
// blocks (█) traced by a line outline (╗ ╔ ═ ║ ╝ ╚) that echoes each letter's
// silhouette toward the bottom-right.
var rcWordRows = []string{
	" ██╗  ██╗ ██████╗",
	" ██║ ██╔╝ ╚════██╗",
	" █████╔╝   █████╔╝",
	" ██╔═██╗  ██╔═══╝ ",
	" ██║  ██╗ ███████╗",
	" ╚═╝  ╚═╝ ╚══════╝",
}

// rcBannerRows is the full startup wordmark: the peak logo followed by "rc",
// laid out side by side with a fixed gap. Rows are fixed-width so the horizontal
// gradient lines up column by column.
var rcBannerRows = buildRcBannerRows()

// buildRcBannerRows places the peak logo to the left of the "rc" wordmark,
// padding each peak row to a common width so the wordmark shares a single left
// edge across all rows.
func buildRcBannerRows() []string {
	const gap = "   "

	peakWidth := 0
	for _, row := range rcPeakRows {
		if w := len([]rune(row)); w > peakWidth {
			peakWidth = w
		}
	}

	rows := make([]string, len(rcWordRows))
	for i, word := range rcWordRows {
		peak := ""
		if i < len(rcPeakRows) {
			peak = rcPeakRows[i]
		}
		pad := strings.Repeat(" ", peakWidth-len([]rune(peak)))
		rows[i] = peak + pad + gap + word
	}
	return rows
}

// renderRcBanner paints the wordmark with a left-to-right gradient across the rc
// brand orange ramp (deep orange -> amber gold); both the solid blocks and the
// outline strokes are colored so the whole mark reads as one orange wordmark.
func renderRcBanner() string {
	width := 0
	for _, row := range rcBannerRows {
		if w := len([]rune(row)); w > width {
			width = w
		}
	}

	painted := make([]string, len(rcBannerRows))
	for i, row := range rcBannerRows {
		var b strings.Builder
		for x, r := range []rune(row) {
			if r == ' ' {
				b.WriteRune(' ')
				continue
			}
			t := 0.0
			if width > 1 {
				t = float64(x) / float64(width-1)
			}
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(gradientHexAt(t)))
			b.WriteString(style.Render(string(r)))
		}
		painted[i] = b.String()
	}
	return lipgloss.JoinVertical(lipgloss.Left, painted...)
}

// renderStartupTips renders the "Tips for getting started" block shown beneath
// the banner, mirroring the Gemini CLI splash but with rc-specific guidance.
func renderStartupTips() string {
	header := lipgloss.NewStyle().Bold(true).Foreground(charmtheme.ColorBrand)
	num := lipgloss.NewStyle().Bold(true).Foreground(charmtheme.ColorAccent)
	text := lipgloss.NewStyle().Foreground(charmtheme.ColorMuted)
	cmd := lipgloss.NewStyle().Foreground(charmtheme.ColorAccentAlt)

	tip := func(n, before, command, after string) string {
		return num.Render(n+".") + " " + text.Render(before) + cmd.Render(command) + text.Render(after)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header.Render("Tips for getting started:"),
		tip("1", "Run ", "`rc setup`", " to install rc skills for your AI agents."),
		tip("2", "Use ", "`rc tasks run`", " to execute a PRD task workflow."),
		tip("3", "Run ", "`rc --help`", " to see every available command."),
	)
}

// renderStatusFooter renders the bottom status bar (working directory, git
// branch, and rc version) under a thin separator, sized to the given width.
func renderStatusFooter(width int) string {
	if width <= 0 {
		width = defaultSplashWidth
	}

	cwd, homeDir := displayRoots()
	pathStyle := lipgloss.NewStyle().Foreground(charmtheme.ColorAccentAlt)
	branchStyle := lipgloss.NewStyle().Foreground(charmtheme.ColorDim)
	versionStyle := lipgloss.NewStyle().Foreground(charmtheme.ColorMuted)
	sepStyle := lipgloss.NewStyle().Foreground(charmtheme.ColorBorder)

	version := rcVersionLabel()
	branchText := ""
	if branch := currentGitBranch(cwd); branch != "" {
		branchText = " (" + branch + ")"
	}

	// Truncate the working-directory path so the bar never overflows the
	// terminal width (reserve one column for the gap before the version label).
	pathText := shortenPath(cwd, "", homeDir)
	if budget := width - len([]rune(version)) - len([]rune(branchText)) - 1; budget > 0 {
		pathText = truncateToWidth(pathText, budget)
	} else {
		pathText = ""
	}

	leftWidth := len([]rune(pathText)) + len([]rune(branchText))
	gap := max(width-leftWidth-len([]rune(version)), 1)

	bar := pathStyle.Render(pathText) +
		branchStyle.Render(branchText) +
		strings.Repeat(" ", gap) +
		versionStyle.Render(version)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		sepStyle.Render(strings.Repeat("─", width)),
		bar,
	)
}

// renderRcSplash assembles the full Gemini-style startup screen: optional
// eyebrow/greeting lines, the block-art banner, the tips block, and the status
// footer. width sizes the footer separator; <= 0 falls back to a default.
func renderRcSplash(width int, eyebrow, greeting string) string {
	parts := make([]string, 0, 8)
	if eyebrow != "" {
		parts = append(parts, lipgloss.NewStyle().Bold(true).Foreground(charmtheme.ColorBrand).Render(eyebrow))
	}
	if greeting != "" {
		parts = append(parts, lipgloss.NewStyle().Bold(true).Foreground(charmtheme.ColorFgBright).Render(greeting))
	}
	if len(parts) > 0 {
		parts = append(parts, "")
	}
	parts = append(parts,
		renderRcBanner(),
		"",
		renderStartupTips(),
		"",
		renderStatusFooter(width),
	)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

type bannerRGB struct{ r, g, b int }

func (c bannerRGB) hex() string {
	return fmt.Sprintf("#%02X%02X%02X", c.r, c.g, c.b)
}

// gradientHexAt samples the rc brand orange ramp at t in [0,1] and returns the
// interpolated color as a hex string. The stops are the brand orange shades.
func gradientHexAt(t float64) string {
	stops := []bannerRGB{
		{0xF2, 0x6B, 0x21}, // ColorBrand
		{0xF3, 0x70, 0x21}, // ColorAccentDeep
		{0xFB, 0xB0, 0x34}, // ColorAccent
		{0xFD, 0xB8, 0x13}, // ColorAccentAlt
	}
	switch {
	case t <= 0:
		return stops[0].hex()
	case t >= 1:
		return stops[len(stops)-1].hex()
	}

	seg := t * float64(len(stops)-1)
	i := int(seg)
	f := seg - float64(i)
	a, b := stops[i], stops[i+1]
	return bannerRGB{
		r: lerpChannel(a.r, b.r, f),
		g: lerpChannel(a.g, b.g, f),
		b: lerpChannel(a.b, b.b, f),
	}.hex()
}

func lerpChannel(a, b int, t float64) int {
	return a + int(math.Round(float64(b-a)*t))
}

// currentGitBranch walks up from startDir to find a .git directory and returns
// the checked-out branch name, a short detached-HEAD sha, or "" when not in a
// regular git repository (the bottom bar omits the branch in that case).
func currentGitBranch(startDir string) string {
	dir := startDir
	for dir != "" {
		gitPath := filepath.Join(dir, ".git")
		info, err := os.Stat(gitPath)
		if err == nil {
			if !info.IsDir() {
				// .git file (submodule/worktree pointer) — not worth resolving here.
				return ""
			}
			return readGitHeadBranch(filepath.Join(gitPath, "HEAD"))
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
	return ""
}

func readGitHeadBranch(headPath string) string {
	data, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(string(data))
	if branch, ok := strings.CutPrefix(line, "ref: refs/heads/"); ok {
		return branch
	}
	if len(line) >= 7 {
		return line[:7]
	}
	return ""
}

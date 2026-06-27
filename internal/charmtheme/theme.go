package charmtheme

import "charm.land/lipgloss/v2"

const (
	ProgressGradientStart = "#F26B21"
	ProgressGradientEnd   = "#FBB034"
)

var (
	ColorBgBase    = lipgloss.Color("#0C0A09")
	ColorBgSurface = lipgloss.Color("#1C1917")
	ColorBgOverlay = lipgloss.Color("#292524")

	ColorBrand      = lipgloss.Color("#F26B21")
	ColorAccent     = lipgloss.Color("#FBB034")
	ColorAccentAlt  = lipgloss.Color("#FDB813")
	ColorAccentDeep = lipgloss.Color("#F37021")

	ColorSuccess = lipgloss.Color("#10B981")
	ColorError   = lipgloss.Color("#EF4444")
	ColorWarning = lipgloss.Color("#F59E0B")
	ColorInfo    = lipgloss.Color("#3B82F6")

	ColorFgBright = lipgloss.Color("#E7E5E4")
	ColorMuted    = lipgloss.Color("#A8A29E")
	ColorDim      = lipgloss.Color("#78716C")

	ColorBorder      = lipgloss.Color("#44403C")
	ColorBorderFocus = ColorAccent
)

var TechBorder = lipgloss.Border{
	Top:         "─",
	Bottom:      "─",
	Left:        "│",
	Right:       "│",
	TopLeft:     "┌",
	TopRight:    "┐",
	BottomLeft:  "└",
	BottomRight: "┘",
}

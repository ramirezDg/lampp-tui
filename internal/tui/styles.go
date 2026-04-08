package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// ─── palette ─────────────────────────────────────────────────────────────────

var (
	// Fixed accent colors — unchanged regardless of terminal theme.
	colorTitle   = lipgloss.Color("#F27127")
	colorSuccess = lipgloss.Color("#27F271")
	colorError   = lipgloss.Color("#F22727")

	colorHighlightFg = lipgloss.Color("#F27127")

	// Adaptive colors — set in init() from terminal background detection.
	colorText        lipgloss.Color
	colorMuted       lipgloss.Color
	colorHighlightBg lipgloss.Color
	colorBorder      lipgloss.Color
	colorPanelBg     lipgloss.Color
	colorPanelFg     lipgloss.Color
)

func init() {
	dark := termenv.NewOutput(os.Stdout).HasDarkBackground()
	if dark {
		colorText        = lipgloss.Color("#DDDDDD")
		colorMuted       = lipgloss.Color("#888888")
		colorHighlightBg = lipgloss.Color("#3A2A18")
		colorBorder      = lipgloss.Color("#555555")
		colorPanelBg     = lipgloss.Color("#1E1E1E")
		colorPanelFg     = lipgloss.Color("#DDDDDD")
	} else {
		colorText        = lipgloss.Color("#222222")
		colorMuted       = lipgloss.Color("#666666")
		colorHighlightBg = lipgloss.Color("#FFF0E0")
		colorBorder      = lipgloss.Color("#CCCCCC")
		colorPanelBg     = lipgloss.Color("#F8F8F8")
		colorPanelFg     = lipgloss.Color("#222222")
	}
}

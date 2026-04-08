package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ─── title & footer ──────────────────────────────────────────────────────────

var BannerTitleL = lipgloss.NewStyle().
	Foreground(colorTitle).
	Bold(true).
	Render(`
██╗      █████╗ ███╗   ███╗██████╗ ██████╗
██║     ██╔══██╗████╗ ████║██╔══██╗██╔══██╗
██║     ███████║██╔████╔██║██████╔╝██████╔╝
██║     ██╔══██║██║╚██╔╝██║██╔═══╝ ██╔═══╝
███████╗██║  ██║██║ ╚═╝ ██║██║     ██║
╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝     ╚═╝     Linux
`)

func title() string {
	return lipgloss.NewStyle().
		Foreground(colorTitle).
		Bold(true).
		Align(lipgloss.Center).
		Render(BannerTitleL)
}

func footer() string {
	sep := lipgloss.NewStyle().Foreground(colorBorder).Render("│")
	key := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	desc := lipgloss.NewStyle().Foreground(colorMuted)

	hint := func(keys, action string) string {
		return key.Render(keys) + desc.Render(" "+action)
	}

	line1 := strings.Join([]string{
		hint("↑↓←→ / wasd", "Navigate"),
		hint("Enter / Space", "Action"),
		hint("q", "Quit"),
	}, "  "+sep+"  ")

	return lipgloss.NewStyle().Foreground(colorMuted).Render(line1)
}

func RenderTitle(width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, title())
}

func RenderFooter(width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, footer())
}

// ─── action buttons row ──────────────────────────────────────────────────────

func RenderOptions(width int) string {
	btn := lipgloss.NewStyle().
		Foreground(colorMuted).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 2).
		Margin(0, 1)

	actions := []struct{ key, label string }{
		{"e", "Start"},
		{"x", "Stop"},
		{"r", "Restart"},
	}

	keyStyle := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	var parts []string
	for _, a := range actions {
		parts = append(parts, btn.Render(keyStyle.Render(a.key)+" "+a.label))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Center, parts...)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, row)
}

// TextArea renders content in the log/action area style.
func TextArea(content string) string {
	return lipgloss.NewStyle().
		Foreground(colorText).
		Padding(1).
		Align(lipgloss.Left).
		Render(content)
}

// ─── service table ───────────────────────────────────────────────────────────

const columnWidth = 18

func RenderTable(m Model) string {
	col := func() lipgloss.Style {
		return lipgloss.NewStyle().
			Width(columnWidth).
			Align(lipgloss.Center).
			Foreground(colorText)
	}
	highlight := lipgloss.NewStyle().
		Foreground(colorHighlightFg).
		Background(colorHighlightBg).
		Bold(true).
		Width(columnWidth).
		Align(lipgloss.Center)

	headerStyle := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Center).
		Foreground(colorTitle).
		Bold(true).
		Underline(true)

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		headerStyle.MarginBottom(1).Render("Service"),
		headerStyle.MarginBottom(1).Render("PID"),
		headerStyle.MarginBottom(1).Render("Port"),
		headerStyle.MarginBottom(1).Render("Config"),
	)

	rows := make([]string, len(m.choices))
	for i, svc := range m.choices {
		running := m.isRunning(i)

		// Status dot + service name (dot takes 2 chars: dot + space).
		dot := "○"
		statusColor := colorError
		if running {
			dot = "●"
			statusColor = colorSuccess
		}
		label := dot + " " + svc

		svcCell := col().Foreground(statusColor).Render(truncateOrPad(label, columnWidth))
		if m.cursorRow == i && m.cursorCol == 0 {
			svcCell = highlight.Render(truncateOrPad(label, columnWidth))
		}

		pidStr := truncateOrPad(fmt.Sprintf("%d", m.pids[i]), columnWidth)
		portStr := truncateOrPad(m.ports[i], columnWidth)
		cfgStr := truncateOrPad(m.config[i], columnWidth)

		var pidCell, portCell, cfgCell string
		if running {
			pidCell = col().Render(pidStr)
			portCell = col().Render(portStr)
			cfgCell = col().Render(cfgStr)
		} else {
			pidCell = col().Foreground(colorBorder).Render(truncateOrPad("—", columnWidth))
			portCell = col().Foreground(colorBorder).Render(truncateOrPad("—", columnWidth))
			cfgCell = col().Render(truncateOrPad(m.config[i], columnWidth))
		}

		if m.cursorRow == i && m.cursorCol == 1 {
			pidCell = highlight.Render(pidStr)
		}
		if m.cursorRow == i && m.cursorCol == 2 {
			portCell = highlight.Render(portStr)
		}
		if m.cursorRow == i && m.cursorCol == 3 {
			cfgCell = highlight.Render(cfgStr)
		}

		rows[i] = lipgloss.JoinHorizontal(lipgloss.Top, svcCell, pidCell, portCell, cfgCell)
	}

	tableContent := lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)

	// Wrap the table in a subtle border box.
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 2)

	return box.Render(tableContent)
}

// ─── version table ───────────────────────────────────────────────────────────

type versionTableData struct {
	Versions        []string
	SelectedVersion int
}

func RenderVersionTable(d versionTableData) string {
	const numCols = 4
	const colW = 20

	col := func() lipgloss.Style {
		return lipgloss.NewStyle().
			Width(colW).
			Align(lipgloss.Center).
			Foreground(colorText)
	}
	hl := lipgloss.NewStyle().
		Foreground(colorHighlightFg).
		Background(colorHighlightBg).
		Bold(true).
		Width(colW).
		Align(lipgloss.Center)

	n := len(d.Versions)
	numRows := (n + numCols - 1) / numCols
	cells := make([][]string, numRows)
	for i := range cells {
		cells[i] = make([]string, numCols)
		for j := 0; j < numCols; j++ {
			idx := i + j*numRows
			if idx < n {
				ver := d.Versions[idx]
				if ver == "" {
					ver = "-"
				}
				if idx == d.SelectedVersion {
					cells[i][j] = hl.Render(ver)
				} else {
					cells[i][j] = col().Render(ver)
				}
			} else {
				cells[i][j] = col().Render("")
			}
		}
	}

	rowStrings := make([]string, numRows)
	for i := range cells {
		rowStrings[i] = lipgloss.JoinHorizontal(lipgloss.Top, cells[i]...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rowStrings...)
}

// ─── version info panel ──────────────────────────────────────────────────────

func RenderVersionInfoPanel(downloadURL string, selectedButton int) string {
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorTitle).
		Padding(0, 2).
		Background(colorPanelBg).
		Foreground(colorPanelFg)

	label := lipgloss.NewStyle().Foreground(colorPanelFg).Bold(true)
	value := lipgloss.NewStyle().Foreground(colorTitle)

	info := label.Render("Download URL: ") + value.Render(downloadURL) + "\n" +
		label.Render("Destination:  ") + value.Render("./downloads/")

	btn := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Foreground(colorMuted)

	btnActive := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorTitle).
		Foreground(colorPanelBg).
		Background(colorTitle).
		Bold(true)

	installBtn := btn.Render("  Install  ")
	quitBtn := btn.Render("  Quit  ")
	if selectedButton == 0 {
		installBtn = btnActive.Render("  Install  ")
	} else {
		quitBtn = btnActive.Render("  Quit  ")
	}

	buttons := lipgloss.PlaceHorizontal(
		lipgloss.Width(info),
		lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Top, installBtn, quitBtn),
	)

	return panel.Render(info + "\n\n" + buttons)
}

// ─── list ────────────────────────────────────────────────────────────────────

// RenderList renders a vertical menu list. Pass a nil selected map when
// checkboxes are not needed.
func RenderList(options []string, cursor int, selected map[int]struct{}) string {
	activeStyle := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(colorText)
	checkActive := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	checkNormal := lipgloss.NewStyle().Foreground(colorMuted)

	var s string
	for i, choice := range options {
		cur := " "
		if cursor == i {
			cur = "▶"
		}
		check := " "
		if selected != nil {
			if _, ok := selected[i]; ok {
				check = "x"
			}
		}

		bracket := fmt.Sprintf("[%s]", check)

		if cursor == i {
			s += activeStyle.Render(cur) + " " +
				checkActive.Render(fmt.Sprintf("[%s]", check)) + " " +
				activeStyle.Render(choice) + "\n"
		} else {
			s += normalStyle.Render(" ") + " " +
				checkNormal.Render(bracket) + " " +
				normalStyle.Render(choice) + "\n"
		}
	}
	return s
}

// ─── log panel ───────────────────────────────────────────────────────────────

// RenderLogPanel renders the recent-activity log inside a border box.
// innerWidth is the desired inner content width (matches the service table).
// visibleLines controls how many log rows are shown (newest at the bottom).
func RenderLogPanel(logs []string, visibleLines, innerWidth int) string {
	if visibleLines < 1 {
		visibleLines = 1
	}
	if innerWidth < 10 {
		innerWidth = 10
	}

	// ── collect the last N entries ──────────────────────────────────────────
	visible := logs
	if len(visible) > visibleLines {
		visible = visible[len(visible)-visibleLines:]
	}

	tsStyle := lipgloss.NewStyle().Foreground(colorMuted)
	msgStyle := lipgloss.NewStyle().Foreground(colorText)
	emptyStyle := lipgloss.NewStyle().Foreground(colorBorder)

	lines := make([]string, visibleLines)
	// pad empty rows at the top so newest entries sit at the bottom
	offset := visibleLines - len(visible)
	for i := 0; i < offset; i++ {
		lines[i] = emptyStyle.Render("—")
	}
	for i, entry := range visible {
		parts := strings.SplitN(entry, "  ", 2)
		if len(parts) == 2 {
			lines[offset+i] = tsStyle.Render(parts[0]) + "  " + msgStyle.Render(parts[1])
		} else {
			lines[offset+i] = msgStyle.Render(entry)
		}
	}

	content := strings.Join(lines, "\n")

	// ── header + separator ──────────────────────────────────────────────────
	headerStyle := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	sepStyle := lipgloss.NewStyle().Foreground(colorBorder)

	header := headerStyle.Render("Recent Activity")
	sep := sepStyle.Render(strings.Repeat("─", innerWidth))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1).
		Width(innerWidth)

	return box.Render(header + "\n" + sep + "\n" + content)
}

// ─── progress bar ────────────────────────────────────────────────────────────

// RenderProgressBar renders a filled progress bar of the given inner width.
// pct must be in [0.0, 1.0].
func RenderProgressBar(pct float64, innerWidth int) string {
	if innerWidth < 4 {
		innerWidth = 4
	}
	filled := int(pct * float64(innerWidth))
	if filled > innerWidth {
		filled = innerWidth
	}

	filledStyle := lipgloss.NewStyle().Foreground(colorTitle)
	emptyStyle := lipgloss.NewStyle().Foreground(colorBorder)
	pctStyle := lipgloss.NewStyle().Foreground(colorText).Bold(true)

	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", innerWidth-filled))

	return bar + pctStyle.Render(fmt.Sprintf("  %.1f%%", pct*100))
}

// ─── string helpers ──────────────────────────────────────────────────────────

func truncateOrPad(s string, width int) string {
	runes := []rune(s)
	if len(runes) > width {
		return string(runes[:width])
	}
	pad := width - len(runes)
	left := pad / 2
	right := pad - left
	return fmt.Sprintf("%s%s%s", spaces(left), s, spaces(right))
}

func spaces(n int) string {
	return string(make([]rune, n))
}

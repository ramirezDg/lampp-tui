package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ─── widgets ─────────────────────────────────────────────────────────────────

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
	text := "← / ↑ / → / ↓ - Navigate | Enter - Action | q - Quit\n" +
		"a / w / s / d - Navigate | Space - Action\n" +
		"Press 'h' for help"
	return lipgloss.NewStyle().
		Foreground(colorText).
		Align(lipgloss.Left).
		MarginTop(-6).
		Render(text)
}

// TextArea renders content in the log/action area style.
func TextArea(content string) string {
	return lipgloss.NewStyle().
		Foreground(colorText).
		Padding(1).
		Align(lipgloss.Left).
		Render(content)
}

func RenderTitle(width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, title())
}

func RenderFooter(width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Left, footer())
}

func RenderOptions(width int) string {
	optionStyle := lipgloss.NewStyle().Padding(0, 4).Align(lipgloss.Center)
	labels := []string{"[e] Start", "[x] Stop", "[r] Restart"}
	row := ""
	for i, label := range labels {
		row += optionStyle.Render(label)
		if i < len(labels)-1 {
			row += " "
		}
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, row)
}

// ─── service table ───────────────────────────────────────────────────────────

const columnWidth = 17

func RenderTable(m Model) string {
	col := func() lipgloss.Style {
		return lipgloss.NewStyle().Width(columnWidth).Align(lipgloss.Center)
	}
	highlight := lipgloss.NewStyle().
		Foreground(colorHighlightFg).
		Background(colorHighlightBg).
		Bold(true).
		Width(columnWidth).
		Align(lipgloss.Center)

	green := lipgloss.Color("#27F271")
	red := lipgloss.Color("#F22727")

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		col().Bold(true).Underline(true).MarginBottom(1).Render("Service"),
		col().Bold(true).Underline(true).MarginBottom(1).Render("PID"),
		col().Bold(true).Underline(true).MarginBottom(1).Render("Port"),
		col().Bold(true).Underline(true).MarginBottom(1).Render("Config"),
	)

	rows := make([]string, len(m.choices))
	for i, svc := range m.choices {
		running := m.isRunning(i)

		// Service name cell — green when running, red when stopped.
		statusColor := red
		if running {
			statusColor = green
		}
		svcCell := col().Foreground(statusColor).Render(truncateOrPad(svc, columnWidth))
		if m.cursorRow == i && m.cursorCol == 0 {
			svcCell = highlight.Render(truncateOrPad(svc, columnWidth))
		}

		// Data cells — populated only when the service is running.
		pidStr := truncateOrPad(fmt.Sprintf("%d", m.pids[i]), columnWidth)
		portStr := truncateOrPad(m.ports[i], columnWidth)
		cfgStr := truncateOrPad(m.config[i], columnWidth)

		var pidCell, portCell, cfgCell string
		if running {
			pidCell = col().Render(pidStr)
			portCell = col().Render(portStr)
			cfgCell = col().Render(cfgStr)
		} else {
			pidCell = col().Render(truncateOrPad("", columnWidth))
			portCell = col().Render(truncateOrPad("", columnWidth))
			cfgCell = col().Render(truncateOrPad("", columnWidth))
		}

		// Cursor highlight overrides the running check for selected cells.
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

	table := lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)
	w := lipgloss.Width(lipgloss.NewStyle().Render(table))
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, table)
}

// ─── version table ───────────────────────────────────────────────────────────

// versionTableData is the render-only view of the version selector — it holds
// only what the renderer needs, keeping Model details out of render functions.
type versionTableData struct {
	Versions        []string
	SelectedVersion int
}

func RenderVersionTable(d versionTableData) string {
	const numCols = 4
	const colW = 20

	col := func() lipgloss.Style {
		return lipgloss.NewStyle().Width(colW).Align(lipgloss.Center)
	}
	hl := lipgloss.NewStyle().Foreground(colorTitle).Bold(true).Width(colW).Align(lipgloss.Center)

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
		Padding(0, 1).
		Background(lipgloss.Color("#222222")).
		Foreground(lipgloss.Color("#F7F7F7"))

	label := lipgloss.NewStyle().Bold(true)
	value := lipgloss.NewStyle().Foreground(colorTitle)

	info := label.Render("Download URL: ") + value.Render(downloadURL) + "\n" +
		label.Render("Destination: ") + value.Render("./downloads/")

	btn := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#888888"))

	btnActive := btn.Copy().
		Foreground(lipgloss.Color("#fff")).
		Background(colorTitle).
		BorderForeground(colorTitle).
		Bold(true)

	installBtn := btn.Render("Install")
	quitBtn := btn.Render("Quit")
	if selectedButton == 0 {
		installBtn = btnActive.Render("Install")
	} else {
		quitBtn = btnActive.Render("Quit")
	}

	return panel.Render(info + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, installBtn, quitBtn))
}

// ─── list ────────────────────────────────────────────────────────────────────

// RenderList renders a vertical menu list. Pass a nil selected map when
// checkboxes are not needed.
func RenderList(options []string, cursor int, selected map[int]struct{}) string {
	var s string
	for i, choice := range options {
		cur := " "
		if cursor == i {
			cur = ">"
		}
		check := " "
		if selected != nil {
			if _, ok := selected[i]; ok {
				check = "x"
			}
		}
		s += fmt.Sprintf("%s [%s] %s\n", cur, check, choice)
	}
	return s
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

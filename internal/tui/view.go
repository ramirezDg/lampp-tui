package tui

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

func (m Model) View() tea.View {
	w, h := terminalSize()

	// Reserve space for the footer so it never overlaps content.
	footerStr := RenderFooter(w)
	footerH := lipgloss.Height(footerStr)
	mainH := h - footerH - 2 // -2 for the \n\n separator before the footer
	if mainH < 10 {
		mainH = 10
	}

	var content string
	switch {
	case m.downloading:
		content = downloadPane(m, w, mainH)
	case m.ShowNewView:
		content = installPane(m, w, mainH)
	default:
		content = adminPane(m, w, mainH)
	}

	return tea.NewView(content + "\n\n" + footerStr)
}

// paneHeader renders the title with a small top margin and returns the height
// remaining for the pane body below it. The 2-line top pad keeps the title
// from sitting flush against the top edge of the terminal.
func paneHeader(w, h int) (titleStr string, belowH int) {
	titleStr = "\n\n" + RenderTitle(w)
	titleH := lipgloss.Height(titleStr)
	belowH = h - titleH - 2 // -2 for the "\n\n" gap before body
	if belowH < 1 {
		belowH = 1
	}
	return
}

// ─── panes ───────────────────────────────────────────────────────────────────

// adminPane renders the main service management screen.
func adminPane(m Model, w, h int) string {
	titleStr, belowH := paneHeader(w, h)

	rawTable := RenderTable(m)
	tableW := lipgloss.Width(rawTable)
	tableStr := lipgloss.PlaceHorizontal(w, lipgloss.Center, rawTable)
	optStr := RenderOptions(w)

	tableH := lipgloss.Height(tableStr)
	optH := lipgloss.Height(optStr)

	// Fixed area: table + blank line + options + blank line before log panel.
	fixedH := tableH + 1 + optH + 1

	// Give remaining space to the log panel, capped to keep it compact.
	logPanelH := belowH - fixedH
	const minLogPanel = 5
	const maxLogPanel = 8 // border(2) + header(1) + sep(1) + 4 log rows
	if logPanelH < minLogPanel {
		logPanelH = minLogPanel
	}
	if logPanelH > maxLogPanel {
		logPanelH = maxLogPanel
	}

	// Inner content rows = total panel height minus border(2) + header(1) + sep(1).
	logVisible := logPanelH - 4
	if logVisible < 1 {
		logVisible = 1
	}

	// Log panel inner width matches the service table box.
	// Table box outer width = tableW; inner = tableW - border(2) - padding(2*2) = tableW - 6.
	logInnerW := tableW - 6
	if logInnerW < 20 {
		logInnerW = 20
	}

	logStr := lipgloss.PlaceHorizontal(w, lipgloss.Center,
		RenderLogPanel(m.logs, logVisible, logInnerW))

	body := lipgloss.JoinVertical(lipgloss.Left,
		tableStr,
		"",
		optStr,
		"",
		logStr,
	)

	// Top-align so the content starts right below the title gap and the log
	// panel naturally fills down from the service table.
	below := lipgloss.Place(w, belowH, lipgloss.Center, lipgloss.Top, body)
	return titleStr + "\n\n" + below
}

// installPane renders either the welcome/install-options screen or the version
// picker, depending on m.installing.
func installPane(m Model, w, h int) string {
	titleStr, belowH := paneHeader(w, h)

	var content string
	if m.installing {
		content = versionPickerContent(m)
	} else {
		content = welcomeContent(m)
	}

	below := lipgloss.Place(w, belowH, lipgloss.Center, lipgloss.Center, content)
	return titleStr + "\n\n" + below
}

// downloadPane renders the active download progress screen.
func downloadPane(m Model, w, h int) string {
	titleStr, belowH := paneHeader(w, h)

	barWidth := 50
	bar := RenderProgressBar(m.downloadProgress, barWidth)

	label := lipgloss.NewStyle().Foreground(colorText).Bold(true).
		Render(fmt.Sprintf("Downloading XAMPP %s...", m.downloadVersion))

	var statusLine string
	switch {
	case m.downloadError != "":
		statusLine = lipgloss.NewStyle().Foreground(colorError).
			Render("Error: " + m.downloadError)
	case m.downloadProgress >= 1.0:
		statusLine = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).
			Render("Download complete!")
	default:
		statusLine = lipgloss.NewStyle().Foreground(colorMuted).Render("Please wait…")
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		label,
		"",
		bar,
		"",
		statusLine,
	)

	below := lipgloss.Place(w, belowH, lipgloss.Center, lipgloss.Center,
		lipgloss.PlaceHorizontal(w, lipgloss.Center, content))
	return titleStr + "\n\n" + below
}

// ─── screen content builders ─────────────────────────────────────────────────

func versionPickerContent(m Model) string {
	numCols := 4
	n := len(m.xamppVersions)
	numRows := (n + numCols - 1) / numCols
	selectedIdx := m.cursorVersionRow + m.cursorVersionCol*numRows

	var names []string
	for _, v := range m.xamppVersions {
		names = append(names, v.Name)
	}

	heading := lipgloss.NewStyle().Foreground(colorMuted).Bold(true).
		Render("Select the XAMPP version:")

	table := RenderVersionTable(versionTableData{
		Versions:        names,
		SelectedVersion: selectedIdx,
	})

	parts := []string{heading, "", table}

	if m.showVersionInfoPanel {
		url := "(select a valid version)"
		if n > 0 && selectedIdx < n {
			url = m.xamppVersions[selectedIdx].DownloadURL
		}
		parts = append(parts, "", RenderVersionInfoPanel(url, m.cursorVersionButton))
	}

	return lipgloss.JoinVertical(lipgloss.Center, parts...)
}

func welcomeContent(m Model) string {
	selected := map[int]struct{}{m.cursorInstall: {}}
	optionsList := RenderList(m.optionsInstallation, m.cursorInstall, selected)

	text := lipgloss.NewStyle().Foreground(colorText).Bold(true).Render("Welcome to XAMPP-TUI.") + "\n" +
		lipgloss.NewStyle().Foreground(colorMuted).Render("XAMPP is not installed on your system.") + "\n\n" +
		lipgloss.NewStyle().Foreground(colorMuted).Render("Options:") + "\n" +
		optionsList

	return lipgloss.NewStyle().Align(lipgloss.Left).Render(text)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// terminalSize returns the current terminal dimensions, falling back to 80×24
// when the size cannot be determined (e.g. in tests or pipes).
func terminalSize() (width, height int) {
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		return w, h
	}
	return 80, 24
}

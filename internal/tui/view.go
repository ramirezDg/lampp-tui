package tui

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

func (m Model) View() tea.View {
	w, h := terminalSize()

	var content string
	if m.ShowNewView {
		content = installPane(m, w, h)
	} else {
		content = adminPane(m, w, h)
	}

	return tea.NewView(content + "\n\n" + RenderFooter(w))
}

// adminPane renders the main service management screen.
func adminPane(m Model, w, h int) string {
	titleStr := RenderTitle(w)
	body := titleStr + "\n" +
		lipgloss.PlaceHorizontal(w, lipgloss.Center, RenderTable(m)) + "\n" +
		RenderOptions(w) + "\n" +
		lipgloss.PlaceHorizontal(w, lipgloss.Center,
			lipgloss.NewStyle().Align(lipgloss.Left).Render(TextArea("Logs De Acciones")))

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, body)
}

// installPane renders either the welcome/install-options screen or the version
// picker table, depending on m.installing.
func installPane(m Model, w, h int) string {
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	titleStr := RenderTitle(w)

	var body string
	if m.installing {
		body = titleStr + "\n\n" +
			lipgloss.PlaceHorizontal(w, lipgloss.Center,
				gray.Bold(true).Render("Select the XAMPP version:")) + "\n\n"

		numCols := 4
		n := len(m.xamppVersions)
		numRows := (n + numCols - 1) / numCols
		selectedIdx := m.cursorVersionRow + m.cursorVersionCol*numRows

		var names []string
		for _, v := range m.xamppVersions {
			names = append(names, v.Name)
		}
		body += lipgloss.PlaceHorizontal(w, lipgloss.Center,
			RenderVersionTable(versionTableData{
				Versions:        names,
				SelectedVersion: selectedIdx,
			})) + "\n"

		if m.showVersionInfoPanel {
			url := "(select a valid version)"
			if n > 0 && selectedIdx < n {
				url = m.xamppVersions[selectedIdx].DownloadURL
			}
			body += lipgloss.PlaceHorizontal(w, lipgloss.Center,
				RenderVersionInfoPanel(url, m.cursorVersionButton)) + "\n"
		}
	} else {
		optionsList := RenderList(m.optionsInstallation, m.cursorInstall, nil)
		welcome := lipgloss.NewStyle().Align(lipgloss.Left).Render(
			gray.Render("Welcome to XAMPP-TUI.") + "\n" +
				gray.Render("XAMPP is not installed on your system.") + "\n\n" +
				gray.Render("Options:") + "\n" +
				gray.Render(optionsList) + "\n",
		)
		body = titleStr + "\n\n" +
			lipgloss.PlaceHorizontal(w, lipgloss.Center, welcome) + "\n"
	}

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, body)
}

// terminalSize returns the current terminal dimensions, falling back to 80×24
// when the size cannot be determined (e.g. in tests or pipes).
func terminalSize() (width, height int) {
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		return w, h
	}
	return 80, 24
}

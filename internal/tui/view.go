package tui

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

func (m Model) View() tea.View {
	terminalWidth, terminalHeight := 80, 24
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		terminalWidth, terminalHeight = w, h
	}

	var mainContent string
	if m.ShowNewView {
		mainContent = InstallPane(m, terminalWidth, terminalHeight)
	} else {
		mainContent = AdminPane(m, terminalWidth, terminalHeight)
	}

	content := renderTitle(terminalWidth) + "\n" + mainContent + "\n\n" + renderFooter(terminalWidth)
	return tea.NewView(content)
}

func AdminPane(m Model, terminalWidth, terminalHeight int) string {
	content := lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, RenderTable(m)) + "\n"
	content += renderOptions(terminalWidth) + "\n"
	content += lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, TextArea("Logs De Acciones"))

	return lipgloss.Place(
		terminalWidth, terminalHeight,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func InstallPane(m Model, terminalWidth, terminalHeight int) string {
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	content := ""

	if m.showVersionList {
		content += "\n\n" + lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, gray.Bold(true).Render("Select the XAMPP version:")) + "\n\n"
		versionTable := RenderVersionTable(VersionTableModel{
			Versiones:       m.versiones,
			SelectedVersion: m.selectedVersion,
		})
		content += lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, versionTable) + "\n"
		if m.installing {
			//var url string
			if m.selectedVersion >= 0 && m.selectedVersion < len(m.versiones) {
				//url = ObtenerURLDeVersion(m.versiones[m.selectedVersion])
			}
			//content += "\n" + ShowDownloadXAMPP(url).(string) + "\n"
		}
	} else {
		installed := Validate()
		if installed.Installed {
			leftStyle := lipgloss.NewStyle().Align(lipgloss.Left)
			welcomeBlock := leftStyle.Render(
				gray.Render("Welcome to XAMPP-TUI.") + "\n" +
					gray.Render("XAMPP is not installed on your system.") + "\n\n" +
					gray.Render("Options:") + "\n" +
					gray.Render("  [I]nstall XAMPP") + "\n" +
					gray.Render("  [Q]uit/exit") + "\n",
			)
			content += "\n\n" + lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, welcomeBlock) + "\n"
		}
	}

	return lipgloss.Place(
		terminalWidth, terminalHeight,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func renderTitle(width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, Title())
}

func renderFooter(width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Left, Footer())
}

func renderOptions(width int) string {
	optionLabels := []string{"[e] Start", "[x] Stop", "[r] Restart"}
	optionStyle := lipgloss.NewStyle().Padding(0, 4).Align(lipgloss.Center)
	optionRow := ""
	for i, label := range optionLabels {
		option := optionStyle.Render(label)
		if i < len(optionLabels)-1 {
			optionRow += option + " "
		} else {
			optionRow += option
		}
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, optionRow)
}

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

	content := RenderTitle(terminalWidth) + "\n" + mainContent + "\n\n" + RenderFooter(terminalWidth)
	return tea.NewView(content)
}

func AdminPane(m Model, terminalWidth, terminalHeight int) string {
	title := RenderTitle(terminalWidth)
	content := title + "\n" + lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, RenderTable(m)) + "\n"
	content += RenderOptions(terminalWidth) + "\n"
	leftStyle := lipgloss.NewStyle().Align(lipgloss.Left)
	content += lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, leftStyle.Render(TextArea("Logs De Acciones"))) + "\n"

	return lipgloss.Place(
		terminalWidth, terminalHeight,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func InstallPane(m Model, terminalWidth, terminalHeight int) string {
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	title := RenderTitle(terminalWidth)
	content := title + "\n"

	if m.installing {
		content += "\n\n" + lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, gray.Bold(true).Render("Select the XAMPP version:")) + "\n\n"
		numCols := 4
		n := len(m.xamppVersions)
		numRows := (n + numCols - 1) / numCols
		selectedIdx := m.cursorVersionRow + m.cursorVersionCol*numRows
		var nombres []string
		for _, v := range m.xamppVersions {
			nombres = append(nombres, v.Name)
		}
		versionTable := RenderVersionTable(VersionTableModel{
			Versiones:       nombres,
			SelectedVersion: selectedIdx,
		})
		content += lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, versionTable) + "\n"

		// Panel de información de descarga solo si está activo
		if m.showVersionInfoPanel {
			var downloadURL string
			if n > 0 && selectedIdx < n {
				downloadURL = m.xamppVersions[selectedIdx].DownloadURL
			} else {
				downloadURL = "(select a valid version)"
			}
			versionInfoPanel := RenderVersionInfoPanel(downloadURL, m.cursorVersionButton)
			content += lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, versionInfoPanel) + "\n"
		}
	} else {
		leftStyle := lipgloss.NewStyle().Align(lipgloss.Left)
		optionsList := RenderList(m.optionsInstallation, m.cursorInstall, nil)
		welcomeBlock := leftStyle.Render(
			gray.Render("Welcome to XAMPP-TUI.") + "\n" +
				gray.Render("XAMPP is not installed on your system.") + "\n\n" +
				gray.Render("Options:") + "\n" +
				gray.Render(optionsList) + "\n",
		)
		content += "\n\n" + lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, welcomeBlock) + "\n"
	}

	return lipgloss.Place(
		terminalWidth, terminalHeight,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

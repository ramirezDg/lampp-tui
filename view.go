package main

import (
	"io"
	"os"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type progressWriter struct {
	total      int
	downloaded int
	file       *os.File
	reader     io.Reader
	onProgress func(float64)
}

type model struct {
	osName          string
	installed       bool
	status          []string
	showVersionList bool
	versiones       []string
	selectedVersion int
	installing      bool
	pw              *progressWriter
	progress        progress.Model
}

type ValidationResult struct {
	OSName    string
	Installed bool
}

func isLAMPInstalled() bool {
	services := []string{
		"/opt/lampp/apache2",
		"/opt/lampp/mysql",
		"/opt/lampp/sbin/proftpd",
	}
	for _, path := range services {
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}

func Validate() ValidationResult {
	installed := isLAMPInstalled()
	return ValidationResult{
		OSName:    "linux",
		Installed: installed,
	}
}

func initialModelValidation() model {
	return model{
		status: []string{"Installed", "Not Installed"},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "Q":
			return m, tea.Quit
		case "i", "I":
			m.showVersionList = true
			m.installing = false
			versiones, err := ObtenerVersiones()
			if err != nil {
				m.versiones = []string{"Error al obtener versiones"}
			} else {
				m.versiones = versiones
			}
			m.selectedVersion = 0
			return m, nil
		case "up", "k":
			if m.showVersionList {
				numCols := 4
				numRows := (len(m.versiones) + numCols - 1) / numCols
				row := m.selectedVersion % numRows
				if row > 0 {
					m.selectedVersion--
				}
			}
		case "down", "j":
			if m.showVersionList {
				numCols := 4
				numRows := (len(m.versiones) + numCols - 1) / numCols
				row := m.selectedVersion % numRows
				if row < numRows-1 && m.selectedVersion+1 < len(m.versiones) {
					m.selectedVersion++
				}
			}
		case "left", "a":
			if m.showVersionList {
				numCols := 4
				numRows := (len(m.versiones) + numCols - 1) / numCols
				col := m.selectedVersion / numRows
				if col > 0 {
					newIdx := m.selectedVersion - numRows
					if newIdx >= 0 {
						m.selectedVersion = newIdx
					}
				}
			}
		case "right", "d":
			if m.showVersionList {
				numCols := 4
				numRows := (len(m.versiones) + numCols - 1) / numCols
				col := m.selectedVersion / numRows
				if col < numCols-1 {
					newIdx := m.selectedVersion + numRows
					if newIdx < len(m.versiones) {
						m.selectedVersion = newIdx
					}
				}
			}
		case "enter":
			if m.showVersionList && len(m.versiones) > 0 {
				m.installing = true
				var url string
				if m.selectedVersion >= 0 && m.selectedVersion < len(m.versiones) {
					url = ObtenerURLDeVersion(m.versiones[m.selectedVersion])
				}
			}
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	terminalWidth, terminalHeight := 80, 24
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		terminalWidth, terminalHeight = w, h
	}

	title := lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, Title())
	content := title

	if m.showVersionList {
		content += "\n\n" + lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, gray.Bold(true).Render("Select the XAMPP version:")) + "\n\n"
		versionTable := RenderVersionTable(VersionTableModel{
			Versiones:       m.versiones,
			SelectedVersion: m.selectedVersion,
		})
		content += lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, versionTable) + "\n"
		if m.installing {
			var url string
			if m.selectedVersion >= 0 && m.selectedVersion < len(m.versiones) {
				url = ObtenerURLDeVersion(m.versiones[m.selectedVersion])
			}
			content += "\n" + ShowDownloadXAMPP(url) + "\n"
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

	footerBlock := lipgloss.NewStyle().Align(lipgloss.Left).Render(Footer())
	footerHeight := lipgloss.Height(footerBlock)
	contentHeight := terminalHeight - footerHeight
	mainContent := lipgloss.Place(
		terminalWidth, contentHeight,
		lipgloss.Center, lipgloss.Center,
		content,
	)
	final := lipgloss.JoinVertical(lipgloss.Left, mainContent, lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Left, footerBlock))
	return tea.NewView(final)
}

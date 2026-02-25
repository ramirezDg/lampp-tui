package main

import (
	"os"
	"runtime"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type model struct {
	osName           string
	installed        bool
	status           []string
	showVersionList  bool
	versiones        []string
	selectedVersion  int
	installing       bool
}

var osName = runtime.GOOS

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

func isXAMPPInstalled() bool {
	var xamppPath string
	if runtime.GOOS == "windows" {
		xamppPath = "C:\\xampp\\xampp-control.exe"
	} else {
		xamppPath = "/opt/lampp/xampp"
	}
	_, err := os.Stat(xamppPath)
	return err == nil
}

func Validate() ValidationResult {
	var installed bool

	switch osName {
	case "linux":
		installed = isLAMPInstalled()
	case "windows":
		installed = isXAMPPInstalled()
	default:
		installed = false
	}

	return ValidationResult{
		OSName:    osName,
		Installed: installed,
	}
}

func initialModelValidation() model {
	return model{
		status:  []string{"Installed", "Not Installed"},
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
			m.versiones = getXAMPPVersions()
			m.selectedVersion = 0
			return m, nil
		case "up", "k":
			if m.showVersionList && m.selectedVersion > 0 {
				m.selectedVersion--
			}
		case "down", "j":
			if m.showVersionList && m.selectedVersion < len(m.versiones)-1 {
				m.selectedVersion++
			}
		case "enter":
			if m.showVersionList && len(m.versiones) > 0 {
				m.installing = true
				go func(version string) {
					InstalarXAMPPConVersion(version)
				}(m.versiones[m.selectedVersion])
			}
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	terminalWidth, terminalHeight := 80, 24
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		terminalWidth, terminalHeight = w, h
	}

	title := lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, Title())
	content := title

	if m.showVersionList {
		content += "\n\n" + lipgloss.NewStyle().Bold(true).Render("Select the XAMPP version:") + "\n"
		for i, v := range m.versiones {
			style := lipgloss.NewStyle()
			if i == m.selectedVersion {
				style = style.Foreground(lipgloss.Color("#F27127")).Bold(true)
			}
			content += style.Render(v) + "\n"
		}
		if m.installing {
			content += "\nInstalando XAMPP...\n"
		}
	} else {
		installed := Validate()
		gray := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		if !installed.Installed {
			content += "\n\n" + gray.Render("Welcome to XAMPP-TUI.") + "\n"
			content += gray.Render("XAMPP is not installed on your system.") + "\n\n"
			content += gray.Render("Options:") + "\n"
			content += gray.Render("  [I]nstall XAMPP") + "\n"
			content += gray.Render("  [Q]uit/exit") + "\n"
		} else {
			content += "\n\n" + gray.Render("XAMPP is already installed.") + "\n"
			content += gray.Render("  [Q]uit/exit") + "\n"
		}
	}

	centered := lipgloss.Place(
		terminalWidth, terminalHeight,
		lipgloss.Center, lipgloss.Center,
		content,
	)

	return tea.NewView(centered)
}




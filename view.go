package main

import (
	"os"
	"runtime"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type model struct {
	osName    string
	installed bool
	status	[]string
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
		case "ctrl+c", "q":
			return m, tea.Quit
		case "e", "E":
			// Acción para iniciar servicio
		case "x", "X":
			// Acción para detener servicio
		case "r", "R":
			// Acción para reiniciar servicio
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

	installed := Validate()
	if !installed.Installed {
	} else {
		gray := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		content += "\n\n" + gray.Render("Welcome to XAMPP-TUI.") + "\n"
		content += gray.Render("XAMPP is not installed on your system.") + "\n\n"
		if osName == "linux" {
			content += gray.Render("Options:") + "\n"
			content += gray.Render("  [I]nstall XAMPP") + "\n"
			content += gray.Render("  [Q]uit/exit") + "\n"
		} else {
			content += gray.Render("Options:") + "\n"
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




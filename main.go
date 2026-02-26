package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var p *tea.Program

type Model struct {
	choices   []string
	pids      []int
	ports     []string
	config    []string
	cursorRow int      // fila seleccionada
	cursorCol int      // columna seleccionada: 0=choices, 1=port, 2=config
	status    []string // "running" or "stopped"
}

func initialModel() Model {
	return Model{
		choices: []string{"Apache", "MySql", "FTP"},
		pids:    []int{0, 0, 0},
		ports:   []string{"", "", ""},
		config:  []string{"httpd.conf", "my.ini", "vsftpd.conf"},
		status:  []string{"stopped", "stopped", "stopped"},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "a", "A", "←":
			if m.cursorCol > 0 {
				m.cursorCol--
			}
		case "right", "d", "D", "→":
			if m.cursorCol < 3 {
				m.cursorCol++
			}
		case "up", "w", "W", "↑":
			if m.cursorRow > 0 {
				m.cursorRow--
			}
		case "down", "s", "S", "↓":
			if m.cursorRow < len(m.choices)-1 {
				m.cursorRow++
			}
		case "enter", "space":
			// Solo permitir acción en columna 0 (choices), 1 (port), 2 (config)
			if m.cursorCol == 0 {
				service := m.choices[m.cursorRow]
				// Mapear a los nombres esperados por ControlXAMPPService
				var serviceKey string
				switch service {
				case "Apache":
					serviceKey = "apache"
				case "MySql":
					serviceKey = "mysql"
				case "FTP":
					serviceKey = "ftp"
				default:
					serviceKey = service
				}
				if m.status[m.cursorRow] == "running" {
					ControlXAMPPService(serviceKey, "stop")
					m.status[m.cursorRow] = "stopped"
				} else {
					ControlXAMPPService(serviceKey, "start")
					m.status[m.cursorRow] = "running"
				}
			} else if m.cursorCol == 1 {
				// Acción para port (ejemplo: editar puerto)
				// m.ports[m.cursorRow] = "nuevo puerto"
			} else if m.cursorCol == 2 {
				// Acción para config (ejemplo: abrir config)
				// m.config[m.cursorRow] = "nuevo config"
			}
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

func (m Model) View() tea.View {
	terminalWidth, terminalHeight := 80, 24
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		terminalWidth, terminalHeight = w, h
	}

	title := lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, Title())
	table := lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, RenderTable(m))

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
	optionsCentered := lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, optionRow)

	logs := lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, TextArea("Logs De Acciones"))

	content := title + "\n" + table + "\n" + optionsCentered + "\n" + logs

	centered := lipgloss.Place(
		terminalWidth, terminalHeight,
		lipgloss.Center, lipgloss.Center,
		content,
	)

	return tea.NewView(centered + "\n\n" + Footer())
}

func main() {
	result := Validate()
	if !result.Installed {
		p := tea.NewProgram(initialModel())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	} else {
		p := tea.NewProgram(initialModelValidation())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}
	}
}

package main

import (
	"fmt"
	"os"

	"golang.org/x/term"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	choices   []string
	pids      []int
	ports     []string
	config    []string
	cursorRow int      // fila seleccionada
	cursorCol int      // columna seleccionada: 0=choices, 1=port, 2=config
	status    []string // "running" or "stopped"
}

func initialModel() model {
	return model{
		choices: []string{"Apache", "MySql", "FTP"},
		pids:    []int{0, 0, 0},
		ports:   []string{"", "", ""},
		config:  []string{"httpd.conf", "my.ini", "vsftpd.conf"},
		status:  []string{"stopped", "stopped", "stopped"},
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
				// Ejemplo: cambiar status
				if m.status[m.cursorRow] == "running" {
					m.status[m.cursorRow] = "stopped"
				} else {
					m.status[m.cursorRow] = "running"
				}
			} else if m.cursorCol == 1 {
				// Acción para port (ejemplo: editar puerto)
				// m.ports[m.cursorRow] = "nuevo puerto"
			} else if m.cursorCol == 2 {
				// Acción para config (ejemplo: abrir config)
				// m.config[m.cursorRow] = "nuevo config"
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

	centered := lipgloss.Place(
		terminalWidth, terminalHeight,
		lipgloss.Center, lipgloss.Center,
		Title()+"\n"+RenderTable(m),
	)

	return tea.NewView(centered + "\n\n" + TextArea("Logs De Acciones") + "\n\n" + Footer())
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

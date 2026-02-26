package tui

import (
	"xampp-tui/internal/services"

	tea "charm.land/bubbletea/v2"
)

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
					services.ControlXAMPPService(serviceKey, "stop")
					m.status[m.cursorRow] = "stopped"
				} else {
					services.ControlXAMPPService(serviceKey, "start")
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

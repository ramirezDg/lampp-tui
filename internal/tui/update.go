package tui

import (
	"xampp-tui/internal/services"

	tea "charm.land/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Si estamos en la vista de instalación, navegamos optionsInstallation
		if m.ShowNewView {
			// Si está instalando, navegamos el menú de versiones
			if m.installing {
				num := len(m.versiones)
				switch msg.String() {
				case "up", "w", "W", "↑":
					if m.cursorVersion > 0 {
						m.cursorVersion--
					}
				case "down", "s", "S", "↓":
					if m.cursorVersion < num-1 {
						m.cursorVersion++
					}
				case "q", "esc":
					m.installing = false // Salir del menú de versiones
				case "enter", "space":
					m.selectedVersion = m.cursorVersion
					// Aquí podrías iniciar la descarga/instalación
				}
				return m, nil
			}
			switch msg.String() {
			case "up", "w", "W", "↑":
				if m.cursorInstall > 0 {
					m.cursorInstall--
				}
			case "down", "s", "S", "↓":
				if m.cursorInstall < len(m.optionsInstallation)-1 {
					m.cursorInstall++
				}
			case "left", "a", "A", "←":
				if m.cursorCol > 0 {
					m.cursorCol--
				}
			case "right", "d", "D", "→":
				if m.cursorCol < 3 {
					m.cursorCol++
				}
			case "enter", "space":
				// Acción según opción seleccionada
				switch m.cursorInstall {
				case 0: // "Install XAMPP"
					if len(m.versiones) == 0 {
						versiones, err := services.ObtenerVersiones()
						if err == nil {
							m.versiones = versiones
						} else {
							m.versiones = []string{"Error obteniendo versiones"}
						}
						m.cursorVersion = 0
						m.selectedVersion = 0
					}
					m.installing = true
				case 1: // "Quit/Exit"
					return m, tea.Quit
				}
			}
			return m, nil
		}
		// Si no, lógica normal de admin
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
			// Si está instalando, navegamos el menú de versiones
			if m.installing {
				// Obtener el número de versiones (debería estar en m.versiones)
				num := len(m.versiones)
				switch msg.String() {
				case "up", "w", "W", "↑":
					if m.cursorVersion > 0 {
						m.cursorVersion--
					}
				case "down", "s", "S", "↓":
					if m.cursorVersion < num-1 {
						m.cursorVersion++
					}
				case "q", "esc":
					m.installing = false // Salir del menú de versiones
				case "enter", "space":
					// Seleccionar versión
					m.selectedVersion = m.cursorVersion
					// Aquí podrías iniciar la descarga/instalación
				}
				return m, nil
			}
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

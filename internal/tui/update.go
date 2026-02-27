package tui

import (
	"xampp-tui/internal/services"

	tea "charm.land/bubbletea/v2"
)

func handleNavigation(key string, row, col, maxRow, maxCol int) (newRow, newCol int, quit bool) {
	newRow, newCol = row, col
	switch key {
	case "up", "w", "W", "↑":
		if newRow > 0 {
			newRow--
		}
	case "down", "s", "S", "↓":
		if newRow < maxRow-1 {
			newRow++
		}
	case "left", "a", "A", "←":
		if newCol > 0 {
			newCol--
		}
	case "right", "d", "D", "→":
		if newCol < maxCol-1 {
			newCol++
		}
	case "ctrl+c", "q":
		quit = true
	}
	return
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		key := msg.String()
		// --- Selección de versión (tabla bidimensional) ---
		if m.ShowNewView && m.installing {
			// Si el panel de info está activo, navegar entre botones
			if m.showVersionInfoPanel {
				// Solo dos botones: 0=Install, 1=Quit
				switch key {
				case "left", "a", "A", "←":
					if m.cursorVersionButton > 0 {
						m.cursorVersionButton--
					}
				case "right", "d", "D", "→":
					if m.cursorVersionButton < 1 {
						m.cursorVersionButton++
					}
				case "q", "esc":
					m.showVersionInfoPanel = false
				case "enter", "space":
					// Acción según botón
					if m.cursorVersionButton == 0 {
						services.InstalarXAMPP(m.xamppVersions[m.selectedVersion].Name)
						m.showVersionInfoPanel = false
					} else {
						m.showVersionInfoPanel = false
					}
				}
				return m, nil
			}
			// Parámetros de la tabla de versiones
			numCols := 4
			n := len(m.xamppVersions)
			numRows := (n + numCols - 1) / numCols
			row, col, quit := handleNavigation(key, m.cursorVersionRow, m.cursorVersionCol, numRows, numCols)
			// Limitar el cursor a celdas válidas
			idx := row + col*numRows
			if idx >= n {
				// Si la celda está fuera de rango, no mover el cursor
				row, col = m.cursorVersionRow, m.cursorVersionCol
			}
			m.cursorVersionRow, m.cursorVersionCol = row, col
			m.selectedVersion = idx
			if key == "q" || key == "esc" {
				m.installing = false
			} else if key == "enter" || key == "space" {
				// Activar panel de info
				m.showVersionInfoPanel = true
				m.cursorVersionButton = 0
			}
			if quit {
				return m, tea.Quit
			}
			return m, nil
		}
		// --- Menú de instalación ---
		if m.ShowNewView {
			row, col, quit := handleNavigation(key, m.cursorInstall, m.cursorCol, len(m.optionsInstallation), 4)
			m.cursorInstall, m.cursorCol = row, col
			if quit {
				return m, tea.Quit
			}
			if key == "enter" || key == "space" {
				switch m.cursorInstall {
				case 0: // "Install XAMPP"
					if len(m.xamppVersions) == 0 {
						versiones, err := services.ObtenerVersiones()
						if err == nil {
							m.xamppVersions = versiones
						} else {
							m.xamppVersions = []services.XAMPPVersion{{Name: "Error obteniendo versiones", DownloadURL: ""}}
						}
						m.cursorVersionRow = 0
						m.cursorVersionCol = 0
						m.selectedVersion = 0
					}
					m.installing = true
				case 1: // "Quit/Exit"
					return m, tea.Quit
				}
			}
			return m, nil
		}
		// --- Menú principal (servicios) ---
		row, col, quit := handleNavigation(key, m.cursorRow, m.cursorCol, len(m.choices), 4)
		m.cursorRow, m.cursorCol = row, col
		if quit {
			return m, tea.Quit
		}
		if key == "enter" || key == "space" {
			if m.installing {
				// Ya no se usa aquí, la selección de versión está arriba
				return m, nil
			}
			if m.cursorCol == 0 {
				service := m.choices[m.cursorRow]
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
				// Acción para port
			} else if m.cursorCol == 2 {
				// Acción para config
			}
		}
		// Acciones rápidas
		switch key {
		case "e", "E":
			// Acción para iniciar servicio
		case "x", "X":
			// Acción para detener servicio
		case "r", "R":
			// Acción para reiniciar servicio
		}
		return m, nil
	}
	return m, nil
}

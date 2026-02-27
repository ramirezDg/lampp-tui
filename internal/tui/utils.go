package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var BannerTitleL = lipgloss.NewStyle().
	Foreground(colorTitle).
	Bold(true).
	Render(`
██╗      █████╗ ███╗   ███╗██████╗ ██████╗ 
██║     ██╔══██╗████╗ ████║██╔══██╗██╔══██╗
██║     ███████║██╔████╔██║██████╔╝██████╔╝
██║     ██╔══██║██║╚██╔╝██║██╔═══╝ ██╔═══╝ 
███████╗██║  ██║██║ ╚═╝ ██║██║     ██║     
╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝     ╚═╝     Linux
`)

func Title() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(colorTitle).
		Bold(true).
		Align(lipgloss.Center)

	return titleStyle.Render(BannerTitleL)
}

func TextArea(content string) string {
	textStyle := lipgloss.NewStyle().
		Foreground(colorText).
		Padding(1).
		Align(lipgloss.Left)

	return textStyle.Render(content)
}

func Footer() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(colorText).
		Align(lipgloss.Left).
		MarginTop(-6)

	footerText := "← / ↑ / → / ↓ - Navigate | Enter - Action | q - Quit\na / w / s / d - Navigate | Space - Action\nPress 'h' for help"

	return footerStyle.Render(footerText)
}

func RenderTitle(width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, Title())
}

func RenderFooter(width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Left, Footer())
}

func RenderOptions(width int) string {
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

const columnWidth = 17

func RenderTable(m Model) string {
	colStyle := func() lipgloss.Style {
		return lipgloss.NewStyle().Width(columnWidth).Align(lipgloss.Center)
	}
	highlight := lipgloss.NewStyle().
		Foreground(colorHighlightFg).
		Background(colorHighlightBg).
		Bold(true).
		Width(columnWidth).
		Align(lipgloss.Center)

	// Colores para estados
	green := lipgloss.Color("#27F271")
	red := lipgloss.Color("#F22727")

	leftTitle := colStyle().Bold(true).Underline(true).MarginBottom(1).Render("Service")
	pidTitle := colStyle().Bold(true).Underline(true).MarginBottom(1).Render("PID")
	portTitle := colStyle().Bold(true).Underline(true).MarginBottom(1).Render("Port")
	configTitle := colStyle().Bold(true).Underline(true).MarginBottom(1).Render("Config")

	rows := make([]string, len(m.choices))
	for i := range m.choices {
		servicio := truncateOrPad(m.choices[i], columnWidth)
		var servicioCell string
		var color lipgloss.Color
		switch m.choices[i] {
		case "apache", "Apache":
			if m.ApacheStatus {
				color = green
			} else {
				color = red
			}
		case "mysql", "MySQL":
			if m.MySQLStatus {
				color = green
			} else {
				color = red
			}
		case "ftp", "FTP":
			if m.FTPStatus {
				color = green
			} else {
				color = red
			}
		default:
			color = colorText
		}
		servicioStyle := colStyle().Foreground(color)
		if m.cursorRow == i && m.cursorCol == 0 {
			servicioCell = highlight.Render(servicio)
		} else {
			servicioCell = servicioStyle.Render(servicio)
		}
		var pidCell, portCell, configCell string
		if m.ApacheStatus && (m.choices[i] == "apache" || m.choices[i] == "Apache") {
			pid := truncateOrPad(fmt.Sprintf("%d", m.pids[i]), columnWidth)
			port := truncateOrPad(m.ports[i], columnWidth)
			config := truncateOrPad(m.config[i], columnWidth)
			pidCell = colStyle().Render(pid)
			portCell = colStyle().Render(port)
			configCell = colStyle().Render(config)
		} else if m.MySQLStatus && (m.choices[i] == "mysql" || m.choices[i] == "MySQL") {
			pid := truncateOrPad(fmt.Sprintf("%d", m.pids[i]), columnWidth)
			port := truncateOrPad(m.ports[i], columnWidth)
			config := truncateOrPad(m.config[i], columnWidth)
			pidCell = colStyle().Render(pid)
			portCell = colStyle().Render(port)
			configCell = colStyle().Render(config)
		} else if m.FTPStatus && (m.choices[i] == "ftp" || m.choices[i] == "FTP") {
			pid := truncateOrPad(fmt.Sprintf("%d", m.pids[i]), columnWidth)
			port := truncateOrPad(m.ports[i], columnWidth)
			config := truncateOrPad(m.config[i], columnWidth)
			pidCell = colStyle().Render(pid)
			portCell = colStyle().Render(port)
			configCell = colStyle().Render(config)
		} else {
			pidCell = colStyle().Render(truncateOrPad("", columnWidth))
			portCell = colStyle().Render(truncateOrPad("", columnWidth))
			configCell = colStyle().Render(truncateOrPad("", columnWidth))
		}

		// Resaltar si el cursor está en la celda
		if m.cursorRow == i && m.cursorCol == 1 {
			pidCell = highlight.Render(truncateOrPad(fmt.Sprintf("%d", m.pids[i]), columnWidth))
		}
		if m.cursorRow == i && m.cursorCol == 2 {
			portCell = highlight.Render(truncateOrPad(m.ports[i], columnWidth))
		}
		if m.cursorRow == i && m.cursorCol == 3 {
			configCell = highlight.Render(truncateOrPad(m.config[i], columnWidth))
		}

		rows[i] = lipgloss.JoinHorizontal(lipgloss.Top, servicioCell, pidCell, portCell, configCell)
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top, leftTitle, pidTitle, portTitle, configTitle)
	table := lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinVertical(lipgloss.Left, rows...))

	terminalWidth := lipgloss.Width(lipgloss.NewStyle().Render(table))
	centeredTable := lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, table)

	return centeredTable
}

func RenderList(options []string, cursor int, selected map[int]struct{}) string {
	var s string
	for i, choice := range options {
		cur := " "
		if cursor == i {
			cur = ">"
		}
		check := " "
		if selected != nil {
			if _, ok := selected[i]; ok {
				check = "x"
			}
		}
		s += fmt.Sprintf("%s [%s] %s\n", cur, check, choice)
	}
	return s
}

// Funciones de Validación e Instalación de XAMPP

func Validate() ValidationResult {
	installed := isLAMPInstalled()
	return ValidationResult{
		OSName:    "linux",
		Installed: installed,
	}
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

func RenderVersionTable(m VersionTableModel) string {
	numCols := 4
	colWidth := 20
	colStyle := func() lipgloss.Style {
		return lipgloss.NewStyle().Width(colWidth).Align(lipgloss.Center)
	}
	highlight := lipgloss.NewStyle().Foreground(lipgloss.Color("#F27127")).Bold(true).Width(colWidth).Align(lipgloss.Center)

	n := len(m.Versiones)
	numRows := (n + numCols - 1) / numCols
	cells := make([][]string, numRows)
	for i := 0; i < numRows; i++ {
		cells[i] = make([]string, numCols)
		for j := 0; j < numCols; j++ {
			idx := i + j*numRows
			if idx < n {
				ver := m.Versiones[idx]
				if ver == "" {
					ver = "-"
				}
				if idx == m.SelectedVersion {
					cells[i][j] = highlight.Render(ver)
				} else {
					cells[i][j] = colStyle().Render(ver)
				}
			} else {
				cells[i][j] = colStyle().Render("")
			}
		}
	}
	var rows []string
	for i := 0; i < numRows; i++ {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells[i]...))
	}
	table := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return table
}

func truncateOrPad(s string, width int) string {
	runes := []rune(s)
	length := len(runes)
	if length > width {
		return string(runes[:width])
	}
	padding := width - length
	left := padding / 2
	right := padding - left
	return fmt.Sprintf("%s%s%s", spaces(left), s, spaces(right))
}

func spaces(n int) string {
	return string(make([]rune, n))
}

func RenderVersionInfoPanel(downloadURL string, selectedButton int) string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#F27127")).
		Padding(0, 1).
		Background(lipgloss.Color("#222222")).
		Foreground(lipgloss.Color("#F7F7F7"))

	labelStyle := lipgloss.NewStyle().Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F27127"))

	info := labelStyle.Render("URL Descarga: ") + valueStyle.Render(downloadURL) + "\n" +
		labelStyle.Render("Destino: ") + valueStyle.Render("./downloads/")

	// Botones
	btnStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#888888"))

	btnActive := btnStyle.Copy().
		Foreground(lipgloss.Color("#fff")).
		Background(lipgloss.Color("#F27127")).
		BorderForeground(lipgloss.Color("#F27127")).
		Bold(true)

	installBtn := btnStyle.Render("Install")
	quitBtn := btnStyle.Render("Quit")
	if selectedButton == 0 {
		installBtn = btnActive.Render("Install")
	} else {
		quitBtn = btnActive.Render("Quit")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, installBtn, quitBtn)
	content := info + "\n" + buttons

	return panelStyle.Render(content)
}

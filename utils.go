package main

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/lipgloss"
)

// Colores globales
var (
	colorTitle       = lipgloss.Color("#F27127")
	colorText        = lipgloss.Color("#333333")
	colorHighlightFg = lipgloss.Color("#F27127")
	colorHighlightBg = lipgloss.Color("7")
)

var BannerTitle = lipgloss.NewStyle().
	Foreground(colorTitle).
	Bold(true).
	Render(`
██╗  ██╗ █████╗ ███╗   ███╗██████╗ ██████╗ 
╚██╗██╔╝██╔══██╗████╗ ████║██╔══██╗██╔══██╗
 ╚███╔╝ ███████║██╔████╔██║██████╔╝██████╔╝
 ██╔██╗ ██╔══██║██║╚██╔╝██║██╔═══╝ ██╔═══╝ 
██╔╝ ██╗██║  ██║██║ ╚═╝ ██║██║     ██║     
╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝     ╚═╝     Windows
`)

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

	var banner string
	if osName == "windows" {
		banner = BannerTitle
	} else {
		banner = BannerTitleL
	}

	return titleStyle.Render(banner)
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

	leftTitle := colStyle().Bold(true).Underline(true).MarginBottom(1).Render("Servicio")
	pidTitle := colStyle().Bold(true).Underline(true).MarginBottom(1).Render("PID")
	portTitle := colStyle().Bold(true).Underline(true).MarginBottom(1).Render("Puerto")
	configTitle := colStyle().Bold(true).Underline(true).MarginBottom(1).Render("Config")

	rows := make([]string, len(m.choices))
	for i := range m.choices {
		// Servicio
		servicio := truncateOrPad(m.choices[i], columnWidth)
		var servicioCell string
		if m.cursorRow == i && m.cursorCol == 0 {
			servicioCell = highlight.Render(servicio)
		} else {
			servicioCell = colStyle().Render(servicio)
		}

		// PID
		pid := truncateOrPad(fmt.Sprintf("%d", m.pids[i]), columnWidth)
		var pidCell string
		if m.cursorRow == i && m.cursorCol == 1 {
			pidCell = highlight.Render(pid)
		} else {
			pidCell = colStyle().Render(pid)
		}

		// Puerto
		port := truncateOrPad(m.ports[i], columnWidth)
		var portCell string
		if m.cursorRow == i && m.cursorCol == 2 {
			portCell = highlight.Render(port)
		} else {
			portCell = colStyle().Render(port)
		}

		// Config
		config := truncateOrPad(m.config[i], columnWidth)
		var configCell string
		if m.cursorRow == i && m.cursorCol == 3 {
			configCell = highlight.Render(config)
		} else {
			configCell = colStyle().Render(config)
		}

		rows[i] = lipgloss.JoinHorizontal(lipgloss.Top, servicioCell, pidCell, portCell, configCell)
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top, leftTitle, pidTitle, portTitle, configTitle)
	table := lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinVertical(lipgloss.Left, rows...))

	terminalWidth := lipgloss.Width(lipgloss.NewStyle().Render(table))
	centeredTable := lipgloss.PlaceHorizontal(terminalWidth, lipgloss.Center, table)

	return centeredTable
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


func InstalarXAMPP() error {
	// Obtiene las versiones disponibles de XAMPP
	cmdVersiones := exec.Command("bash", "-c", "curl -s https://www.apachefriends.org/download.html | grep -oP 'xampp-linux-x64-\\K[0-9.]+' | sort -V | uniq")
	out, err := cmdVersiones.Output()
	if err != nil {
		return fmt.Errorf("error al obtener versiones de XAMPP: %v", err)
	}
	versiones := string(out)
	fmt.Println("Versiones disponibles de XAMPP:")
	fmt.Println(versiones)
	fmt.Print("Ingrese la versión que desea instalar: ")
	var version string
	fmt.Scanln(&version)
	url := fmt.Sprintf("https://www.apachefriends.org/xampp-files/%s/xampp-linux-x64-%s-0-installer.run", version, version)
	// Descarga el instalador
	cmdDescarga := exec.Command("wget", url, "-O", "xampp-installer.run")
	cmdDescarga.Stdout = nil
	cmdDescarga.Stderr = nil
	if err := cmdDescarga.Run(); err != nil {
		return fmt.Errorf("error al descargar XAMPP: %v", err)
	}
	// Da permisos de ejecución
	cmdPermisos := exec.Command("chmod", "+x", "xampp-installer.run")
	if err := cmdPermisos.Run(); err != nil {
		return fmt.Errorf("error al dar permisos: %v", err)
	}
	// Ejecuta el instalador con sudo
	cmdInstalar := exec.Command("sudo", "./xampp-installer.run")
	cmdInstalar.Stdout = nil
	cmdInstalar.Stderr = nil
	if err := cmdInstalar.Run(); err != nil {
		return fmt.Errorf("error al instalar XAMPP: %v", err)
	}
	return nil
}

// Instala XAMPP con la versión seleccionada
func InstalarXAMPPConVersion(version string) error {
	url := fmt.Sprintf("https://www.apachefriends.org/xampp-files/%s/xampp-linux-x64-%s-0-installer.run", version, version)
	cmdDescarga := exec.Command("wget", url, "-O", "xampp-installer.run")
	if err := cmdDescarga.Run(); err != nil {
		return fmt.Errorf("error al descargar XAMPP: %v", err)
	}
	cmdPermisos := exec.Command("chmod", "+x", "xampp-installer.run")
	if err := cmdPermisos.Run(); err != nil {
		return fmt.Errorf("error al dar permisos: %v", err)
	}
	cmdInstalar := exec.Command("sudo", "./xampp-installer.run")
	if err := cmdInstalar.Run(); err != nil {
		return fmt.Errorf("error al instalar XAMPP: %v", err)
	}
	return nil
}

func getXAMPPVersions() []string {
	cmd := exec.Command("bash", "-c", "curl -s https://www.apachefriends.org/download.html | grep -oP 'xampp-linux-x64-\\K[0-9.]+' | sort -V | uniq")
	out, err := cmd.Output()
	if err != nil {
		return []string{"Error al obtener versiones"}
	}
	return splitLines(string(out))
}

func splitLines(s string) []string {
	var res []string
	curr := ""
	for _, c := range s {
		if c == '\n' {
			if curr != "" {
				res = append(res, curr)
				curr = ""
			}
		} else {
			curr += string(c)
		}
	}
	if curr != "" {
		res = append(res, curr)
	}
	return res
}

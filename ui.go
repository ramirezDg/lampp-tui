package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Colores globales
var (
	colorTitle       = lipgloss.Color("#F27127")
	colorText        = lipgloss.Color("#333333")
	colorHighlightFg = lipgloss.Color("#F27127")
	colorHighlightBg = lipgloss.Color("7")
)

var bannerTitle = `
██╗  ██╗ █████╗ ███╗   ███╗██████╗ ██████╗ 
╚██╗██╔╝██╔══██╗████╗ ████║██╔══██╗██╔══██╗
 ╚███╔╝ ███████║██╔████╔██║██████╔╝██████╔╝
 ██╔██╗ ██╔══██║██║╚██╔╝██║██╔═══╝ ██╔═══╝ 
██╔╝ ██╗██║  ██║██║ ╚═╝ ██║██║     ██║     
╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝     ╚═╝     
`

func Title() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(colorTitle).
		Bold(true).
		Align(lipgloss.Center)

	return titleStyle.Render(bannerTitle)
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

func RenderTable(m model) string {
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

package main

import (
	"github.com/charmbracelet/lipgloss"
)

var bannerTitle = `
██╗  ██╗ █████╗ ███╗   ███╗██████╗ ██████╗ 
╚██╗██╔╝██╔══██╗████╗ ████║██╔══██╗██╔══██╗
 ╚███╔╝ ███████║██╔████╔██║██████╔╝██████╔╝
 ██╔██╗ ██╔══██║██║╚██╔╝██║██╔═══╝ ██╔═══╝ 
██╔╝ ██╗██║  ██║██║ ╚═╝ ██║██║     ██║     
╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝     ╚═╝     
`

// GridElement representa un elemento en el grid
type GridElement struct {
	Content  string
	Row      int
	Col      int
	Selected bool
}

// Grid genera un grid visual usando lipgloss
func Grid(elements []GridElement, rows, cols int) string {
	//cellWidth := 15
	//cellHeight := 3
	grid := make([][]string, rows)
	for i := range grid {
		grid[i] = make([]string, cols)
	}

	for _, el := range elements {
		if el.Row >= 0 && el.Row < rows && el.Col >= 0 && el.Col < cols {
			grid[el.Row][el.Col] = Box("", el.Selected)
		}
	}

	var renderedRows []string
	for _, row := range grid {
		// Si la celda está vacía, renderiza una caja vacía
		for i, cell := range row {
			if cell == "" {
				row[i] = Box("", false)
			}
		}
		// Une horizontalmente las celdas
		rowStr := lipgloss.JoinHorizontal(lipgloss.Center, row...)
		renderedRows = append(renderedRows, rowStr)
	}

	// Une verticalmente las filas
	return lipgloss.JoinVertical(lipgloss.Center, renderedRows...)
}

func Box(title string, selected bool) string {
	var (
		borderColor   = lipgloss.Color("#E5E7EB")
		titleColor    = lipgloss.Color("#FFFFFF")
		selectedColor = lipgloss.Color("#F27127")
	)

	if selected {
		titleColor = selectedColor
		borderColor = selectedColor
	}

	boxWidth := 15
	titleLen := lipgloss.Width(title)
	leftPad := (boxWidth - titleLen) / 2
	rightPad := boxWidth - titleLen - leftPad
	if leftPad < 0 {
		leftPad = 0
	}
	if rightPad < 0 {
		rightPad = 0
	}
	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true).
		Padding(0, leftPad, 0, rightPad).
		MarginBottom(0).
		Align(lipgloss.Center)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0).
		Width(15).
		Align(lipgloss.Center)

	box := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(title),
	)

	return boxStyle.Render(box)
}

func Title() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F27127")).
		Bold(true).
		Align(lipgloss.Center)

	return titleStyle.Render(bannerTitle)
}

func TextArea(content string) string {
	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333")).
		Padding(1).
		Align(lipgloss.Left)

	return textStyle.Render(content)
}

func Footer() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333")).
		Align(lipgloss.Center).
		MarginTop(-6)

	footerText := "[q, ctrl+c] quit | [↑, w, k] up | [↓, s, j] down | [enter, space] toggle state"

	return footerStyle.Render(footerText)
}

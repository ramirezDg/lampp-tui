package tui

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

var (
	colorTitle       = lipgloss.Color("#F27127")
	colorText        = lipgloss.Color("#333333")
	colorHighlightFg = lipgloss.Color("#F27127")
	colorHighlightBg = lipgloss.Color("7")
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

// getResponse realiza una petición HTTP y retorna la respuesta
func getResponse(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("error en getResponse:", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("receiving status of %d for url: %s", resp.StatusCode, url)
	}
	return resp, nil
}

func (pw *progressWriter) Start() {
	_, err := io.Copy(pw.file, io.TeeReader(pw.reader, pw))
	if err != nil {
		log.Println("error en progressWriter.Start:", err)
	}
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	pw.downloaded += len(p)
	if pw.total > 0 && pw.onProgress != nil {
		pw.onProgress(float64(pw.downloaded) / float64(pw.total))
	}
	return len(p), nil
}

func ShowDownloadXAMPP(url string) error {
	resp, err := getResponse(url)
	if err != nil {
		return fmt.Errorf("could not get response: %v", err)
	}
	defer resp.Body.Close()

	if resp.ContentLength <= 0 {
		return fmt.Errorf("can't parse content length, aborting download")
	}

	filename := filepath.Base(url)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create file: %v", err)
	}
	defer file.Close()

	pw := &progressWriter{
		total:  int(resp.ContentLength),
		file:   file,
		reader: resp.Body,
		onProgress: func(ratio float64) {
			p.Send(progressMsg(ratio))
		},
	}

	m := Model{
		pw:       pw,
		progress: progress.New(progress.WithDefaultBlend()),
	}
	p := tea.NewProgram(m)
	go pw.Start()
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %v", err)
	}
	return nil
}
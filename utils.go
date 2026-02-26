package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// Colores globales
var (
	colorTitle       = lipgloss.Color("#F27127")
	colorText        = lipgloss.Color("#333333")
	colorHighlightFg = lipgloss.Color("#F27127")
	colorHighlightBg = lipgloss.Color("7")
)

type progressMsg float64


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

type VersionTableModel struct {
	Versiones       []string
	SelectedVersion int
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

type SFResponse struct {
	Children []struct {
		Name string `json:"name"`
	} `json:"children"`
}

func ObtenerVersiones() ([]string, error) {
	bashScript := `
		curl -s https://sourceforge.net/projects/xampp/files/XAMPP%20Linux/ | \
		gawk '
			BEGIN { ver=""; count=""; link=""; }
			/<tr title=/ { ver=""; count=""; link=""; }
			/<a href="\/projects\/xampp\/files\/XAMPP%20Linux\/[0-9.]+\// {
				match($0, /<a href="(\/projects\/xampp\/files\/XAMPP%20Linux\/[0-9.]+\/)"/, arr)
				if (arr[1] != "") link=arr[1]
			}
			/<span class="name">/ {
				match($0, /<span class="name">([^<]+)<\/span>/, arr)
				if (arr[1] != "") ver=arr[1]
			}
			/<span class="count">/ {
				match($0, /<span class="count">([0-9,]+)<\/span>/, arr)
				gsub(",", "", arr[1])
				count=arr[1]
			}
			/<\/tr>/ {
				if (ver != "" && count != "" && count+0 > 5 && link != "") {
					print ver "|https://sourceforge.net" link
				}
			}'
		`
	cmd := exec.Command("bash", "-c", bashScript)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error ejecutando scraping: %v", err)
	}
	raw := splitLines(string(out))
	var versiones []string
	for _, v := range raw {
		if v == "" {
			continue
		}
		// sep := "|"
		idx := -1
		for i, c := range v {
			if c == '|' {
				idx = i
				break
			}
		}
		if idx > 0 {
			versiones = append(versiones, v[:idx])
		} else {
			versiones = append(versiones, v)
		}
	}
	if len(versiones) == 0 {
		return nil, fmt.Errorf("no se encontraron versiones con más de 5 descargas")
	}
	return versiones, nil
}

func InstalarXAMPP() error {
	versiones, err := ObtenerVersiones()
	if err != nil {
		return fmt.Errorf("error al obtener versiones: %v", err)
	}
	fmt.Println("Versiones disponibles de XAMPP:")
	for _, v := range versiones {
		fmt.Println(v)
	}
	fmt.Print("Ingrese la versión que desea instalar: ")
	var version string
	fmt.Scanln(&version)
	url := fmt.Sprintf("https://sourceforge.net/projects/xampp/files/XAMPP%%20Linux/%s/xampp-linux-x64-%s-0-installer.run/download", version, version)
	cmdDescarga := exec.Command("wget", url, "-O", "xampp-installer.run")
	if err := cmdDescarga.Run(); err != nil {
		return fmt.Errorf("error al descargar XAMPP: %v", err)
	}
	cmdPermisos := exec.Command("chmod", "+x", "xampp-installer.run")
	if err := cmdPermisos.Run(); err != nil {
		return fmt.Errorf("error al dar permisos: %v", err)
	}
	var confirm string
	fmt.Print("¿Desea ejecutar el instalador ahora? (s/n): ")
	fmt.Scanln(&confirm)
	if confirm == "s" || confirm == "S" {
		cmdInstalar := exec.Command("sudo", "./xampp-installer.run")
		if err := cmdInstalar.Run(); err != nil {
			return fmt.Errorf("error al instalar XAMPP: %v", err)
		}
	} else {
		fmt.Println("Instalador descargado pero no ejecutado.")
	}
	return nil
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

/* XAMPP Services */
type XAMPPServiceStatus struct {
	Apache bool
	MySQL  bool
	FTP    bool
}

func GetXAMPPServiceStatus() (XAMPPServiceStatus, error) {
	cmd := exec.Command("sudo", "/opt/lampp/lampp", "status")
	out, err := cmd.Output()
	if err != nil {
		return XAMPPServiceStatus{}, fmt.Errorf("error al obtener estado de XAMPP: %v", err)
	}
	status := string(out)
	return XAMPPServiceStatus{
		Apache: contains(status, "Apache is running"),
		MySQL:  contains(status, "MySQL is running"),
		FTP:    contains(status, "ProFTPD is running"),
	}, nil
}

func ControlXAMPPService(service, action string) error {
	var cmd *exec.Cmd
	switch service {
	case "apache":
		cmd = exec.Command("/opt/lampp/lampp", action+"apache")
	case "mysql":
		cmd = exec.Command("/opt/lampp/lampp", action+"mysql")
	case "ftp":
		cmd = exec.Command("/opt/lampp/lampp", action+"ftp")
	case "all":
		cmd = exec.Command("/opt/lampp/lampp", action)
	default:
		return fmt.Errorf("servicio no soportado: %s", service)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error al ejecutar acción %s en %s: %v", action, service, err)
	}
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (contains(s[1:], substr) || contains(s[:len(s)-1], substr)))) || (len(s) >= len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
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

	m := model{
		pw:       pw,
		progress: progress.New(progress.WithDefaultBlend()),
	}
	p = tea.NewProgram(m)
	go pw.Start()
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %v", err)
	}
	return nil
}

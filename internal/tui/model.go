package tui

import (
	"io"
	"os"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
)

var p *tea.Program

type progressMsg float64

type progressWriter struct {
	total      int
	downloaded int
	file       *os.File
	reader     io.Reader
	onProgress func(float64)
}

type ValidationResult struct {
	OSName    string
	Installed bool
}

type VersionTableModel struct {
	Versiones       []string
	SelectedVersion int
}

type Model struct {
	choices            []string
	pids               []int
	ports              []string
	config             []string
	cursorRow          int
	cursorCol          int
	status             []string
	statusInstallation []string
	ShowNewView        bool
	osName             string
	installed          bool
	showVersionList    bool
	versiones          []string
	selectedVersion    int
	installing         bool
	pw                 *progressWriter
	progress           progress.Model
}

func InitialModel() Model {
	return Model{
		choices:            []string{"Apache", "MySql", "FTP"},
		pids:               []int{0, 0, 0},
		ports:              []string{"", "", ""},
		config:             []string{"httpd.conf", "my.ini", "vsftpd.conf"},
		status:             []string{"stopped", "stopped", "stopped"},
		statusInstallation: []string{"Not Installed", "Installed"},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

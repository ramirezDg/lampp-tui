package tui

import (
	"xampp-tui/internal/services"

	tea "charm.land/bubbletea/v2"
)

var p *tea.Program

type ValidationResult struct {
	OSName    string
	Installed bool
}

type VersionTableModel struct {
	Versiones       []string
	SelectedVersion int
}

type Model struct {
	choices             []string
	pids                []int
	ports               []string
	config              []string
	cursorRow           int
	cursorCol           int
	status              []string
	statusInstallation  []string
	ShowNewView         bool
	osName              string
	installed           bool
	showVersionList     bool
	xamppVersions       []services.XAMPPVersion
	selectedVersion     int
	installing          bool
	optionsInstallation []string
	cursorInstall       int

	cursorVersionRow int
	cursorVersionCol int

	showVersionInfoPanel bool
	cursorVersionButton  int
}

func InitialModel() Model {
	ShowNewView := Validate().Installed
	return Model{
		choices:             []string{"Apache", "MySql", "FTP"},
		pids:                []int{0, 0, 0},
		ports:               []string{"", "", ""},
		config:              []string{"httpd.conf", "my.ini", "vsftpd.conf"},
		status:              []string{"stopped", "stopped", "stopped"},
		statusInstallation:  []string{"Not Installed", "Installed"},
		ShowNewView:         ShowNewView,
		optionsInstallation: []string{"Install XAMPP", "Quit/Exit"},
		cursorVersionRow:    0,
		cursorVersionCol:    0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

package tui

import (
	"xampp-tui/internal/installer"
	"xampp-tui/internal/xampp"

	tea "charm.land/bubbletea/v2"
)

// Model is the central BubbleTea state for the application. It is immutable
// between updates — Update always returns a new copy.
type Model struct {
	// Service runtime state
	ApacheStatus bool
	MySQLStatus  bool
	FTPStatus    bool
	pids         []int
	ports        []string

	// Main service table
	choices []string
	config  []string

	// Cursor for the main service table (row = service, col = column)
	cursorRow int
	cursorCol int

	// Installation flow
	ShowNewView         bool
	installing          bool
	xamppVersions       []installer.Version
	selectedVersion     int
	optionsInstallation []string
	cursorInstall       int

	// Version selection table cursor
	cursorVersionRow int
	cursorVersionCol int

	// Version info panel (shown after selecting a version)
	showVersionInfoPanel bool
	cursorVersionButton  int

	// Download progress
	downloading      bool
	downloadProgress float64 // 0.0–1.0
	downloadVersion  string
	downloadError    string

	// Recent activity log (capped at maxLogs entries)
	logs []string
}

func InitialModel() Model {
	status, _ := xampp.GetServiceStatus(backgroundCtx())
	return Model{
		choices:             []string{"Apache", "MySQL", "FTP"},
		pids:                []int{0, 0, 0},
		ports:               []string{"", "", ""},
		config:              []string{"httpd.conf", "my.ini", "vsftpd.conf"},
		ShowNewView:         !xampp.IsInstalled(),
		optionsInstallation: []string{"Install XAMPP", "Quit/Exit"},
		ApacheStatus:        status.Apache,
		MySQLStatus:         status.MySQL,
		FTPStatus:           status.FTP,
		logs: xampp.RecentLogs(20),
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

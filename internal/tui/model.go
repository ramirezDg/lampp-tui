package tui

import (
	"xampp-tui/internal/installer"
	"xampp-tui/internal/xampp"

	tea "charm.land/bubbletea/v2"
)

// ─── install-menu helpers ─────────────────────────────────────────────────────

// readyToInstall returns downloaded .run files that are not yet installed,
// suitable for immediate installation without re-downloading.
func readyToInstall(installed []xampp.InstalledVersion) []string {
	downloaded := installer.DownloadedVersions()
	installedSet := make(map[string]bool, len(installed))
	for _, v := range installed {
		installedSet[v.Version] = true
	}
	var ready []string
	for _, ver := range downloaded {
		if !installedSet[ver] {
			ready = append(ready, ver)
		}
	}
	return ready
}

// buildInstallOptions constructs the option list for the welcome screen.
// Downloaded-but-not-installed versions appear first as quick-install entries.
func buildInstallOptions(downloaded []string) []string {
	opts := make([]string, 0, len(downloaded)+2)
	for _, ver := range downloaded {
		opts = append(opts, "Install XAMPP "+ver+"  (ready to install)")
	}
	opts = append(opts, "Download new version", "Quit")
	return opts
}

// Model is the central BubbleTea state for the application. It is immutable
// between updates — Update always returns a new copy.
type Model struct {
	// ── Service runtime state ─────────────────────────────────────────────────
	ApacheStatus bool
	MySQLStatus  bool
	FTPStatus    bool
	pids         []int
	ports        []string

	// ── Main service table ────────────────────────────────────────────────────
	choices []string
	config  []string

	// Cursor for the main service table (row = service, col = column)
	cursorRow int
	cursorCol int

	// ── Install / version-picker flow ─────────────────────────────────────────
	ShowNewView         bool
	installing          bool
	xamppVersions       []installer.Version
	selectedVersion     int
	optionsInstallation []string
	cursorInstall       int

	// Downloaded installers that are not yet installed (ready for immediate use).
	downloadedVersions []string

	// Version selection table cursor
	cursorVersionRow int
	cursorVersionCol int

	// Version info panel (shown after selecting a version in the picker)
	showVersionInfoPanel bool
	cursorVersionButton  int

	// ── Download progress ─────────────────────────────────────────────────────
	downloading      bool
	downloadProgress float64 // 0.0–1.0
	downloadVersion  string
	downloadError    string

	// ── Post-download install prompt ──────────────────────────────────────────
	postDownload    bool
	postDownloadBtn int // 0=Install  1=Skip

	// ── XAMPP installer runner ────────────────────────────────────────────────
	runningInstaller    bool
	installerStatus     string
	installerError      string
	installerBackgrounded bool

	// ── Download backgrounding ────────────────────────────────────────────────
	downloadBackgrounded bool

	// ── URL info modal (shown when opening port in browser) ───────────────────
	showURLModal bool
	urlModalSvc  string
	urlModalURL  string

	// ── Installed versions management panel ───────────────────────────────────
	showVersionsPanel  bool
	installedVersions  []xampp.InstalledVersion
	cursorVersionsMgmt int

	// ── Column-action dialog (kill / config / switch_version) ─────────────────
	showDialog bool
	dialogType string // "kill" | "config" | "switch_version"
	dialogBtn  int    // 0=Yes  1=No
	dialogRow  int    // service row or version index that triggered the dialog

	// ── Config-file paths (parallel to choices/config display names) ──────────
	configPaths []string

	// ── Recent activity log (from XAMPP log file) ─────────────────────────────
	logs []string
}

func InitialModel() Model {
	status, _ := xampp.GetServiceStatus(backgroundCtx())
	installed := xampp.ScanInstalledVersions()
	downloaded := readyToInstall(installed)

	return Model{
		choices:             []string{"Apache", "MySQL", "FTP"},
		pids:                []int{0, 0, 0},
		ports:               []string{"", "", ""},
		config:              []string{"httpd.conf", "my.cnf", "proftpd.conf"},
		ShowNewView:         !xampp.IsInstalled(),
		optionsInstallation: buildInstallOptions(downloaded),
		ApacheStatus:        status.Apache,
		MySQLStatus:         status.MySQL,
		FTPStatus:           status.FTP,
		configPaths: []string{
			"/opt/lampp/etc/httpd.conf",
			"/opt/lampp/etc/my.cnf",
			"/opt/lampp/etc/proftpd.conf",
		},
		logs:               xampp.RecentLogs(20),
		installedVersions:  installed,
		downloadedVersions: downloaded,
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

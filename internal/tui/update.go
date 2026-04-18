package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
	"xampp-tui/internal/installer"
	"xampp-tui/internal/platform"
	"xampp-tui/internal/xampp"

	tea "charm.land/bubbletea/v2"
)

// ─── column-action messages & commands ───────────────────────────────────────

type editorClosedMsg struct{ err error }

// openBrowserCmd opens url in the system browser using the platform-specific
// implementation (xdg-open on Linux, cmd /c start on Windows).
func openBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		platform.OpenBrowser(url)
		return nil
	}
}

// openEditorCmd suspends the TUI, opens the platform editor for path, then
// resumes (nano on Linux, notepad on Windows).
func openEditorCmd(path string) tea.Cmd {
	return tea.ExecProcess(platform.EditorCommand(path), func(err error) tea.Msg {
		return editorClosedMsg{err: err}
	})
}

// ─── download messages & commands ────────────────────────────────────────────

type downloadProgressMsg struct{ pct float64 }
type downloadDoneMsg struct{ err error }

// dlCh is the channel through which the download goroutine feeds the TUI.
var dlCh chan tea.Msg

// startDownloadCmd kicks off the download goroutine and returns a cmd that
// blocks until the first message arrives from it.
// dirURL is the SourceForge directory URL for the version; when non-empty it
// is used to resolve the exact installer filename from the directory listing.
func startDownloadCmd(version, dirURL string) tea.Cmd {
	dlCh = make(chan tea.Msg, 200)
	return func() tea.Msg {
		go func() {
			err := installer.Download(version, dirURL, func(done, total int64) {
				if total > 0 {
					dlCh <- downloadProgressMsg{pct: float64(done) / float64(total)}
				}
			})
			dlCh <- downloadDoneMsg{err: err}
		}()
		return <-dlCh
	}
}

// nextDownloadMsgCmd reads the next message from the active download channel.
func nextDownloadMsgCmd() tea.Cmd {
	return func() tea.Msg { return <-dlCh }
}

// ─── installer messages & commands ───────────────────────────────────────────

type installerProgressMsg struct{ status string }
type installerDoneMsg struct{ err error }

var installCh chan tea.Msg

// startInstallerCmd runs the XAMPP .run installer in a goroutine and pipes
// progress/done messages back to the TUI.
func startInstallerCmd(version string) tea.Cmd {
	installCh = make(chan tea.Msg, 50)
	return func() tea.Msg {
		go func() {
			err := installer.RunInstaller(version, func(msg string) {
				installCh <- installerProgressMsg{status: msg}
			})
			installCh <- installerDoneMsg{err: err}
		}()
		return <-installCh
	}
}

func nextInstallerMsgCmd() tea.Cmd {
	return func() tea.Msg { return <-installCh }
}

// ─── background context ───────────────────────────────────────────────────────

func backgroundCtx() context.Context { return context.Background() }

// ─── tick ────────────────────────────────────────────────────────────────────

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m = m.refreshSnapshot()
		m.logs = xampp.RecentLogs(20)
		return m, tickCmd()

	case editorClosedMsg:
		m = m.refreshSnapshot()
		m.logs = xampp.RecentLogs(20)
		return m, nil

	case downloadProgressMsg:
		m.downloadProgress = msg.pct
		return m, nextDownloadMsgCmd()

	case downloadDoneMsg:
		if msg.err != nil {
			// Keep m.downloading = true so downloadPane stays visible with the
			// error. The key handler dismisses it when the user presses q/esc.
			m.downloadError = msg.err.Error()
		} else {
			m.downloading = false
			m.downloadProgress = 1.0
			m.postDownload = true
			m.postDownloadBtn = 0
		}
		return m, nil

	case installerProgressMsg:
		m.installerStatus = msg.status
		return m, nextInstallerMsgCmd()

	case installerDoneMsg:
		m.runningInstaller = false
		if msg.err != nil {
			m.installerError = msg.err.Error()
			m.installerStatus = ""
		} else {
			m.installerStatus = "Installation complete!"
			// Create/update the /opt/lampp symlink to the new installation.
			targetDir := filepath.Join(platform.XAMPPBaseDir(), m.downloadVersion)
			xampp.SwitchVersion(targetDir) //nolint:errcheck
			// Refresh everything.
			m.installedVersions = xampp.ScanInstalledVersions()
			// Remove the just-installed version from the "ready to install" list.
			m.downloadedVersions = readyToInstall(m.installedVersions)
			m.optionsInstallation = buildInstallOptions(m.downloadedVersions)
			m = m.refreshSnapshot()
			m.logs = xampp.RecentLogs(20)
			m.ShowNewView = !xampp.IsInstalled()

			// Automatically add the XAMPP bin dir to the user's shell config so
			// that php/mysql always point to the active XAMPP version.
			binDir := platform.LamppBinDir()
			cfgPath := xampp.DetectShellConfig()
			if cfgPath != "" {
				if added, err := xampp.EnsureLamppInPATH(cfgPath); added {
					m.pathNotice = fmt.Sprintf(
						"%s added to PATH in %s\n"+
							"Run: source %s   (or open a new terminal)", binDir, cfgPath, cfgPath)
				} else if err != nil {
					m.pathNotice = fmt.Sprintf(
						"Could not update %s: %s\n"+
							"Add manually: export PATH=\"%s:$PATH\"", cfgPath, err, binDir)
				} else {
					// Already present — no notice needed.
					m.pathNotice = ""
				}
			}
		}
		return m, nil
	}

	msg2, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	key := msg2.String()

	// URL info modal: any key closes it.
	if m.showURLModal {
		m.showURLModal = false
		return m, nil
	}

	// PATH notice banner: any key dismisses it.
	if m.pathNotice != "" && !m.pathNoticeDone {
		m.pathNoticeDone = true
		return m, nil
	}

	// Installer running (not backgrounded) — q/esc sends to background.
	if m.runningInstaller && !m.installerBackgrounded {
		switch key {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "esc":
			m.installerBackgrounded = true
		}
		return m, nil
	}

	// Downloading (not backgrounded) — q/esc sends to background (or dismisses error).
	if m.downloading && !m.downloadBackgrounded {
		switch key {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "esc":
			if m.downloadError != "" {
				// Error is visible — dismiss and return to the correct panel.
				m.downloading = false
				m.downloadError = ""
				m.ShowNewView = !xampp.IsInstalled()
			} else {
				m.downloadBackgrounded = true
			}
		}
		return m, nil
	}

	// Post-download install prompt.
	if m.postDownload {
		return m.handlePostDownload(key)
	}

	// Versions management panel (with possible dialog overlay).
	if m.showVersionsPanel {
		if m.showDialog {
			return m.handleDialog(key)
		}
		return m.handleVersionsMgmt(key)
	}

	if m.showDialog {
		return m.handleDialog(key)
	}

	switch {
	case m.ShowNewView && m.installing:
		return m.handleVersionSelection(key)
	case m.ShowNewView:
		return m.handleInstallMenu(key)
	default:
		return m.handleMainMenu(key)
	}
}

// ─── sub-handlers ────────────────────────────────────────────────────────────

// handleVersionSelection processes keyboard input while the version-picker
// table (and its info panel) is visible.
func (m Model) handleVersionSelection(key string) (tea.Model, tea.Cmd) {
	if m.showVersionInfoPanel {
		return m.handleVersionInfoPanel(key)
	}

	numCols := 4
	n := len(m.xamppVersions)
	numRows := (n + numCols - 1) / numCols

	row, col, quit := navigate(key, m.cursorVersionRow, m.cursorVersionCol, numRows, numCols)

	// Clamp cursor to valid cells only.
	if idx := row + col*numRows; idx >= n {
		row, col = m.cursorVersionRow, m.cursorVersionCol
	}
	m.cursorVersionRow, m.cursorVersionCol = row, col
	m.selectedVersion = m.cursorVersionRow + m.cursorVersionCol*numRows

	switch key {
	case "q", "esc":
		// Return early so the quit flag from navigate() doesn't trigger tea.Quit.
		m.installing = false
		m.ShowNewView = !xampp.IsInstalled()
		return m, nil
	case "enter", " ":
		m.showVersionInfoPanel = true
		m.cursorVersionButton = 0
	}
	if quit {
		return m, tea.Quit
	}
	return m, nil
}

// handleVersionInfoPanel processes keyboard input while the download-info
// panel overlay is shown.
func (m Model) handleVersionInfoPanel(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "a", "A", "←":
		if m.cursorVersionButton > 0 {
			m.cursorVersionButton--
		}
	case "right", "d", "D", "→":
		if m.cursorVersionButton < 1 {
			m.cursorVersionButton++
		}
	case "q", "esc":
		m.showVersionInfoPanel = false
	case "enter", " ":
		if m.cursorVersionButton == 0 {
			ver := m.xamppVersions[m.selectedVersion].Name
			m.showVersionInfoPanel = false
			m.installing = false
			m.ShowNewView = false

			// Check if already downloaded — skip download, go straight to install prompt.
			alreadyDownloaded := false
			for _, v := range m.downloadedVersions {
				if v == ver {
					alreadyDownloaded = true
					break
				}
			}

			if alreadyDownloaded {
				m.downloadVersion = ver
				m.downloadProgress = 1.0
				m.postDownload = true
				m.postDownloadBtn = 0
			} else {
				m.downloading = true
				m.downloadProgress = 0
				m.downloadVersion = ver
				m.downloadError = ""
				dirURL := m.xamppVersions[m.selectedVersion].DownloadURL
				return m, startDownloadCmd(ver, dirURL)
			}
		}
		m.showVersionInfoPanel = false
	}
	return m, nil
}

// handleInstallMenu processes keyboard input on the "XAMPP not installed"
// welcome screen. Options are built dynamically: downloaded-but-not-installed
// versions appear first, followed by "Download new version" and "Quit".
func (m Model) handleInstallMenu(key string) (tea.Model, tea.Cmd) {
	row, _, quit := navigate(key, m.cursorInstall, 0, len(m.optionsInstallation), 1)
	m.cursorInstall = row

	if quit {
		return m, tea.Quit
	}

	if key == "enter" || key == " " {
		nDownloaded := len(m.downloadedVersions)

		switch {
		case m.cursorInstall < nDownloaded:
			// Install an already-downloaded version directly.
			ver := m.downloadedVersions[m.cursorInstall]
			m.downloadVersion = ver
			m.downloadProgress = 1.0
			m.runningInstaller = true
			m.installerStatus = "Starting installer…"
			m.installerError = ""
			return m, startInstallerCmd(ver)

		case m.cursorInstall == nDownloaded:
			// "Download new version" — open the version picker.
			// Refresh downloads in case files were added externally.
			m.downloadedVersions = readyToInstall(m.installedVersions)
			m.optionsInstallation = buildInstallOptions(m.downloadedVersions)
			if len(m.xamppVersions) == 0 {
				versions, err := installer.FetchVersions()
				if err != nil {
					versions = []installer.Version{{Name: "Error fetching versions", DownloadURL: ""}}
				}
				m.xamppVersions = versions
				m.cursorVersionRow = 0
				m.cursorVersionCol = 0
				m.selectedVersion = 0
			}
			m.installing = true

		default:
			// Quit
			return m, tea.Quit
		}
	}
	return m, nil
}

// handleMainMenu processes keyboard input on the main service management
// screen.
func (m Model) handleMainMenu(key string) (tea.Model, tea.Cmd) {
	row, col, quit := navigate(key, m.cursorRow, m.cursorCol, len(m.choices), 4)
	m.cursorRow, m.cursorCol = row, col

	if quit {
		return m, tea.Quit
	}

	if key == "enter" || key == " " {
		switch col {
		case 0: // Toggle service start/stop.
			service := m.choices[m.cursorRow]
			if m.isRunning(m.cursorRow) {
				xampp.Control(service, "stop")
			} else {
				xampp.Control(service, "start")
			}
			m = m.refreshSnapshot()
			m.logs = xampp.RecentLogs(20)

		case 1: // PID → ask to kill process.
			if m.pids[m.cursorRow] > 0 {
				m.showDialog = true
				m.dialogType = "kill"
				m.dialogBtn = 1 // default No (safe)
				m.dialogRow = m.cursorRow
			}

		case 2: // Port → open in browser + show URL modal for reference.
			port := m.ports[m.cursorRow]
			if m.isRunning(m.cursorRow) && port != "" && port != "N/A" {
				url := "http://localhost:" + port
				m.showURLModal = true
				m.urlModalSvc = m.choices[m.cursorRow]
				m.urlModalURL = url
				return m, openBrowserCmd(url)
			}

		case 3: // Config → ask to open in editor.
			m.showDialog = true
			m.dialogType = "config"
			m.dialogBtn = 0 // default Yes (non-destructive)
			m.dialogRow = m.cursorRow
		}
		return m, nil
	}

	switch key {
	case "e", "E":
		xampp.Control("all", "restart")
		m = m.refreshSnapshot()
		m.logs = xampp.RecentLogs(20)
	case "x", "X":
		xampp.Control("all", "stop")
		m = m.refreshSnapshot()
		m.logs = xampp.RecentLogs(20)
	case "r", "R":
		xampp.Control("all", "start")
		m = m.refreshSnapshot()
		m.logs = xampp.RecentLogs(20)
	case "v", "V":
		// Open the installed-versions management panel.
		m.showVersionsPanel = true
		m.installedVersions = xampp.ScanInstalledVersions()
		m.cursorVersionsMgmt = 0
	case "i", "I":
		// Open version picker to download/install a new XAMPP version.
		// Refresh downloadedVersions so any files added since startup appear.
		m.downloadedVersions = readyToInstall(m.installedVersions)
		m.optionsInstallation = buildInstallOptions(m.downloadedVersions)
		if len(m.xamppVersions) == 0 {
			versions, err := installer.FetchVersions()
			if err != nil {
				versions = []installer.Version{{Name: "Error fetching versions", DownloadURL: ""}}
			}
			m.xamppVersions = versions
			m.cursorVersionRow = 0
			m.cursorVersionCol = 0
			m.selectedVersion = 0
		}
		m.ShowNewView = true
		m.installing = true
	}

	return m, nil
}

// handleVersionsMgmt processes keyboard input in the installed-versions panel.
func (m Model) handleVersionsMgmt(key string) (tea.Model, tea.Cmd) {
	n := len(m.installedVersions)

	switch key {
	case "up", "w", "W", "↑":
		if m.cursorVersionsMgmt > 0 {
			m.cursorVersionsMgmt--
		}
	case "down", "s", "S", "↓":
		if m.cursorVersionsMgmt < n-1 {
			m.cursorVersionsMgmt++
		}
	case "q", "esc":
		m.showVersionsPanel = false
	case "enter", " ":
		if n > 0 && m.cursorVersionsMgmt < n {
			ver := m.installedVersions[m.cursorVersionsMgmt]
			if !ver.IsActive {
				m.showDialog = true
				m.dialogType = "switch_version"
				m.dialogBtn = 0
				m.dialogRow = m.cursorVersionsMgmt
			}
		}
	case "d", "D":
		if n > 0 && m.cursorVersionsMgmt < n {
			ver := m.installedVersions[m.cursorVersionsMgmt]
			if !ver.IsActive {
				m.showDialog = true
				m.dialogType = "uninstall"
				m.dialogBtn = 1 // default No — destructive action
				m.dialogRow = m.cursorVersionsMgmt
			}
		}
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// handlePostDownload processes keyboard input on the post-download install
// prompt screen.
func (m Model) handlePostDownload(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "a", "A", "←":
		if m.postDownloadBtn > 0 {
			m.postDownloadBtn--
		}
	case "right", "d", "D", "→":
		if m.postDownloadBtn < 1 {
			m.postDownloadBtn++
		}
	case "q", "esc":
		m.postDownload = false
		m.ShowNewView = !xampp.IsInstalled()
	case "enter", " ":
		if m.postDownloadBtn == 0 { // Install Now
			m.postDownload = false
			m.runningInstaller = true
			m.installerStatus = "Starting installer…"
			m.installerError = ""
			return m, startInstallerCmd(m.downloadVersion)
		}
		// Skip
		m.postDownload = false
		m.ShowNewView = !xampp.IsInstalled()
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// ─── dialog handler ──────────────────────────────────────────────────────────

func (m Model) handleDialog(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "a", "A", "←":
		if m.dialogBtn > 0 {
			m.dialogBtn--
		}
	case "right", "d", "D", "→":
		if m.dialogBtn < 1 {
			m.dialogBtn++
		}
	case "esc", "q":
		m.showDialog = false
	case "enter", " ":
		m.showDialog = false
		if m.dialogBtn == 0 { // Yes
			return m.executeDialogAction()
		}
	}
	return m, nil
}

// executeDialogAction runs the confirmed action for the active dialog.
func (m Model) executeDialogAction() (tea.Model, tea.Cmd) {
	switch m.dialogType {
	case "kill":
		pid := fmt.Sprintf("%d", m.pids[m.dialogRow])
		platform.KillProcess(pid)
		m = m.refreshSnapshot()
		m.logs = xampp.RecentLogs(20)

	case "config":
		return m, openEditorCmd(m.configPaths[m.dialogRow])

	case "switch_version":
		if m.dialogRow < len(m.installedVersions) {
			ver := m.installedVersions[m.dialogRow]
			if err := xampp.SwitchVersion(ver.Path); err == nil {
				m.installedVersions = xampp.ScanInstalledVersions()
				m = m.refreshSnapshot()
				m.logs = xampp.RecentLogs(20)
			}
		}

	case "uninstall":
		if m.dialogRow < len(m.installedVersions) {
			ver := m.installedVersions[m.dialogRow]
			if err := xampp.UninstallVersion(ver.Path); err == nil {
				m.installedVersions = xampp.ScanInstalledVersions()
				m.downloadedVersions = readyToInstall(m.installedVersions)
				m.optionsInstallation = buildInstallOptions(m.downloadedVersions)
				// Adjust cursor if it's now out of bounds.
				if m.cursorVersionsMgmt >= len(m.installedVersions) {
					m.cursorVersionsMgmt = max(0, len(m.installedVersions)-1)
				}
			}
		}
	}
	return m, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (m Model) refreshSnapshot() Model {
	snap, err := xampp.GetSnapshot(backgroundCtx())
	if err != nil {
		return m
	}
	m.ApacheStatus = snap.Status.Apache
	m.MySQLStatus = snap.Status.MySQL
	m.FTPStatus = snap.Status.FTP

	for i, svc := range m.choices {
		info, ok := snap.Details[svc]
		if !ok {
			m.pids[i] = 0
			m.ports[i] = ""
			continue
		}
		var pid int
		fmt.Sscanf(info.PID, "%d", &pid)
		m.pids[i] = pid
		m.ports[i] = info.Port
	}
	return m
}

func (m Model) isRunning(row int) bool {
	switch m.choices[row] {
	case "Apache":
		return m.ApacheStatus
	case "MySQL":
		return m.MySQLStatus
	case "FTP":
		return m.FTPStatus
	}
	return false
}

// navigate handles directional key input and returns updated row/col and a
// quit flag. Single source of truth for keyboard navigation across all screens.
func navigate(key string, row, col, maxRow, maxCol int) (newRow, newCol int, quit bool) {
	newRow, newCol = row, col
	switch key {
	case "up", "w", "W", "↑":
		if newRow > 0 {
			newRow--
		}
	case "down", "s", "S", "↓":
		if newRow < maxRow-1 {
			newRow++
		}
	case "left", "a", "A", "←":
		if newCol > 0 {
			newCol--
		}
	case "right", "d", "D", "→":
		if newCol < maxCol-1 {
			newCol++
		}
	case "ctrl+c", "q":
		quit = true
	}
	return
}

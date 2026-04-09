package tui

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"xampp-tui/internal/installer"
	"xampp-tui/internal/xampp"

	tea "charm.land/bubbletea/v2"
)

// ─── column-action messages & commands ───────────────────────────────────────

type editorClosedMsg struct{ err error }

// openBrowserCmd opens url in the system browser, detached from the TUI's
// terminal session so it doesn't interfere with raw-mode input handling.
func openBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		for _, browser := range []string{"xdg-open", "sensible-browser", "x-www-browser"} {
			if _, err := exec.LookPath(browser); err != nil {
				continue
			}
			cmd := exec.Command(browser, url)
			cmd.Stdin = nil
			cmd.Stdout = nil
			cmd.Stderr = nil
			// Setsid creates a new session so the child is fully detached from
			// the TUI's controlling terminal (raw mode). Without this xdg-open
			// can silently fail inside Bubble Tea.
			cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
			if cmd.Start() == nil {
				return nil
			}
		}
		return nil
	}
}

// openEditorCmd suspends the TUI, opens nano for path, then resumes.
func openEditorCmd(path string) tea.Cmd {
	return tea.ExecProcess(exec.Command("nano", path), func(err error) tea.Msg {
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
func startDownloadCmd(version string) tea.Cmd {
	dlCh = make(chan tea.Msg, 200)
	return func() tea.Msg {
		go func() {
			err := installer.Download(version, func(done, total int64) {
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
		m.downloading = false
		m.downloadProgress = 1.0
		if msg.err != nil {
			m.downloadError = msg.err.Error()
		} else {
			// Offer to install the downloaded version.
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
			targetDir := filepath.Join(installer.XAMPPBaseDir, m.downloadVersion)
			xampp.SwitchVersion(targetDir) //nolint:errcheck
			// Refresh everything.
			m.installedVersions = xampp.ScanInstalledVersions()
			m = m.refreshSnapshot()
			m.logs = xampp.RecentLogs(20)
			m.ShowNewView = !xampp.IsInstalled()
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

	// Downloading (not backgrounded) — q/esc sends to background.
	if m.downloading && !m.downloadBackgrounded {
		switch key {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "esc":
			m.downloadBackgrounded = true
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
			m.downloading = true
			m.downloadProgress = 0
			m.downloadVersion = ver
			m.downloadError = ""
			return m, startDownloadCmd(ver)
		}
		m.showVersionInfoPanel = false
	}
	return m, nil
}

// handleInstallMenu processes keyboard input on the "XAMPP not installed"
// welcome screen.
func (m Model) handleInstallMenu(key string) (tea.Model, tea.Cmd) {
	row, _, quit := navigate(key, m.cursorInstall, 0, len(m.optionsInstallation), 1)
	m.cursorInstall = row

	if quit {
		return m, tea.Quit
	}

	if key == "enter" || key == " " {
		switch m.cursorInstall {
		case 0: // Install XAMPP
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
		case 1: // Quit
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
		exec.Command("kill", pid).Run() //nolint:errcheck
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
	}
	return m, nil
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

package tui

import (
	"context"
	"fmt"
	"time"
	"xampp-tui/internal/installer"
	"xampp-tui/internal/xampp"

	tea "charm.land/bubbletea/v2"
)

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

// backgroundCtx returns a plain background context. Centralised here so that
// service calls throughout the TUI package are easy to swap for a cancelable
// context later.
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

	case downloadProgressMsg:
		m.downloadProgress = msg.pct
		return m, nextDownloadMsgCmd()

	case downloadDoneMsg:
		m.downloading = false
		m.downloadProgress = 1.0
		if msg.err != nil {
			m.downloadError = msg.err.Error()
		}
		return m, nil
	}

	msg2, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	key := msg2.String()

	// While downloading, only allow quitting.
	if m.downloading {
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil
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
		m.installing = false
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

// handleInstallMenu processes keyboard input on the "XAMPP not installed" welcome
// screen.
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
		if col == 0 {
			service := m.choices[m.cursorRow]
			if m.isRunning(m.cursorRow) {
				xampp.Control(service, "stop")
			} else {
				xampp.Control(service, "start")
			}
			m = m.refreshSnapshot()
			m.logs = xampp.RecentLogs(20)
		}
		// col 1 (port) and col 2 (config) have no action yet.
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
	}

	return m, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// refreshSnapshot calls GetSnapshot once and updates all service state fields.
// This replaces the three separate status-refresh blocks that were scattered
// across the old Update function.
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

// isRunning reports whether the service at the given table row is currently
// running, using the authoritative status fields rather than the local
// m.status slice that the old code maintained manually.
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
// quit flag. It is the single source of truth for keyboard navigation across
// all screens.
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

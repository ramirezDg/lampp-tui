package tui

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

func (m Model) View() tea.View {
	w, h := terminalSize()

	footerStr := contextFooter(m, w)
	footerH := lipgloss.Height(footerStr)
	mainH := h - footerH - 2 // -2 for the \n\n separator before the footer
	if mainH < 10 {
		mainH = 10
	}

	var content string
	switch {
	case m.runningInstaller && !m.installerBackgrounded:
		content = installerPane(m, w, mainH)
	case m.postDownload:
		content = postDownloadPane(m, w, mainH)
	case m.downloading && !m.downloadBackgrounded:
		content = downloadPane(m, w, mainH)
	case m.showVersionsPanel:
		content = versionsMgmtPane(m, w, mainH)
	case m.ShowNewView:
		content = installPane(m, w, mainH)
	default:
		content = adminPane(m, w, mainH)
	}

	return tea.NewView(content + "\n\n" + footerStr)
}

// paneHeader renders the title with a small top margin and returns the height
// remaining for the pane body below it.
func paneHeader(w, h int) (titleStr string, belowH int) {
	titleStr = "\n\n" + RenderTitle(w)
	titleH := lipgloss.Height(titleStr)
	belowH = h - titleH - 2
	if belowH < 1 {
		belowH = 1
	}
	return
}

// ─── context-aware footer ─────────────────────────────────────────────────────

func contextFooter(m Model, w int) string {
	sep := lipgloss.NewStyle().Foreground(colorBorder).Render("·")
	key := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	desc := lipgloss.NewStyle().Foreground(colorMuted)

	hint := func(k, d string) string {
		return key.Render(k) + desc.Render(" "+d)
	}
	join := func(hints ...string) string {
		return strings.Join(hints, "  "+sep+"  ")
	}

	navLine := lipgloss.PlaceHorizontal(w, lipgloss.Center,
		join(hint("↑↓←→/wasd", "Navigate"), hint("Enter/Space", "Action"), hint("q", "Quit")))

	// Background task indicator (right-aligned, above nav hints).
	var bgLine string
	if m.downloading && m.downloadBackgrounded {
		badge := lipgloss.NewStyle().Foreground(colorTitle).
			Render(fmt.Sprintf("⟳  DL %.0f%%", m.downloadProgress*100))
		bgLine = lipgloss.PlaceHorizontal(w, lipgloss.Right, badge) + "\n"
	} else if m.runningInstaller && m.installerBackgrounded {
		badge := lipgloss.NewStyle().Foreground(colorTitle).Render("⟳  Installing…")
		bgLine = lipgloss.PlaceHorizontal(w, lipgloss.Right, badge) + "\n"
	}

	switch {
	case m.showVersionsPanel:
		extra := lipgloss.PlaceHorizontal(w, lipgloss.Center,
			join(hint("Enter", "Switch version"), hint("q/Esc", "Back")))
		return bgLine + navLine + "\n" + extra

	case !m.ShowNewView && !m.postDownload:
		extra := lipgloss.PlaceHorizontal(w, lipgloss.Center,
			join(hint("e", "Start all"), hint("x", "Stop all"), hint("r", "Restart all"),
				hint("v", "Versions"), hint("i", "Install")))
		return bgLine + navLine + "\n" + extra

	case m.downloading && !m.downloadBackgrounded,
		m.runningInstaller && !m.installerBackgrounded:
		extra := lipgloss.PlaceHorizontal(w, lipgloss.Center,
			hint("q/Esc", "Send to background"))
		return navLine + "\n" + extra
	}

	return bgLine + navLine
}

// ─── panes ───────────────────────────────────────────────────────────────────

// adminPane renders the main service management screen.
func adminPane(m Model, w, h int) string {
	titleStr, belowH := paneHeader(w, h)

	// Active version info bar (may be empty if no versions are scanned yet).
	versionBar := RenderActiveVersionBar(m.installedVersions, w)
	versionBarH := lipgloss.Height(versionBar)

	rawTable := RenderTable(m)
	tableW := lipgloss.Width(rawTable)
	tableStr := lipgloss.PlaceHorizontal(w, lipgloss.Center, rawTable)

	tableH := lipgloss.Height(tableStr)

	// Fixed area: version bar + blank line + table + blank line before log.
	fixedH := versionBarH + 1 + tableH + 1

	logPanelH := belowH - fixedH
	const minLogPanel = 5
	const maxLogPanel = 8
	if logPanelH < minLogPanel {
		logPanelH = minLogPanel
	}
	if logPanelH > maxLogPanel {
		logPanelH = maxLogPanel
	}

	logVisible := logPanelH - 4
	if logVisible < 1 {
		logVisible = 1
	}

	logInnerW := w - 20
	if logInnerW < tableW {
		logInnerW = tableW
	}

	logStr := lipgloss.PlaceHorizontal(w, lipgloss.Center,
		RenderLogPanel(m.logs, logVisible, logInnerW))

	// ── URL info modal (port → open browser) ─────────────────────────────
	if m.showURLModal {
		modalStr := lipgloss.PlaceHorizontal(w, lipgloss.Center,
			RenderURLModal(m.urlModalSvc, m.urlModalURL))
		body := lipgloss.JoinVertical(lipgloss.Left, versionBar, tableStr, "", modalStr)
		below := lipgloss.Place(w, belowH, lipgloss.Center, lipgloss.Top, body)
		return titleStr + "\n\n" + below
	}

	// ── dialog overlay replaces log panel when active ──────────────────────
	if m.showDialog {
		dlgTitle, dlgBody := dialogTitleBody(m)
		dialogStr := lipgloss.PlaceHorizontal(w, lipgloss.Center,
			RenderActionDialog(dlgTitle, dlgBody, m.dialogBtn))

		body := lipgloss.JoinVertical(lipgloss.Left,
			versionBar, tableStr, "", dialogStr)
		below := lipgloss.Place(w, belowH, lipgloss.Center, lipgloss.Top, body)
		return titleStr + "\n\n" + below
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		versionBar,
		tableStr,
		"",
		logStr,
	)

	below := lipgloss.Place(w, belowH, lipgloss.Center, lipgloss.Top, body)
	return titleStr + "\n\n" + below
}

// versionsMgmtPane renders the installed-versions management screen.
func versionsMgmtPane(m Model, w, h int) string {
	titleStr, belowH := paneHeader(w, h)

	heading := lipgloss.NewStyle().Foreground(colorMuted).Bold(true).
		Render("Installed XAMPP Versions")

	pathNote := lipgloss.NewStyle().Foreground(colorBorder).Italic(true).
		Render("Switching updates /opt/lampp symlink only — no shell or PATH config is modified.")

	table := RenderInstalledVersionsTable(m)

	var body string
	if m.showDialog {
		dlgTitle, dlgBody := dialogTitleBody(m)
		dlg := lipgloss.PlaceHorizontal(w, lipgloss.Center,
			RenderActionDialog(dlgTitle, dlgBody, m.dialogBtn))
		body = lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.PlaceHorizontal(w, lipgloss.Center, heading),
			"",
			lipgloss.PlaceHorizontal(w, lipgloss.Center, table),
			"",
			dlg,
		)
	} else {
		body = lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.PlaceHorizontal(w, lipgloss.Center, heading),
			lipgloss.PlaceHorizontal(w, lipgloss.Center, pathNote),
			"",
			lipgloss.PlaceHorizontal(w, lipgloss.Center, table),
		)
	}

	below := lipgloss.Place(w, belowH, lipgloss.Center, lipgloss.Top, body)
	return titleStr + "\n\n" + below
}

// installPane renders either the welcome/install-options screen or the version
// picker, depending on m.installing.
func installPane(m Model, w, h int) string {
	titleStr, belowH := paneHeader(w, h)

	var content string
	if m.installing {
		content = versionPickerContent(m)
	} else {
		content = welcomeContent(m)
	}

	below := lipgloss.Place(w, belowH, lipgloss.Center, lipgloss.Center, content)
	return titleStr + "\n\n" + below
}

// downloadPane renders the active download progress screen.
func downloadPane(m Model, w, h int) string {
	titleStr, belowH := paneHeader(w, h)

	barWidth := 50
	bar := RenderProgressBar(m.downloadProgress, barWidth)

	label := lipgloss.NewStyle().Foreground(colorText).Bold(true).
		Render(fmt.Sprintf("Downloading XAMPP %s…", m.downloadVersion))

	var statusLine string
	switch {
	case m.downloadError != "":
		statusLine = lipgloss.NewStyle().Foreground(colorError).
			Render("Error: " + m.downloadError)
	case m.downloadProgress >= 1.0:
		statusLine = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).
			Render("Download complete!")
	default:
		statusLine = lipgloss.NewStyle().Foreground(colorMuted).Render("Please wait…")
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		label, "", bar, "", statusLine,
	)

	below := lipgloss.Place(w, belowH, lipgloss.Center, lipgloss.Center,
		lipgloss.PlaceHorizontal(w, lipgloss.Center, content))
	return titleStr + "\n\n" + below
}

// postDownloadPane renders the "install now?" prompt after a download completes.
func postDownloadPane(m Model, w, h int) string {
	titleStr, belowH := paneHeader(w, h)

	label := lipgloss.NewStyle().Foreground(colorText).Bold(true).
		Render(fmt.Sprintf("Download complete: XAMPP %s", m.downloadVersion))

	sublabel := lipgloss.NewStyle().Foreground(colorMuted).
		Render("Would you like to install it now?")

	destination := lipgloss.NewStyle().Foreground(colorMuted).
		Render(fmt.Sprintf("Target: /opt/xampp/%s/", m.downloadVersion))

	btn := lipgloss.NewStyle().
		Padding(0, 2).Margin(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Foreground(colorMuted)

	btnActive := lipgloss.NewStyle().
		Padding(0, 2).Margin(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorTitle).
		Foreground(colorPanelBg).
		Background(colorTitle).
		Bold(true)

	installBtn := btn.Render("  Install Now  ")
	skipBtn := btn.Render("  Skip  ")
	if m.postDownloadBtn == 0 {
		installBtn = btnActive.Render("  Install Now  ")
	} else {
		skipBtn = btnActive.Render("  Skip  ")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, installBtn, skipBtn)

	content := lipgloss.JoinVertical(lipgloss.Center,
		label, "", sublabel, destination, "", buttons,
	)

	below := lipgloss.Place(w, belowH, lipgloss.Center, lipgloss.Center,
		lipgloss.PlaceHorizontal(w, lipgloss.Center, content))
	return titleStr + "\n\n" + below
}

// installerPane renders the progress screen while the XAMPP installer runs.
func installerPane(m Model, w, h int) string {
	titleStr, belowH := paneHeader(w, h)

	label := lipgloss.NewStyle().Foreground(colorText).Bold(true).
		Render(fmt.Sprintf("Installing XAMPP %s…", m.downloadVersion))

	var statusLine string
	if m.installerError != "" {
		statusLine = lipgloss.NewStyle().Foreground(colorError).
			Render("Error: " + m.installerError)
	} else {
		statusLine = lipgloss.NewStyle().Foreground(colorMuted).Render(m.installerStatus)
	}

	notice := lipgloss.NewStyle().Foreground(colorTitle).Bold(true).
		Render("Please wait, this may take a few minutes…")

	content := lipgloss.JoinVertical(lipgloss.Center,
		label, "", statusLine, "", notice,
	)

	below := lipgloss.Place(w, belowH, lipgloss.Center, lipgloss.Center,
		lipgloss.PlaceHorizontal(w, lipgloss.Center, content))
	return titleStr + "\n\n" + below
}

// ─── screen content builders ─────────────────────────────────────────────────

func versionPickerContent(m Model) string {
	numCols := 4
	n := len(m.xamppVersions)
	numRows := (n + numCols - 1) / numCols
	selectedIdx := m.cursorVersionRow + m.cursorVersionCol*numRows

	var names []string
	for _, v := range m.xamppVersions {
		names = append(names, v.Name)
	}

	heading := lipgloss.NewStyle().Foreground(colorMuted).Bold(true).
		Render("Select a XAMPP version to download:")

	table := RenderVersionTable(versionTableData{
		Versions:        names,
		SelectedVersion: selectedIdx,
	})

	parts := []string{heading, "", table}

	if m.showVersionInfoPanel {
		url := "(select a valid version)"
		if n > 0 && selectedIdx < n {
			url = m.xamppVersions[selectedIdx].DownloadURL
		}
		parts = append(parts, "", RenderVersionInfoPanel(url, m.cursorVersionButton))
	}

	return lipgloss.JoinVertical(lipgloss.Center, parts...)
}

func welcomeContent(m Model) string {
	selected := map[int]struct{}{m.cursorInstall: {}}
	optionsList := RenderList(m.optionsInstallation, m.cursorInstall, selected)

	text := lipgloss.NewStyle().Foreground(colorText).Bold(true).Render("Welcome to XAMPP-TUI.") + "\n" +
		lipgloss.NewStyle().Foreground(colorMuted).Render("XAMPP is not installed on your system.") + "\n\n" +
		lipgloss.NewStyle().Foreground(colorMuted).Render("Options:") + "\n" +
		optionsList

	return lipgloss.NewStyle().Align(lipgloss.Left).Render(text)
}

// ─── dialog helpers ───────────────────────────────────────────────────────────

// dialogTitleBody returns the title and body text for the active dialog.
func dialogTitleBody(m Model) (title, body string) {
	svc := ""
	if m.dialogRow < len(m.choices) {
		svc = m.choices[m.dialogRow]
	}

	switch m.dialogType {
	case "kill":
		return fmt.Sprintf("Kill %s process?", svc),
			fmt.Sprintf("PID %d will receive SIGTERM.", m.pids[m.dialogRow])
	case "config":
		return fmt.Sprintf("Edit %s config?", svc),
			m.configPaths[m.dialogRow]
	case "switch_version":
		if m.dialogRow < len(m.installedVersions) {
			ver := m.installedVersions[m.dialogRow]
			return fmt.Sprintf("Switch to XAMPP %s?", ver.Version),
				fmt.Sprintf(
					"Only /opt/lampp symlink is updated.\n"+
						"No PATH or shell config is modified.\n\n"+
						"Stop running services before switching,\n"+
						"then restart them with the new version.\n\n"+
						"Path: %s", ver.Path)
		}
	}
	return "Confirm?", ""
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func terminalSize() (width, height int) {
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		return w, h
	}
	return 80, 24
}

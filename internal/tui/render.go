package tui

import (
	"fmt"
	"strings"
	"xampp-tui/internal/xampp"

	"github.com/charmbracelet/lipgloss"
)

// ─── title & footer ──────────────────────────────────────────────────────────

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

func title() string {
	return lipgloss.NewStyle().
		Foreground(colorTitle).
		Bold(true).
		Align(lipgloss.Center).
		Render(BannerTitleL)
}

func footer() string {
	sep := lipgloss.NewStyle().Foreground(colorBorder).Render("│")
	key := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	desc := lipgloss.NewStyle().Foreground(colorMuted)

	hint := func(keys, action string) string {
		return key.Render(keys) + desc.Render(" "+action)
	}

	line1 := strings.Join([]string{
		hint("↑↓←→ / wasd", "Navigate"),
		hint("Enter / Space", "Action"),
		hint("q", "Quit"),
	}, "  "+sep+"  ")

	return lipgloss.NewStyle().Foreground(colorMuted).Render(line1)
}

func RenderTitle(width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, title())
}

func RenderFooter(width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, footer())
}

// ─── action buttons row ──────────────────────────────────────────────────────

func RenderOptions(width int) string {
	btn := lipgloss.NewStyle().
		Foreground(colorMuted).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 2).
		Margin(0, 1)

	actions := []struct{ key, label string }{
		{"e", "Start"},
		{"x", "Stop"},
		{"r", "Restart"},
	}

	keyStyle := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	var parts []string
	for _, a := range actions {
		parts = append(parts, btn.Render(keyStyle.Render(a.key)+" "+a.label))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Center, parts...)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, row)
}

// TextArea renders content in the log/action area style.
func TextArea(content string) string {
	return lipgloss.NewStyle().
		Foreground(colorText).
		Padding(1).
		Align(lipgloss.Left).
		Render(content)
}

// ─── active version bar ───────────────────────────────────────────────────────

// RenderActiveVersionBar renders a subtle one-line bar showing the active
// XAMPP version's PHP and MySQL versions. Returns an empty string if no active
// version is found.
func RenderActiveVersionBar(versions []xampp.InstalledVersion, width int) string {
	for _, v := range versions {
		if !v.IsActive {
			continue
		}
		dot := lipgloss.NewStyle().Foreground(colorSuccess).Render("●")
		label := lipgloss.NewStyle().Foreground(colorMuted).
			Render(fmt.Sprintf(" XAMPP %s   PHP %s   MySQL %s",
				v.Version, v.PHPVersion, v.MySQLVersion))
		line := dot + label
		return lipgloss.PlaceHorizontal(width, lipgloss.Center, line)
	}
	return ""
}

// ─── service table ───────────────────────────────────────────────────────────

const columnWidth = 18

func RenderTable(m Model) string {
	col := func() lipgloss.Style {
		return lipgloss.NewStyle().
			Width(columnWidth).
			Align(lipgloss.Center).
			Foreground(colorText)
	}
	highlight := lipgloss.NewStyle().
		Foreground(colorHighlightFg).
		Background(colorHighlightBg).
		Bold(true).
		Width(columnWidth).
		Align(lipgloss.Center)

	headerStyle := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Center).
		Foreground(colorTitle).
		Bold(true).
		Underline(true)

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		headerStyle.MarginBottom(1).Render("Service"),
		headerStyle.MarginBottom(1).Render("PID"),
		headerStyle.MarginBottom(1).Render("Port"),
		headerStyle.MarginBottom(1).Render("Config"),
	)

	rows := make([]string, len(m.choices))
	for i, svc := range m.choices {
		running := m.isRunning(i)

		dot := "○"
		statusColor := colorError
		if running {
			dot = "●"
			statusColor = colorSuccess
		}
		label := dot + " " + svc

		svcCell := col().Foreground(statusColor).Render(truncateOrPad(label, columnWidth))
		if m.cursorRow == i && m.cursorCol == 0 {
			svcCell = highlight.Render(truncateOrPad(label, columnWidth))
		}

		pidStr := truncateOrPad(fmt.Sprintf("%d", m.pids[i]), columnWidth)
		portStr := truncateOrPad(m.ports[i], columnWidth)
		cfgStr := truncateOrPad(m.config[i], columnWidth)

		var pidCell, portCell, cfgCell string
		if running {
			pidCell = col().Render(pidStr)
			portCell = col().Render(portStr)
			cfgCell = col().Render(cfgStr)
		} else {
			pidCell = col().Foreground(colorBorder).Render(truncateOrPad("—", columnWidth))
			portCell = col().Foreground(colorBorder).Render(truncateOrPad("—", columnWidth))
			cfgCell = col().Render(truncateOrPad(m.config[i], columnWidth))
		}

		if m.cursorRow == i && m.cursorCol == 1 {
			pidCell = highlight.Render(pidStr)
		}
		if m.cursorRow == i && m.cursorCol == 2 {
			portCell = highlight.Render(portStr)
		}
		if m.cursorRow == i && m.cursorCol == 3 {
			cfgCell = highlight.Render(cfgStr)
		}

		rows[i] = lipgloss.JoinHorizontal(lipgloss.Top, svcCell, pidCell, portCell, cfgCell)
	}

	tableContent := lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 2)

	return box.Render(tableContent)
}

// ─── installed versions table ─────────────────────────────────────────────────

const (
	verColW    = 12
	phpColW    = 12
	mysqlColW  = 12
	pathColW   = 28
	statusColW = 12
)

// RenderInstalledVersionsTable renders a table of installed XAMPP versions with
// PHP/MySQL info and active status indicator.
func RenderInstalledVersionsTable(m Model) string {
	col := func(w int) lipgloss.Style {
		return lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Foreground(colorText)
	}
	hl := func(w int) lipgloss.Style {
		return lipgloss.NewStyle().
			Width(w).Align(lipgloss.Center).
			Foreground(colorHighlightFg).Background(colorHighlightBg).Bold(true)
	}
	hdr := func(w int, label string) string {
		return lipgloss.NewStyle().
			Width(w).Align(lipgloss.Center).
			Foreground(colorTitle).Bold(true).Underline(true).
			Render(label)
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		hdr(verColW, "Version"),
		hdr(phpColW, "PHP"),
		hdr(mysqlColW, "MySQL"),
		hdr(pathColW, "Path"),
		hdr(statusColW, "Status"),
	)

	if len(m.installedVersions) == 0 {
		empty := lipgloss.NewStyle().Foreground(colorMuted).Padding(1, 0).
			Render("No XAMPP versions found. Press i to install one.")
		content := lipgloss.JoinVertical(lipgloss.Left, header, empty)
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 2).
			Render(content)
	}

	rows := make([]string, len(m.installedVersions))
	for i, v := range m.installedVersions {
		isHL := m.cursorVersionsMgmt == i

		statusStr := "○ Inactive"
		statusColor := colorMuted
		if v.IsActive {
			statusStr = "● Active"
			statusColor = colorSuccess
		}

		if isHL {
			rows[i] = lipgloss.JoinHorizontal(lipgloss.Top,
				hl(verColW).Render(truncateOrPad(v.Version, verColW)),
				hl(phpColW).Render(truncateOrPad(v.PHPVersion, phpColW)),
				hl(mysqlColW).Render(truncateOrPad(v.MySQLVersion, mysqlColW)),
				hl(pathColW).Render(truncateOrPad(v.Path, pathColW)),
				hl(statusColW).Render(truncateOrPad(statusStr, statusColW)),
			)
		} else {
			rows[i] = lipgloss.JoinHorizontal(lipgloss.Top,
				col(verColW).Render(truncateOrPad(v.Version, verColW)),
				col(phpColW).Render(truncateOrPad(v.PHPVersion, phpColW)),
				col(mysqlColW).Render(truncateOrPad(v.MySQLVersion, mysqlColW)),
				col(pathColW).Render(truncateOrPad(v.Path, pathColW)),
				col(statusColW).Foreground(statusColor).Render(truncateOrPad(statusStr, statusColW)),
			)
		}
	}

	tableContent := lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 2).
		Render(tableContent)
}

// ─── version table (download picker) ─────────────────────────────────────────

type versionTableData struct {
	Versions           []string
	SelectedVersion    int
	DownloadedVersions []string // versions already downloaded (show ⬇ indicator)
}

func RenderVersionTable(d versionTableData) string {
	const numCols = 4
	const colW = 20

	downloadedSet := make(map[string]bool, len(d.DownloadedVersions))
	for _, v := range d.DownloadedVersions {
		downloadedSet[v] = true
	}

	col := func() lipgloss.Style {
		return lipgloss.NewStyle().
			Width(colW).
			Align(lipgloss.Center).
			Foreground(colorText)
	}
	hl := lipgloss.NewStyle().
		Foreground(colorHighlightFg).
		Background(colorHighlightBg).
		Bold(true).
		Width(colW).
		Align(lipgloss.Center)
	dlStyle := lipgloss.NewStyle().Foreground(colorSuccess)

	n := len(d.Versions)
	numRows := (n + numCols - 1) / numCols
	cells := make([][]string, numRows)
	for i := range cells {
		cells[i] = make([]string, numCols)
		for j := 0; j < numCols; j++ {
			idx := i + j*numRows
			if idx < n {
				ver := d.Versions[idx]
				if ver == "" {
					ver = "-"
				}
				label := ver
				if downloadedSet[ver] {
					label = ver + " ⬇"
				}
				if idx == d.SelectedVersion {
					if downloadedSet[ver] {
						// Highlight with green tint to show it's ready
						cells[i][j] = lipgloss.NewStyle().
							Foreground(colorHighlightFg).
							Background(colorHighlightBg).
							Bold(true).
							Width(colW).
							Align(lipgloss.Center).
							Render(label)
					} else {
						cells[i][j] = hl.Render(label)
					}
				} else {
					if downloadedSet[ver] {
						cells[i][j] = col().Foreground(colorSuccess).Render(truncateOrPad(label, colW))
					} else {
						cells[i][j] = col().Render(ver)
					}
				}
			} else {
				cells[i][j] = col().Render("")
			}
		}
	}

	rowStrings := make([]string, numRows)
	for i := range cells {
		rowStrings[i] = lipgloss.JoinHorizontal(lipgloss.Top, cells[i]...)
	}

	legend := dlStyle.Render("⬇") + lipgloss.NewStyle().Foreground(colorMuted).Render(" = ready to install")
	return lipgloss.JoinVertical(lipgloss.Left, rowStrings...) + "\n\n" + legend
}

// ─── version info panel ──────────────────────────────────────────────────────

func RenderVersionInfoPanel(downloadURL string, selectedButton int, alreadyDownloaded bool) string {
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorTitle).
		Padding(0, 2).
		Background(colorPanelBg).
		Foreground(colorPanelFg)

	label := lipgloss.NewStyle().Foreground(colorPanelFg).Bold(true)
	value := lipgloss.NewStyle().Foreground(colorTitle)

	var info string
	if alreadyDownloaded {
		info = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).Render("⬇ Already downloaded — ready to install") + "\n" +
			label.Render("Destination:  ") + value.Render("./downloads/")
	} else {
		info = label.Render("Download URL: ") + value.Render(downloadURL) + "\n" +
			label.Render("Destination:  ") + value.Render("./downloads/")
	}

	btn := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Foreground(colorMuted)

	btnActive := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorTitle).
		Foreground(colorPanelBg).
		Background(colorTitle).
		Bold(true)

	actionLabel := "  Download  "
	if alreadyDownloaded {
		actionLabel = "  Install Now  "
	}

	actionBtn := btn.Render(actionLabel)
	cancelBtn := btn.Render("  Cancel  ")
	if selectedButton == 0 {
		actionBtn = btnActive.Render(actionLabel)
	} else {
		cancelBtn = btnActive.Render("  Cancel  ")
	}

	buttons := lipgloss.PlaceHorizontal(
		lipgloss.Width(info),
		lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Top, actionBtn, cancelBtn),
	)

	return panel.Render(info + "\n\n" + buttons)
}

// ─── PATH notice banner ───────────────────────────────────────────────────────

// renderPathNotice renders a dismissible banner shown after /opt/lampp/bin is
// automatically added to the user's shell config during installation.
func renderPathNotice(notice string, _ int) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorSuccess).
		Padding(0, 3).
		Foreground(colorSuccess).
		Bold(true).
		Render("PATH updated\n\n" +
			lipgloss.NewStyle().Foreground(colorText).Bold(false).Render(notice) +
			"\n\n" +
			lipgloss.NewStyle().Foreground(colorMuted).Render("Press any key to dismiss"))
}

// ─── URL info modal ───────────────────────────────────────────────────────────

// RenderURLModal renders a minimal overlay showing the service URL after the
// user presses Enter on the Port column. Press any key to dismiss.
func RenderURLModal(svc, url string) string {
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorTitle).
		Padding(1, 4).
		Background(colorPanelBg).
		Foreground(colorPanelFg)

	svcStyle := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	urlStyle := lipgloss.NewStyle().Foreground(colorText).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(colorMuted)

	content := lipgloss.JoinVertical(lipgloss.Center,
		svcStyle.Render(svc),
		"",
		urlStyle.Render(url),
		"",
		hintStyle.Render("Opening in browser…   any key to close"),
	)

	return panel.Render(content)
}

// ─── action dialog ───────────────────────────────────────────────────────────

func RenderActionDialog(title, body string, selectedBtn int) string {
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorTitle).
		Padding(1, 3).
		Background(colorPanelBg).
		Foreground(colorPanelFg)

	titleStyle := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	bodyStyle := lipgloss.NewStyle().Foreground(colorPanelFg)

	btn := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Foreground(colorMuted)

	btnActive := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorTitle).
		Foreground(colorPanelBg).
		Background(colorTitle).
		Bold(true)

	yesBtn := btn.Render("  Yes  ")
	noBtn := btn.Render("  No   ")
	if selectedBtn == 0 {
		yesBtn = btnActive.Render("  Yes  ")
	} else {
		noBtn = btnActive.Render("  No   ")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, yesBtn, noBtn)
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		"",
		bodyStyle.Render(body),
		"",
		buttons,
	)
	return panel.Render(content)
}

// ─── list ────────────────────────────────────────────────────────────────────

func RenderList(options []string, cursor int, selected map[int]struct{}) string {
	activeStyle := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(colorText)
	checkActive := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	checkNormal := lipgloss.NewStyle().Foreground(colorMuted)

	var s string
	for i, choice := range options {
		cur := " "
		if cursor == i {
			cur = "▶"
		}
		check := " "
		if selected != nil {
			if _, ok := selected[i]; ok {
				check = "x"
			}
		}

		bracket := fmt.Sprintf("[%s]", check)

		if cursor == i {
			s += activeStyle.Render(cur) + " " +
				checkActive.Render(fmt.Sprintf("[%s]", check)) + " " +
				activeStyle.Render(choice) + "\n"
		} else {
			s += normalStyle.Render(" ") + " " +
				checkNormal.Render(bracket) + " " +
				normalStyle.Render(choice) + "\n"
		}
	}
	return s
}

// ─── log panel ───────────────────────────────────────────────────────────────

func RenderLogPanel(logs []string, visibleLines, innerWidth int) string {
	if visibleLines < 1 {
		visibleLines = 1
	}
	if innerWidth < 10 {
		innerWidth = 10
	}

	visible := logs
	if len(visible) > visibleLines {
		visible = visible[len(visible)-visibleLines:]
	}

	tsStyle := lipgloss.NewStyle().Foreground(colorMuted)
	msgStyle := lipgloss.NewStyle().Foreground(colorText)
	emptyStyle := lipgloss.NewStyle().Foreground(colorBorder)

	lines := make([]string, visibleLines)
	offset := visibleLines - len(visible)
	for i := 0; i < offset; i++ {
		lines[i] = emptyStyle.Render("—")
	}
	for i, entry := range visible {
		parts := strings.SplitN(entry, "  ", 2)
		if len(parts) == 2 {
			lines[offset+i] = tsStyle.Render(parts[0]) + "  " + msgStyle.Render(parts[1])
		} else {
			lines[offset+i] = msgStyle.Render(entry)
		}
	}

	content := strings.Join(lines, "\n")

	headerStyle := lipgloss.NewStyle().Foreground(colorTitle).Bold(true)
	sepStyle := lipgloss.NewStyle().Foreground(colorBorder)

	header := headerStyle.Render("Recent Activity")
	sep := sepStyle.Render(strings.Repeat("─", innerWidth))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1).
		Width(innerWidth)

	return box.Render(header + "\n" + sep + "\n" + content)
}

// ─── progress bar ────────────────────────────────────────────────────────────

func RenderProgressBar(pct float64, innerWidth int) string {
	if innerWidth < 4 {
		innerWidth = 4
	}
	filled := int(pct * float64(innerWidth))
	if filled > innerWidth {
		filled = innerWidth
	}

	filledStyle := lipgloss.NewStyle().Foreground(colorTitle)
	emptyStyle := lipgloss.NewStyle().Foreground(colorBorder)
	pctStyle := lipgloss.NewStyle().Foreground(colorText).Bold(true)

	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", innerWidth-filled))

	return bar + pctStyle.Render(fmt.Sprintf("  %.1f%%", pct*100))
}

// ─── string helpers ──────────────────────────────────────────────────────────

func truncateOrPad(s string, width int) string {
	runes := []rune(s)
	if len(runes) > width {
		return string(runes[:width])
	}
	pad := width - len(runes)
	left := pad / 2
	right := pad - left
	return fmt.Sprintf("%s%s%s", spaces(left), s, spaces(right))
}

func spaces(n int) string {
	return string(make([]rune, n))
}

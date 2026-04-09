# xampp-tui

A terminal user interface (TUI) for managing XAMPP/LAMPP on Linux — built with Go, Bubble Tea and Lipgloss.

Manage services, install multiple XAMPP versions, switch between them, and monitor your stack — all without leaving the terminal.

---

## Features

- **Service control** — Start, stop, and restart Apache, MySQL, and FTP individually or all at once
- **Live status** — PID, listening port, and running state refreshed every 5 seconds
- **Open in browser** — Press `Enter` on a port to open `http://localhost:{port}` directly
- **Edit config** — Open `httpd.conf`, `my.cnf`, or `proftpd.conf` in `nano` without leaving the TUI
- **Kill processes** — Send SIGTERM to any service process with a confirmation dialog
- **Multi-version XAMPP** — Install multiple XAMPP versions side by side under `/opt/xampp/{version}/`
- **Version switching** — Switch the active version by updating the `/opt/lampp` symlink (no PATH or shell config changes)
- **Version info** — Shows PHP and MySQL version for each installed XAMPP
- **Background downloads** — Send downloads or installations to the background with `q`/`Esc`, monitor progress in the corner
- **Recent activity log** — Shows the last entries from Apache's error log
- **Adaptive theme** — Automatically adjusts to dark or light terminal backgrounds

---

## Requirements

| Dependency | Purpose |
| --- | --- |
| Go 1.21+ | Build from source |
| `sudo` access | Control XAMPP services and manage `/opt/lampp` |
| `gawk` | Scrape XAMPP version list from SourceForge |
| `curl` | Fetch version list |
| `xdg-open` / `sensible-browser` | Open URLs in the system browser |
| `ss` | Detect service ports (`iproute2` package) |
| `nano` | Edit configuration files |

Install dependencies on Debian/Ubuntu:

```bash
sudo apt install gawk curl iproute2
```

On Arch/Manjaro:

```bash
sudo pacman -S gawk curl iproute2
```

---

## Installation

### Option A — Build from source

```bash
git clone https://github.com/ramirezDg/xampp-tui.git
cd xampp-tui
make install
```

This compiles an optimized binary and copies it to `/usr/local/bin/xampp-tui`.

### Option B — Download a pre-built release

Go to the [Releases page](https://github.com/ramirezDg/xampp-tui/releases), download the latest `xampp-tui-linux-amd64.tar.gz`, then:

```bash
tar xzf xampp-tui-linux-amd64.tar.gz
sudo install -m 755 xampp-tui /usr/local/bin/
```

### Option C — Run directly without installing

```bash
git clone https://github.com/ramirezDg/xampp-tui.git
cd xampp-tui
go run ./cmd/lampp-tui
```

---

## Usage

```bash
xampp-tui
```

If XAMPP is not installed, the tool will guide you through downloading and installing it.

---

## Keyboard Shortcuts

### Main panel

| Key | Action |
| --- | --- |
| `↑` `↓` `←` `→` / `wasd` | Navigate the service table |
| `Enter` / `Space` | Execute action for the selected cell |
| `e` | Start all services |
| `x` | Stop all services |
| `r` | Restart all services |
| `v` | Open installed versions panel |
| `i` | Install a new XAMPP version |
| `q` / `Ctrl+C` | Quit |

### Cell actions (press `Enter` on each column)

| Column | Action |
| --- | --- |
| **Service** | Toggle start / stop |
| **PID** | Confirmation dialog → kill process (SIGTERM) |
| **Port** | Open `http://localhost:{port}` in browser + show URL modal |
| **Config** | Confirmation dialog → open config file in `nano` |

### Download / Install screens

| Key | Action |
| --- | --- |
| `q` / `Esc` | Send to background (continues running) |
| `Ctrl+C` | Quit the application |

A small `⟳ DL 67%` or `⟳ Installing…` indicator appears in the bottom-right corner when a task is running in the background.

### Versions panel (`v`)

| Key | Action |
| --- | --- |
| `↑` `↓` | Navigate installed versions |
| `Enter` | Switch to selected version (confirmation required) |
| `q` / `Esc` | Back to main panel |

---

## Multi-version XAMPP

xampp-tui supports installing and running multiple XAMPP versions side by side.

### How it works

Each version is installed to its own directory:

```
/opt/xampp/
  8.2.12/    ← XAMPP 8.2.12 (PHP 8.2, MySQL 8.0)
  8.1.6/     ← XAMPP 8.1.6  (PHP 8.1, MySQL 5.7)
```

The **active version** is determined by the `/opt/lampp` symlink:

```
/opt/lampp  →  /opt/xampp/8.2.12/
```

Switching versions updates only this symlink — **no shell config, PATH, or `.zshrc` is ever modified**.

### Installing a new version

1. Press `i` from the main panel
2. Select a version from the grid (fetched from SourceForge)
3. Confirm the download
4. When complete, choose **Install Now** or **Skip**
5. The installer runs unattended; when done, `/opt/lampp` is updated automatically

### Switching versions

1. Press `v` from the main panel
2. Navigate to the desired version
3. Press `Enter` and confirm
4. Manually stop current services and restart with the new version

### PATH behaviour

If `/opt/lampp/bin` is **not** in your PATH (the default), switching versions has **zero effect** on your shell environment.

If you add it to your PATH:

```bash
# In ~/.zshrc or ~/.bashrc
export PATH="/opt/lampp/bin:$PATH"
```

Then `php`, `mysql`, etc. will always point to the currently active XAMPP version — which is the intended behaviour for version management.

---

## Project structure

```
xampp-tui/
├── cmd/lampp-tui/
│   ├── main.go            # Entry point
│   └── downloads/         # Downloaded .run installers
├── internal/
│   ├── tui/
│   │   ├── model.go       # Application state (Bubble Tea Model)
│   │   ├── update.go      # Event handling and keyboard input
│   │   ├── view.go        # Screen routing and layout
│   │   ├── render.go      # Component rendering
│   │   └── styles.go      # Adaptive colour palette
│   ├── xampp/
│   │   ├── service.go     # Service control and status (start/stop/PID/port)
│   │   ├── multiver.go    # Multi-version scanning and switching
│   │   ├── logs.go        # Apache error log parser
│   │   └── validator.go   # XAMPP installation detection
│   ├── installer/
│   │   ├── downloader.go  # HTTP download with progress callback
│   │   ├── runner.go      # BitRock .run installer execution
│   │   └── versions.go    # SourceForge version scraper
│   └── logger/
│       └── logger.go      # Append-only file logger
├── Makefile
├── install.sh
└── README.md
```

---

## Building

```bash
# Development build
make build

# Optimised release binary (stripped, no debug info)
make release

# Install to /usr/local/bin
make install

# Clean build artefacts
make clean
```

The release target produces:

- `xampp-tui-linux-amd64.tar.gz` — binary + install script
- `xampp-tui-linux-amd64.tar.gz.sha256` — checksum

---

## Publishing a GitHub Release

1. Tag the commit:

   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. The GitHub Actions workflow (`.github/workflows/release.yml`) will automatically:
   - Build an optimised binary
   - Create a release tarball with checksum
   - Publish it to the GitHub Releases page

---

## Sudo configuration

xampp-tui uses `sudo` to control XAMPP services (`/opt/lampp/lampp`) and manage the `/opt/lampp` symlink. To avoid password prompts, add a sudoers rule:

```bash
sudo visudo
```

Add (replace `youruser` with your username):

```text
youruser ALL=(ALL) NOPASSWD: /opt/lampp/lampp, /bin/ln, /usr/bin/ln
```

---

## License

MIT — see [LICENSE](LICENSE).

---

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

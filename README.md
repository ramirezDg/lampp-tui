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
- **Version switching** — Switch the active version by updating the `/opt/lampp` symlink
- **Version uninstall** — Remove an installed version directly from the TUI
- **Auto PATH setup** — Adds `/opt/lampp/bin` to your shell config after installation so `php` and `mysql` always reflect the active version
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
| `curl` | Fetch version list and check for versions |
| `xdg-open` / `sensible-browser` | Open URLs in the system browser |
| `ss` | Detect service ports (`iproute2` package) |
| `nano` | Edit configuration files |

Install dependencies on Debian/Ubuntu/Pop!_OS:

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
git clone <repository-url>
cd xampp-tui
make install
```

This compiles an optimised binary and copies it to `/usr/local/bin/xampp-tui`.

### Option B — Download a pre-built release

Download the latest `xampp-tui-linux-amd64.tar.gz` from the Releases page, then:

```bash
tar xzf xampp-tui-linux-amd64.tar.gz
sudo install -m 755 xampp-tui /usr/local/bin/
```

### Option C — Run without installing

```bash
git clone <repository-url>
cd xampp-tui
go run ./cmd/lampp-tui
```

---

## Usage

```bash
xampp-tui             # launch the TUI
xampp-tui --version   # print version and exit
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
| `d` | Uninstall selected version (confirmation required) |
| `q` / `Esc` | Back to main panel |

> Note: the active version cannot be switched to itself or uninstalled. Switch to a different version first.

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

```text
/opt/lampp  →  /opt/xampp/8.2.12/
```

Switching versions updates only this symlink — no other files are modified.

### Installing a new version

1. Press `i` from the main panel
2. Select a version from the grid (fetched from SourceForge)
   - Versions marked with `⬇` are already downloaded and ready to install
3. Confirm the download (or press Install Now if already downloaded)
4. When complete, choose **Install Now** or **Skip**
5. The installer runs unattended; when done, `/opt/lampp` is updated automatically

### Switching versions

1. Press `v` from the main panel
2. Navigate to the desired version
3. Press `Enter` and confirm
4. Stop the current services and restart them with the new version

### Uninstalling a version

1. Press `v` from the main panel
2. Navigate to the version to remove
3. Press `d` and confirm
4. The directory `/opt/xampp/{version}/` is permanently deleted

### PATH setup

After a successful installation, xampp-tui automatically adds `/opt/lampp/bin` to your shell startup file (`~/.zshrc`, `~/.bashrc`, or `~/.profile` depending on your shell). This means `php`, `mysql`, and other XAMPP binaries will always point to the currently active version.

To apply the change in your current terminal session:

```bash
source ~/.zshrc   # or ~/.bashrc
```

---

## File locations

| Path | Purpose |
| --- | --- |
| `~/.local/share/xampp-tui/downloads/` | Downloaded XAMPP installers |
| `~/.local/share/xampp-tui/logs/` | Application log |
| `/opt/xampp/{version}/` | Installed XAMPP versions |
| `/opt/lampp` | Symlink to the active XAMPP version |

---

## Project structure

```text
xampp-tui/
├── cmd/lampp-tui/
│   └── main.go              # Entry point (--version flag)
├── internal/
│   ├── tui/
│   │   ├── model.go         # Application state (Bubble Tea Model)
│   │   ├── update.go        # Event handling and keyboard input
│   │   ├── view.go          # Screen routing and layout
│   │   ├── render.go        # Component rendering
│   │   └── styles.go        # Adaptive colour palette
│   ├── xampp/
│   │   ├── service.go       # Service control and status (start/stop/PID/port)
│   │   ├── multiver.go      # Multi-version scanning, switching, and uninstall
│   │   ├── shell.go         # Shell config detection and PATH setup
│   │   ├── logs.go          # Apache error log parser
│   │   └── validator.go     # XAMPP installation detection
│   ├── installer/
│   │   ├── downloader.go    # HTTP download with progress callback and validation
│   │   ├── runner.go        # BitRock .run installer execution
│   │   └── versions.go      # SourceForge version scraper
│   └── logger/
│       └── logger.go        # Append-only file logger (XDG-compliant path)
├── Makefile
└── install.sh
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

---

## Sudo configuration

xampp-tui uses `sudo` to control XAMPP services and manage the `/opt/lampp` symlink. To avoid password prompts, add a sudoers rule:

```bash
sudo visudo
```

Add (replace `youruser` with your actual username):

```text
youruser ALL=(ALL) NOPASSWD: /opt/lampp/lampp, /bin/ln, /usr/bin/ln, /bin/rm, /usr/bin/rm
```

---

## License

MIT

# Changelog

All notable changes to xampp-tui are documented here.

## [Unreleased]

### Added
- Downloaded-but-not-installed versions appear as quick-install options in the welcome screen
- Version picker grid shows `⬇` indicator for already-downloaded versions (green tint)
- Version info panel shows "Install Now" button (skips download) for already-downloaded versions
- URL modal overlay when opening a port in the browser — shows the URL even if the browser fails
- Download and installer tasks can be sent to background with `q`/`Esc`; progress indicator shown in footer
- Multi-version XAMPP support: install multiple versions to `/opt/xampp/{version}/`
- Active version shown via `/opt/lampp` symlink — no shell or PATH config is modified
- Installed versions management panel (`v`) with PHP/MySQL version info per installation
- Post-download install prompt — offered immediately after a download completes
- `--version` / `-v` flag on the binary
- Makefile with `build`, `run`, `install`, `uninstall`, `release`, `clean` targets
- `install.sh` script for building and installing from source
- GitHub Actions release workflow with separate test/vet job and auto-generated release notes

### Changed
- Install menu options are built dynamically (downloaded versions first, then "Download new version", then "Quit")
- `q`/`Esc` in the version picker now returns to the install menu instead of quitting the app
- Start/Stop/Restart button row removed from the admin panel (keyboard shortcuts `e`/`x`/`r` remain)
- Browser opening uses `Setsid: true` to detach from TUI raw-mode terminal

### Fixed
- Port column `Enter` now opens the browser correctly inside Bubble Tea raw mode
- `q` in the version picker no longer quits the entire application

## [0.1.0] — Initial release

- TUI for managing XAMPP (Apache, MySQL, FTP) services on Linux
- Service status table with PID, port, and config file columns
- Log panel showing recent XAMPP activity
- Version downloader with progress bar
- Adaptive dark/light terminal theme detection

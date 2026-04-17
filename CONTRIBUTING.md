# Contributing to lampp-tui

Thank you for your interest in contributing to **lampp-tui** ❤️

**lampp-tui** is a keyboard-driven TUI for managing **XAMPP/LAMPP on Linux**.
It interacts with real system services, real processes, real files under `/opt`, and requires careful design to remain safe, predictable, and terminal-friendly.

Because of this, contributions must follow the architectural and UX principles of the project.

---

## 🌿 Active Development Branch: `multiplatform-support`

> ⚠️ Important for contributors

The branch **`multiplatform-support`** is where the project is currently evolving towards a **platform-agnostic architecture**.

This means:

* Decoupling Linux-specific logic from the core
* Abstracting XAMPP interactions behind interfaces
* Preparing the codebase to support **non-Linux environments** in the future
* Refactoring `internal/xampp` into platform adapters
* Making the TUI independent from OS assumptions

### What this implies for contributors

If your contribution touches:

* Service control
* File paths (`/opt/lampp`, `/opt/xampp`)
* Process detection
* Shell commands
* Installer logic

➡️ **You should target the `multiplatform-support` branch**, not `main`.

If your contribution is purely:

* TUI rendering
* Styles
* Keyboard UX
* Logger improvements
* Non-OS specific refactors

➡️ You may target `main`.

When in doubt, open an Issue first.

---

## 🚀 Getting Started

```bash
git clone https://github.com/ramirezDg/lampp-tui.git
cd lampp-tui
make build
./lampp-tui
```

You must test with:

* Linux
* A real XAMPP installation under `/opt/lampp`
* `sudo` properly configured
* Terminal size at **80x24**

---

## 🧠 Understand the Architecture First

Before changing anything, read the project structure:

```
internal/
 ├─ tui/        → Bubble Tea model, update loop, rendering, styles
 ├─ xampp/      → All system & service logic (start/stop/PID/ports/logs)
 ├─ installer/  → Download, SourceForge scraping, BitRock installer runner
 └─ logger/     → Append-only XDG logger
```
---

## 🖥️ What This Project Is (and Is Not)

This is:

* A terminal dashboard
* Keyboard-first
* Safe wrapper around `/opt/lampp`
* Process/status inspector
* Multi-version manager via symlinks

This is NOT:

* A GUI
* A mouse-driven app
* A systemd manager
* A config editor beyond opening files in `nano`

---

## 🐛 Reporting Bugs

Open an issue and include:

* Linux distro
* Terminal emulator
* XAMPP version
* Output of: `/opt/lampp/lampp status`
* Steps to reproduce
* Expected vs actual behavior

Screenshots of the terminal are extremely helpful.

---

## 💡 Feature Requests

Good feature requests:

* Improve visibility of service state
* Improve log reading
* Improve version management
* Improve keyboard UX
* Improve safety around destructive actions

Bad feature requests:

* Mouse support
* GUI ideas
* Non-terminal workflows
* Anything that breaks 80x24 compatibility

---

## 🛠️ Development Guidelines

### Tech Stack

* Go 1.21+
* Bubble Tea
* Lipgloss
* Pure shell interaction with the system (no daemons, no DBus, no systemd APIs)

---

### Code Rules

* Small files, small functions
* No global state
* No hidden side effects
* Explicit error handling
* No panic in normal flows
* All system calls must be centralized in `internal/xampp` or `internal/installer`

---

### TUI Rules (very important)

* Must render correctly at **80x24**
* No color assumptions (respect adaptive theme)
* All actions must be keyboard accessible
* Confirmation dialogs for destructive actions
* Never block the UI during downloads or installs (use background tasks)

---

## 🔐 Working with sudo and /opt

This app modifies:

* `/opt/lampp` symlink
* `/opt/xampp/{version}`
* Runs `/opt/lampp/lampp`
* Kills real processes

Be extremely careful.

Any PR touching these areas must be tested on a real machine.

---

## 🧪 Testing Your Changes

Before opening a PR, you must test:

* Start/stop/restart services
* Kill PID flow
* Open config in nano
* Install a version
* Switch versions
* Uninstall versions
* Background download flow (`q` / `Esc`)
* Terminal resized to 80x24

---

## 📝 Commit Convention

Use Conventional Commits:

```
feat: add apache access log viewer
fix: prevent crash when lampp symlink is broken
refactor: isolate port detection logic
docs: update README for multi-version flow
```

---

## 🔀 Pull Requests

1. Create a descriptive branch:

```
feat/log-improvements
fix/symlink-validation
refactor/tui-render
```

2. One logical change per PR
3. Explain **why** the change is needed
4. Include terminal screenshots if UI changed
5. Ensure `make build` succeeds

---

## 🚫 Reasons a PR Will Be Rejected

* Mixing TUI and system logic
* Breaking keyboard workflow
* Not working at 80x24
* Unsafe operations on `/opt`
* Over-engineering
* Adding unnecessary dependencies

---

## 🤝 Code of Conduct

Be respectful. Be constructive. Help keep the project clean and focused.

---

## ❓ Questions?

Open an Issue with the `question` label before starting large changes.

---

Thank you for helping make **lampp-tui** a serious terminal tool for Linux developers.

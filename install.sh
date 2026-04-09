#!/usr/bin/env bash
# install.sh — Build and install xampp-tui
set -euo pipefail

BINARY="xampp-tui"
INSTALL_DIR="/usr/local/bin"
MAIN="./cmd/lampp-tui"

# ── colours ──────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'
ORANGE='\033[0;33m'
RESET='\033[0m'

info()    { echo -e "${ORANGE}→${RESET} $*"; }
success() { echo -e "${GREEN}✓${RESET} $*"; }

# ── checks ───────────────────────────────────────────────────────────────────
if ! command -v go &>/dev/null; then
  echo "Error: Go is not installed. Install it from https://go.dev/dl/" >&2
  exit 1
fi

# ── build ────────────────────────────────────────────────────────────────────
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
info "Building xampp-tui ${VERSION}…"
go build -ldflags="-s -w -X main.version=${VERSION}" -o "${BINARY}" "${MAIN}"
success "Build complete"

# ── install ──────────────────────────────────────────────────────────────────
info "Installing to ${INSTALL_DIR}/${BINARY}…"
sudo install -m 755 "${BINARY}" "${INSTALL_DIR}/${BINARY}"
rm -f "${BINARY}"
success "Installed ${INSTALL_DIR}/${BINARY}"

# ── done ─────────────────────────────────────────────────────────────────────
echo ""
echo -e "  ${GREEN}xampp-tui is ready.${RESET}"
echo -e "  Run with: ${ORANGE}xampp-tui${RESET}"
echo ""
echo "  Tip: add your user to sudoers for passwordless service control:"
echo "  sudo visudo"
echo "  → youruser ALL=(ALL) NOPASSWD: /opt/lampp/lampp, /bin/ln, /usr/bin/ln"

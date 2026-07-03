#!/usr/bin/env bash
# Installs gee-bee binary + systemd timer.
# Tries to download a pre-built binary from GitHub Releases; falls back to
# building from source if Go is available.
# Run as root: sudo ./install.sh
set -euo pipefail

INSTALL_DIR=/usr/local/bin
CONFIG_DIR=/etc/gee-bee
DATA_DIR=/var/lib/gee-bee
UNIT_DIR=/etc/systemd/system
REPO=rorpage/gee-bee
SERVICE_USER=gee-bee
# When piped from curl, BASH_SOURCE[0] is unset; fall back to empty string.
SCRIPT_DIR=""
if [[ -n "${BASH_SOURCE[0]:-}" && -f "${BASH_SOURCE[0]}" ]]; then
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fi

if [ "$EUID" -ne 0 ]; then
  echo "Please run as root: sudo $0"
  exit 1
fi

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)        GOARCH=amd64 ;;
  aarch64|arm64) GOARCH=arm64 ;;
  *) GOARCH="" ;;
esac

BINARY_INSTALLED=false

# Try downloading a pre-built release binary
if [ -n "$GOARCH" ] && command -v curl &>/dev/null; then
  echo "Fetching latest release from GitHub..."
  LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null \
    | grep '"tag_name"' | cut -d'"' -f4 || true)

  if [ -n "$LATEST" ]; then
    BINARY_URL="https://github.com/$REPO/releases/download/$LATEST/gee-bee-linux-$GOARCH"
    echo "Downloading $LATEST binary for linux/$GOARCH..."
    if curl -fsSL "$BINARY_URL" -o "$INSTALL_DIR/gee-bee"; then
      chmod +x "$INSTALL_DIR/gee-bee"
      echo "Installed $LATEST to $INSTALL_DIR/gee-bee"
      BINARY_INSTALLED=true
    else
      echo "Download failed, will try building from source."
    fi
  else
    echo "No releases found, will try building from source."
  fi
fi

# Fall back to building from source
if [ "$BINARY_INSTALLED" = false ]; then
  if ! command -v go &>/dev/null; then
    echo ""
    echo "Go is not installed and no pre-built binary was found."
    echo "Options:"
    echo "  1. Install Go from https://go.dev/dl/ and re-run this script"
    echo "  2. Download a binary directly from https://github.com/$REPO/releases"
    echo "     and copy it to $INSTALL_DIR/gee-bee"
    exit 1
  fi

  echo "Building from source with Go..."
  REPO_ROOT=""
  if [[ -n "$SCRIPT_DIR" ]]; then
    REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
  fi
  if [[ -n "$REPO_ROOT" && -f "$REPO_ROOT/go.mod" ]]; then
    (cd "$REPO_ROOT" && CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/gee-bee" ./cmd/geebee)
  else
    BUILD_DIR=$(mktemp -d)
    trap "rm -rf $BUILD_DIR" EXIT
    git clone --depth 1 "https://github.com/$REPO.git" "$BUILD_DIR"
    (cd "$BUILD_DIR" && CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/gee-bee" ./cmd/geebee)
  fi
  echo "Built and installed to $INSTALL_DIR/gee-bee"
fi

# Create dedicated system user
if ! id "$SERVICE_USER" &>/dev/null; then
  useradd --system --no-create-home --shell /usr/sbin/nologin "$SERVICE_USER"
  echo "Created system user: $SERVICE_USER"
fi

# Create directories with appropriate ownership
mkdir -p "$CONFIG_DIR" "$DATA_DIR"
chown "root:$SERVICE_USER" "$CONFIG_DIR"
chown "$SERVICE_USER:$SERVICE_USER" "$DATA_DIR"
chmod 750 "$DATA_DIR" "$CONFIG_DIR"

RAW_BASE="https://raw.githubusercontent.com/$REPO/main/deploy/systemd"

# Install unit files (from local clone or downloaded from GitHub)
for UNIT_FILE in gee-bee.service gee-bee.timer; do
  if [[ -n "$SCRIPT_DIR" && -f "$SCRIPT_DIR/$UNIT_FILE" ]]; then
    install -m 644 "$SCRIPT_DIR/$UNIT_FILE" "$UNIT_DIR/"
  else
    echo "Downloading $UNIT_FILE from GitHub..."
    curl -fsSL "$RAW_BASE/$UNIT_FILE" -o "$UNIT_DIR/$UNIT_FILE"
    chmod 644 "$UNIT_DIR/$UNIT_FILE"
  fi
done

# Set up env file from example if not present
if [ ! -f "$CONFIG_DIR/env" ]; then
  if [[ -n "$SCRIPT_DIR" && -f "$SCRIPT_DIR/env.example" ]]; then
    install -m 640 "$SCRIPT_DIR/env.example" "$CONFIG_DIR/env"
  else
    echo "Downloading env.example from GitHub..."
    curl -fsSL "$RAW_BASE/env.example" -o "$CONFIG_DIR/env"
    chmod 640 "$CONFIG_DIR/env"
  fi
  chown "root:$SERVICE_USER" "$CONFIG_DIR/env"
  echo ""
  echo "  -> Edit $CONFIG_DIR/env with your tail numbers and webhook URL(s)."
fi

if [ ! -f "$CONFIG_DIR/config.yml" ]; then
  echo "  -> Or copy your config.yml to $CONFIG_DIR/config.yml instead."
fi

echo ""
systemctl daemon-reload
systemctl enable gee-bee.timer
echo "Timer enabled. Once your env file (and/or config.yml) is in place:"
echo "  sudo systemctl start gee-bee.timer"
echo ""
echo "Commands:"
echo "  systemctl status gee-bee.timer    -- next scheduled run"
echo "  systemctl start gee-bee.service   -- run immediately"
echo "  journalctl -u gee-bee -f          -- follow logs"

#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-latest}"
INSTALL_DIR="$HOME/Library/Application Support/LocalLaunch"
PLIST_DIR="$HOME/Library/LaunchAgents"
PLIST_NAME="com.locallaunch.plist"
BINARY_NAME="locallaunch"
REPO_URL="https://github.com/rhymn/locallaunch"

require_command() {
    if ! command -v "$1" >/dev/null 2>&1; then
        echo "Error: Required command not found: $1" >&2
        exit 1
    fi
}

resolve_version() {
    if [ "$VERSION" != "latest" ]; then
        echo "$VERSION"
        return
    fi

    local latest_url
    latest_url="$(curl -fsSIL -o /dev/null -w '%{url_effective}' "${REPO_URL}/releases/latest")"
    local latest_tag
    latest_tag="$(basename "$latest_url")"

    if [ -z "$latest_tag" ]; then
        echo "Error: Unable to resolve latest release version." >&2
        exit 1
    fi

    echo "${latest_tag#v}"
}

detect_arch() {
    case "$(uname -m)" in
        x86_64)  echo "amd64" ;;
        arm64)   echo "arm64" ;;
        *)       echo "unsupported" ;;
    esac
}

detect_os() {
    case "$(uname -s)" in
        Darwin) echo "darwin" ;;
        Linux)  echo "linux" ;;
        *)      echo "unsupported" ;;
    esac
}

OS=$(detect_os)
ARCH=$(detect_arch)

if [ "$OS" = "unsupported" ]; then
    echo "Error: Unsupported operating system"
    exit 1
fi

if [ "$ARCH" = "unsupported" ]; then
    echo "Error: Unsupported architecture"
    exit 1
fi

if [ "$OS" = "linux" ]; then
    INSTALL_DIR="$HOME/.local/share/locallaunch"
fi

require_command curl
INSTALL_VERSION="$(resolve_version)"

echo "Installing LocalLaunch v${INSTALL_VERSION} (${OS}/${ARCH})..."

mkdir -p "$INSTALL_DIR"

BINARY_URL="${REPO_URL}/releases/download/v${INSTALL_VERSION}/locallaunch-${OS}-${ARCH}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
TMP_BINARY="$TMP_DIR/$BINARY_NAME"

echo "Downloading from: ${BINARY_URL}"
curl -fL --retry 3 --retry-delay 2 --connect-timeout 15 "$BINARY_URL" -o "$TMP_BINARY"
install -m 0755 "$TMP_BINARY" "$INSTALL_DIR/$BINARY_NAME"

echo "Binary installed to: $INSTALL_DIR/$BINARY_NAME"

if [ "$OS" = "darwin" ]; then
    require_command launchctl
    mkdir -p "$PLIST_DIR"
    cat > "$PLIST_DIR/$PLIST_NAME" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.locallaunch</string>
    <key>ProgramArguments</key>
    <array>
        <string>${INSTALL_DIR}/${BINARY_NAME}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>${HOME}/Library/Logs/locallaunch.log</string>
    <key>StandardErrorPath</key>
    <string>${HOME}/Library/Logs/locallaunch.log</string>
</dict>
</plist>
EOF

    launchctl bootout "gui/$(id -u)/com.locallaunch" 2>/dev/null || true
    launchctl bootstrap "gui/$(id -u)" "$PLIST_DIR/$PLIST_NAME"
    launchctl enable "gui/$(id -u)/com.locallaunch"
    launchctl kickstart -k "gui/$(id -u)/com.locallaunch"
    echo "LaunchAgent installed and started."
elif [ "$OS" = "linux" ]; then
    require_command systemctl
    SERVICE_DIR="$HOME/.config/systemd/user"
    mkdir -p "$SERVICE_DIR"
    cat > "$SERVICE_DIR/locallaunch.service" <<EOF
[Unit]
Description=LocalLaunch Process Launcher
After=network.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/${BINARY_NAME}
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF

    systemctl --user daemon-reload
    systemctl --user enable --now locallaunch.service
    echo "Systemd user service installed and started."
fi

echo ""
echo "Installation complete!"
echo "Config: $INSTALL_DIR/config.json"
echo ""
echo "Usage:"
echo "  $INSTALL_DIR/$BINARY_NAME           # Start server"
echo "  $INSTALL_DIR/$BINARY_NAME token     # Show auth token"
echo "  $INSTALL_DIR/$BINARY_NAME version   # Show version"

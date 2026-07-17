#!/bin/bash
set -e

VERSION="${VERSION:-0.1.0}"
INSTALL_DIR="$HOME/Library/Application Support/LocalLaunch"
PLIST_DIR="$HOME/Library/LaunchAgents"
PLIST_NAME="com.locallaunch.plist"
BINARY_NAME="locallaunch"

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
    echo "For Linux, consider using the systemd service generator below."
fi

echo "Installing LocalLaunch v${VERSION} (${OS}/${ARCH})..."

mkdir -p "$INSTALL_DIR"

BINARY_URL="https://github.com/rhymn/locallaunch/releases/download/v${VERSION}/locallaunch-${OS}-${ARCH}"
echo "Downloading from: ${BINARY_URL}"
curl -fsSL "$BINARY_URL" -o "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo "Binary installed to: $INSTALL_DIR/$BINARY_NAME"

if [ "$OS" = "darwin" ]; then
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

    launchctl unload "$PLIST_DIR/$PLIST_NAME" 2>/dev/null || true
    launchctl load "$PLIST_DIR/$PLIST_NAME"
    echo "LaunchAgent installed and started."
elif [ "$OS" = "linux" ]; then
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
    systemctl --user enable locallaunch
    systemctl --user start locallaunch
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

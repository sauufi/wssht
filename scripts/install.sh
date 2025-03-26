#!/bin/bash

# Script to install go-websocket-ssh-proxy as a systemd service
# Run this script as root (sudo)

# Exit on any error
set -e

# Default installation paths
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
CONFIG_DIR="/etc/wssht"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root (use sudo)"
  exit 1
fi

# Go to the project root directory
cd "$(dirname "$0")/.."

# Build the binary if it doesn't exist
if [ ! -f "bin/wssht" ]; then
  echo "Binary not found, building..."
  ./scripts/build.sh
fi

# Create directories if they don't exist
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"

# Copy binary
echo "Installing binary to $INSTALL_DIR..."
cp -f bin/wssht "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/wssht"

# Create default config if it doesn't exist
if [ ! -f "$CONFIG_DIR/config" ]; then
  echo "Creating default configuration..."
  cat > "$CONFIG_DIR/config" << EOF
# Configuration for WSSHTunnel
# Override default settings

# Listening address and port
BIND_ADDR=0.0.0.0
BIND_PORT=80

# Default target host (used when X-Real-Host header is missing)
DEFAULT_HOST=127.0.0.1:143

# Set password (optional)
# PASSWORD=your_secure_password
EOF
fi

# Install systemd service
echo "Installing systemd service..."
cat > "$SERVICE_DIR/wssht.service" << EOF
[Unit]
Description=WSSHTunnel - WebSocket SSH Proxy
After=network.target

[Service]
Type=simple
User=root
EnvironmentFile=-/etc/wssht/config
ExecStart=/usr/local/bin/wssht -b \${BIND_ADDR} -p \${BIND_PORT} -t \${DEFAULT_HOST} \${PASSWORD:+-pass \$PASSWORD}
Restart=on-failure
RestartSec=5s
KillMode=process
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

# Enable and start the service
echo "Enabling and starting service..."
systemctl enable wssht.service
systemctl restart wssht.service

# Show status
echo "Service status:"
systemctl status wssht.service --no-pager

echo ""
echo "Installation completed successfully!"
echo "Configuration file: $CONFIG_DIR/config"
echo ""
echo "To start/stop the service:"
echo "  systemctl start wssht.service"
echo "  systemctl stop wssht.service"
echo ""
echo "To view logs:"
echo "  journalctl -u wssht.service -f"
echo ""
echo "To uninstall:"
echo "  systemctl stop wssht.service"
echo "  systemctl disable wssht.service"
echo "  rm $SERVICE_DIR/wssht.service"
echo "  rm $INSTALL_DIR/wssht"
echo "  rm -rf $CONFIG_DIR"
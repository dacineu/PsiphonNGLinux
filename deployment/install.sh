#!/usr/bin/env bash
# Install PsiphonNG as a user service with optional foreground client mode

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Installation paths
BIN_DIR="${HOME}/.local/bin"
CONFIG_DIR="${HOME}/.config/psiphond-ng"
SYSTEMD_DIR="${HOME}/.config/systemd/user"

echo "Installing PsiphonNG..."

# Create directories
mkdir -p "$BIN_DIR" "$CONFIG_DIR" "$SYSTEMD_DIR"

# Check for binary
BINARY_SOURCE="${PROJECT_ROOT}/build/psiphond-ng"
if [ ! -f "$BINARY_SOURCE" ]; then
    echo "Error: Binary not found at $BINARY_SOURCE"
    echo "Please build first:"
    echo "  cd $PROJECT_ROOT && make build"
    exit 1
fi

# Copy binary
echo "Installing binary to $BIN_DIR"
cp "$BINARY_SOURCE" "$BIN_DIR/"
chmod +x "$BIN_DIR/psiphond-ng"

# Install default config if not exists
if [ ! -f "$CONFIG_DIR/psiphond-ng.conf" ]; then
    echo "Installing default config to $CONFIG_DIR/psiphond-ng.conf"
    # Copy example config from deployment directory
    if [ -f "$SCRIPT_DIR/psiphond-ng.conf.example" ]; then
        # Replace placeholders with actual paths
        sed "s|{{HOME}}|$HOME|g" "$SCRIPT_DIR/psiphond-ng.conf.example" > "$CONFIG_DIR/psiphond-ng.conf"
        echo "Created config with paths for your home directory."
        echo "IMPORTANT: Edit $CONFIG_DIR/psiphond-ng.conf and set your propagation_channel_id and sponsor_id."
    else
        echo "Warning: No example config found. Please create $CONFIG_DIR/psiphond-ng.conf manually."
    fi
else
    echo "Config already exists at $CONFIG_DIR/psiphond-ng.conf, skipping."
fi

# Install systemd service
echo "Installing systemd user service to $SYSTEMD_DIR"
cp "$SCRIPT_DIR/psiphond-ng.service" "$SYSTEMD_DIR/"

# Install control script
echo "Installing control script to $BIN_DIR"
cp "$SCRIPT_DIR/psiphon-ctl" "$BIN_DIR/"
chmod +x "$BIN_DIR/psiphon-ctl"

# Reload systemd daemon and enable service (daemon mode)
echo ""
read -p "Install and enable systemd daemon service? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    systemctl --user daemon-reload
    systemctl --user enable psiphond-ng.service
    echo ""
    echo "Daemon service installed and enabled."
    echo "To start it now: psiphon-ctl start"
    echo "Or: systemctl --user start psiphond-ng"
    echo ""
else
    echo "Skipping service installation."
fi

echo "Installation complete!"
echo ""
echo "Available commands:"
echo "  psiphon-ctl start      # Start as daemon (background, auto-restart)"
echo "  psiphon-ctl stop       # Stop daemon"
echo "  psiphon-ctl status     # Check daemon status"
echo "  psiphon-ctl restart    # Restart daemon"
echo "  psiphon-ctl run        # Run in foreground (client mode)"
echo ""
echo "Config: $CONFIG_DIR/psiphond-ng.conf"
echo "Logs: journalctl --user-unit psiphond-ng -f (daemon mode)"
echo "      or view output directly in 'run' mode"
echo ""
echo "Note: To enable the daemon to start at login, run: systemctl --user enable psiphond-ng"

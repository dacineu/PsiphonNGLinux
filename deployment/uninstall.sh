#!/usr/bin/env bash
# Uninstall PsiphonNG user service

set -euo pipefail

# Installation paths
BIN_DIR="${HOME}/.local/bin"
CONFIG_DIR="${HOME}/.config/psiphond-ng"
SYSTEMD_DIR="${HOME}/.config/systemd/user"

echo "Uninstalling PsiphonNG..."

# Stop and disable service if installed
if command -v systemctl &> /dev/null; then
    if systemctl --user list-units --full --type=service | grep -q "^${SYSTEMD_DIR}/psiphond-ng.service"; then
        echo "Stopping and disabling systemd service..."
        systemctl --user disable --now psiphond-ng.service 2>/dev/null || true
    fi
fi

# Remove service file
if [ -f "$SYSTEMD_DIR/psiphond-ng.service" ]; then
    echo "Removing service file..."
    rm -f "$SYSTEMD_DIR/psiphond-ng.service"
    systemctl --user daemon-reload 2>/dev/null || true
fi

# Remove control script
if [ -f "$BIN_DIR/psiphon-ctl" ]; then
    echo "Removing control script..."
    rm -f "$BIN_DIR/psiphon-ctl"
fi

# Remove binary
if [ -f "$BIN_DIR/psiphond-ng" ]; then
    echo "Removing binary..."
    rm -f "$BIN_DIR/psiphond-ng"
fi

# Ask about config removal
if [ -d "$CONFIG_DIR" ]; then
    read -p "Remove config directory ($CONFIG_DIR)? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Removing config directory..."
        rm -rf "$CONFIG_DIR"
    fi
fi

echo "Uninstall complete."

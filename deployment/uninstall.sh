#!/usr/bin/env bash
# Uninstall PsiphonNG user service
#
# Usage: ./uninstall.sh [--full]
#   --full    Remove ALL traces including config, data, and logs (no prompts)

set -euo pipefail

# Installation paths
BIN_DIR="${HOME}/.local/bin"
CONFIG_DIR="${HOME}/.config/psiphond-ng"
SYSTEMD_DIR="${HOME}/.config/systemd/user"
DATA_DIR="${HOME}/.local/var/lib/psiphon"
LOG_DIR="${HOME}/.local/var/log/psiphon"

FULL_UNINSTALL=false
if [[ "${1:-}" == "--full" ]]; then
    FULL_UNINSTALL=true
fi

echo "=== PsiphonNG Uninstaller ==="
echo ""

# Stop and disable service if installed
if command -v systemctl &> /dev/null; then
    if systemctl --user list-units --full --type=service 2>/dev/null | grep -q "psiphond-ng.service"; then
        echo "→ Stopping and disabling systemd service..."
        systemctl --user disable --now psiphond-ng.service 2>/dev/null || true
    fi
fi

# Remove service file
if [ -f "$SYSTEMD_DIR/psiphond-ng.service" ]; then
    echo "→ Removing systemd service file..."
    rm -f "$SYSTEMD_DIR/psiphond-ng.service"
    systemctl --user daemon-reload 2>/dev/null || true
fi

# Remove binary
if [ -f "$BIN_DIR/psiphond-ng" ]; then
    echo "→ Removing binary..."
    rm -f "$BIN_DIR/psiphond-ng"
fi

# Remove control script (legacy, if present)
if [ -f "$BIN_DIR/psiphon-ctl" ]; then
    echo "→ Removing control script..."
    rm -f "$BIN_DIR/psiphon-ctl"
fi

# Handle config, data, and logs
if [ "$FULL_UNINSTALL" = true ]; then
    echo "→ Full uninstall: removing all user data..."
    rm -rf "$CONFIG_DIR" 2>/dev/null || true
    rm -rf "$DATA_DIR" 2>/dev/null || true
    rm -rf "$LOG_DIR" 2>/dev/null || true
    echo "  - Config: $CONFIG_DIR"
    echo "  - Data: $DATA_DIR"
    echo "  - Logs: $LOG_DIR"
else
    # Interactive mode
    echo ""
    echo "The following directories may contain user data:"
    echo "  - Config: $CONFIG_DIR"
    echo "  - Data:   $DATA_DIR"
    echo "  - Logs:   $LOG_DIR"
    echo ""

    # Ask about config
    if [ -d "$CONFIG_DIR" ] || [ -f "$CONFIG_DIR" ]; then
        read -p "Remove config directory? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "  → Removing config..."
            rm -rf "$CONFIG_DIR"
        fi
    fi

    # Ask about data
    if [ -d "$DATA_DIR" ] || [ -f "$DATA_DIR" ]; then
        read -p "Remove data directory? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "  → Removing data..."
            rm -rf "$DATA_DIR"
        fi
    fi

    # Ask about logs
    if [ -d "$LOG_DIR" ] || [ -f "$LOG_DIR" ]; then
        read -p "Remove log directory? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "  → Removing logs..."
            rm -rf "$LOG_DIR"
        fi
    fi
fi

echo ""
echo "✓ Uninstall complete."
echo ""
if [ "$FULL_UNINSTALL" = false ]; then
    echo "To remove everything (including config, data, logs) without prompts:"
    echo "  $0 --full"
fi

#!/bin/bash
#
# PsiphonNGLinux Uninstallation Script
#
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EID -ne 0 ]]; then
   log_error "This script must be run as root (use sudo)"
   exit 1
fi

log_info "Uninstalling PsiphonNGLinux..."

# Stop and disable service
if systemctl is-active --quiet psiphond-ng; then
    log_info "Stopping psiphond-ng service..."
    systemctl stop psiphond-ng
fi

if systemctl is-enabled --quiet psiphond-ng; then
    log_info "Disabling psiphond-ng service..."
    systemctl disable psiphond-ng
fi

# Remove systemd service
SERVICE_FILE="/etc/systemd/system/psiphond-ng.service"
if [[ -f "$SERVICE_FILE" ]]; then
    log_info "Removing systemd service..."
    rm -f "$SERVICE_FILE"
    systemctl daemon-reload
    systemctl reset-failed psiphond-ng 2>/dev/null || true
fi

# Ask about removing data
read -p "Remove all Psiphon data (/var/lib/psiphon)? This cannot be undone. (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Removing data directory..."
    rm -rf /var/lib/psiphon
fi

# Ask about removing logs
read -p "Remove all Psiphon logs (/var/log/psiphon)? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Removing log directory..."
    rm -rf /var/log/psiphon
fi

# Ask about removing config
read -p "Remove configuration file (/etc/psiphon/psiphond-ng.conf)? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Removing configuration..."
    rm -f /etc/psiphon/psiphond-ng.conf
    rmdir /etc/psiphon 2>/dev/null || true
fi

# Remove binary
BINARY="/usr/local/bin/psiphond-ng"
if [[ -f "$BINARY" ]]; then
    read -p "Remove psiphond-ng binary? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Removing binary..."
        rm -f "$BINARY"
    fi
fi

# Ask about removing user
read -p "Remove psiphon system user? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if id -u psiphon &>/dev/null; then
        log_info "Removing psiphon user..."
        userdel -r psiphon 2>/dev/null || true
    fi
fi

log_info "Uninstallation complete!"

#!/bin/bash
#
# PsiphonNGLinux Installation Script
# Installs the psiphond-ng systemd service
#
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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
if [[ $EUID -ne 0 ]]; then
   log_error "This script must be run as root (use sudo)"
   exit 1
fi

log_info "Installing PsiphonNGLinux..."

# Define paths
BINARY_SOURCE="psiphond-ng"
BINARY_DEST="/usr/local/bin/psiphond-ng"
CONFIG_SOURCE="config/psiphond-ng.conf"
CONFIG_DEST="/etc/psiphon/psiphond-ng.conf"
SERVICE_SOURCE="config/psiphond-ng.service"
SERVICE_DEST="/etc/systemd/system/psiphond-ng.service"

# Check if binary exists in current directory
if [[ ! -f "$BINARY_SOURCE" ]]; then
    log_error "Binary not found: $BINARY_SOURCE"
    log_info "Please build the binary first: go build -o psiphond-ng ./cmd/psiphond-ng"
    exit 1
fi

# Create psiphon user if it doesn't exist
if ! id -u psiphon &>/dev/null; then
    log_info "Creating psiphon user and group..."
    useradd -r -s /usr/sbin/nologin -d /var/lib/psiphon psiphon
fi

# Create directories
log_info "Creating directories..."
mkdir -p /etc/psiphon
mkdir -p /var/lib/psiphon
mkdir -p /var/log/psiphon

# Set ownership
chown -R psiphon:psiphon /var/lib/psiphon
chown -R psiphon:psiphon /var/log/psiphon

# Install binary
log_info "Installing binary to $BINARY_DEST..."
install -m 0755 "$BINARY_SOURCE" "$BINARY_DEST"
chown root:root "$BINARY_DEST"

# Install config
if [[ ! -f "$CONFIG_DEST" ]]; then
    log_info "Installing configuration to $CONFIG_DEST..."
    install -m 0644 "$CONFIG_SOURCE" "$CONFIG_DEST"
    chown root:root "$CONFIG_DEST"
else
    log_warn "Configuration already exists at $CONFIG_DEST, not overwriting"
    log_info "You should review and update your configuration"
fi

# Install systemd service
log_info "Installing systemd service..."
install -m 0644 "$SERVICE_SOURCE" "$SERVICE_DEST"
systemctl daemon-reload

# Enable and start service
read -p "Do you want to enable and start psiphond-ng now? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Enabling and starting psiphond-ng..."
    systemctl enable psiphond-ng
    systemctl start psiphond-ng
    sleep 2

    if systemctl is-active --quiet psiphond-ng; then
        log_info "psiphond-ng is running"
        log_info "Check status: systemctl status psiphond-ng"
        log_info "View logs: journalctl -u psiphond-ng -f"
    else
        log_error "psiphond-ng failed to start"
        log_info "Check logs: journalctl -u psiphond-ng -n 50"
    fi
fi

log_info "Installation complete!"
log_info ""
log_info "Next steps:"
log_info "  1. Edit configuration: sudo nano $CONFIG_DEST"
log_info "     - Set propagation_channel_id and sponsor_id if you have them"
log_info "     - Adjust tunnel_mode (portforward or packet)"
log_info "     - Configure region, proxies, etc."
log_info "  2. Reload systemd: sudo systemctl daemon-reload"
log_info "  3. Start service: sudo systemctl start psiphond-ng"
log_info "  4. Check logs: sudo journalctl -u psiphond-ng -f"
log_info ""
log_info "For port forward mode, configure your applications to use:"
log_info "  SOCKS5: 127.0.0.1:1080"
log_info "  HTTP:   127.0.0.1:8080"
log_info ""
log_info "For packet (TUN) mode, ensure CAP_NET_ADMIN capability:"
log_info "  sudo setcap cap_net_admin+ep /usr/local/bin/psiphond-ng"

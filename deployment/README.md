# PsiphonNG Control Integration

This directory contains scripts and configuration to install PsiphonNG as a user service with flexible control options.

## Features

- **Daemon mode**: Run PsiphonNG as a systemd user service with auto-restart and journald logging
- **Client mode**: Run PsiphonNG in the foreground for debugging or interactive use
- **Easy control**: `psiphon-ctl` script provides simple start/stop/status/run commands
- **User-space install**: No root required; installs to `~/.local/bin` and `~/.config/`
- **Integration-ready**: Includes DOIP approval server integration for in-proxy mode

## Installation

1. **Build PsiphonNG** (if not already built):
   ```bash
   cd /home/dacineu/dev/PsiphonNGLinux
   make build
   ```

2. **Run the installer**:
   ```bash
   cd deployment
   ./install.sh
   ```

   The installer will:
   - Copy `psiphond-ng` binary to `~/.local/bin/`
   - Create config directory `~/.config/psiphond-ng/`
   - Install a default `psiphond-ng.conf` with user-writable paths
   - Install systemd user service to `~/.config/systemd/user/`
   - Install `psiphon-ctl` control script to `~/.local/bin/`
   - Optionally enable the daemon service

3. **Edit the config**:
   ```bash   nano ~/.config/psiphond-ng/psiphond-ng.conf
   ```
   - Set your `propagation_channel_id` and `sponsor_id` (get from Psiphon)
   - Choose tunnel mode (`portforward` for SOCKS/HTTP proxy, `packet` for TUN)
   - If using in-proxy mode with approval:
     - Set `"inproxy_mode": "proxy"` (to run as in-proxy server)
     - Configure `"inproxy_approval_websocket_url": "ws://localhost:8443/approval"`
     - Adjust `inproxy_approval_timeout` as needed

## Usage

### Control Script: `psiphon-ctl`

Make sure `~/.local/bin` is in your PATH, or use full path `~/.local/bin/psiphon-ctl`.

#### Start as daemon (background service)
```bash
psiphon-ctl start
```
This enables and starts the systemd user service. The daemon will auto-restart on failure and log to journald.

#### Stop daemon
```bash
psiphon-ctl stop
```

#### Restart daemon
```bash
psiphon-ctl restart
```

#### Check status
```bash
psiphon-ctl status
```
Shows whether the service is enabled/active and recent log lines.

#### Run in foreground (client mode)
```bash
psiphon-ctl run
```
Runs PsiphonNG directly in the terminal. Press Ctrl+C to stop. This mode is useful for:
- Debugging
- Seeing live logs
- Temporary sessions
- Testing config changes

You can specify a custom config:
```bash
psiphon-ctl run -c /path/to/alt-config.conf
```

Or via environment variable:
```bash
PSIPHON_CONFIG=~/my-config.conf psiphon-ctl run
```

#### View logs

For daemon mode:
```bash
journalctl --user-unit psiphond-ng -f
```

For foreground mode, logs appear directly in the terminal.

## DOIP Approval Server Integration

When running in in-proxy server mode (`inproxy_mode: "proxy"`), PsiphonNG will:
- Listen for client connection requests
- For each connection, send an approval request to the DOIP approval server at `inproxy_approval_websocket_url`
- Wait for approval response (up to `inproxy_approval_timeout`)
- Accept or reject the connection based on server response
- Log warnings if approval fails or fields are missing

The DOIP approval server must be running and accessible. Example endpoints:
- WebSocket: `ws://localhost:8443/approval`
- Admin API: `http://localhost:8080/admin/strict-fields`

You can dynamically adjust which metadata fields are considered "strict" via the DOIP admin API, and PsiphonNG will enforce those requirements in real-time.

## Uninstall

```bash
cd deployment
./uninstall.sh
```

This will stop the service, remove binaries and service files, and optionally remove config.

## Manual Systemd Commands (Alternative to psiphon-ctl)

If you prefer direct systemd control:
```bash
# Start daemon
systemctl --user start psiphond-ng
systemctl --user enable psiphond-ng   # auto-start on login

# Stop daemon
systemctl --user stop psiphond-ng
systemctl --user disable psiphond-ng  # stop auto-start

# Status
systemctl --user status psiphond-ng

# Logs
journalctl --user-unit psiphond-ng -f
```

## Troubleshooting

### Port already in use
If port 1080 or 8080 is already in use, either stop the conflicting process or change the ports in the config.

### TUN mode requires privileges
If using `tunnel_mode: "packet"`, you may need additional capabilities. The service file includes a commented CapabilityBoundingSet. Alternatively, ensure your user has permission to create TUN devices (typically via `group` membership in `netdev` or running with appropriate capabilities).

### Config file validation errors
Check the logs for specific errors. Common issues:
- Invalid JSON syntax
- Invalid URL for `inproxy_approval_websocket_url`
- Missing required fields

### DOIP approval server not running
If in-proxy mode with approval is enabled but the approval server is unreachable, connections will be rejected after timeout. Check that the DOIP server is running on the expected address/port.

### Permissions
Ensure `~/.local/bin` is in your PATH and that config directories are writable by your user.

## File Locations

- Binary: `~/.local/bin/psiphond-ng`
- Control script: `~/.local/bin/psiphon-ctl`
- Config: `~/.config/psiphond-ng/psiphond-ng.conf`
- Data directory: `~/.local/var/lib/psiphon/` (created automatically)
- Log file: `~/.local/var/log/psiphon/psiphond-ng.log` (created automatically)
- Systemd service: `~/.config/systemd/user/psiphond-ng.service`

## Next Steps

Consider also installing the DOIP approval server if you plan to use in-proxy mode with approval:
- See `/home/dacineu/dev/doip-approval-server/deployment/install.sh`
- Start the approval server: `systemctl --user start doip-approval-server`
- Configure strict fields via `curl -X POST http://localhost:8080/admin/strict-fields`

Enjoy secure, flexible Psiphon tunneling with easy user-space management!

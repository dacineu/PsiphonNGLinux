# PsiphonNGLinux

**A full-featured, production-ready Psiphon client for Linux**

PsiphonNGLinux is a native Linux implementation of the Psiphon client, built directly on the official `psiphon-tunnel-core` Go library. It provides advanced features like systemd integration, TUN/tap packet tunneling, automatic server discovery, broker support, and dynamic configuration via tactics — all the capabilities of the PsiphonClient mobile app, adapted for Linux.

---

## Features

✅ **Full Psiphon protocol support** - Uses official psiphon-tunnel-core library
✅ **Systemd service** - Auto-start, restart, journald logging
✅ **TUN/tap packet tunnel mode** - System-wide VPN-like routing
✅ **Port forwarding mode** - Local SOCKS5 and HTTP proxies
✅ **Automatic server discovery** - Remote server lists, DSL, OSL
✅ **Dynamic configuration (Tactics)** - Remote tuning without updates
✅ **Broker/in-proxy support** - Advanced evasion with WebRTC relays
✅ **Connection approval integration** - Optional external authorization via WebSocket or DOIP protocol
✅ **Split tunneling** - Bypass tunnel for local/region traffic
✅ **Automatic updates** - Binary self-update capability
✅ **Performance monitoring** - Stats API, speed tests, feedback
✅ **Multiple tunnel pools** - Parallel connections for redundancy

---

## Quick Start

### Prerequisites

- Linux (tested on Ubuntu 20.04+, Debian 11+, Arch, Fedora)
- Go 1.21+ (for building from source) **or** download pre-built binary
- **No root access required** for proxy mode
- For TUN/packet mode: either root or `setcap cap_net_admin+ep` on binary
- systemd (for service management; most Linux distros have this)

### Installation

PsiphonNGLinux is designed to run as a **user-level service** with XDG directories. No `sudo` required!

#### Option 1: Pre-built binary

```bash
# Download latest release
wget https://github.com/dacineu/PsiphonNGLinux/releases/latest/download/psiphond-ng-linux-amd64
chmod +x psiphond-ng-linux-amd64

# Move to ~/.local/bin (or anywhere in your PATH)
mkdir -p ~/.local/bin
mv psiphond-ng-linux-amd64 ~/.local/bin/psiphond-ng

# Install user systemd service
mkdir -p ~/.config/systemd/user
cp config/psiphond-ng-user.service ~/.config/systemd/user/psiphond-ng.service

# Reload systemd user daemon
systemctl --user daemon-reload

# Enable and start the service
systemctl --user enable --now psiphond-ng

# Check status
systemctl --user status psiphond-ng
```

#### Option 2: Build from source

```bash
git clone https://github.com/dacineu/PsiphonNGLinux.git
cd PsiphonNGLinux
go mod download
go build -o psiphond-ng ./cmd/psiphond-ng

# Install user systemd service
mkdir -p ~/.config/systemd/user
cp config/psiphond-ng-user.service ~/.config/systemd/user/psiphond-ng.service
systemctl --user daemon-reload
systemctl --user enable --now psiphond-ng
```

#### Option 3: Run directly (no service)

For running without root privileges using **systemd user services**:

```bash
# 1. Build or obtain the binary
go build -o psiphond-ng ./cmd/psiphond-ng

# 2. Install user service
mkdir -p ~/.config/systemd/user
cp config/psiphond-ng-user.service ~/.config/systemd/user/psiphond-ng.service

# 3. Reload systemd user daemon
systemctl --user daemon-reload

# 4. Enable and start the service
systemctl --user enable --now psiphond-ng

# 5. Check status
systemctl --user status psiphond-ng

# 6. View logs
journalctl --user -u psiphond-ng -f
```

**Notes:**
- The service runs entirely within your user session
- Configuration is auto-created at `~/.config/psiphond-ng/psiphond-ng.conf`
- Data directory: `~/.local/var/lib/psiphon`
- Log file: `~/.local/var/log/psiphon/psiphond-ng.log`
- The default config includes the official Psiphon server list signature key, so remote server fetching works out of the box
- TUN mode requires root or `setcap cap_net_admin+ep` on the binary (see below)

#### Run directly (no systemd)

```bash
# First time: creates default config at ~/.config/psiphond-ng/psiphond-ng.conf
./psiphond-ng

# Or specify custom config location
./psiphond-ng -config /path/to/custom.conf

# Run in foreground (logs to console)
./psiphond-ng -config ~/.config/psiphond-ng/psiphond-ng.conf -log-level debug
```


## Configuration

### Configuration File

**Default location** (user-level): `~/.config/psiphond-ng/psiphond-ng.conf`

This file is auto-created with sensible defaults the first time you run the binary or start the user service.

```json
{
  "data_directory": "~/.local/var/lib/psiphon",
  "log_file": "~/.local/var/log/psiphon/psiphond-ng.log",
  "log_level": "info",

  // Network identifiers (obtain from Psiphon Network)
  "PropagationChannelId": "YOUR_CHANNEL_ID",
  "SponsorId": "YOUR_SPONSOR_ID",

  // Tunnel mode
  "TunnelMode": "portforward", // "portforward" or "packet"

  // Port forward settings
  "LocalSocksProxyPort": 1080,
  "LocalHttpProxyPort": 8080,

  // Packet tunnel (TUN) settings
  "PacketTunnelDeviceName": "tun-psiphon",
  "PacketTunnelAddressIPv4": "10.0.0.2/30",
  "PacketTunnelDNSIPv4": "8.8.8.8",
  "PacketTunnelGatewayIPv4": "10.0.0.1",

  // Region selection (ISO 3166-1 alpha-2)
  "EgressRegion": "",

  // Split tunneling
  "SplitTunnelOwnRegion": false,
  "SplitTunnelInclude": [],
  "SplitTunnelExclude": [],

  // In-proxy (broker) mode
  "InproxyMode": "", // "client" for client mode, "proxy" for server mode
  "InproxyBrokerServerAddresses": [],
  "InproxyCompartmentID": "",
  "InproxyApprovalWebSocketURL": "", // optional: for connection approval
  "InproxyApprovalTimeout": "10s",

  // Rate limiting
  "LimitUpstreamBytesPerSecond": 0,
  "LimitDownstreamBytesPerSecond": 0,

  // Connection parameters
  "EstablishTunnelTimeout": "5m",
  "ConnectionWorkerPoolSize": 0, // 0 = default
  "TunnelPoolSize": 1,

  // Remote server list fetching
  "DisableRemoteServerListFetcher": false,
  "RemoteServerListURLs": [
    "https://s3.amazonaws.com/psiphon/web/server_list_compressed"
  ],
  "RemoteServerListSignaturePublicKey": "...",

  // Obfuscated server lists
  "ObfuscatedServerListRootURLs": [],

  // Tactics
  "DisableTactics": false,
  "TacticsWaitPeriod": "10s",
  "TacticsRetryPeriod": "1m",

  // Logging
  "UseNoticeFiles": true,
  "NoticeRotationSize": 10485760,
  "NoticeRotationCount": 1,

  // Advanced
  "UserAgent": "",
  "CustomHeaders": {},
  "UpstreamProxyURL": "",
  "LimitTunnelProtocols": [],
  "InitialLimitTunnelProtocols": [],
  "InitialLimitTunnelProtocolsCandidateCount": 0
}
```

---

## Usage

### Starting the user service

```bash
# Start (if not already started with --now)
systemctl --user start psiphond-ng

# Enable auto-start on login/boot
systemctl --user enable psiphond-ng

# Check status
systemctl --user status psiphond-ng

# View logs
journalctl --user -u psiphond-ng -f
journalctl --user -u psiphond-ng --since "1 hour ago"
```

### Running manually (for debugging)

```bash
# Run with default user config (~/.config/psiphond-ng/psiphond-ng.conf)
./psiphond-ng

# With custom config file
./psiphond-ng -config /path/to/custom.conf

# With debug logging
./psiphond-ng -config ~/.config/psiphond-ng/psiphond-ng.conf -log-level debug
```

### When running in port forward mode:

1. **Proxy ports:**
   - SOCKS5: `127.0.0.1:1080` (configured via `LocalSocksProxyPort`)
   - HTTP: `127.0.0.1:8080` (configured via `LocalHttpProxyPort`)

2. **Configure applications:**
   - Set browser/system proxy to the above ports
   - Or use `proxychains` to force applications through Psiphon
   ```bash
   proxychains curl https://checkip.amazonaws.com
   ```

### When running in packet tunnel mode:

1. The TUN device (`tun-psiphon` by default) is created and configured
2. All traffic routed through the tunnel (configure via `PacketTunnelGatewayIPv4`)
3. Set up NAT on the host if sharing the tunnel with other devices/containers:
   ```bash
   sudo iptables -t nat -A POSTROUTING -o tun-psiphon -j MASQUERADE
   ```
   (This requires root; skip if only routing your own traffic)
4. DNS queries are intercepted and sent through the tunnel

---

## In-Proxy Mode (Broker-Based)

For advanced deployments using the in-proxy system:

### Running as a client (connects via broker)

```json
{
  "InproxyMode": "client",
  "InproxyBrokerServerAddresses": ["broker.example.com:443"],
  "InproxyCompartmentID": "personal-compartment-id-if-any"
}
```

The client will:
1. Contact the broker to get a matched proxy
2. Perform WebRTC handshake
3. Relay traffic through the proxy to the destination Psiphon server

### Running as a proxy (relay for other clients)

```json
{
  "InproxyMode": "proxy",
  "InproxyBrokerServerAddresses": ["broker.example.com:443"],
  "InproxyMaxCommonClients": 10,
  "InproxyMaxPersonalClients": 5,
  "InproxySessionPrivateKey": "base64-encoded-ed25519-key",
  "InproxySessionRootObfuscationSecret": "base64-encoded-secret",
  "InproxyEgressInterface": "eth0", // for TUN mode
  "InproxyMTU": 1500,
  // Optional: connection approval
  "InproxyApprovalWebSocketURL": "ws://localhost:8080/approve",  // or DOIP approval server
  "InproxyApprovalTimeout": "5s"
}
```

See `docs/inproxy-tun-approval-integration.md` (from psiphon-tunnel-core) for detailed setup.

---

## Integration with DOIP Approval Server

For production deployments requiring robust connection authorization, the **DOIP Approval Server**
provides a standards-compliant, dynamically-configurable solution.

### What is DOIP Approval Server?

`doip-approval-server` is a standalone Go service that validates connection approval requests
against a configurable set of strict fields. It features:

- **WebSocket endpoint** (`/approval` or `/ws`) for real-time validation
- **Admin HTTP API** (`/admin/strict-fields`) for dynamic configuration
- **JSON config file** with hot-reloading (no restart needed)
- **Systemd user service** support
- **Strict field enforcement**: requires specific metadata fields in each request

### Usage with PsiphonNGLinux

1. **Deploy the DOIP approval server** (see its README):
   ```bash
   cd doip-approval-server
   go build -o doip-approval-server ./cmd/doip-approval-server
   ./doip-approval-server -config ./config.json
   ```

2. **Configure PsiphonNGLinux** to use it:

   ```json
   {
     "inproxy_approval_websocket_url": "ws://localhost:8443/approval",
     "inproxy_approval_timeout": "10s"
   }
   ```

   Note: The `InproxyApprovalWebSocketURL` works with any WebSocket server implementing
   the approval protocol (JSON request/response). The DOIP server uses the same protocol.

3. **Configure strict fields** (optional) via DOIP admin API:
   ```bash
   curl -X POST http://localhost:8080/admin/strict-fields \
     -H "Content-Type: application/json" \
     -d '{"strict_fields": ["connection_id", "client_region", "destination", "timestamp"]}'
   ```

   The approval server will then require these fields in every approval request from Psiphon.
   Psiphon's `ClientConnectionInfo` includes: `connection_id`, `client_region`, `destination`,
   `network_protocol`. Additional fields (timestamp, daemon_version, daemon_platform) would
   need to be added via custom code if required.

### Benefits over Simple WebSocket Server

| Feature | Simple ws-approval-server | DOIP Approval Server |
|---------|--------------------------|---------------------|
| Dynamic config reload | ❌ No | ✅ Yes (via file watch) |
| Admin API for config | ❌ No | ✅ Yes |
| Strict field validation | ❌ No (auto-approves) | ✅ Yes |
| Production-ready logging | ❌ Basic | ✅ Extensible |
| Systemd service | ❌ Manual | ✅ Included |

### When to Use Which?

- **ws-approval-server**: Development, testing, proof-of-concept
- **DOIP approval server**: Production, multi-tenant environments, audit requirements

---

---

## Region Selection

Set `EgressRegion` to an ISO 3166-1 alpha-2 country code:

```json
"EgressRegion": "DE"
```

Available regions: AT, BE, BG, CA, CH, CZ, DE, DK, EE, ES, FI, FR, GB, HU, IE, IN, IT, JP, LV, NL, NO, PL, RO, RS, SE, SG, SK, US, and others depending on current Psiphon infrastructure.

Leave blank for automatic best-performance selection.

---

## Split Tunneling

Configure which traffic goes through the tunnel:

```json
{
  "SplitTunnelOwnRegion": true,
  "SplitTunnelInclude": ["192.168.0.0/16", "10.0.0.0/8"],
  "SplitTunnelExclude": ["8.8.8.8", "1.1.1.1"]
}
```

- `SplitTunnelOwnRegion`: Direct-connect to servers in your own country
- `SplitTunnelInclude`: CIDR ranges that bypass tunnel (local network)
- `SplitTunnelExclude`: Specific IPs that always go through tunnel (even if in include list)

---

## Logs and Diagnostics

### Log files

If `UseNoticeFiles` is enabled, notices are written to:

- Main notice file: `~/.local/var/log/psiphon/psiphond-ng.notices`
- Rotated file: `~/.local/var/log/psiphon/psiphond-ng.notices.1` (if size > `NoticeRotationSize`)

### Journald (systemd user service)

```bash
# View recent logs
journalctl --user -u psiphond-ng -n 100

# Follow live logs
journalctl --user -u psiphond-ng -f

# View logs with specific priority
journalctl --user -u psiphond-ng -p info

# Export logs
journalctl --user -u psiphond-ng --since "2024-01-01" > psiphon.log
```

### Debug logging

Set `"LogLevel": "debug"` in config. Verbose logs include:
- Tunnel establishment attempts
- Server scoring details
- API request/response details
- Protocol negotiation
- Obfuscation performance

---

## Troubleshooting

### Tunnel won't establish

1. Check logs:
   ```bash
   journalctl --user -u psiphond-ng -n 50
   ```

2. Common issues:
   - **No server entries**: Check network connectivity; wait for remote server list fetch
   - **All servers timing out**: Check firewall allows outbound connections (port 443, others)
   - **Tactics timeout**: Ensure `PsiphonAPIRequestTimeout` sufficient; check DNS
   - **Region restriction**: Try different `EgressRegion` or leave blank

3. Enable debug logging temporarily:
   ```bash
   systemctl --user edit psiphond-ng
   # Add:
   [Service]
   Environment="LOG_LEVEL=debug"
   systemctl --user restart psiphond-ng
   ```

### TUN device not created

- Ensure CAP_NET_ADMIN capability or run as root
  - For user-level service, grant capability: `sudo setcap cap_net_admin+ep /usr/local/bin/psiphond-ng`
  - Or run the binary with sudo (not recommended for production)
- Check TUN module loaded: `sudo modprobe tun`
- Verify no conflicting TUN devices

### Port forward not working

- Verify proxy ports are listening: `ss -tlnp | grep psiphond-ng`
- Check application proxy configuration
- Test with `curl`: `curl -x http://127.0.0.1:8080 https://checkip.amazonaws.com`

### High latency/packet loss

- Check `EgressRegion` (try different region)
- Review server scoring: logs show server RTTs
- Limit protocols: set `LimitTunnelProtocols` to exclude high-latency ones like meek
- Adjust `ConnectionWorkerPoolSize` for parallelization

### Connection drops frequently

- Review logs for SSH keepalive failures
- Check server-side issues (stats.psiphon.ca)
- Enable `SSHKeepAliveSpeedTestSampleProbability` for proactive testing
- Consider increasing `EstablishTunnelTimeout` on slow networks

---

## Advanced Configuration

### Custom server list (for testing)

```json
{
  "TargetServerEntry": "base64-encoded-server-entry",
  "DisableRemoteServerListFetcher": true,
  "DisableTactics": true
}
```

### Protocol restrictions

```json
{
  "LimitTunnelProtocols": ["OSSH", "TLS-OSSH"],
  "InitialLimitTunnelProtocols": ["OSSH"],
  "InitialLimitTunnelProtocolsCandidateCount": 5
}
```

Valid protocols: SSH, OSSH, TLS-OSSH, QUIC-OSSH, FRONTED-MEEK-OSSH, UNFRONTED-MEEK-OSSH, TAPDANCE-OSSH, CONJURE-OSSH, SHADOWSOCKS-OSSH.

### Obfuscation tuning

See `psiphon/common/parameters/parameters.go` for all tunable parameters. Examples:

```json
{
  "ObfuscatedSSHMinPadding": 100,
  "ObfuscatedSSHMaxPadding": 2000,
  "FragmentorProbability": 0.1,
  "FragmentorMinTotalBytes": 1000,
  "FragmentorMaxTotalBytes": 5000,
  "UseIndistinguishableTLS": true
}
```

---

## Security Considerations

PsiphonNGLinux is designed to run without root privileges. Here are security best practices:

- **Configuration files** contain sensitive data. For user-level installation:
  ```bash
  chmod 600 ~/.config/psiphond-ng/psiphond-ng.conf
  ```

- **Log files** may contain metadata. Ensure proper permissions:
  ```bash
  chmod 700 ~/.local/var/log/psiphon
  ```

- **Run as non-root** by default (user service). No dedicated system user needed.

- **TUN mode** requires elevated privileges:
  - Option 1: Run binary with `sudo` (not recommended for ongoing use)
  - Option 2: Grant capability: `sudo setcap cap_net_admin+ep $(which psiphond-ng)` (preferred)
  - Option 3: Use root-only system-wide service (legacy approach)

- **Firewall**: Allow outbound to Psiphon servers (typically port 443). Restrict inbound access to local proxies (127.0.0.1 only by default).

- **Process isolation**: The user service runs with security hardening (NoNewPrivileges, PrivateTmp, etc.) automatically.

- **Binary verification**: Verify downloads with checksums when possible.

---

## Updating

### Manual update (user-level)

```bash
# Download new binary to ~/.local/bin
wget -O ~/.local/bin/psiphond-ng https://github.com/dacineu/PsiphonNGLinux/releases/latest/download/psiphond-ng-linux-amd64
chmod +x ~/.local/bin/psiphond-ng

# Restart service
systemctl --user restart psiphond-ng
```


### Automatic updates (watchtower)

If using the optional watchtower integration:
```bash
# Watchtower will automatically check for new images (if containerized)
# or replace binary if configured
```

---

## Uninstallation

### User-level installation

```bash
# Stop and disable user service
systemctl --user disable --now psiphond-ng

# Remove user service file
rm ~/.config/systemd/user/psiphond-ng.service
systemctl --user daemon-reload

# Remove binary (if installed to ~/.local/bin)
rm ~/.local/bin/psiphond-ng

# Remove config and data
rm -rf ~/.config/psiphond-ng ~/.local/var/lib/psiphon ~/.local/var/log/psiphon
```

### System-wide installation (if you used sudo)

```bash
sudo systemctl disable --now psiphond-ng
sudo rm /usr/local/bin/psiphond-ng
sudo rm -rf /etc/psiphon /var/lib/psiphon /var/log/psiphon
sudo rm /etc/systemd/system/psiphond-ng.service
sudo systemctl daemon-reload
```

---

## Development

### Project structure

```
PsiphonNGLinux/
├── cmd/
│   └── psiphond-ng/         # Main daemon binary
│       └── main.go
├── internal/
│   ├── config/              # Config parsing and validation
│   ├── daemon/              # Systemd integration
│   ├── tunnel/              # Tunnel management wrapper
│   └── metrics/             # Prometheus/StatsD metrics
├── config/
│   ├── psiphond-ng.conf         # Default config (system-wide template)
│   ├── psiphond-ng-user.service # User-level systemd service
│   └── defaults/                # Default parameter sets
├── scripts/
│   ├── install.sh           # System-wide installation script
│   ├── uninstall.sh         # Removal script
│   └── update.sh            # Update check script
├── docs/
│   ├── TUN-setup.md
│   ├── inproxy-setup.md
│   ├── troubleshooting.md
│   └── configuration-reference.md
├── vendor/                  # Vendored dependencies (optional)
├── go.mod
├── go.sum
└── README.md
```

### Building

```bash
# Development build (with local psiphon-tunnel-core)
go build -o psiphond-ng ./cmd/psiphond-ng

# Production build (static binary)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o psiphond-ng ./cmd/psiphond-ng

# With upx compression
upx --best psiphond-ng
```

### Installing after build (user-level)

```bash
# Copy binary to ~/.local/bin
mkdir -p ~/.local/bin
cp psiphond-ng ~/.local/bin/

# Install user service
mkdir -p ~/.config/systemd/user
cp config/psiphond-ng-user.service ~/.config/systemd/user/psiphond-ng.service
systemctl --user daemon-reload
systemctl --user enable --now psiphond-ng
```

---

## Differences from PsiphonLinux

| Aspect | PsiphonLinux | PsiphonNGLinux |
|--------|--------------|----------------|
| **Implementation** | Wrapper script around official binary | Native Go client using psiphon-tunnel-core library |
| **Build** | Downloads pre-built binary | Can be built from source |
| **Configuration** | Static single-server config | Dynamic multi-source server discovery |
| **Updates** | Manual binary replacement | Optional automatic updates; tactics for config |
| **Tactics support** | Limited (binary's built-in only) | Full remote configuration via Psiphon API |
| **In-proxy** | No | Full support (client and server modes) |
| **TUN mode** | No | Yes (systemd integration, network setup) |
| **Metrics** | Minimal | Prometheus/StatsD, feedback API |
| **Service management** | Simple init script | Full systemd unit with watchdog, journald |
| **Compartment IDs** | No | Yes (for broker rate limits) |
| **Connection approval** | No | Optional WebSocket hook (compatible with DOIP approval server) |
| **Log rotation** | Manual | Built-in rotation + journald |
| **Split tunneling** | No | Yes (own region, CIDR include/exclude) |

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests (if applicable)
4. Submit a PR with description and testing instructions

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## License

This project is licensed under the same terms as psiphon-tunnel-core (GPLv3). See [LICENSE](LICENSE).

---

## Acknowledgments

- [Psiphon Inc.](https://psiphon.ca) for the original Psiphon technology
- [Psiphon-Labs/psiphon-tunnel-core](https://github.com/Psiphon-Labs/psiphon-tunnel-core) for the Go implementation
- The Psiphon Network and all its operators

---

**Warning:** Psiphon is a circumvention tool. Usage may be restricted in some jurisdictions. Users are responsible for complying with local laws.

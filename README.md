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
✅ **Connection approval integration** - Optional external authorization
✅ **Split tunneling** - Bypass tunnel for local/region traffic
✅ **Automatic updates** - Binary self-update capability
✅ **Performance monitoring** - Stats API, speed tests, feedback
✅ **Multiple tunnel pools** - Parallel connections for redundancy

---

## Quick Start

### Prerequisites

- Linux (tested on Ubuntu 20.04+, Debian 11+, Arch, Fedora)
- Go 1.21+ (for building from source)
- Root access (for TUN mode; optional for proxy-only mode)
- Systemd (for service mode)

### Installation

#### Option 1: Pre-built binary

```bash
# Download latest release
wget https://github.com/dacineu/PsiphonNGLinux/releases/latest/download/psiphond-ng-linux-amd64
sudo mv psiphond-ng-linux-amd64 /usr/local/bin/psiphond-ng
sudo chmod +x /usr/local/bin/psiphond-ng

# Install systemd service
sudo cp config/psiphond-ng.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now psiphond-ng
```

#### Option 2: Build from source

```bash
git clone https://github.com/dacineu/PsiphonNGLinux.git
cd PsiphonNGLinux
go mod download
go build -o psiphond-ng ./cmd/psiphond-ng

# Install
sudo cp psiphond-ng /usr/local/bin/
sudo cp config/psiphond-ng.service /etc/systemd/system/
sudo cp config/psiphond-ng.conf /etc/psiphon/
sudo systemctl daemon-reload
sudo systemctl enable --now psiphond-ng
```

---

## Configuration

### Configuration File

Default location: `/etc/psiphon/psiphond-ng.conf`

```json
{
  "DataDirectory": "/var/lib/psiphon",
  "LogFile": "/var/log/psiphon/psiphond-ng.log",
  "LogLevel": "info",

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

### Starting the service

```bash
# Start
sudo systemctl start psiphond-ng

# Enable auto-start on boot
sudo systemctl enable psiphond-ng

# Check status
sudo systemctl status psiphond-ng

# View logs
sudo journalctl -u psiphond-ng -f
sudo journalctl -u psiphond-ng --since "1 hour ago"
```

### Running manually (for debugging)

```bash
# With config file
sudo psiphond-ng -config /etc/psiphon/psiphond-ng.conf

# With data directory (for testing)
mkdir -p ~/.psiphon
psiphond-ng -datadir ~/.psiphon -config ./dev.conf
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
3. Set up NAT on the host if needed:
   ```bash
   sudo iptables -t nat -A POSTROUTING -o tun-psiphon -j MASQUERADE
   ```
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
  "InproxyApprovalWebSocketURL": "ws://localhost:8080/approve",
  "InproxyApprovalTimeout": "5s"
}
```

See `docs/inproxy-tun-approval-integration.md` (from psiphon-tunnel-core) for detailed setup.

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

- Main notice file: `/var/log/psiphon/psiphond-ng.notices`
- Rotated file: `/var/log/psiphon/psiphond-ng.notices.1` (if size > `NoticeRotationSize`)

### Journald (systemd)

```bash
# View recent logs
sudo journalctl -u psiphond-ng -n 100

# Follow live logs
sudo journalctl -u psiphond-ng -f

# View logs with specific priority
sudo journalctl -u psiphond-ng -p info

# Export logs
sudo journalctl -u psiphond-ng --since "2024-01-01" > psiphon.log
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
   sudo journalctl -u psiphond-ng -n 50
   ```

2. Common issues:
   - **No server entries**: Check network connectivity; wait for remote server list fetch
   - **All servers timing out**: Check firewall allows outbound connections (port 443, others)
   - **Tactics timeout**: Ensure `PsiphonAPIRequestTimeout` sufficient; check DNS
   - **Region restriction**: Try different `EgressRegion` or leave blank

3. Enable debug logging temporarily:
   ```bash
   sudo systemctl edit psiphond-ng
   # Add:
   [Service]
   Environment="LOG_LEVEL=debug"
   sudo systemctl restart psiphond-ng
   ```

### TUN device not created

- Ensure CAP_NET_ADMIN capability or run as root
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

- **Configuration files** contain sensitive data (server entries, keys). Set permissions:
  ```bash
  sudo chmod 600 /etc/psiphon/psiphond-ng.conf
  sudo chown root:root /var/lib/psiphon
  ```

- **Log files** may contain metadata. Secure log access:
  ```bash
  sudo chown psiphon:psiphon /var/log/psiphon/
  sudo chmod 750 /var/log/psiphon
  ```

- **Run as non-root** when using port forward mode (no TUN). Create dedicated user:
  ```bash
  sudo useradd -r -s /usr/sbin/nologin psiphon
  sudo chown -R psiphon:psiphon /var/lib/psiphon /var/log/psiphon
  ```

- **TUN mode** requires root or CAP_NET_ADMIN. Consider setcap:
  ```bash
  sudo setcap cap_net_admin+ep /usr/local/bin/psiphond-ng
  ```

- **Firewall**: Allow outbound to Psiphon servers (typically port 443). Restrict inbound access to local proxies.

---

## Updating

### Manual update

```bash
# Download new binary
sudo wget -O /usr/local/bin/psiphond-ng https://github.com/dacineu/PsiphonNGLinux/releases/latest/download/psiphond-ng-linux-amd64
sudo chmod +x /usr/local/bin/psiphond-ng
sudo systemctl restart psiphond-ng
```

### Automatic updates (watchtower)

If using the optional watchtower integration:
```bash
# Watchtower will automatically check for new images (if containerized)
# or replace binary if configured
```

---

## Uninstallation

```bash
# Stop and disable service
sudo systemctl disable --now psiphond-ng

# Remove binary
sudo rm /usr/local/bin/psiphond-ng

# Remove config and data
sudo rm -rf /etc/psiphon /var/lib/psiphon /var/log/psiphon

# Remove systemd unit
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
│   ├── psiphond-ng.conf     # Default config
│   ├── psiphond-ng.service  # Systemd unit file
│   └── defaults/            # Default parameter sets
├── scripts/
│   ├── install.sh           # Installation script
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
| **Connection approval** | No | Optional WebSocket hook |
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

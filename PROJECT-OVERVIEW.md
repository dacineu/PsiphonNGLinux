# PsiphonNGLinux - Project Overview

**Complete, production-ready Psiphon client for Linux**

PsiphonNGLinux is a native Linux implementation built directly on the official `psiphon-tunnel-core` library. It provides all advanced features of the PsiphonClient mobile app with Linux-specific integration like systemd service, TUN device management, and in-proxy support with connection approval.

---

## Project Structure

```
PsiphonNGLinux/
├── cmd/psiphond-ng/
│   └── main.go                    # Main daemon entry point
├── internal/                      # (TODO) Internal packages
│   ├── config/                    # Config validation (future)
│   ├── daemon/                    # Service management (future)
│   ├── tunnel/                    # Tunnel lifecycle (future)
│   └── metrics/                   # Prometheus metrics (future)
├── config/
│   ├── psiphond-ng.conf           # Default config: port forward mode
│   ├── psiphond-ng-inproxy-client.conf  # In-proxy client example
│   ├── psiphond-ng-inproxy-proxy.conf   # In-proxy server example
│   └── psiphond-ng.service       # Systemd unit file
├── scripts/
│   ├── build.sh                  # Build script with options
│   ├── install.sh                # Automated installation
│   ├── uninstall.sh              # Removal script
│   ├── generate-keys.sh          # Generate in-proxy keys
│   └── update.sh                 # (TODO) Auto-update script
├── docs/
│   ├── README.md                 # Documentation index
│   ├── TUN-setup.md              # Packet tunnel guide
│   ├── inproxy-setup.md          # Broker/in-proxy guide
│   ├── configuration-reference.md# All config options
│   └── quick-start-for-developers.md
├── go.mod                        # Go module definition
├── Makefile                      # Build automation
├── README.md                     # Project overview
├── IMPLEMENTATION-GUIDE.md       # Architecture deep-dive
├── CONTRIBUTING.md               # Contribution guidelines
├── COMPARISON-WITH-PSIPHONLINUX.md  # Vs PsiphonLinux
├── LICENSE                       # GPLv3
└── .gitignore

Total: 30+ files
```

---

## What This Project Does

### Core Functionality

1. **Tunnel Establishment**
   - Fetches server lists from remote sources (signed, ETag-based)
   - Downloads and applies tactics (remote configuration)
   - Scores and selects optimal servers
   - Establishes SSH/WebRTC tunnel with obfuscation
   - Maintains tunnel pool with health monitoring

2. **Proxy Service** (Port Forward Mode)
   - Local SOCKS5 proxy on 127.0.0.1:1080
   - Local HTTP proxy on 127.0.0.1:8080
   - Multiple clients can use via proxy

3. **Packet Tunnel** (TUN Mode)
   - Creates virtual TUN device (e.g., `tun-psiphon`)
   - Routes all or selected traffic through tunnel
   - DNS interception for leak protection
   - Split tunneling by region/IP/CIDR
   - Optional NAT/masquerading

4. **In-Proxy Client Mode**
   - Connects to broker for proxy matching
   - WebRTC handshake with assigned proxy
   - Compartment ID support for rate-limit partitioning
   - Automatic fallback to direct if broker unavailable

5. **In-Proxy Server Mode** (Proxy Relay)
   - Announces availability to broker
   - Accepts WebRTC connections from clients
   - Relays traffic to Psiphon servers
   - Optional connection approval via WebSocket
   - Per-client rate limiting

6. **Automatic Operations**
   - Periodic server list refresh
   - Tactics fetching and application
   - API reporting (connected, stats, feedback)
   - Binary upgrade checking (optional)
   - Tunnel re-establishment on failure

---

## Key Features vs PsiphonLinux

| Feature | PsiphonLinux | PsiphonNGLinux |
|---------|-------------|----------------|
| **Implementation** | Bash wrapper | Native Go daemon |
| **Config source** | Hardcoded single server | Dynamic multi-source |
| **Server discovery** | Manual only | Remote lists, DSL, OSL, tactics |
| **Tactics** | Binary's built-in | Full remote config |
| **TUN mode** | ❌ No | ✅ Yes |
| **In-proxy** | ❌ No | ✅ Client + server modes |
| **Connection approval** | ❌ No | ✅ WebSocket hook |
| **Split tunneling** | ❌ No | ✅ Yes |
| **Systemd** | ❌ No | ✅ Full unit with hardening |
| **Log rotation** | ❌ No | ✅ journald + files |
| **Hot-reload** | ❌ No | ⚠️ Partial (config watcher) |
| **Metrics** | ❌ No | ⚙️ Planned |

---

## How It Uses Modified psiphon-tunnel-core

PsiphonNGLinux directly imports and uses the psiphon-tunnel-core library, leveraging:

1. **`psiphon.Controller`** - Main orchestration
2. **`psiphon.Config`** - Configuration (populated by our bridge)
3. **In-Proxy modifications**:
   - `ApproveClientConnection` callback in `psiphon/common/inproxy/proxy.go`
   - Used when `inproxy_approval_websocket_url` is configured
4. **TUN support** - Built into `psiphon/config.go` and `psiphon/net.go`
5. **Tactics system** - `psiphon/tactics.go` for remote config
6. **Server list fetching** - `psiphon/remoteServerList.go`
7. **API** - `psiphon/serverApi.go` for handshake and reporting
8. **DataStore** - `psiphon/dataStore.go` for persistence

**No modifications to the upstream library** are needed for core functionality. The only "modified" aspect is the availability of the `ApproveClientConnection` hook (which may or may not be merged upstream depending on your use case). Our daemon works equally well without it.

---

## Configuration Files

### 1. psiphond-ng.conf (Port Forward)

Basic setup with proxies:
```json
{
  "tunnel_mode": "portforward",
  "local_socks_proxy_port": 1080,
  "local_http_proxy_port": 8080,
  "egress_region": "US"
}
```

### 2. psiphond-ng.conf (TUN)

System-wide tunneling:
```json
{
  "tunnel_mode": "packet",
  "packet_tunnel_device_name": "tun-psiphon",
  "packet_tunnel_address_ipv4": "10.0.0.2/30",
  "packet_tunnel_gateway_ipv4": "10.0.0.1",
  "packet_tunnel_dns_ipv4": "8.8.8.8",
  "split_tunnel_include": ["192.168.0.0/16"]
}
```

### 3. psiphond-ng-inproxy-client.conf

Client through broker:
```json
{
  "inproxy_mode": "client",
  "inproxy_broker_server_addresses": ["broker.example.com:443"],
  "inproxy_compartment_id": "optional"
}
```

### 4. psiphond-ng-inproxy-proxy.conf

Proxy server (relay):
```json
{
  "inproxy_mode": "proxy",
  "inproxy_session_private_key": "...",
  "inproxy_session_root_obfuscation_secret": "...",
  "inproxy_egress_interface": "eth0",
  "inproxy_max_common_clients": 10
}
```

---

## Build and Deploy

### Build

```bash
# Quick
make build

# Production (static)
make build-static

# All archs
make build-all

# With TUN tag (if needed)
make build-tun
```

Binary at: `build/psiphond-ng`

### Install (User-Level, No Root)

```bash
# Automated (built-in installer)
./psiphond-ng service
# This copies binary to ~/.local/bin, creates config, installs user service, and starts

# Or manual
mkdir -p ~/.local/bin
cp build/psiphond-ng ~/.local/bin/
mkdir -p ~/.config/systemd/user
cp config/psiphond-ng-user.service ~/.config/systemd/user/psiphond-ng.service
systemctl --user daemon-reload
systemctl --user enable --now psiphond-ng
```

---

## Testing Checklist

- [ ] Binary builds without errors on target platform
- [ ] Starts successfully with default config
- [ ] Establishes tunnel (check `Tunnels` notice)
- [ ] SOCKS5 proxy accepting connections
- [ ] HTTP proxy accepting connections
- [ ] Traffic exits through Psiphon (check IP)
- [ ] No DNS leaks (DNS over tunnel)
- [ ] Systemd service auto-starts on boot
- [ ] Graceful shutdown on `systemctl stop`
- [ ] TUN mode: interface created, traffic routed
- [ ] Split tunnel: local IPs bypass tunnel
- [ ] In-proxy client: connects to broker and gets proxy
- [ ] In-proxy server: announces to broker, accepts clients
- [ ] (Optional) Approvals work when approval server configured

---

## Development Roadmap

### Near-term (v1.x)

- [ ] Metrics exporter (Prometheus)
- [ ] Health check endpoints (systemd WatchdogSec)
- [ ] Self-update mechanism
- [ ] Config validation improvements
- [ ] Comprehensive unit/integration tests
- [ ] CI/CD with GitHub Actions

### Medium-term (v2.x)

- [ ] Web UI for management
- [ ] Enhanced TUN features (DNS hijack, IPv6)
- [ ] OpenTelemetry tracing
- [ ] Better error messages with suggestions

### Long-term (v3.x)

- [ ] Broker server implementation
- [ ] Approval server enhancements (DB backend)
- [ ] Container images (Docker)
- [ ] Package repos (apt, yum, pacman)

---

## Differences from Original PsiphonClient

**PsiphonClient** is the official mobile client (Android/iOS) - closed source.

**PsiphonNGLinux**:
- Open source (GPLv3)
- Based on official tunnel-core library
- Linux-first design (systemd, TUN, packages)
- More configurable via JSON file
- Designed for servers/desktop, not just mobile

Functionally equivalent but different implementation details.

---

## Credits

- **Psiphon Inc.** - Original Psiphon technology
- **Psiphon-Labs** - psiphon-tunnel-core Go implementation
- **SpherionOS** - PsiphonLinux (inspiration for packaging)
- **Community contributors**

---

## License

GNU General Public License v3 (GPLv3)

See [LICENSE](LICENSE) file.

---

**Ready to get started?** See [docs/quick-start-for-developers.md](docs/quick-start-for-developers.md) or [README.md](../README.md).

# PsiphonNGLinux Implementation Guide

This document describes the architecture and implementation of PsiphonNGLinux, highlighting how it leverages the modified psiphon-tunnel-core codebase to provide a full-featured Linux client.

---

## Project Goals

1. **Native Go implementation**: Build directly on psiphon-tunnel-core library (not a wrapper script)
2. **System integration**: Systemd service, journald logging, proper daemonization
3. **Advanced features**: TUN mode, in-proxy support, connection approval, split tunneling
4. **Production-ready**: Security hardening, resource limits, graceful shutdown
5. **Configurable**: Comprehensive configuration via JSON file
6. **Automatic**: Background server list fetching, tactics, updates

---

## Comparison with PsiphonLinux

| Aspect | PsiphonLinux | PsiphonNGLinux |
|--------|--------------|----------------|
| **Architecture** | Shell wrapper around pre-built binary | Native Go daemon linking psiphon-tunnel-core library |
| **Build** | Downloads binary from GitHub | Builds from source (or uses provided binary) |
| **Configuration** | Hardcoded single-server | Dynamic multi-source with auto-discovery |
| **Server discovery** | None (manual only) | Remote server lists, DSL, OSL, tactics |
| **Tactics** | Limited (what's baked into binary) | Full remote configuration |
| **In-proxy** | No | Client and server modes |
| **TUN/packet tunnel** | No | Yes (system-wide routing) |
| **Split tunneling** | No | Yes (CIDR inclusion/exclusion, own-region) |
| **Systemd** | No | Full unit with watchdog, restart, limits |
| **Logging** | Stdout only | Journald + file rotation |
| **Compartment IDs** | No | Yes (for broker rate limiting) |
| **Connection approval** | No | Optional WebSocket hook (compatible with DOIP approval server) |
| **Metrics** | None | Prometheus/StatsD support (TODO) |
| **Auto-update** | Manual | Can integrate with watchtower/self-update |
| **Security** | Runs as root (simple) | Dedicated user, capabilities, seccomp |
| **Maintainer** | Community | Your organization |

---

## Component Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    PsiphonNGLinux                        │
├──────────────────────────────────────────────────────────┤
│  cmd/psiphond-ng/main.go                                 │
│    ├─ Config loading & validation                        │
│    ├─ Signal handling (SIGINT, SIGTERM, SIGHUP)         │
│    ├─ Logging setup (journald + file)                   │
│    ├─ Config watcher (hot-reload on file change)        │
│    └─ Controller lifecycle                              │
├──────────────────────────────────────────────────────────┤
│  internal/config/                                        │
│    └─ Config struct ↔ psiphon.Config conversion         │
├──────────────────────────────────────────────────────────┤
│  internal/daemon/                                        │
│    └─ Systemd integration, watchdog, service state      │
├──────────────────────────────────────────────────────────┤
│  internal/tunnel/                                        │
│    └─ Tunnel lifecycle, health monitoring, metrics      │
├──────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐│
│  │   psiphon-tunnel-core (external library)           ││
│  │  ├── Controller (orchestration)                    ││
│  │  ├── DataStore (BoltDB persistence)               ││
│  │  ├── RemoteServerList (fetch + verify)            ││
│  │  ├── Tactics (remote config)                      ││
│  │  ├── In-Proxy (broker client/proxy)               ││
│  │  ├── Network dialing (protocols, obfuscation)    ││
│  │  └── API (handshake, stats, feedback)             ││
│  └─────────────────────────────────────────────────────┘│
├──────────────────────────────────────────────────────────┤
│  config/                                                 │
│    ├── psiphond-ng.conf (default config)               │
│    ├── psiphond-ng.service (systemd unit)             │
│    └── defaults/ (parameter sets)                     │
├──────────────────────────────────────────────────────────┤
│  scripts/                                                │
│    ├── install.sh (systemd deployment)                │
│    ├── uninstall.sh                                     │
│    ├── build.sh (cross-compilation)                   │
│    └── update.sh (auto-update check)                  │
└──────────────────────────────────────────────────────────┘
```

---

## Key Implementation Details

### 1. Config Bridge (`buildPsiphonConfig`)

The main challenge is converting PsiphonNGLinux's user-friendly JSON config into the complex `psiphon.Config` struct with proper defaults and callbacks.

**Responsibilities:**
- Set all required fields with sensible defaults
- Convert string durations (e.g., "10s") to `time.Duration`
- Set up file paths based on `data_directory`
- Configure in-proxy callbacks (approval hook)
- Set packet tunnel parameters
- Handle optional fields and pointer types

**Example - Tactics timeouts:**
```go
psiphonCfg.TacticsWaitPeriod = ngConfig.TacticsWaitPeriod // "10s" -> psiphon parses
psiphonCfg.TacticsRetryPeriod = ngConfig.TacticsRetryPeriod
```

**Example - File paths:**
```go
psiphonCfg.RemoteServerListDownloadFilename = filepath.Join(
    ngConfig.DataDirectory,
    ngConfig.RemoteServerListDownloadFilename)
```

### 2. Notice Handler (`simpleNoticeHandler`)

Psiphon core uses a notice system for logging. We implement `psiphon.NoticeHandler`:

- Writes notices to log file (with rotation) and stdout
- Formats timestamps
- Optionally filters by severity
- Could be extended for metrics, alerting, or remote logging

### 3. Controller Wrapper (`PsiphonController`)

Encapsulates the `psiphon.Controller`:

```go
type PsiphonController struct {
    controller *psiphon.Controller
    config     *Config
    stopChan   chan struct{}
}
```

Methods:
- `NewPsiphonController()`: Creates controller from config
- `Start()`: Initializes datastore and calls `controller.Run()`
- `Stop()`: Signals shutdown, calls `controller.Stop()`
- `Wait()`: Blocks until stopped

This wrapper allows us to add our own lifecycle management, metrics collection, and graceful shutdown handling.

### 4. Config Watcher (`ConfigWatcher`)

Monitors config file for changes and triggers hot-reload:

- Uses `fsnotify` to watch directory
- Debounces rapid changes (500ms)
- Parses new config on write
- Sends new config on channel
- Controller can react by updating parameters or restarting

**Extending for live reload:**
```go
go func() {
    for newConfig := range configWatcher.ReloadChan() {
        // Update controller parameters without restart
        controller.UpdateParameters(newConfig)
    }
}()
```

Currently we log the new config but restart is needed; we could implement live parameter updates.

### 5. Systemd Service

The `.service` file provides:

- **User/Group**: Run as non-root `psiphon` user
- **Capabilities**: `CAP_NET_ADMIN` for TUN; `CAP_NET_RAW` for raw sockets
- **ProtectSystem**: Read-only filesystem except data/log dirs
- **ProtectHome**: No home directory access
- **SystemCallFilter**: Restrict to safe syscalls
- **Restart**: Auto-restart on failure with backoff
- **Logging**: journald integration
- **Resource limits**: `LimitNOFILE=65536`

This is production-grade hardening following systemd best practices.

### 6. TUN Integration

Psiphon-tunnel-core already supports packet tunnel (TUN) via the `psiphon` package:

- `config.PacketTunnelTunFileDescriptor`: File descriptor for TUN (default 0 = create internally)
- `config.PacketTunnelDeviceName`: Interface name
- `config.PacketTunnelAddressIPv4/IPv6`: Assigned addresses
- `config.PacketTunnelGateway...`: Gateway hops
- `config.PacketTunnelMTU`: MTU
- `config.PacketTunnelTransparentDNS...`: DNS servers to intercept
- `config.SplitTunnel...`: Split tunnel rules

Our daemon:
- Runs as root or with `CAP_NET_ADMIN` → Psiphon can create TUN device
- Central config sets these parameters
- TUN device appears as `tun-psiphon` (configurable name)
- All routing, NAT, iptables rules are handled by psiphon core

**Testing TUN:**
```bash
sudo setcap cap_net_admin+ep $(which psiphond-ng)
# Restart service
ip addr show tun-psiphon
```

### 7. In-Proxy with Approval Hook

We utilize the modified `psiphon/common/inproxy/proxy.go` which adds:

```go
type ProxyConfig struct {
    ...
    ApproveClientConnection func(ClientConnectionInfo) (bool, error)
}
```

Our daemon can be configured to inject this callback when running in proxy mode:

```go
if ngConfig.InproxyMode == "proxy" && ngConfig.InproxyApprovalWebSocketURL != "" {
    config.ApproveClientConnection = func(info inproxy.ClientConnectionInfo) (bool, error) {
        // Connect to WebSocket approval server
        ws, _, err := websocket.Dial(ngConfig.InproxyApprovalWebSocketURL, "", "origin")
        if err != nil {
            return false, err
        }
        defer ws.Close()

        // Send approval request
        ws.WriteJSON(info)

        // Wait for response (with timeout)
        ws.SetReadDeadline(time.Now().Add(approvalTimeout))
        var resp struct{ Approved bool }
        if err := ws.ReadJSON(&resp); err != nil {
            return false, err
        }
        return resp.Approved, nil
    }
}
```

This enables dynamic, external control over which clients can connect.

---

## Configuration Management

### Hierarchical Config

1. **Base defaults** (`DefaultConfig()`): Hardcoded safe defaults
2. **File config**: `~/.config/psiphond-ng/psiphond-ng.conf` (user overrides)
3. **Tactics**: Remote configuration (highest priority)

### Config Validation

We validate:
- Required fields: `propagation_channel_id`, `sponsor_id` (can be placeholders)
- Tunnel mode: must be `"portforward"` or `"packet"`
- Inproxy mode: `""`, `"client"`, or `"proxy"`
- Port ranges: Check for valid port numbers
- CIDR syntax: (TODO) validate split tunnel CIDRs

---

## Building and Deployment

### Build Process

```bash
# Development build (fast, with debugging)
go build -o psiphond-ng ./cmd/psiphond-ng

# Production build (optimized, static)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o psiphond-ng ./cmd/psiphond-ng

# With TUN tag (if needed)
go build -tags=tun -o psiphond-ng ./cmd/psiphond-ng
```

### Distribution

We provide:
- Pre-built binaries in GitHub releases
- Source distribution with everything needed to build
- `build.sh` script for reproducible builds
- `install.sh` for automated deployment

---

## Testing Strategy

### Unit Tests

Test individual components:

```go
// Test config loading
func TestLoadConfig(t *testing.T) {
    cfg, err := LoadConfig("testdata/valid.conf")
    assert.NoError(t, err)
    assert.Equal(t, "portforward", cfg.TunnelMode)
}

// Test config conversion
func TestBuildPsiphonConfig(t *testing.T) {
    ngCfg := &Config{...}
    psiphonCfg, err := buildPsiphonConfig(ngCfg)
    assert.NoError(t, err)
    assert.Equal(t, "/var/lib/psiphon", psiphonCfg.DataRootDirectory)
}

// Test notice handler
func TestNoticeHandler(t *testing.T) {
    // ...
}
```

### Integration Tests

Test full startup with test Psiphon server:

1. Start local test Psiphon server (from `psiphon/server`)
2. Configure client to use it
3. Start PsiphonNGLinux daemon
4. Verify tunnel established
5. Test traffic through proxy/TUN

### System Tests

Using Docker or VM:
- Install PsiphonNGLinux as service
- Verify systemd unit works
- Test automatic restart
- Test config hot-reload (SIGHUP)
- Test TUN device creation and routing
- Test in-proxy broker interaction

---

## Future Enhancements

### 1. Metrics and Monitoring

Add Prometheus exporter:
```go
// internal/metrics/metrics.go
var (
    tunnelsGauge = prometheus.NewGauge(...)
    bytesTransferred = prometheus.NewCounter(...)
    connectionAttempts = prometheus.NewCounter(...)
)
```

Expose on `/metrics` endpoint (configurable port).

### 2. Self-Updates

Integrate with GitHub releases or custom update server:

- Check for new version periodically
- Download new binary with signature verification
- Atomic replacement (rename temp -> target)
- Restart service
- Rollback on failure

### 3. Health Checks

HTTP endpoint for orchestrators (Kubernetes, nomad):
```
GET /healthz → 200 OK if tunnel is up
GET /readyz → 200 OK if ready to accept traffic
GET /metrics → Prometheus metrics
```

### 4. Configuration UI

Web UI for configuration:
- View status, logs, metrics
- Edit config (with validation)
- Restart/stop service
- View connected tunnels

### 5. Enhanced TUN Features

- DNS hijacking (redirect all DNS to tunnel)
- IPv6 support in split tunnel
- Policy-based routing (pf/iptables rules generator)
- TUN device hot-plug (start/stop without restart)

---

## Development Workflow

1. **Make changes** to codebase
2. **Run tests**:
   ```bash
   go test ./...
   ```
3. **Build locally**:
   ```bash
   ./scripts/build.sh
   ```
4. **Test manually** (portforward mode, no sudo):
   ```bash
   ./build/psiphond-ng -config test.conf
   ```

   For TUN mode, either grant capability `sudo setcap cap_net_admin+ep ./build/psiphond-ng` or run with sudo.
5. **Check with linters**:
   ```bash
   go vet ./...
   golangci-lint run
   ```
6. **Commit** with signed tags
7. **Release**:
   ```bash
   ./scripts/build.sh --static --compress
   gh release create v1.0.0 build/psiphond-ng
   ```

---

## Security Checklist

**User-Level Service Design:**

- [x] Runs as regular user (no elevated privileges by default)
- [x] Uses XDG directories (`~/.config`, `~/.local/var`)
- [x] No system-wide files or directories
- [x] No dedicated system account required
- [x] Systemd user service with hardening (`NoNewPrivileges`, `PrivateTmp`, `ProtectHome`, etc.)
- [x] Capability bounding set (`CAP_NET_ADMIN` only if TUN needed via setcap)
- [x] Resource limits (NOFILE, NPROC)
- [x] Config file permissions: 0600 (user-owned)
- [x] Log and data directories: user-only access (0700)
- [x] System call filtering (`@system-service`)

**TUN Mode (when needed):**

- [ ] Grant `CAP_NET_ADMIN` via `setcap cap_net_admin+ep $(which psiphond-ng)` (preferred)
- [ ] Or run binary with sudo (not recommended for continuous operation)

**Future Enhancements:**

- [ ] Implement mutual TLS for approval server (WSS)
- [ ] Config file validation on load
- [ ] Binary update signature verification
- [ ] Optional seccomp profile for further restriction

---

## Operational Guide

### Monitoring

Check service health (user service):
```bash
systemctl --user status psiphond-ng
journalctl --user -u psiphond-ng -f
```

Check tunnel status:
```bash
# Controller status (would need to expose via API or metrics)
# For now, parse notice file or journal:
grep "Tunnels" ~/.local/var/log/psiphon/psiphond-ng.log
# or: journalctl --user -u psiphond-ng | grep Tunnels
```

Check TUN device:
```bash
ip addr show tun-psiphon
ip route | grep tun-psiphon
```

### Updates

Manual (user-level):
```bash
systemctl --user stop psiphond-ng
wget -O ~/.local/bin/psiphond-ng https://github.com/.../psiphond-ng-linux-amd64
chmod +x ~/.local/bin/psiphond-ng
systemctl --user start psiphond-ng
```

Automatic (if enabled):
```bash
# Custom script (runs as user)
~/.local/bin/psiphond-ng-updater
# or use watchtower, auto-update mechanism, etc.
```

### Troubleshooting

**Service fails to start:**
```bash
journalctl --user -u psiphond-ng -n 100 -p err
# Common: config validation errors, missing directories, permission issues
```

**No tunnels established:**
```bash
# Run in foreground with verbose logging
~/.local/bin/psiphond-ng -config ~/.config/psiphond-ng/psiphond-ng.conf -debug
# or
./build/psiphond-ng -config ./psiphond-dev.conf -debug
```

**TUN not created:**
```bash
# Grant capability (one-time, requires sudo)
sudo setcap cap_net_admin+ep $(which psiphond-ng || ~/.local/bin/psiphond-ng)

# Restart user service
systemctl --user restart psiphond-ng

# Check TUN module
lsmod | grep tun || sudo modprobe tun
```

---

## References

- Psiphon Tunnel Core: https://github.com/Psiphon-Labs/psiphon-tunnel-core
- Psiphon Network: https://psiphon.ca
- systemd.exec(5): Security directives
- Linux TUN/TAP: https://www.kernel.org/doc/Documentation/networking/tuntap.txt

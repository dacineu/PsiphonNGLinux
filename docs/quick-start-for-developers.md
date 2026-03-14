# Quick Start Guide for Developers

This guide helps you get PsiphonNGLinux up and running quickly for development and testing.

---

## Prerequisites

- **Go 1.21+**: https://go.dev/dl/
- **Linux**: Any modern distro (Ubuntu, Debian, Fedora, Arch, etc.)
- **Git**: For cloning the repository
- **Root/sudo**: For TUN mode testing (optional)

---

## 1. Clone and Build

```bash
# Clone repository
git clone https://github.com/yourusername/PsiphonNGLinux.git
cd PsiphonNGLinux

# Download dependencies
go mod download
go mod verify

# Build
./scripts/build.sh
# Binary is at: build/psiphond-ng
```

**Quick build (without script):**
```bash
go build -o psiphond-ng ./cmd/psiphond-ng
```

---

## 2. Basic Configuration

Copy and edit the default config:

```bash
cp config/psiphond-ng.conf psiphond-dev.conf
nano psiphond-dev.conf
```

**Minimum required fields:**

```json
{
  "tunnel_mode": "portforward",
  "propagation_channel_id": "FFFFFFFFFFFFFFFF",
  "sponsor_id": "FFFFFFFFFFFFFFFF"
}
```

**Note:** These F's are placeholders. For production, obtain real IDs from Psiphon.

---

## 3. Run in Port Forward Mode (No Root)

This is the simplest way to test:

```bash
./psiphond-ng -config psiphond-dev.conf
```

**What happens:**
- Creates local SOCKS5 proxy on 127.0.0.1:1080
- Creates local HTTP proxy on 127.0.0.1:8080
- Fetches remote server list
- Establishes tunnel to a Psiphon server
- Logs to stdout

**Test connectivity:**
```bash
# In another terminal, test the proxy
curl -x http://127.0.0.1:8080 https://checkip.amazonaws.com
# Should return an IP different from your real IP

# Test SOCKS
curl --socks5 127.0.0.1:1080 https://checkip.amazonaws.com
```

**Check logs:**
You should see notices like:
```
[INFO] ListeningHttpProxyPort: {"port":8080}
[INFO] ListeningSocksProxyPort: {"port":1080}
[INFO] Tunnels: {"count":1}
```

---

## 4. Run in TUN Mode (Packet Tunnel)

TUN mode requires elevated privileges. You have two options:

### Option A: Grant capability (recommended)

```bash
# Grant CAP_NET_ADMIN to binary (one-time)
sudo setcap cap_net_admin+ep $(which psiphond-ng 2>/dev/null || ./psiphond-ng)

# Run as normal user
./psiphond-ng -config psiphond-dev.conf
```

Edit `psiphond-dev.conf` first:
```json
{
  "tunnel_mode": "packet",
  "packet_tunnel_device_name": "tun-psiphon",
  "packet_tunnel_address_ipv4": "10.0.0.2/30",
  "packet_tunnel_gateway_ipv4": "10.0.0.1",
  "packet_tunnel_dns_ipv4": "8.8.8.8"
}
```

### Option B: Run with sudo (not recommended for regular use)

```bash
sudo ./psiphond-ng -config psiphond-dev.conf
```

**Test TUN interface:**

```bash
# Check interface
ip addr show tun-psiphon
# Should show: inet 10.0.0.2/30

# Check routes
ip route | grep tun

# Force traffic through TUN
curl --interface tun-psiphon https://checkip.amazonaws.com
```

---

## 5. Systemd User Service (Recommended)

PsiphonNGLinux is designed to run as a user service with XDG directories.

### Option A: Built-in installer (easiest)

```bash
# From project root (after building)
./psiphond-ng service
```

This will:
1. Copy binary to `~/.local/bin/psiphond-ng`
2. Create config at `~/.config/psiphond-ng/psiphond-ng.conf`
3. Install systemd user service to `~/.config/systemd/user/`
4. Enable and start the service
5. Create data and log directories under `~/.local/var/`

### Option B: Manual install

```bash
mkdir -p ~/.local/bin
cp build/psiphond-ng ~/.local/bin/
mkdir -p ~/.config/systemd/user
cp config/psiphond-ng-user.service ~/.config/systemd/user/psiphond-ng.service
systemctl --user daemon-reload
systemctl --user enable --now psiphond-ng
```

**Manage service:**
```bash
systemctl --user status psiphond-ng
systemctl --user start psiphond-ng
systemctl --user stop psiphond-ng
journalctl --user -u psiphond-ng -f
```

**Note:** The old `scripts/install.sh` (system-wide) is deprecated. Use `./psiphond-ng service` for user-level installation.

---

## 6. Debugging

### Enable Debug Logging

Edit config:
```json
{
  "log_level": "debug"
}
```

Or set environment variable:
```bash
LOG_LEVEL=debug ./psiphond-ng -config psiphond-dev.conf
```

Verbose logs show:
- Server candidate list
- Connection attempts and failures
- Protocol negotiation
- API requests/responses
- Obfuscation details

### Increase Connection Worker Pool

Faster debugging of connection establishment:

```json
{
  "connection_worker_pool_size": 20,
  "establish_tunnel_timeout_seconds": 180
}
```

### Use Local Test Server

Instead of connecting to real Psiphon network, test with local psiphond:

1. Start local server (from psiphon-tunnel-core Server/):
```bash
cd ../psiphon-tunnel-core/Server
go build -o psiphond ./cmd/psiphond
./psiphond -ipaddress 127.0.0.1 -protocol OSSH:9999 listen
```

2. Get server entry from output:
```json
{
  "tag": "test",
  "sshPort": 9999,
  "sshHostKey": "..."
}
```

3. Create `TargetServerEntry` in config:
```go
// Encode as base64
echo '{"tag":"test","sshPort":9999,"sshHostKey":"..."}' | base64 -w0
```

4. Use in config:
```json
{
  "target_server_entry": "base64-here",
  "disable_remote_server_list_fetcher": true,
  "disable_tactics": true
}
```

---

## 7. Working with Tactics

Tactics are remote configuration payloads delivered by Psiphon servers.

**Enable tactics in dev:**
```json
{
  "disable_tactics": false,
  "tactics_wait_period": "5s"
}
```

**Trigger tactics fetch:**
1. Start with no tactics
2. Controller will fetch tactics via API once tunnel is established
3. Check logs: `Applying tactics payload`

**View received tactics:**
Check datastore:
```bash
ls /var/lib/psiphon/datastore/
# Contains tactics.bbolt
```

---

## 8. In-Proxy Mode (Broker)

For testing in-proxy, you need:

1. **Broker server** (deploy separately or use local test)
2. **Proxy server** (run psiphond-ng in proxy mode)
3. **Client** (run another psiphond-ng in client mode)

### Quick Test with Approval Server

1. **Start approval server** (from psiphon-tunnel-core):
```bash
cd ../psiphon-tunnel-core/scripts/ws-approval-server
go run ./main.go
# Runs on ws://localhost:8080/approve
```

2. **Configure proxy**:
```json
{
  "inproxy_mode": "proxy",
  "inproxy_broker_server_addresses": ["localhost:443"],  // your broker
  "inproxy_max_common_clients": 5,
  "inproxy_approval_websocket_url": "ws://localhost:8080/approve"
}
```

3. **Configure client**:
```json
{
  "inproxy_mode": "client",
  "inproxy_broker_server_addresses": ["localhost:443"],
  "inproxy_compartment_id": "test-compartment"
}
```

4. Run both and watch logs.

---

## 9. Common Development Tasks

### Add a new config field

1. Add to `Config` struct
2. Update `DefaultConfig()` with default value
3. Update `buildPsiphonConfig()` to map to `psiphon.Config`
4. Update README and config example
5. Add validation in `LoadConfig()` if needed

### Modify in-proxy behavior

- See `psiphon/common/inproxy/proxy.go` for approval hook
- See `psiphon/common/inproxy/broker.go` for client/broker protocol
- See `psiphon/common/inproxy/session.go` for session management

### Add metrics

Create `internal/metrics/`:
```go
type Metrics struct {
    ActiveTunnels prometheus.Gauge
    BytesSent prometheus.Counter
    // ...
}

// Expose HTTP endpoint
func (m *Metrics) Handler() http.HandlerFunc { ... }

// Export Prometheus format
```

Modify `main.go` to start metrics server.

---

## 10. Testing Checklist

Before submitting PR, test:

- [ ] Binary builds without errors
- [ ] Starts without crashing (check config)
- [ ] Establishes tunnel in portforward mode
- [ ] Establishes tunnel in TUN mode (with capability or sudo)
- [ ] User service installs and starts
- [ ] Logs written to file and journald
- [ ] Graceful shutdown on SIGTERM
- [ ] Config hot-reload on SIGHUP (TODO)
- [ ] Proxy ports listening and accepting connections
- [ ] TUN device created with correct IP
- [ ] Traffic routes correctly (no leaks)

---

## Troubleshooting

### "Failed to create tun device: operation not permitted"

Run as root or:
```bash
sudo setcap cap_net_admin+ep psiphond-ng
```

### "No server entries available"

- Check network connectivity
- Increase `EstablishTunnelTimeout`
- Set `DisableRemoteServerListFetcher: false`
- Add a `TargetServerEntry` to config for debugging
- Check logs for "fetching remote server list"

### Ports already in use

Change `LocalHttpProxyPort` and `LocalSocksProxyPort` to unused ports:
```bash
ss -tlnp | grep LISTEN
```

### Binary won't start after install

Check permissions (user service):
```bash
ls -l ~/.local/bin/psiphond-ng
# Should be -rwxr-xr-x (0755)
```

Check systemd user journal:
```bash
journalctl --user -u psiphond-ng -n 50
```

---

## Resources

- **psiphon-tunnel-core**: https://github.com/Psiphon-Labs/psiphon-tunnel-core
- **GoDoc** (if published): https://pkg.go.dev/github.com/yourusername/PsiphonNGLinux
- **Systemd docs**: https://www.freedesktop.org/software/systemd/man/
- **TUN/TAP**: https://www.kernel.org/doc/Documentation/networking/tuntap.txt

---

## Got Help?

- Open an issue: https://github.com/yourusername/PsiphonNGLinux/issues
- Check existing documentation in `docs/`
- Review psiphon-tunnel-core issues and PRs for context

---

**Happy coding!**

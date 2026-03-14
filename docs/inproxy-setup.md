# In-Proxy Mode Setup Guide

In-proxy (or "broker") mode allows PsiphonNGLinux to operate as either:
1. A **client** that connects to a broker to get matched with an available proxy
2. A **proxy server** that relays traffic for clients via WebRTC

This architecture provides enhanced censorship resistance by using intermediary proxies.

---

## Architecture

```
Client                          Broker                          Proxy
  │                                │                               │
  │─── ClientOffer ───────────────>│                               │
  │<── ProxyAnnounce ──────────────│                               │
  │                                │─── ProxyAnswer ──────────────>│
  │                                │<─────────(WebRTC offer)──────│
  │<─────────────────WebRTC───────────────────────────────────────>│
  │────────────────Request (dest)─────────────────────────────────>│
  │<──────────────────Response────────────────────────────────────│
```

**Key points:**
- Broker knows only proxy ID, not final destination
- Client and proxy communicate directly via WebRTC (with broker signaling)
- Proxy sees destination but not client identity (if configured properly)
- All communication is obfuscated and padded

---

## Broker Server

The broker is a separate service (not included in psiphon-tunnel-core). You need to deploy a broker server. See `psiphon/common/inproxy/broker.go` for the broker protocol specification.

Typical broker deployment:
```bash
# Broker runs on a server with public IP
./broker -address 0.0.0.0:443
```

---

## Running as a Client

### Configuration

```json
{
  "tunnel_mode": "portforward",
  "inproxy_mode": "client",
  "inproxy_broker_server_addresses": ["broker.example.com:443"],
  "inproxy_compartment_id": "personal-compartment-id",  // optional
  "propagation_channel_id": "YOUR_CHANNEL_ID",
  "sponsor_id": "YOUR_SPONSOR_ID",
  "egress_region": "US"
}
```

### Deploying (User-Level)

```bash
# Build
./scripts/build.sh

# Install as user service (no sudo!)
./psiphond-ng service

# Or use the deployment installer:
# cd deployment && ./install.sh

# Edit config
nano ~/.config/psiphond-ng/psiphond-ng.conf
# Set inproxy_mode and broker addresses

# Start service
systemctl --user start psiphond-ng
```

### Client Behavior

1. Psiphon contacts broker via `ClientOffer`
2. Broker matches client to a proxy (based on compartment, region, load)
3. Client gets proxy info and performs WebRTC handshake
4. Client requests proxy to connect to destination (Psiphon server)
5. Traffic relayed through proxy to actual Psiphon server

**Note:** The client still needs to know which Psiphon servers are available. This can come from:
- Embedded server entries
- Remote server list fetch (tunneled through the proxy or directly)
- Tactics delivered via API (after tunnel established)

---

## Running as a Proxy (In-Proxy Server)

### Prerequisites

1. **Psiphon server**: You need access to a Psiphon server (psiphond). The proxy will forward client traffic to this server.
2. **Public IP**: Proxy must have a public IP address reachable by broker and clients.
3. **Broker registration**: Proxy's public key must be registered with the broker.

### Configuration

```json
{
  "inproxy_mode": "proxy",

  // Broker
  "inproxy_broker_server_addresses": ["broker.example.com:443"],

  // Proxy identity (generated keys)
  "inproxy_session_private_key": "base64-encoded-ed25519-private-key",
  "inproxy_session_root_obfuscation_secret": "base64-encoded-32-byte-secret",

  // Limits
  "inproxy_max_common_clients": 10,
  "inproxy_max_personal_clients": 5,
  "inproxy_max_common_client_compartment_ids": [],
  "inproxy_max_personal_client_compartment_ids": [],
  "inproxy_rate_limiting": true,

  // Egress
  "inproxy_egress_interface": "eth0",  // interface to connect to Psiphon server

  // Optional: approval hook
  "inproxy_approval_websocket_url": "ws://localhost:8080/approve",
  "inproxy_approval_timeout": "5s"
}
```

### Setup Steps

#### 1. Generate Keys

```go
// Generate a key pair (private key for proxy, public key to register with broker)
import (
    "crypto/ed25519"
    "encoding/base64"
)

privateKey, publicKey, err := ed25519.GenerateKey(rand.Reader)
if err != nil {
    log.Fatal(err)
}

// Encode for config
privateKeyB64 := base64.StdEncoding.EncodeToString(privateKey.Seed())
// Public key will be derived from private key, or you can export:
publicKeyBytes := publicKey.Bytes()
publicKeyB64 := base64.StdEncoding.EncodeToString(publicKeyBytes)
```

Register `publicKeyB64` with your broker.

#### 2. Generate Root Obfuscation Secret

```go
secret := make([]byte, 32)
if _, err := rand.Read(secret); err != nil {
    log.Fatal(err)
}
secretB64 := base64.StdEncoding.EncodeToString(secret)
```

Add both to config.

#### 3. Configure Egress

Set `inproxy_egress_interface` to the network interface that should be used to connect to Psiphon servers.

This interface should have:
- Internet connectivity
- Ability to reach Psiphon servers (ports 443, etc.)
- Proper routing

#### 4. Build and Install

```bash
./scripts/build.sh -t inproxy

# User-level install (no sudo!)
./psiphond-ng service
# Or manual:
# mkdir -p ~/.local/bin
# cp build/psiphond-ng ~/.local/bin/
# mkdir -p ~/.config/systemd/user
# cp config/psiphond-ng-user.service ~/.config/systemd/user/psiphond-ng.service
# systemctl --user daemon-reload
# systemctl --user enable --now psiphond-ng

#### 5. Verify Proxy is Running

Check logs:
```bash
journalctl --user -u psiphond-ng -f
```

Look for:
```
Proxy available with public key: ...
```

And broker logs should show the proxy connecting.

---

## Connection Approval (Optional)

For access control, you can configure the proxy to approve each client connection via a WebSocket server.

### Setup Approval Server

You have two options for the approval server:

#### Option 1: Simple WebSocket Server (for testing)

Use the WebSocket approval server from `scripts/ws-approval-server/` (included in psiphon-tunnel-core repository).

```bash
cd /path/to/psiphon-tunnel-core/scripts/ws-approval-server
go mod download
go build -o approval-server main.go
./approval-server -addr :8080
```

This auto-approves all connections by default. Customize the code for your approval logic.

#### Option 2: DOIP Approval Server (production-ready)

For production deployments with dynamic configuration and strict field validation, use the standalone DOIP approval server (included in this repository's parent directory or available separately):

```bash
# Build
go build -o doip-approval-server ./doip-approval-server/cmd/doip-approval-server

# Run with config
./doip-approval-server -config ./config.json -admin-port :8080 -ws-port :8443

# Or install as systemd user service
cd doip-approval-server/deployment
./install.sh
systemctl --user start doip-approval-server
```

The DOIP server provides:
- Dynamic strict field configuration via JSON file
- Admin HTTP API for runtime updates
- Persistent logging (when configured)
- Systemd service management

See `doip-approval-server/README.md` for details.

### Configure Psiphon Proxy

```json
{
  "inproxy_approval_websocket_url": "ws://localhost:8080/approve",
  "inproxy_approval_timeout": "5s"
}
```

### Approval Request Format

The WebSocket server receives JSON:
```json
{
  "connection_id": "abc123",
  "client_region": "US",
  "destination": "server.example.com:443",
  "network_protocol": "TCP"
}
```

And must respond:
```json
{
  "approved": true,
  "timestamp": "2024-01-15T10:30:00Z",
  "connection_id": "abc123"
}
```

If no response or `approved: false`, the connection is rejected.

### Implement Custom Approval Logic

Your approval server can:
- Check client region against allowlist
- Rate limit by IP or compartment ID
- Integrate with authentication systems
- Log to database
- Implement time-based access

Example:
```go
// Simplified approval server
func approveHandler(conn net.Conn) {
    decoder := json.NewDecoder(conn)
    for {
        var req ApprovalRequest
        if err := decoder.Decode(&req); err != nil {
            break
        }

        approved := false
        if req.ClientRegion == "US" || req.ClientRegion == "CA" {
            approved = true
        }

        resp := ApprovalResponse{
            Approved: approved,
            Timestamp: time.Now().UTC().Format(time.RFC3339),
            ConnectionID: req.ConnectionID,
        }

        json.NewEncoder(conn).Encode(resp)
    }
}
```

---

## Troubleshooting

### Proxy won't connect to broker

- Check broker address/port is correct
- Verify broker is running and accessible: `nc -zv broker.example.com 443`
- Check firewall allows outbound to broker
- Ensure public key is registered with broker

### Client can't find proxy

- Broker may have no available proxies (all at capacity)
- Check region restrictions: client's `egress_region` must match proxy's advertised regions
- Verify compartment IDs if used
- Check broker logs

### WebRTC fails

- Clients/proxies may be behind symmetric NAT
- Broker should provide TURN servers if needed
- Check `iceTransportPolicy` in broker/client configuration
- Enable verbose logging: `"log_level": "debug"`

### Approval timeouts

- Ensure approval server is running and reachable from proxy
- Check `inproxy_approval_timeout` is sufficient
- Approval server should respond quickly (<2s) to avoid client timeouts
- Verify WebSocket URL scheme (`ws://` or `wss://`)

### High proxy CPU/memory

- Reduce `inproxy_max_common_clients` and `inproxy_max_personal_clients`
- Enable rate limiting
- Check for misbehaving clients (excessive bandwidth)
- Consider running multiple proxy instances behind a load balancer

---

## Production Considerations

### High Availability

- Run multiple proxies with different keys registered to broker
- Broker will distribute clients across available proxies
- Monitor proxy health via logs/metrics

### Monitoring

Key metrics:
- `connectedClients` (via ActivityUpdater callback)
- Tunnel establishment success rate
- Approval decisions (approval server logs)
- Network I/O per client

### Security

- Run proxy under dedicated unprivileged user
- Restrict egress interface with firewall
- Enable approval hook to control client access
- Use WSS (secure WebSocket) for approval server
- Rotate keys periodically

### Scaling

- Horizontal: Add more proxy instances
- Vertical: Increase machine resources (CPU, memory, network)
- Broker capacity: Ensure broker can handle proxy announcements from all proxies

---

## References

- Broker protocol: `psiphon/common/inproxy/broker.go`
- Session management: `psiphon/common/inproxy/session.go`
- Proxy server implementation: `scripts/psiphon-inproxy/main.go`
- Approval integration: `docs/inproxy-tun-approval-integration.md`
- Systemd service: `config/psiphond-ng.service`

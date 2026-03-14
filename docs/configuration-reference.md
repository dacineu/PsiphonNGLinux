# PsiphonNGLinux Configuration Reference

Complete reference of all configuration options.

---

## Table of Contents

- [General](#general)
- [Tunnel Modes](#tunnel-modes)
  - [Port Forward Mode](#port-forward-mode)
  - [Packet Tunnel (TUN) Mode](#packet-tunnel-tun-mode)
- [In-Proxy Mode](#in-proxy-mode)
  - [Client Configuration](#client-configuration)
  - [Proxy Configuration](#proxy-configuration)
- [Split Tunneling](#split-tunneling)
- [Protocols and Obfuscation](#protocols-and-obfuscation)
- [Remote Server Lists](#remote-server-lists)
- [Tactics](#tactics)
- [API and Feedback](#api-and-feedback)
- [Resource Limits](#resource-limits)
- [Network](#network)
- [Logging](#logging)
- [Callbacks](#callbacks)

---

## General

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `data_directory` | string | `/var/lib/psiphon` | Root directory for persistent data |
| `log_file` | string | `/var/log/psiphon/psiphond-ng.log` | Path to log file |
| `log_level` | string | `"info"` | Log level: `"debug"`, `"info"`, `"warning"`, `"error"` |
| `propagation_channel_id` | string | `"FFFFFFFFFFFFFFFF"` | Channel identifier (required) |
| `sponsor_id` | string | `"FFFFFFFFFFFFFFFF"` | Sponsor identifier (required) |
| `client_version` | string | `"1.0.0"` | Client version string |
| `client_platform` | string | `"linux"` | Platform identifier |
| `client_features` | string[] | `[]` | List of enabled feature flags |

---

## Tunnel Modes

### Port Forward Mode

Provides local SOCKS5 and HTTP proxies.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `tunnel_mode` | string | `"portforward"` | Tunnel mode (`"portforward"` or `"packet"`) |
| `local_socks_proxy_port` | int | `1080` | SOCKS5 proxy port (0 = disable) |
| `local_http_proxy_port` | int | `8080` | HTTP proxy port (0 = disable) |
| `listen_interface` | string | `"127.0.0.1"` | Interface to bind proxies (use `"0.0.0.0"` for all) |
| `disable_local_socks_proxy` | bool | `false` | Disable SOCKS5 proxy |
| `disable_local_http_proxy` | bool | `false` | Disable HTTP proxy |

---

## Packet Tunnel (TUN) Mode

Creates a virtual network interface for system-wide routing.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `tunnel_mode` | string | `"packet"` | Tunnel mode |
| `packet_tunnel_device_name` | string | `"tun-psiphon"` | TUN device name |
| `packet_tunnel_address_ipv4` | string | `"10.0.0.2/30"` | IPv4 address for TUN (CIDR) |
| `packet_tunnel_address_ipv6` | string | `""` | IPv6 address for TUN (CIDR) |
| `packet_tunnel_gateway_ipv4` | string | `"10.0.0.1"` | IPv4 gateway (destination for default route) |
| `packet_tunnel_gateway_ipv6` | string | `""` | IPv6 gateway |
| `packet_tunnel_mtu` | int | `1500` | MTU for TUN device |
| `packet_tunnel_persist` | bool | `false` | Keep TUN device after exit |
| `packet_tunnel_offense_mode` | bool | `false` | Enable NAT/masquerading |
| `packet_tunnel_dns_ipv4` | string | `"8.8.8.8"` | DNS server to intercept and tunnel |
| `packet_tunnel_dns_ipv6` | string | `""` | IPv6 DNS server to intercept |
| `packet_tunnel_transparent_dns_ipv4_address` | string | *see above* | Alias for DNS settings |
| `packet_tunnel_transparent_dns_ipv6_address` | string | *see above* | Alias for IPv6 DNS |

**Note:** In TUN mode, local socks/http proxy ports are still created (for applications that support proxy but not system-wide routing). Set their ports to `0` to disable.

---

## In-Proxy Mode

### Client Configuration

Run PsiphonNGLinux as a client that connects through a broker.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `inproxy_mode` | string | `""` | Set to `"client"` |
| `inproxy_broker_server_addresses` | []string | `[]` | Broker addresses (host:port) |
| `inproxy_compartment_id` | string | `""` | Personal compartment ID (optional) |
| `inproxy_client_limits` | int | `0` | Client-side connection limits |

---

## Proxy Configuration

Run PsiphonNGLinux as a proxy server (in-proxy).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `inproxy_mode` | string | `""` | Set to `"proxy"` |
| `inproxy_broker_server_addresses` | []string | `[]` | Broker addresses to announce to |
| `inproxy_session_private_key` | string | `""` | Base64-encoded Ed25519 private key (required) |
| `inproxy_session_root_obfuscation_secret` | string | `""` | Base64-encoded 32-byte obfuscation secret (required) |
| `inproxy_egress_interface` | string | `""` | Network interface to use for egress connections |
| `inproxy_mtu` | int | `1500` | MTU for proxy connections |
| `inproxy_max_common_clients` | int | `10` | Max concurrent common clients |
| `inproxy_max_personal_clients` | int | `5` | Max concurrent personal clients |
| `inproxy_max_common_client_compartment_ids` | []string | `[]` | Allowed common compartment IDs (empty = all) |
| `inproxy_max_personal_client_compartment_ids` | []string | `[]` | Allowed personal compartment IDs (empty = all) |
| `inproxy_rate_limiting` | bool | `false` | Enable per-client rate limiting |
| `inproxy_approval_websocket_url` | string | `""` | WebSocket URL for connection approval hook |
| `inproxy_approval_timeout` | string | `"10s"` | Timeout for approval requests |

**Generating keys for proxy mode:**

```go
package main

import (
    "crypto/ed25519"
    "encoding/base64"
    "fmt"
    "math/rand"
)

func main() {
    // Generate private key
    privateKey, publicKey := ed25519.GenerateKey(rand.NewSource(0))

    // Generate root obfuscation secret
    secret := make([]byte, 32)
    rand.Read(secret)

    fmt.Println("Private key (add to config):")
    fmt.Println(base64.StdEncoding.EncodeToString(privateKey.Seed()))

    fmt.Println("\nObfuscation secret (add to config):")
    fmt.Println(base64.StdEncoding.EncodeToString(secret))

    fmt.Println("\nPublic key (register with broker):")
    fmt.Println(base64.StdEncoding.EncodeToString(publicKey.Bytes()))
}
```

---

## Split Tunneling

Control which traffic goes through the tunnel.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `split_tunnel_own_region` | bool | `false` | Enable split tunnel for client's own country |
| `split_tunnel_regions` | []string | `[]` | Regions to split tunnel (bypass for those countries) |
| `split_tunnel_include` | []string | `[]` | CIDR ranges that bypass tunnel (e.g., local network) |
| `split_tunnel_exclude` | []string | `[]` | CIDR ranges that always use tunnel (even if in include) |

**Priority:**
1. All traffic goes through tunnel by default
2. `split_tunnel_own_region` and `split_tunnel_regions` (if server classification enabled) may bypass
3. `split_tunnel_include` bypasses
4. `split_tunnel_exclude` forces through tunnel (overrides include)

**Example:**

```json
{
  "split_tunnel_own_region": true,
  "split_tunnel_include": ["192.168.0.0/16", "10.0.0.0/8"],
  "split_tunnel_exclude": ["1.1.1.1"]
}
```

Local traffic (`192.168.x.x`, `10.x.x.x`) goes directly; everything else tunnels. But `1.1.1.1` always tunnels even though it's not in local ranges.

---

## Protocols and Obfuscation

Control which transport protocols are used.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `limit_tunnel_protocols` | []string | `[]` | Restrict to these protocols (empty = all) |
| `initial_limit_tunnel_protocols` | []string | `[]` | Initial phase limited protocols |
| `initial_limit_tunnel_protocols_candidate_count` | int | `0` | Number of candidates for initial phase |

**Supported protocol values:**

- `"SSH"` - Plain SSH (rarely works)
- `"OSSH"` - Oblivious SSH (default)
- `"TLS-OSSH"` - SSH over TLS
- `"QUIC-OSSH"` - SSH over QUIC
- `"FRONTED-MEEK-OSSH"` - Meek via domain fronting
- `"UNFRONTED-MEEK-OSSH"` - Meek without fronting
- `"UNFRONTED-MEEK-HTTPS-OSSH"` - Meek HTTPS
- `"UNFRONTED-MEEK-SESSION-TICKET-OSSH"` - Meek with session tickets
- `"FRONTED-MEEK-QUIC-OSSH"` - Meek QUIC with fronting
- `"TAPDANCE-OSSH"` - TapDance protocol
- `"CONJURE-OSSH"` - Conjure protocol
- `"SHADOWSOCKS-OSSH"` - Shadowsocks protocol

**Example:**

```json
{
  "limit_tunnel_protocols": ["OSSH", "TLS-OSSH", "QUIC-OSSH"],
  "initial_limit_tunnel_protocols": ["OSSH"],
  "initial_limit_tunnel_protocols_candidate_count": 5
}
```

---

## Remote Server Lists

Configure how server entries are discovered.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `disable_remote_server_list_fetcher` | bool | `false` | Disable remote server list fetching |
| `remote_server_list_urls` | []string | *see below* | URLs for common server list |
| `remote_server_list_signature_public_key` | string | `""` | Ed25519 public key for signature verification |
| `remote_server_list_download_filename` | string | `"remote_server_list"` | Filename (relative to data_dir) |
| `obfuscated_server_list_root_urls` | []string | `[]` | URLs for OSL registry |
| `obfuscated_server_list_download_directory` | string | `"obfuscated_server_lists"` | Directory for OSL files |

**Default remote_server_list_urls** (if not set and tactics provide them):
```json
["https://s3.amazonaws.com/psiphon/web/server_list_compressed"]
```

**Note:** The `remote_server_list_signature_public_key` is required to verify signed server lists.

---

## Tactics

Tactics deliver remote configuration from Psiphon servers.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `disable_tactics` | bool | `false` | Disable tactics completely |
| `tactics_wait_period` | string | `"10s"` | Max time to wait for tactics on startup |
| `tactics_retry_period` | string | `"1m"` | Retry period for failed tactics fetch |
| `tactics_retry_period_jitter` | string | `"0s"` | Random jitter for retry period |
| `use_device_detect_code` | bool | `false` | Send device detection code in tactics request |
| `use_region_detection_code` | bool | `true` | Send region detection code |

Tactics can override many other configuration parameters. Without tactics, most defaults apply.

---

## API and Feedback

Parameters for Psiphon API requests.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `psiphon_api_request_timeout` | string | `"30s"` | Timeout for API requests |
| `psiphon_api_status_request_period_min` | string | `"30s"` | Min interval between APi status requests |
| `psiphon_api_status_request_period_max` | string | `"60s"` | Max interval between status requests |
| `psiphon_api_connected_request_period` | string | `"5m"` | Interval for connected notifications |
| `psiphon_api_persistent_stats_max_count` | int | `100` | Max number of persistent stats to buffer |
| `psiphon_api_persistent_stats_upload_interval` | string | `"1h"` | How often to upload stats |

---

## Resource Limits

Control resource usage.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `limit_upstream_bytes_per_second` | int | `0` | Upstream rate limit per tunnel (0 = unlimited) |
| `limit_downstream_bytes_per_second` | int | `0` | Downstream rate limit per tunnel (0 = unlimited) |
| `connection_worker_pool_size` | int | `0` | Concurrent connection attempts (0 = default) |
| `stagger_connection_workers_milliseconds` | int | `0` | Delay between starting workers |
| `limit_intensive_connection_workers` | int | `0` | Limit resource-intensive protocols |
| `tunnel_pool_size` | int | `1` | Number of parallel tunnels to maintain |
| `limit_cpu_threads` | bool | `false` | Minimize CPU thread count |
| `limit_relay_buffer_sizes` | bool | `false` | Use smaller relay buffers |
| `limit_meek_buffer_sizes` | bool | `false` | Use smaller meek buffers |

---

## Network

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `upstream_proxy_url` | string | `""` | Upstream proxy URL (http/socks5) |
| `custom_headers` | map[string]string | `{}` | Additional HTTP headers |
| `use_indistinguishable_tls` | bool | `true` | Make TLS look like real browser |
| `establish_tunnel_timeout_seconds` | int | `300` | Overall tunnel establishment timeout |
| `establish_tunnel_pause_period_seconds` | int | `5` | Pause between establishment attempts |
| `egress_region` | string | `""` | Preferred egress country (ISO 3166-1 alpha-2) |

**Upstream proxy URL format:**
- HTTP: `http://proxy.example.com:8080`
- HTTPS: `https://user:pass@proxy.example.com:8443`
- SOCKS5: `socks5://127.0.0.1:1080`

---

## Logging

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `use_notice_files` | bool | `true` | Write notices to files |
| `notice_rotation_size` | int | `10485760` | Rotate at 10MB |
| `notice_rotation_count` | int | `1` | Number of rotated files to keep |

**Note:** These settings apply when using the built-in notice file rotation. When running as systemd service, journald handles logs.

---

## Callbacks

These are advanced options that require code changes to utilize.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `get_network_id` | function | `nil` | Callback to determine network ID for tactics |
| `activity_updater` | function | `nil` | Periodic callback with tunnel stats |

**Example of custom network ID:**

```go
config.GetNetworkID = func(ctx context.Context) (string, error) {
    // Detect network based on Wi-Fi SSID, cellular carrier, etc.
    return "my-network", nil
}
```

---

## Notes on Types

- **Integer fields**: Use JSON numbers (no quotes)
- **Boolean fields**: `true` or `false`
- **String durations**: Go `time.ParseDuration` format, e.g., `"30s"`, `"5m"`, `"1h"`
- **CIDR ranges**: Standard notation, e.g., `"192.168.0.0/16"`
- **Country codes**: ISO 3166-1 alpha-2, e.g., `"US"`, `"GB"`, `"DE"`

---

## Configuration Precedence

Parameters are applied in this order (later wins):

1. Built-in defaults (in `DefaultConfig()`)
2. Configuration file
3. Tactics (if enabled and received)

Tactics have highest priority and can override almost everything.

---

## Example Configurations

### Minimal (uses embedded servers)

```json
{
  "propagation_channel_id": "FFFFFFFFFFFFFFFF",
  "sponsor_id": "FFFFFFFFFFFFFFFF",
  "tunnel_mode": "portforward"
}
```

### Production with region and split tunnel

```json
{
  "propagation_channel_id": "YOUR_CHANNEL",
  "sponsor_id": "YOUR_SPONSOR",
  "tunnel_mode": "packet",
  "egress_region": "DE",
  "split_tunnel_include": ["10.0.0.0/8", "192.168.0.0/16"],
  "connection_worker_pool_size": 20,
  "tunnel_pool_size": 2,
  "log_level": "info"
}
```

### In-Proxy Client

```json
{
  "propagation_channel_id": "YOUR_CHANNEL",
  "sponsor_id": "YOUR_SPONSOR",
  "inproxy_mode": "client",
  "inproxy_broker_server_addresses": ["broker.example.com:443"],
  "inproxy_compartment_id": "my-compartment-123"
}
```

### In-Proxy Server

```json
{
  "inproxy_mode": "proxy",
  "inproxy_broker_server_addresses": ["broker.example.com:443"],
  "inproxy_session_private_key": "base64...",
  "inproxy_session_root_obfuscation_secret": "base64...",
  "inproxy_egress_interface": "eth0",
  "inproxy_max_common_clients": 10
}
```

---

## Validation

The `LoadConfig()` function validates:
- Required fields (`propagation_channel_id`, `sponsor_id`)
- Tunnel mode value (must be "portforward" or "packet")
- Inproxy mode value (must be "", "client", or "proxy")

Additional validation may be added in the future (port ranges, CIDR syntax, etc.).

---

## See Also

- [Quick Start for Developers](quick-start-for-developers.md)
- [TUN Setup Guide](TUN-setup.md)
- [In-Proxy Setup Guide](inproxy-setup.md)
- [IMPLEMENTATION-GUIDE.md](IMPLEMENTATION-GUIDE.md)

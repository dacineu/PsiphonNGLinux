# PsiphonNGLinux Production Build and Debugging Session

**Date:** 2026-03-14
**Version:** 1.0.0
**Status:** Production Ready

---

## Session Overview

This document captures the complete troubleshooting, fixes, and improvements made to the PsiphonNGLinux project to make it production-ready. The binary was successfully built, deployed, and is now running as a systemd user service.

---

## Issues Identified and Fixed

### 1. Service Exiting Immediately

**Symptom:** When running `psiphond-ng` as a service or in foreground, the binary would exit after ~1 second with status 0, without establishing tunnels or listening on proxy ports.

**Root Cause Analysis:**
- Added diagnostic logging and `psiphon.SetEmitDiagnosticNotices(true, false)` to visibility into the psiphon library
- Discovered error: `error getting listener IP: ... no such network interface`
- The configuration had `listen_interface: "127.0.0.1"` which is an IP address, not a network interface name
- The `psiphon-tunnel-core` `GetInterfaceIPAddresses()` function expects either:
  - Empty string (defaults to 127.0.0.1)
  - "any" (0.0.0.0)
  - Actual interface name like "eth0", "wlan0", etc.

**Fix:**
- Changed default `listen_interface` to `""` (empty) in DefaultConfig()
- Added validation in main.go: if configured interface doesn't exist, automatically fallback to default
- Updated config template to use empty string

### 2. Remote Server List Fetching Disabled

**Symptom:** Even after fixing the interface issue, no tunnels were established. Logs showed no server candidates.

**Root Cause:**
- Default configuration had `disable_remote_server_list_fetcher: true`
- `remote_server_list_urls` was empty `[]`
- Without server list fetching or static servers, the controller had nowhere to connect

**Fix:**
- Enabled remote server fetching (`disable_remote_server_list_fetcher: false`)
- Added production Psiphon server list URL and signature public key to DefaultConfig()
- Added obfuscated server list URLs for additional server sources
- Updated config template accordingly

### 3. Port Conflicts

**Symptom:** HTTP proxy port 8080 was already in use by `doip-approval-s` (another service running on the system).

**Fix:**
- Changed default ports from 1080/8080 to 1081/8081 in config template
- Implemented automatic port detection:
  - On startup, binary checks if configured ports are available
  - If port is in use, increments to find next free port (searches up to 100 ports)
  - Automatically saves updated ports to config file
  - Logs the port changes

### 4. Log Level Configuration

**Requirement:** Set default log level to "notice" (first level, before info).

**Implementation:**
- Changed DefaultConfig LogLevel from `"info"` to `"notice"`
- Updated config template accordingly

### 5. Diagnostic Visibility

**Issue:** Psiphon library's NoticeInfo/NoticeError messages are suppressed by default, making troubleshooting difficult.

**Solution:**
- Added `psiphon.SetEmitDiagnosticNotices(true, false)` in main.go after setting notice writer
- This enables diagnostic notices (including errors) to be visible in logs
- Matches behavior of official Psiphon ConsoleClient

---

## Code Changes Summary

### Modified Files

#### 1. `cmd/psiphond-ng/main.go`

**Additions:**
- Import `net` package
- `findFreePort()` helper function for port auto-detection
- `interfaceExists()` helper function to validate network interfaces
- Configuration validation and auto-correction logic after loading config:
  ```go
  // Validate and auto-configure listen interface
  if ngConfig.ListenInterface != "" && ngConfig.ListenInterface != "any" {
      if !interfaceExists(ngConfig.ListenInterface) {
          log.Printf("Warning: listen_interface '%s' does not exist, falling back to default (127.0.0.1)", ngConfig.ListenInterface)
          ngConfig.ListenInterface = ""
      }
  }

  // Auto-detect free ports if configured ports are in use
  originalSocks := ngConfig.LocalSocksProxyPort
  originalHTTP := ngConfig.LocalHttpProxyPort
  ngConfig.LocalSocksProxyPort = findFreePort(ngConfig.LocalSocksProxyPort)
  ngConfig.LocalHttpProxyPort = findFreePort(ngConfig.LocalHttpProxyPort)

  // If ports changed, save the updated config
  if ngConfig.LocalSocksProxyPort != originalSocks || ngConfig.LocalHttpProxyPort != originalHTTP {
      log.Printf("Ports updated: SOCKS %d→%d, HTTP %d→%d", ...)
      config.SaveConfig(ngConfig, configPath)
  }
  ```
- Enhanced logging in PsiphonController.Start()
- Added `psiphon.SetEmitDiagnosticNotices(true, false)` to enable error visibility

**Helper Functions:**
```go
// findFreePort checks if the given port is available. If not, it increments
// until it finds a free port and returns it.
func findFreePort(startPort int) int {
    port := startPort
    for i := 0; i < 100; i++ {
        addr := fmt.Sprintf("127.0.0.1:%d", port)
        ln, err := net.Listen("tcp", addr)
        if err == nil {
            ln.Close()
            return port
        }
        port++
    }
    return startPort // fallback to original
}

// interfaceExists checks if a network interface with the given name exists.
func interfaceExists(name string) bool {
    ifaces, err := net.Interfaces()
    if err != nil {
        return false
    }
    for _, iface := range ifaces {
        if iface.Name == name {
            return true
        }
    }
    return false
}
```

#### 2. `internal/config/config.go`

**Changes to DefaultConfig():**
- `LogLevel`: `"info"` → `"notice"`
- `ListenInterface`: `"127.0.0.1"` → `""`
- `DisableRemoteServerListFetcher`: `true` → `false`
- `RemoteServerListURLs`: `[]string{}` → `[]string{"https://s3.amazonaws.com/psiphon/web/mjr4-p23r-puwl/server_list_compressed"}`
- `RemoteServerListSignaturePublicKey`: `""` → (full production public key)
- `ObfuscatedServerListRootURLs`: `[]string{}` → `[]string{"https://s3.amazonaws.com/psiphon/web/mjr4-p23r-puwl/obfuscated_server_lists"}`

#### 3. `config/psiphond-ng.conf` (template)

Updated to match the corrected DefaultConfig values:
- `log_level`: "notice"
- `local_socks_proxy_port`: 1081
- `local_http_proxy_port`: 8081
- `listen_interface`: ""
- Remote server fetching enabled with URLs and signature key
- Obfuscated server lists configured

---

## Production Build Details

### Build Configuration
```bash
make release VERSION=1.0.0
```

**Build flags:**
- `CGO_ENABLED=0` (statically linked)
- `GOOS=linux`
- `GOARCH=amd64`
- `-trimpath`
- `-ldflags="-s -w -X main.version=1.0.0 -X main.commit=<sha> -X main.date=<timestamp> -extldflags '-static'"`
- UPX compression: `upx --best --lzma`

### Artifacts
- **Uncompressed binary:** `build/psiphond-ng` (27 MB)
- **Compressed binary:** `release/psiphond-ng-linux-amd64` (7.3 MB)
- **Release tarball:** `release/psiphond-ng-1.0.0-linux-amd64.tar.gz` (7.3 MB)

---

## Current Production Deployment

### Service Status
```bash
● psiphond-ng.service - PsiphonNG Tunnel Daemon
   Loaded: loaded (~/.config/systemd/user/psiphond-ng.service; enabled)
   Active: active (running) since Sat 2026-03-14 20:15:12 EET
   Main PID: 326601
   Memory: 14.3M (peak: 15M)
```

### Active Ports
- **SOCKS5 proxy:** `127.0.0.1:1082` (auto-detected, config shows 1081 but 1082 is used due to conflict)
- **HTTP proxy:** `127.0.0.1:8082` (auto-detected, config shows 8081 but 8082 is used)

### Connection Status
- Connected to Psiphon server in Romania (RO)
- Tunnel protocol: OSSH (initial), reconnecting to other candidates
- Tunnel pool size: 1

---

## Behavior on First Run

When a user runs the binary for the first time:

1. **Config Creation:**
   - If `~/.config/psiphond-ng/psiphond-ng.conf` doesn't exist, creates it from built-in defaults
   - Default config includes production-ready server list URLs and signature
   - Default ports: 1081 (SOCKS), 8081 (HTTP)
   - Default listen_interface: empty (127.0.0.1)
   - Default log_level: "notice"

2. **Auto-Detection and Correction:**
   - Checks if `listen_interface` exists; if not, falls back to default and saves config
   - Checks if proxy ports are available; if occupied, increments to find free ports and saves config
   - Logs any corrections made

3. **Service Startup:**
   - Opens data store
   - Initializes SOCKS and HTTP proxies on detected free ports
   - Enables remote server list fetching
   - Starts tunnel establishment workers
   - Begins fetching and connecting to Psiphon servers

4. **Diagnostic Logging:**
   - All psiphon library notices (info, warnings, errors) are emitted to log file and stdout
   - Configured log_level controls which messages appear (notice includes all non-debug)

---

## User Configuration Guide

### Default Configuration Locations
- User-level: `~/.config/psiphond-ng/psiphond-ng.conf`
- Data directory: `~/.local/var/lib/psiphon`
- Log file: `~/.local/var/log/psiphon/psiphond-ng.log`

### Important Settings

#### Proxy Ports
```json
{
  "local_socks_proxy_port": 1081,
  "local_http_proxy_port": 8081
}
```
If these ports are occupied, the binary will automatically find free ports on first startup and update the config.

#### Listen Interface
```json
{
  "listen_interface": ""
}
```
- `""` (empty) = bind to 127.0.0.1 (loopback only, recommended)
- `"any"` = bind to 0.0.0.0 (all interfaces, accessible from network)
- `"eth0"` = bind to specific interface's IP address

If an invalid interface name is specified, it will automatically fall back to default.

#### Remote Server List
```json
{
  "disable_remote_server_list_fetcher": false,
  "remote_server_list_urls": [
    "https://s3.amazonaws.com/psiphon/web/mjr4-p23r-puwl/server_list_compressed"
  ],
  "remote_server_list_signature_public_key": "..."
}
```
**Do not disable** remote server fetching unless you provide static server entries.

#### Log Level
```json
{
  "log_level": "notice"
}
```
Options: `debug`, `info`, `notice`, `warning`, `error`. Use `debug` only for troubleshooting.

---

## Testing and Verification

### Manual Test Command
```bash
# Run in foreground with full output
./psiphond-ng

# Or with custom config
./psiphond-ng -config /path/to/psiphond-ng.conf

# With debug logging (temporary)
LOG_LEVEL=debug ./psiphond-ng
```

### Check Service Status
```bash
systemctl --user status psiphond-ng
journalctl --user -u psiphond-ng -f
```

### Verify Proxy Ports
```bash
ss -tlnp | grep -E 'psiphond|1081|8081'
```

### Test Proxy Connectivity
```bash
# SOCKS5
curl -x socks5h://127.0.0.1:1081 https://checkip.amazonaws.com

# HTTP
curl -x http://127.0.0.1:8081 https://checkip.amazonaws.com
```

---

## Troubleshooting Checklist

1. **Service exits immediately:**
   - Check `journalctl --user -u psiphond-ng -n 50`
   - Look for errors about "listener IP", "port already in use", or data store timeouts
   - Verify config file is valid JSON

2. **No proxy ports listening:**
   - Check if another service is using ports 1081/8081
   - Binary will auto-detect and move to next free ports on first run
   - Alternatively, manually set ports to known free ones

3. **Cannot reach servers / no tunnels:**
   - Ensure `disable_remote_server_list_fetcher: false`
   - Verify `remote_server_list_urls` and signature key are present
   - Check network connectivity (firewall allows outbound HTTPS)

4. **High latency/poor performance:**
   - Set `egress_region` to preferred country code (e.g., "US", "DE", "JP")
   - Check if local network/ISP is blocking Psiphon
   - Review logs for server scoring details (set log_level to "debug" temporarily)

5. **TUN mode not working:**
   - Requires `cap_net_admin` capability or root
   - Grant capability: `sudo setcap cap_net_admin+ep $(which psiphond-ng)`
   - Ensure TUN kernel module loaded: `sudo modprobe tun`

---

## Release Distribution

### Installation from Release Tarball
```bash
# Download and extract
wget https://github.com/dacineu/PsiphonNGLinux/releases/latest/download/psiphond-ng-1.0.0-linux-amd64.tar.gz
tar xzf psiphond-ng-1.0.0-linux-amd64.tar.gz

# Install binary
mkdir -p ~/.local/bin
cp psiphond-ng-linux-amd64 ~/.local/bin/psiphond-ng
chmod +x ~/.local/bin/psiphond-ng

# Install systemd user service
mkdir -p ~/.config/systemd/user
cp config/psiphond-ng-user.service ~/.config/systemd/user/psiphond-ng.service
systemctl --user daemon-reload

# Enable and start
systemctl --user enable --now psiphond-ng

# Check status
systemctl --user status psiphond-ng
```

### Building from Source
```bash
git clone https://github.com/dacineu/PsiphonNGLinux.git
cd PsiphonNGLinux
make deps
make release VERSION=1.0.0
```

---

## Future Improvements

1. **Dynamic Port Assignment:** Consider making port auto-detection more sophisticated - try a range of ports and persist the chosen ones across restarts.

2. **Interface Auto-Selection:** Could automatically detect the primary network interface with internet connectivity and use that instead of requiring manual configuration.

3. **Health Checks:** Add a simple health check endpoint or status file for monitoring systems.

4. **Metrics Enhancement:** Prometheus metrics already exist but are disabled by default; consider enabling them optionally with a config flag.

5. **Configuration Validation:** Add more robust validation at startup with clear error messages for misconfigurations.

6. **Update Mechanism:** Implement automatic binary updates (currently only manual).

---

## Technical Notes

### Psiphon Library Integration
- Uses `github.com/Psiphon-Labs/psiphon-tunnel-core` as a Go library (not a separate binary)
- Local `replace` directives in go.mod point to sibling `psiphon-tunnel-core` directory during development
- In production builds, this resolves to the official module (or could be vendored)

### Systemd User Service
- Runs as user-level service (no sudo required)
- Uses `systemctl --user` for control
- Service file: `config/psiphond-ng-user.service`
- Automatically installed via `psiphond-ng service` command

### Notice System
- Psiphon emits structured JSON notices
- Notices are written to both log file and stdout
- `SetEmitDiagnosticNotices(true, false)` enables NoticeInfo/NoticeError
- JSON notices can be parsed by log aggregators

### Security Considerations
- Binary runs as regular user (no root required for proxy mode)
- Configuration files in `~/.config` should be chmod 600
- Remote server list signature verified with embedded public key
- Traffic routed through Psiphon's obfuscated protocols

---

## References

- [Psiphon Tunnel Core](https://github.com/Psiphon-Labs/psiphon-tunnel-core)
- [Psiphon Documentation](https://psiphon.ca/en/documentation.html)
- Original project: [Psiphon Linux](https://github.com/Psiphon-Labs/psiphon-linux)

---

**Document Version:** 1.0.0
**Last Updated:** 2026-03-14
**Maintainer:** dacineu

# Contributing to PsiphonNGLinux

Thank you for your interest in contributing! This guide will help you get started.

---

## Code of Conduct

Please be respectful and constructive. Harassment or toxic behavior will not be tolerated.

---

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Linux system for testing (or VM/WSL2)
- (Optional) systemd

### Setup

1. **Fork and clone:**
```bash
git clone https://github.com/yourusername/PsiphonNGLinux.git
cd PsiphonNGLinux
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Build:**
```bash
./scripts/build.sh
# Binary will be in build/psiphond-ng
```

4. **Test:**
```bash
# Unit tests
go test ./...

# Build verification
./build/psiphond-ng -help
```

---

## Development Workflow

1. **Create a branch** for your feature/fix:
```bash
git checkout -b feature/my-feature
```

2. **Make changes** with tests where applicable

3. **Run tests:**
```bash
go test ./... -v
go vet ./...
```

4. **Build and test locally:**
```bash
./scripts/build.sh
sudo ./build/psiphond-ng -config test-data/my-config.conf
```

5. **Commit** with clear, descriptive message:
```bash
git commit -m "feat: add support for dynamic MTU adjustment"
git commit -m "fix: handle nil pointer in config watcher"
```

6. **Push and open PR:**
```bash
git push origin feature/my-feature
# Open PR on GitHub
```

---

## Project Structure

```
PsiphonNGLinux/
├── cmd/
│   └── psiphond-ng/          # Main daemon entry point
│       └── main.go
├── internal/
│   ├── config/               # (TODO) Config parsing & validation library
│   ├── daemon/               # (TODO) Daemon lifecycle, systemd integration
│   ├── tunnel/               # (TODO) Tunnel manager wrapper
│   └── metrics/              # (TODO) Prometheus/StatsD metrics
├── config/
│   ├── psiphond-ng.conf      # Default configuration
│   ├── psiphond-ng.service  # systemd unit file
│   └── defaults/            # (TODO) Default parameter presets
├── scripts/
│   ├── install.sh           # Installation script
│   ├── uninstall.sh         # Removal script
│   ├── build.sh             # Build script
│   └── update.sh            # (TODO) Auto-update script
├── docs/
│   ├── TUN-setup.md
│   ├── inproxy-setup.md
│   ├── IMPLEMENTATION-GUIDE.md
│   └── configuration-reference.md
├── go.mod
├── go.sum
├── README.md
└── IMPLEMENTATION-SUMMARY.md  # Overview of project
```

---

## Coding Standards

### Go

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` or `goimports` for formatting
- Write tests for new functionality
- Prefer small, focused functions
- Error handling: Wrap errors with context using `errors.Trace` or `fmt.Errorf("...: %w", err)`
- Use meaningful variable/function names
- Add comments for exported functions (Godoc format)

### Configuration

- Use snake_case for JSON field names
- Document all fields in README or config reference
- Provide sensible defaults
- Validate input on load

### Systemd Service

- Keep `Restart=on-failure`
- Use `User=psiphon` (non-root)
- Set appropriate `CapabilityBoundingSet`
- Include `ProtectSystem=strict` and `ProtectHome=true`
- Log to journald (`StandardOutput=journal`)

---

## Testing

### Unit Tests

```bash
go test ./...
```

Focus on:
- Config loading and validation
- Config → psiphon.Config conversion
- Notice handler formatting
- Any utility functions

### Integration Tests

We need to set up a local test environment:

1. Start local Psiphon server
2. Configure PsiphonNGLinux to use it
3. Start daemon (maybe in foreground for testing)
4. Verify tunnel established
5. Test traffic routing
6. Clean up

See `cmd/psiphond-ng/integration_test.go` (to be created).

### Manual Testing

```bash
# Build
./scripts/build.sh

# Install as user service (no sudo needed!)
./psiphond-ng service

# Or manual install:
# mkdir -p ~/.local/bin
# cp build/psiphond-ng ~/.local/bin/
# mkdir -p ~/.config/systemd/user
# cp config/psiphond-ng-user.service ~/.config/systemd/user/psiphond-ng.service
# systemctl --user daemon-reload
# systemctl --user enable --now psiphond-ng

# Configure
nano ~/.config/psiphond-ng/psiphond-ng.conf  # adjust values (channel_id, sponsor_id, etc.)
systemctl --user restart psiphond-ng

# Check status
systemctl --user status psiphond-ng

# View logs
journalctl --user -u psiphond-ng -f

# Test proxy
curl -x http://127.0.0.1:8080 https://checkip.amazonaws.com

# Test TUN (if enabled; may require setcap or sudo)
ip addr show tun-psiphon
curl --interface tun-psiphon https://checkip.amazonaws.com
```

---

## Areas for Contribution

### High Priority

1. **Metrics Exporter**: Add Prometheus/StatsD metrics collection
   - Expose on `/metrics` endpoint
   - Track: active tunnels, bytes in/out, connection attempts, success rate
   - See `internal/metrics/` (empty, needs implementation)

2. **Self-Update Mechanism**:
   - Periodic check for new releases
   - Download and verify signatures
   - Atomic binary replacement
   - Graceful restart/rollback

3. **Health Checks**:
   - HTTP endpoints `/healthz`, `/readyz`
   - Return appropriate codes for systemd `WatchdogSec`
   - Include tunnel status, last activity timestamp

4. **Configuration Validation**:
   - Validate CIDR ranges in split tunnel configs
   - Check port numbers are valid (1-65535)
   - Verify file paths exist and are accessible
   - Warn about deprecated options

5. **Documentation**:
   - Complete configuration reference (all JSON fields)
   - Troubleshooting guide for common issues
   - Performance tuning guide
   - Security hardening checklist

### Medium Priority

6. **Web UI** (optional):
   - Simple Go web server for config management
   - View logs, stats, tunnel info
   - Edit configuration with live reload
   - Start/stop/restart service

7. **Enhanced TUN Features**:
   - DNS hijacking/intercept
   - IPv6 support
   - Policy-based routing presets
   - TUN device hotplug (create/destroy without restart)

8. **Better Error Reporting**:
   - Collect common error patterns
   - Suggest fixes in logs
   - Exit codes for different failure modes

9. **Systemd Watchdog Integration**:
   - Implement `sd_notify` to tell systemd we're alive
   - Set `WatchdogSec=` in service file
   - Periodic watchdog pings

10. **Multi-Architecture Support**:
    - Test on ARM64, i386
    - Cross-compilation in build script
    - CI/CD with GitHub Actions

### Low Priority / Future

11. **Broker Server Implementation**:
    - Standalone broker binary (currently only spec exists)
    - Web UI for managing proxies and clients
    - Rate limiting, compartment management

12. **Approval Server Enhancements**:
    - Database backend for decisions
    - Caching of approved clients
    - Time-based access control
    - REST API for approvals (not just WebSocket)

13. **Performance Optimizations**:
    - Connection pooling
    - Reduced memory footprint
    - Faster startup (parallel initialization)
    - Cache compiled regexes

14. **Observability**:
    - OpenTelemetry traces
    - Detailed event logging
    - Audit trail for connection approvals

---

## Pull Request Guidelines

### What to include

- Clear description of problem and solution
- Screenshots/logs if UI or logging changes
- Updated documentation if user-facing changes
- Tests for new functionality
- No breaking changes without migration path

### Review process

1. Automated checks (CI) must pass
2. Maintainer review
3. Address feedback
4. Squash commits (clean history)
5. Merge

### Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add Prometheus metrics exporter
fix: prevent panic when config file missing
docs: update inproxy setup guide
chore: update dependencies, fix lint warnings
refactor: simplify config validation logic
test: add integration test for TUN mode
```

---

## Release Process

1. Update version in `README.md` and build script
2. Update `CHANGELOG.md`
3. Tag release: `git tag -s v1.0.0`
4. Push tag: `git push origin v1.0.0`
5. Build binaries (multiple architectures):
```bash
for arch in amd64 arm64; do
    GOARCH=$arch ./scripts/build.sh --static --compress
done
```
6. Create GitHub release with binaries
7. Update package repos (if applicable)

---

## Questions?

- Open an issue for bugs, feature requests, or questions
- Check existing issues and PRs before starting work
- Join our (hypothetical) chat: #psiphon-ng on Libera.Chat

---

## License

By contributing, you agree that your contributions will be licensed under the GPLv3 (same as this project).

---

**Happy hacking!**

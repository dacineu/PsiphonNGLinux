# PsiphonNGLinux - Development Todo List

**Total Tasks:** 25
**Status:** All pending (initial project setup complete, implementation pending)

---

## Legend

- 🔴 High Priority
- 🟡 Medium Priority
- 🟢 Low Priority / Future
- [ ] Not started
- [x] Completed

---

## 📋 Phase 1: Core Implementation (Must-Have)

These tasks are critical for the first working release.

### 1.1 Foundation

- [ ] **Build and test basic PsiphonNGLinux binary**
  - Verify go.mod replace directive works with local psiphon-tunnel-core
  - Build executable without errors
  - Run with simple config and establish tunnel
  - Test both portforward and TUN modes

- [ ] **Implement comprehensive configuration validation**
  - Add schema validation for CIDR ranges, port numbers, durations
  - Validate required fields (propagation_channel_id, sponsor_id)
  - Check inproxy mode dependencies (keys, addresses)
  - Provide clear error messages

- [ ] **Add configuration hot-reload capability**
  - Integrate ConfigWatcher with controller
  - Implement SIGHUP signal handler
  - Apply parameter updates without restart (where possible)
  - Test live config changes

### 1.2 Production Features

- [ ] **Build Prometheus metrics exporter**
  - Create internal/metrics package
  - Define gauges, counters, histograms
  - Track: active_tunnels, bytes_transferred, connection_attempts, errors
  - Expose on /metrics endpoint

- [ ] **Add health check HTTP endpoints**
  - /healthz (overall health)
  - /readyz (ready to accept traffic)
  - /metrics (Prometheus)
  - Consider systemd WatchdogSec integration

- [ ] **Implement comprehensive test suite**
  - Unit tests for config loading/validation/conversion
  - Integration tests using local test Psiphon server
  - Table-driven tests for edge cases
  - Achieve 60%+ code coverage

### 1.3 Observations & Error Handling

- [ ] **Implement feedback and statistics reporting**
  - Wire up ActivityUpdater callback
  - Implement periodic Connected API calls
  - Batch stats for upload
  - Handle offline buffering

- [ ] **Improve error messages with actionable hints**
  - Detect common misconfigurations
  - Suggest specific fixes and config changes
  - Include docs links (e.g., "See TUN-setup.md")
  - Add context to wrapped errors

---

## 🟡 Phase 2: Advanced Features (Important)

Enhance functionality and reliability.

### 2.1 In-Proxy & Approval

- [ ] **Implement approval hook integration for in-proxy mode**
  - [x] Add `ApproveClientConnection` callback field to `psiphon.Config`
  - [x] Integrate callback into controller's proxy initialization
  - [x] Implement WebSocket client with timeout in `BuildPsiphonConfig`
  - [x] Handle approval failures (log warnings, reject connection)
  - [x] Add configurable metadata (`approval_logging` include flags)
  - [x] Build extended `approvalRequest` payload based on config
  - [x] Parse server response for `strict_fields` and `missing_strict`
  - [x] Implement notification server (WebSocket) for GUI alerts
  - [x] Detect strictness changes, broadcast via notification server
  - [x] Log local warnings for missing strict fields
  - [ ] Write comprehensive unit tests for approval strictness handling
  - [ ] Add integration test with mock approval + notification servers

- [ ] **Implement server-side admin API to dynamically manage strict fields**
  - REST endpoint to get/set current strict fields list
  - Broadcast updates to connected clients (WebSocket push)
  - Persist configuration on server
  - Validate field names (predefined set)

- [ ] **Add fallback mechanisms for resilience**
  - Broker failure → fallback to direct server connections
  - Tactics fetch failure → use cached
  - Remote server list failure → use local cache
  - Track fallback events in metrics/logs

### 2.2 Observability

- [ ] **Add observability: metrics and tracing**
  - OpenTelemetry instrumentation
  - Export to Jaeger/Zipkin
  - Trace tunnel establishment, API calls, server selection
  - Ensure minimal performance impact

- [ ] **Integrate systemd watchdog**
  - Use sd_notify for READY=, WATCHDOG=1
  - Set WatchdogSec=30s in service file
  - Implement ping loop in daemon
  - Service auto-restarts if unresponsive

### 2.3 TUN Enhancements

- [ ] **Enhance TUN mode with advanced features**
  - DNS hijacking (redirect all DNS to tunnel DNS)
  - IPv6 split tunneling support
  - Policy-based routing presets (for NetworkManager, systemd-networkd)
  - TUN device hotplug (create/destroy without daemon restart)

### 2.4 Self-Update

- [ ] **Implement self-update mechanism**
  - Periodic GitHub releases API check
  - Download new binary with resume support
  - Verify signatures (cosign/sigstore)
  - Atomic replacement and graceful restart
  - Rollback on failure

### 2.5 Dynamic Observer Interface Protocol (DOIP)

- [ ] **Define DOIP specification**
  - Message envelope format (JSON RPC/event stream)
  - Channels/streams: approval requests, notifications, admin
  - Versioning and compatibility
  - Authentication (mutual TLS or token)

- [ ] **Implement DOIP server (daemon side)**
  - Implemented notification server (WebSocket) for GUI alerts
  - Broadcasts: strictness_changed, missing_strict_fields
  - Start/stop via config and graceful shutdown
  - Basic implementation; future: TLS, authentication, hot-reload

- [ ] **Implement DOIP client (GUI side) reference**
  - Reconnection logic with backoff
  - Event handlers for different message types
  - Auto‑subscribe to relevant notification streams
  - Secure storage of connection credentials

- [ ] **Add admin DOIP endpoints for server‑side strictness management**
  - `GET /admin/strict-fields` – retrieve current strict set
  - `POST /admin/strict-fields` – update strict set (list of metadata field names)
  - Broadcast updates to all connected DOIP clients immediately
  - Validate field names against allowed set
  - Log admin actions for audit

- [ ] **Hot‑reload DOIP configuration**
  - Ability to change DOIP port, TLS certs, auth via main config
  - Graceful restart of DOIP server on config change (maintain connections if possible)
  - Validate new TLS certificates before switch

- [ ] **Security hardening**
  - Enforce TLS 1.3+
  - Rate limit admin API
  - Implement token‑based auth for admin API
  - Access logging for admin actions

### 2.6 Approval Server (Standalone)

- [x] **Create separate DOIP Approval Server project**
  - Created `doip-approval-server/` repository with its own go.mod
  - Implements WebSocket `/approval` endpoint for client requests
  - Implements admin HTTP API `/admin/strict-fields` (GET/POST)
  - Optional GUI static file serving (`-gui-dir`)
  - Config via flags: `-ws-port`, `-admin-port`, `-strict`, `-gui-dir`
  - Uses gorilla/websocket only
  - Binary builds and tests pass

- [x] **Add strict field validation on server**
  - [x] Validate admin POST fields against allowed set
  - [x] Reject unknown fields with 400 Bad Request
  - [x] Log admin changes with timestamps for audit (already logging)

- [x] **Write tests for approval server**
  - [x] Unit tests for missing field logic
  - [x] Integration test: WebSocket approval flow
  - [x] Admin API GET/POST and validation

- [ ] **Package and release**
  - [ ] Build binaries for linux/amd64, linux/arm64 (CI)
  - [ ] Provide Dockerfile
  - [ ] Include example config and systemd unit
  - [ ] Document API and usage

---

## 🟢 Phase 3: Quality & Testing (Should-Have)

Ensure reliability, security, and maintainability.

### 3.1 Testing

- [ ] **Create comprehensive test suite** (see Phase 1) ✅ listed above

- [ ] **Expand test coverage to 80%+**
  - Target internal/config, internal/metrics, tunnel manager
  - Add E2E tests with Docker Compose (broker, approval server, test Psiphon)
  - Performance regression tests
  - Chaos testing (kill processes, network partitions)

- [ ] **Conduct thorough integration and load testing**
  - Deploy to staging environment with real Psiphon servers
  - Load test: 100+ concurrent tunnels
  - Verify failover and recovery
  - Monitor resource usage (memory leaks, file descriptor leaks)
  - Test under adverse network conditions (latency, loss, censorship simulation)

### 3.2 Performance

- [ ] **Performance optimization and profiling**
  - Profile CPU with pprof
  - Profile memory allocations
  - Reduce allocations in hot paths (tunnel establishment, data forwarding)
  - Target: <50MB RSS, <5% CPU steady-state
  - Optimize config parsing, JSON marshaling

- [ ] **Optimize build artifacts for size**
  - Use upx compression
  - Strip debug symbols
  - Reduce vendored dependencies
  - Target binary size <10MB (compressed)

### 3.3 Security

- [ ] **Add binary signing and verification**
  - Use sigstore/cosign for signing releases
  - Include SBOMs (Software Bill of Materials)
  - Create sigstore attestations for provenance
  - Verify signatures on self-update

- [ ] **Implement crash reporting**
  - Integrate Sentry or similar
  - Capture panic stack traces with symbols
  - Include context (config, version, OS)
  - Respect user privacy (opt-out, no PII)

- [ ] **Support environment variable and metadata config sources**
  - Override config from env vars (PSIPHON_*)
  - Cloud metadata: AWS IMDS, GCP metadata, Azure IMDS
  - Flag files (drop directory)
  - Precedence: CLI flags > env > file > defaults

- [ ] **Implement advanced TLS/mTLS configuration**
  - Client certificate support for upstream proxies
  - Custom root CA bundles
  - Certificate pinning for Psiphon servers
  - TLS 1.3 enforcement option

---

## 📦 Phase 4: Distribution & Operations (Release)

Prepare for public release and production use.

### 4.1 Packaging

- [ ] **Package for distribution platforms**
  - Debian/Ubuntu: .deb packages, apt repository
  - Fedora/RHEL: .rpm packages, COPR/YUM repo
  - Arch Linux: AUR package (PKGBUILD)
  - Include postinst/prerm/preinst scripts
  - Handle config migrations

- [ ] **Create Docker containerization**
  - Multi-stage Dockerfile (build + runtime)
  - Non-root user in container
  - Health checks
  - Push to Docker Hub / GHCR
  - docker-compose.yml for quick start

- [ ] **Create build automation for multiple architectures**
  - Cross-compile to arm64, amd64, 386
  - Build matrix in CI/CD
  - Upload all arch binaries to GitHub Releases

### 4.2 Documentation

- [ ] **Finalize comprehensive user documentation**
  - Complete README with installation options
  - Configuration reference (all parameters documented)
  - Troubleshooting guide (common issues, solutions)
  - FAQ
  - Migration guide from PsiphonLinux
  - Security best practices
  - Performance tuning

- [ ] **Complete developer and contributor documentation**
  - Architecture ADRs (Architecture Decision Records)
  - Code walkthrough for major components
  - Testing guide (how to add tests)
  - Release process documentation
  - Contributing guidelines (already started)
  - API documentation (if any public APIs)

### 4.3 CI/CD & Release

- [ ] **Set up complete CI/CD pipeline**
  - GitHub Actions workflow:
    - Build and test on PR
    - Security scanning (trivy, gosec)
    - Dependency updates (dependabot)
    - Release automation on tag push
    - Cross-compilation matrix
    - Package building
    - Upload to releases

- [ ] **Implement binary signing and verification** (see Phase 3) ✅ listed above

- [ ] **Enhance region-based routing**
  - Regional DNS server selection
  - GeoIP-based server scoring hints
  - Dynamic region detection improvements
  - Multi-region failover testing

---

## 🎯 Phase 5: Polish & Maintenance (Nice-to-Have)

Refinements for maturity.

### 5.1 User Experience

- [ ] **Develop web management UI** (optional)
  - Simple Go http server (no external deps)
  - Dashboard: status, logs, stats
  - Configuration editor with validation
  - Start/stop/restart controls
  - Authentication (basic auth or mTLS)

- [ ] **Refactor configuration handling**
  - Move config logic to internal/config package
  - Support multiple config sources (file, env, flags)
  - Config diffing for hot-reload decisions
  - Config migration/upgrade tool

### 5.2 Operations

- [ ] **Complete developer and contributor documentation** (see above) ✅ listed

- [ ] **Implement distributed tracing** (see above) ✅ listed

---

## 🏁 Phase 6: Release Checklist (v1.0.0)

All items below must be complete before tagging v1.0.0:

- [ ] **Prepare v1.0.0 stable release**
  - [ ] All tests passing (unit, integration, system)
  - [ ] Security audit complete (no critical vulnerabilities)
  - [ ] Performance benchmarks meet targets
  - [ ] Documentation complete and accurate
  - [ ] Packages built for all target platforms
  - [ ] Binaries signed with sigstore
  - [ ] Release notes written
  - [ ] Migration guide from PsiphonLinux
  - [ ] Post-mortem from beta testing
  - [ ] Support channel established (GitHub Discussions, matrix, etc.)

---

## 📊 Task Summary by Priority

| Priority | Count | Category |
|----------|-------|----------|
| 🔴 High (Phase 1) | 9 | Core functionality, metrics, testing |
| 🟡 Medium (Phase 2) | 8 | Advanced features, observability |
| 🟢 Low (Phase 3) | 15 | Quality, performance, security |
| 📦 Release (Phase 4) | 10 | Packaging, docs, CI/CD |
| 🎯 Polish (Phase 5) | 4 | UX, maintenance |
| **Total** | **46** | — |

---

## 🔄 Dependencies

```
Core Implementation (1.1)
├── Build & test binary
├── Config validation
└── Hot-reload

  ↓

Observability & Health (1.2)
├── Metrics exporter
├── Health endpoints
└── Feedback reporting

  ↓

Advanced Features (2.x)
├── Approval hook (needs core stable)
├── Fallback mechanisms
├── TUN enhancements
├── Self-update
└── Tracing/mTLS

  ↓

Testing & Quality (3.x)
├── Comprehensive tests
├── Load testing (needs features complete)
├── Performance optimization
└── Security hardening

  ↓

Distribution (4.x)
├── Packaging (needs stable binary)
├── Docker
├── CI/CD
├── Documentation
└── Signing

  ↓

Release (6.x)
└── All above complete + final QA
```

---

## 🎯 Current Focus

**Immediate next tasks (in order):**

1. Build and test basic binary with local psiphon-tunnel-core
2. Implement config validation with clear error messages
3. Build metrics exporter and health endpoints
4. Create and expand test suite
5. Implement approval hook for in-proxy mode

---

## 📝 Notes

- This todo list assumes the modified psiphon-tunnel-core with approval hook is available
- If using upstream without modifications, skip/adapt approval-related tasks
- Tasks can be parallelized where independent
- Release should be after 2-3 months of testing in staging

---

**Last updated:** 2025-03-14
**Project:** PsiphonNGLinux
**Repository:** `/home/dacineu/dev/PsiphonNGLinux`

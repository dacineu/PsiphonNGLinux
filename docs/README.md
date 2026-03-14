# PsiphonNGLinux Documentation

Complete documentation for PsiphonNGLinux.

---

## Getting Started

- **[README.md](../README.md)** - Project overview, features, quick install
- **[Quick Start for Developers](quick-start-for-developers.md)** - Set up dev environment and first build

---

## Configuration

- **[Configuration Reference](configuration-reference.md)** - All config options with examples
- **[TUN Setup Guide](TUN-setup.md)** - System-wide tunneling with virtual network interface
- **[In-Proxy Setup Guide](inproxy-setup.md)** - Broker-mediated connections with WebRTC

---

## Development

- **[IMPLEMENTATION-GUIDE.md](../IMPLEMENTATION-GUIDE.md)** - Architecture and implementation details
- **[CONTRIBUTING.md](../CONTRIBUTING.md)** - How to contribute code and report issues

---

## Reference

- **Project Comparison**: See `COMPARISON-WITH-PSIPHONLINUX.md` in parent directory
- **Configuration examples** in `config/` directory:
  - `psiphond-ng.conf` - Default port forward mode
  - `psiphond-ng-inproxy-client.conf` - In-proxy client configuration
  - `psiphond-ng-inproxy-proxy.conf` - In-proxy server configuration
- **Systemd service**: `config/psiphond-ng.service`

---

## Key Concepts

### Tunnel Modes

1. **Port Forward** - Local SOCKS5/HTTP proxies (no root required)
2. **Packet Tunnel** - TUN device for system-wide routing (requires root/CAP_NET_ADMIN)

### In-Proxy Modes

1. **Client** - Connect to broker to get matched with a proxy
2. **Proxy** - Run a relay server that clients connect through

### Split Tunneling

Choose which traffic goes through Psiphon:
- Split by country (`split_tunnel_own_region`, `split_tunnel_regions`)
- Split by IP/CIDR (`split_tunnel_include`, `split_tunnel_exclude`)

---

## Common Tasks

### Install from source

```bash
git clone https://github.com/yourusername/PsiphonNGLinux.git
cd PsiphonNGLinux
./scripts/build.sh
sudo ./scripts/install.sh
```

### Update configuration

```bash
sudo nano /etc/psiphon/psiphond-ng.conf
sudo systemctl restart psiphond-ng
```

### View logs

```bash
sudo journalctl -u psiphond-ng -f
sudo tail -f /var/log/psiphon/psiphond-ng.log
```

### Check status

```bash
sudo systemctl status psiphond-ng
ip addr show tun-psiphon  # if TUN mode
ss -tlnp | grep psiphond  # check listening ports
```

### Generate in-proxy keys

```bash
./scripts/generate-keys.sh
```

---

## Troubleshooting

**Tunnel won't establish:**
- Check `log_level: "debug"` and review logs
- Verify network connectivity to Psiphon servers
- Ensure `propagation_channel_id` and `sponsor_id` are set (use F's for testing)
- Try increasing `establish_tunnel_timeout_seconds`

**Port forward not working:**
- Verify ports are listening: `ss -tlnp | grep 1080`
- Configure application to use 127.0.0.1:1080 (SOCKS) or 127.0.0.1:8080 (HTTP)
- Check firewall allows localhost connections

**TUN mode fails:**
- Run as root or grant capability: `sudo setcap cap_net_admin+ep psiphond-ng`
- Check TUN module: `lsmod | grep tun` or `sudo modprobe tun`
- Verify `tunnel_mode: "packet"` in config

**High CPU/memory:**
- Reduce `tunnel_pool_size` (default 1 is usually fine)
- Reduce `connection_worker_pool_size`
- Limit protocols: `"limit_tunnel_protocols": ["OSSH", "TLS-OSSH"]`
- Increase `stagger_connection_workers_milliseconds`

**DNS leaks:**
- For TUN mode, set `packet_tunnel_dns_ipv4` to tunneled DNS
- Verify `/etc/resolv.conf` points to tunnel DNS or is managed by systemd-resolved
- Use `curl --interface tun-psiphon https://checkip.amazonaws.com` to test

---

## Support

- **Issues**: https://github.com/yourusername/PsiphonNGLinux/issues
- **Psiphon docs**: https://psiphon.ca/en/docs.html
- **Psiphon core**: https://github.com/Psiphon-Labs/psiphon-tunnel-core

---

## License

GPLv3 - See [LICENSE](../LICENSE)

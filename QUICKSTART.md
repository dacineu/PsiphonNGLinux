# Quick Start Guide

## First Time Use

### Option 1: Just Run It (Simplest)

```bash
# Download or build the binary
./psiphond-ng
```

The binary will:
- If config is missing → create `~/.config/psiphond-ng/psiphond-ng.conf` with safe defaults
- If config exists but is invalid → backup it and create fresh default config
- Create data and log directories under `~/.local/var/`
- Start the tunnel (foreground)
- Log to both console and file

**Edit the config** before running for real:
```bash
nano ~/.config/psiphond-ng/psiphond-ng.conf
```
Set your `propagation_channel_id` and `sponsor_id` from your Psiphon account.

### Option 2: Install as Background Service

```bash
# Download or build the binary
./psiphond-ng service [--bin-dir PATH]
```

This will:
- Copy itself to the install directory (default: `~/.local/bin/psiphond-ng`)
  - Override with `--bin-dir /custom/path` or `PSIPHON_BIN_DIR` environment variable
- Generate config at `~/.config/psiphond-ng/psiphond-ng.conf`
- Install systemd user service
- Start the daemon automatically
- If a binary already exists at the destination:
  - Shows prompt: `[O]verwrite, [B]ackup then overwrite, [C]ancel?` (interactive only)
- **Asks if you want to delete the original binary** (the one you ran, interactive only)

After installation, control the service:
```bash
systemctl --user status psiphond-ng
systemctl --user stop psiphond-ng
systemctl --user start psiphond-ng
journalctl --user-unit psiphond-ng -f
```

## After Installation

### Run as Service (recommended)
```bash
systemctl --user start psiphond-ng
# enable for autostart: systemctl --user enable psiphond-ng
```

### Run in Foreground
```bash
~/.local/bin/psiphond-ng
# or with custom config
~/.local/bin/psiphond-ng /path/to/config.conf
```

### Changing Configuration
Edit `~/.config/psiphond-ng/psiphond-ng.conf`. Changes are auto-reloaded (hot-reload).

### Uninstall
```bash
# Stop service
systemctl --user stop psiphond-ng
systemctl --user disable psiphond-ng

# Remove files
rm -rf ~/.local/bin/psiphond-ng
rm -rf ~/.config/psiphond-ng
rm -rf ~/.local/var/lib/psiphon
rm -rf ~/.local/var/log/psiphon
```

## Notes

- Default config uses DOIP approval server at `ws://localhost:8443/approval` – install the DOIP approval server separately if you plan to use in-proxy mode with approval.
- `disable_remote_server_list_fetcher: true` by default – no signature key required. Set to `false` and provide `remote_server_list_signature_public_key` to use remote server lists.
- Systemd user services run in your user session. For system-wide service (all users), you'd need to adapt the service file to `/etc/systemd/system/` and run as root (not recommended for this design).
- Interactive prompts (for install path, binary overwrite/backup, self-delete) only appear when running in a terminal. In scripts, use `--bin-dir` and `PSIPHON_BIN_DIR` to control the install path; the binary will auto-overwrite without prompting.

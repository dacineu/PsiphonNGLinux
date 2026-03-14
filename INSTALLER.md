# PsiphonNG Self-Installing Binary

The `psiphond-ng` binary now includes built-in installation and configuration management.

## Usage Modes

### 1. Foreground/Run Mode (Default)

```bash
./psiphond-ng [config-path]
```

- If config-path is provided, uses that config file.
- If no argument, uses `~/.config/psiphond-ng/psiphond-ng.conf`.
- If the config file does **not** exist:
  - Automatically creates a default configuration with:
    - User-writable data and log paths under `~/.local/var/`
    - DOIP approval server pre-configured (`ws://localhost:8443/approval`)
    - Remote server list fetching disabled (no signature key required)
  - Ensures all required directories exist
  - Continues to run
- If the config file exists but is **invalid or outdated**:
  - Automatically backs it up with timestamp (e.g., `psiphond-ng.conf.bak_20260314-144439`)
  - Creates a fresh default config
  - Logs a warning about the migration
  - Continues running

This makes the binary instantly runnable and resilient to configuration format changes.

### 2. Service Installation Mode

```bash
./psiphond-ng service [--bin-dir PATH]  # or set PSIPHON_BIN_DIR environment variable
```

Installs PsiphonNG as a systemd **user** service:
- Copies the binary to the install directory (default: `~/.local/bin/psiphond-ng`)
- Creates config at `~/.config/psiphond-ng/psiphond-ng.conf`
  - If config already exists, backs it up as `psiphond-ng.conf.bkp`
- Installs systemd unit to `~/.config/systemd/user/psiphond-ng.service`
- Creates data and log directories
- Reloads systemd daemon
- Enables and starts the service
- **Prompts to delete the original downloaded binary** (if interactive)

**Specifying install location:**
- Use `--bin-dir /path/to/bin` to set custom directory
- Or set environment variable `PSIPHON_BIN_DIR=/path/to/bin`
- If neither is set and running in a terminal, you'll be prompted
- Default: `~/.local/bin`

**When a binary already exists at the destination:**
- Interactive mode: prompts with options:
  - **O** = Overwrite (default, just replace)
  - **B** = Backup first (creates `psiphond-ng.old` then copy)
  - **C** = Cancel (abort installation)
- Non-interactive mode (scripts/pipes): overwrites automatically with a log message

### 3. Control After Installation

If installed as a service:

```bash
systemctl --user status psiphond-ng
systemctl --user start psiphond-ng
systemctl --user stop psiphond-ng
journalctl --user-unit psiphond-ng -f
```

Or use the convenience wrapper script (optional):
```bash
~/.local/bin/psiphon-ctl {start|stop|restart|status|run}
```

## Backups During Installation

When running `psiphond-ng service`:
- Existing binary at the install destination → backed up as `psiphond-ng.old`
- Existing config at `~/.config/psiphond-ng/psiphond-ng.conf` → backed up as `psiphond-ng.conf.bkp`

## Default Configuration Details

When a new config is generated (either via `service` install or auto-creation on first run):

- `data_directory`: `~/.local/var/lib/psiphon`
- `log_file`: `~/.local/var/log/psiphon/psiphond-ng.log`
- `tunnel_mode`: `portforward`
- `local_socks_proxy_port`: `1080`
- `local_http_proxy_port`: `8080`
- `inproxy_approval_websocket_url`: `ws://localhost:8443/approval`
- `inproxy_approval_timeout`: `10s`
- `disable_remote_server_list_fetcher`: `true` (no signature key needed)
- Other fields use library defaults

**Important:** Users should edit the config to set their `propagation_channel_id` and `sponsor_id` to get real Psiphon servers. The defaults are placeholder values.

## Advantages

- **Single binary** – no external scripts needed
- **First-run simplicity** – just run and it works
- **Service integration** – one-command install
- **Custom install path** – via `--bin-dir` or `PSIPHON_BIN_DIR`
- **Safe upgrades** – backups previous installation
- **DOIP approval ready** – pre-configured for in-proxy mode with approval
- **Self-cleaning** – optional deletion of original downloaded binary after install

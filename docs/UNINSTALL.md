# Uninstalling PsiphonNGLinux

This document describes how to completely remove PsiphonNGLinux from your system.

---

## Quick Summary

| Command | What it does |
|---------|--------------|
| `make uninstall` | Removes service + binary, **prompts** about config/data/logs |
| `make full-uninstall` | Removes **everything** (service, binary, config, data, logs) without prompts |
| `deployment/uninstall.sh` | Run script directly (same as `make uninstall`) |
| `deployment/uninstall.sh --full` | Run script with full removal |

---

## What Gets Installed

When you install PsiphonNGLinux, the following files and directories are created:

### Systemd User Service
- **Service file:** `~/.config/systemd/user/psiphond-ng.service`
- **Service name:** `psiphond-ng.service` (user-level)
- **Status:** Controlled with `systemctl --user`

### Binary
- **Location:** `~/.local/bin/psiphond-ng`
- **Permissions:** Executable (755)

### Configuration
- **Config file:** `~/.config/psiphond-ng/psiphond-ng.conf`
- **Created:** On first run if not present

### Data Directory
- **Path:** `~/.local/var/lib/psiphon/`
- **Contents:** Datastore (bolt DB), server lists cache, tactics cache, certificates

### Logs
- **Log file:** `~/.local/var/log/psiphon/psiphond-ng.log`
- **Notice file:** `~/.local/var/log/psiphon/psiphond-ng.notices` (if enabled)

---

## Uninstall Options

### Option 1: Standard Uninstall (Interactive)

Run:

```bash
make uninstall
```

Or directly:

```bash
./deployment/uninstall.sh
```

**What it does:**
1. ✓ Stops and disables the systemd user service
2. ✓ Removes the service file (`~/.config/systemd/user/psiphond-ng.service`)
3. ✓ Reloads systemd daemon
4. ✓ Removes the binary (`~/.local/bin/psiphond-ng`)
5. ✓ Asks **interactively** whether to remove:
   - Config directory (`~/.config/psiphond-ng`)
   - Data directory (`~/.local/var/lib/psiphon`)
   - Log directory (`~/.local/var/log/psiphon`)

**Use this** if you want to keep your configuration and data for potential reinstallation.

---

### Option 2: Full Uninstall (Removes Everything)

Run:

```bash
make full-uninstall
```

Or directly:

```bash
./deployment/uninstall.sh --full
```

**What it does:**
1. ✓ Stops and disables the systemd user service
2. ✓ Removes the service file
3. ✓ Reloads systemd daemon
4. ✓ Removes the binary
5. ✓ **Automatically removes WITHOUT prompting:**
   - Config directory
   - Data directory
   - Log directory

**Use this** if you want a clean slate, e.g., before reinstalling or completely removing PsiphonNG.

---

## Manual Uninstall (If Scripts Fail)

If the uninstall scripts don't work, you can manually remove everything:

```bash
# 1. Stop the service
systemctl --user stop psiphond-ng 2>/dev/null || true
systemctl --user disable psiphond-ng 2>/dev/null || true

# 2. Remove service file
rm -f ~/.config/systemd/user/psiphond-ng.service
systemctl --user daemon-reload

# 3. Remove binary
rm -f ~/.local/bin/psiphond-ng

# 4. Remove config (optional - keep if you want to reinstall with same settings)
rm -rf ~/.config/psiphond-ng

# 5. Remove data (optional - this deletes cached server lists, datastore)
rm -rf ~/.local/var/lib/psiphon

# 6. Remove logs (optional)
rm -rf ~/.local/var/log/psiphon
```

---

## Verification

After uninstall, verify everything is removed:

```bash
# Service should not be listed
systemctl --user status psiphond-ng 2>&1 | grep -i "not found\|inactive" || echo "Service still present?"

# Binary should be gone
test -f ~/.local/bin/psiphond-ng && echo "Binary still exists!" || echo "Binary removed"

# Directories should be gone (unless you kept them)
ls -la ~/.config/psiphond-ng 2>/dev/null || echo "Config removed"
ls -la ~/.local/var/lib/psiphon 2>/dev/null || echo "Data removed"
ls -la ~/.local/var/log/psiphon 2>/dev/null || echo "Logs removed"
```

---

## What About Other Locations?

PsiphonNGLinux **only** installs to the user's home directory under XDG directories:
- `~/.local/bin` - binary
- `~/.config/psiphond-ng` - config
- `~/.local/var/lib/psiphon` - data
- `~/.local/var/log/psiphon` - logs
- `~/.config/systemd/user` - service file

**No system directories** are modified (no `/etc`, `/usr/local`, `/var`), and **no root privileges** are required for install or uninstall.

---

## Troubleshooting

### Service Still "Active" After Uninstall

If `systemctl --user status psiphond-ng` shows the service as still active after uninstall:

```bash
# Force stop any running instance
systemctl --user stop psiphond-ng 2>/dev/null || true
systemctl --user kill psiphond-ng 2>/dev/null || true

# Check for orphaned processes
ps aux | grep psiphond-ng | grep -v grep

# If processes remain, kill manually:
#   pkill -f psiphond-ng
# or: killall psiphond-ng
```

### Permission Denied

If you get permission errors, ensure you're running as the same user that installed PsiphonNG. The user-level installation only affects your user account.

### Want to Reinstall?

After uninstall (even full), you can reinstall by simply running the binary again or using `make install`. A fresh config will be created with defaults.

---

## Migration Between Versions

When upgrading from an older version to a newer one:

1. **Stop the service:**
   ```bash
   systemctl --user stop psiphond-ng
   ```

2. **Backup your config:**
   ```bash
   cp -r ~/.config/psiphond-ng ~/.config/psiphond-ng.backup
   ```

3. **Install the new binary** (overwrite existing):
   ```bash
   make install
   # Or manually copy binary to ~/.local/bin/
   ```

4. **Review the new config** (if migrated):
   ```bash
   # The installer/backup may have preserved your old config
   # Check for differences with the new default template
   diff ~/.config/psiphond-ng/psiphond-ng.conf config/psiphond-ng.conf
   ```

5. **Restart the service:**
   ```bash
   systemctl --user start psiphond-ng
   systemctl --user status psiphond-ng
   ```

---

## Cleanup Leftovers

If you've manually removed files or the uninstall script missed something, here's a comprehensive cleanup command:

```bash
# WARNING: This removes ALL PsiphonNG traces permanently
systemctl --user stop psiphond-ng 2>/dev/null || true
systemctl --user disable psiphond-ng 2>/dev/null || true
rm -f ~/.config/systemd/user/psiphond-ng.service
systemctl --user daemon-reload
rm -f ~/.local/bin/psiphond-ng
rm -rf ~/.config/psiphond-ng
rm -rf ~/.local/var/lib/psiphon
rm -rf ~/.local/var/log/psiphon
```

---

## Summary

| Action | Command |
|--------|---------|
| Normal uninstall (interactive) | `make uninstall` |
| Complete removal (no prompts) | `make full-uninstall` |
| Manual cleanup | See section above |
| Verify removal | Check `systemctl --user status psiphond-ng` |

The `--full` flag is recommended for a clean slate before reinstalling or completely removing PsiphonNG from your system.

---

**Last Updated:** 2026-03-14
**Applies to Version:** 1.0.0+

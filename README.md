# Pi-Star MCP (Master Control Program)

The single Go binary that powers Pi-Star v5 — serving as both the **web dashboard** and **process supervisor** for amateur radio digital voice hotspots.

Pi-Star MCP replaces the legacy PHP/lighttpd/log-parsing stack with a modern, lightweight architecture built around MQTT, WebSockets, and a modular UI.

## What It Does

- **Process Supervisor** — spawns and manages Mosquitto, MMDVMHost, gateways, and bridges as child processes. Monitors health, restarts on crash, handles clean shutdown.
- **HTTPS Dashboard** — serves the web UI over TLS with auto-generated self-signed certificates (or user-supplied certs).
- **MQTT Relay** — subscribes to MMDVMHost's MQTT topics and relays messages to the browser via WebSocket for real-time updates.
- **Module System** — dashboard panels are loaded from disk at startup. Community modules can be added without recompiling the binary.
- **Authentication** — authenticates against the system `pi-star` user via `/etc/shadow` (pure Go hash verification with `unix_chkpwd` fallback). API bearer tokens for external access.
- **Theming & i18n** — CSS custom property themes (6 themes, light/dark variants), internationalisation via JSON translation files.

## Architecture

```
Browser ──WebSocket + REST──▶ Pi-Star MCP (Go binary, runs as root)
                                  │
                    ┌─────────────┼─────────────┐
                    ▼             ▼              ▼
               Mosquitto     MMDVMHost      Gateways
               (child)       (child)        (children)
```

The binary is the **only managed service** on the host. It owns the full lifecycle of all child processes — the systemd/OpenRC service definitions shipped in MMDVM packages are for standalone (non-Pi-Star) use only.

### Startup Sequence

1. Load configuration from `/etc/pistar-dashboard/dashboard.ini`
2. Ensure TLS certificates exist (generate self-signed if needed)
3. Start Mosquitto on an available port (localhost-only, no auth)
4. Start MMDVMHost and enabled gateways with matching MQTT config
5. Connect MQTT client, subscribe to module-declared topics
6. Discover UI modules from disk
7. Start HTTPS server

### Key Design Decisions

- **Single static binary** — `CGO_ENABLED=0`, no libc dependency, runs on both glibc (Debian) and musl (Alpine/Pi-Star_OS)
- **Standard paths** — config lives in `/etc/`, persistent data in `/var/lib/`, ephemeral state in `/run/`. No awareness of Pi-Star_OS bind-mounts needed.
- **MQTT only** — no log file parsing. Real-time data comes from MMDVMHost's MQTT output.
- **Managed Mosquitto** — always spawns its own broker instance, never conflicts with user-installed Mosquitto.

## Building

### Prerequisites

- Go 1.22+

### Development Build (native)

```bash
make build
```

### Cross-Compile for Linux Targets

```bash
make linux-arm      # ARMv6 — all Raspberry Pis
make linux-arm64    # ARM64 — Pi 3/4/5 (64-bit OS)
make linux-amd64    # x86_64 — Debian/Ubuntu servers
make all            # All three targets
```

Binaries are written to `build/` with version injected from `git describe`.

All targets produce fully static binaries (~2-3 MB stripped, ~10-12 MB with full deps).

### Clean

```bash
make clean
```

## Configuration

The dashboard reads a single INI file (default: `/etc/pistar-dashboard/dashboard.ini`). If the file doesn't exist, sensible defaults are used — the dashboard is fully functional out of the box.

```bash
pistar-dashboard --config /path/to/dashboard.ini
```

### Example Configuration

```ini
[dashboard]
listen_http=:80
listen_https=:443
modules_dir=/opt/pistar/modules

[security]
auth_user=pi-star
session_timeout=1800

[tls]
cert_file=/etc/pistar-dashboard/certs/server.crt
key_file=/etc/pistar-dashboard/certs/server.key
auto_generate=1
min_version=1.2

[paths]
certs_dir=/etc/pistar-dashboard/certs
db_dir=/var/lib/pistar-dashboard
backup_dir=/var/lib/pistar-dashboard/backups
audit_log=/var/log/pistar-dashboard/audit.log
runtime_dir=/run/pistar

[mqtt]
port=1883
fallback_port=1884
mosquitto_path=/usr/sbin/mosquitto

[services]
mmdvmhost_enabled=1
mmdvmhost_path=/usr/local/bin/MMDVMHost
mmdvmhost_config=/etc/mmdvmhost/MMDVM.ini
dmrgateway_enabled=1
dmrgateway_path=/usr/local/bin/DMRGateway
dmrgateway_config=/etc/dmrclients/DMRGateway.ini
ysfgateway_enabled=0
p25gateway_enabled=0
nxdngateway_enabled=0
```

## Project Structure

```
Pi-Star_MCP/
├── main.go                          # Entry point, CLI flags, startup orchestration
├── internal/
│   ├── config/                      # INI config loading, defaults, validation
│   ├── tlsutil/                     # Self-signed certificate generation
│   ├── auth/                        # /etc/shadow auth, sessions, CSRF, API tokens
│   ├── supervisor/                  # Child process lifecycle, Mosquitto management
│   ├── server/                      # HTTPS server, middleware, routes
│   │   └── handlers/               # API and page request handlers
│   ├── mqttclient/                  # MQTT connection and WebSocket relay
│   ├── wshub/                       # WebSocket hub, connection registry, broadcast
│   ├── modules/                     # Module discovery and manifest parsing
│   └── system/                      # CPU temp, load, uptime, memory
├── web/
│   ├── templates/                   # HTML shell and login page
│   └── static/                      # CSS, JavaScript
├── modules/
│   ├── core/                        # System info panel + themes + i18n
│   └── lastHeard/                   # RF/network activity table
├── i18n/                            # Shell-level translation files
└── Makefile                         # Cross-compile targets
```

## Module System

Modules are self-contained directories under `modules/` containing a `module.json` manifest and associated HTML, JS, CSS, and i18n files. The dashboard discovers them at startup.

### Module Manifest

```json
{
  "name": "lastHeard",
  "displayName": "Last Heard",
  "version": "1.0.0",
  "type": "panel",
  "description": "Displays recent RF and network activity",
  "author": "Pi-Star Team",
  "menu": {
    "label": "Last Heard",
    "icon": "radio",
    "order": 10
  },
  "mqttTopics": ["mmdvm/display/#", "mmdvm/log/#"],
  "entryPoint": "panel.html",
  "scripts": ["panel.js"],
  "styles": ["style.css"]
}
```

### Module Types

- **`panel`** — displayed as a section on the main dashboard (e.g. system info, last heard)
- **`page`** — loaded as a separate page via navigation (e.g. config editor)

### Creating a Module

1. Create a directory under `modules/` with your module name
2. Add a `module.json` manifest declaring type, MQTT topics, and entry points
3. Use `data-i18n` attributes for all user-visible strings
4. Use `--ps-*` CSS custom properties for all colours (never hardcode)
5. Listen for `pistar:message` events on `document` to receive MQTT data

## Theming

Six pre-defined themes, each with light and dark variants:

| Theme | Light | Dark | Description |
|-------|:-----:|:----:|-------------|
| Default | yes | yes | Clean neutral grey/blue |
| Classic | yes | yes | Pi-Star legacy blue/purple |
| Night Ops | — | yes | Red-on-dark for night vision |
| High Contrast | yes | yes | WCAG AAA accessible |
| Midnight | — | yes | Deep navy/teal |
| Emerald | yes | yes | Green tones |

All colours are CSS custom properties (`--ps-bg-primary`, `--ps-accent`, `--ps-danger`, etc.). Theme files live in `modules/core/themes/`.

## API

### Configuration

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/config/{service}` | Read service config |
| PUT | `/api/config/{service}` | Write service config |

### System

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/system/info` | CPU, memory, temp, uptime |
| GET | `/api/system/services` | All managed service statuses |
| POST | `/api/system/services/{svc}` | Start / stop / restart a service |

### Tokens

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/tokens` | List API tokens |
| POST | `/api/tokens` | Create a new token |
| DELETE | `/api/tokens/{id}` | Revoke a token |

### Preferences

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/preferences` | Get user preferences |
| PUT | `/api/preferences` | Update preferences (theme, language) |

### WebSocket

```
wss://pistar.local/ws
```

Receives real-time MQTT messages and system status updates as JSON.

## Platform Support

| Platform | Architecture | Binary |
|----------|-------------|--------|
| All Raspberry Pis | ARMv6 | `pistar-dashboard-linux-arm` |
| Pi 3/4/5 (64-bit) | ARM64 | `pistar-dashboard-linux-arm64` |
| Debian/Ubuntu servers | x86_64 | `pistar-dashboard-linux-amd64` |
| Pi-Star_OS (Alpine/musl) | ARMv6, ARM64 | Same static binaries |

## Packaging

This repo produces the Go binary and module files. Packaging into `.deb` (systemd) and `.apk` (OpenRC) is handled by the [MMDVM_DEB](https://github.com/MW0MWZ/MMDVM_DEB) and [MMDVM_APK](https://github.com/MW0MWZ/MMDVM_APK) repos respectively.

## Dependencies

| Library | Purpose |
|---------|---------|
| [go-chi/chi](https://github.com/go-chi/chi) | HTTP router |
| [nhooyr.io/websocket](https://github.com/nhooyr/websocket) | WebSocket server |
| [paho.golang](https://github.com/eclipse/paho.golang) | MQTT 5.0 client |
| [go-ini/ini](https://github.com/go-ini/ini) | INI config parsing |
| [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) | Shadow hash verification |

All other functionality uses Go stdlib.

## Licence

[GPLv2](LICENSE)

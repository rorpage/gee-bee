# Gee Bee

An app that alerts you if it detects airplane tail numbers in flight!

Gee Bee is a one-shot Go binary for home servers. It polls [ADS-B Exchange](https://www.adsbexchange.com/) for a list of tail numbers you care about and fires a notification the moment one of them shows up on radar. No sighting goes unnoticed; no notification repeats while the aircraft stays in range.

It is a single Go binary with no runtime dependencies. Drop it on any Linux or macOS machine, point it at a config file (or env vars), and schedule it with cron or systemd.

### Quick notes

- This app uses quite a bit of code and logic from [Jetspotter](https://github.com/vvanouytsel/jetspotter). Many thanks to [Vincent](https://github.com/vvanouytsel) for his wonderful work there!
- There is an example config file included that you can modify for your own needs. Make a copy of it and name it `config.yml` or `config.yaml`. Sensible defaults are included in the app if you'd like to test drive it first.

## Features

- Watches any list of tail numbers/registrations, not tied to a specific airline or fleet
- One notification per continuous sighting -- no duplicates while the aircraft stays in range, even across restarts
- Notification fan-out: Slack, Discord, a generic custom webhook, and console logging can all be enabled simultaneously
- Rich notifications include callsign, registration, speed, altitude, heading, aircraft type/owner, a live tracker link, and a photo (via [planespotters.net](https://www.planespotters.net/)) when available
- Config via YAML file or environment variables
- Single static binary, no runtime dependencies, no Docker required

## Installation

### Download a pre-built binary (recommended)

Download the appropriate binary for your platform from the [Releases](https://github.com/rorpage/gee-bee/releases) page:

| Platform | File |
|---|---|
| Linux x86-64 | `gee-bee-linux-amd64` |
| Linux ARM64 (Raspberry Pi, etc.) | `gee-bee-linux-arm64` |
| macOS Intel | `gee-bee-darwin-amd64` |
| macOS Apple Silicon | `gee-bee-darwin-arm64` |
| Windows x86-64 | `gee-bee-windows-amd64.exe` |

```bash
# Example: Linux x86-64
curl -fsSL https://github.com/rorpage/gee-bee/releases/latest/download/gee-bee-linux-amd64 \
  -o /usr/local/bin/gee-bee
chmod +x /usr/local/bin/gee-bee
```

### Build from source

Requires Go 1.23+.

```bash
git clone https://github.com/rorpage/gee-bee.git
cd gee-bee
CGO_ENABLED=0 go build -ldflags="-s -w" -o gee-bee ./cmd/geebee
sudo mv gee-bee /usr/local/bin/
```

Or use the bundled [Task](https://taskfile.dev/) tasks -- see [Development](#development) below.

## Configuration

Create a config file at `~/.config/gee-bee/config.yml`:

```bash
mkdir -p ~/.config/gee-bee
cp config.example.yml ~/.config/gee-bee/config.yml
```

```yaml
tailNumbers: 28000,29000
discordWebhookUrl: https://discord.com/api/webhooks/12345/abcdef-ghijkl
```

See `config.example.yml` for a more complete example with all supported fields.

### Config fields

| Field | Required | Default | Description |
|---|---|---|---|
| `tailNumbers` | No | `28000,29000` | Comma-separated list of tail numbers/registrations to watch for |
| `logPlanesToConsole` | No | `true` | Print spotted-aircraft details to stdout in addition to the "spotted" log line |
| `discordWebhookUrl` | No | -- | Discord incoming webhook URL. Leave blank to disable Discord notifications. |
| `slackWebhookUrl` | No | -- | Slack incoming webhook URL. Leave blank to disable Slack notifications. |
| `customWebhookUrl` | No | -- | Any URL to receive a generic JSON POST. Leave blank to disable. |
| `stateFilePath` | No | `~/.config/gee-bee/state.json` | Where to persist the list of currently-spotted aircraft |

At least one of `discordWebhookUrl`, `slackWebhookUrl`, or `customWebhookUrl` should be set to receive notifications outside the console -- all three, none, or any combination can be active at once.

## Environment Variables

Every config field can be set via environment variable instead of (or as an override for) the config file:

| Variable | Description |
|---|---|
| `TAIL_NUMBERS` | Overrides `tailNumbers` |
| `LOG_PLANES_TO_CONSOLE` | Overrides `logPlanesToConsole` |
| `DISCORD_WEBHOOK_URL` | Overrides `discordWebhookUrl` |
| `SLACK_WEBHOOK_URL` | Overrides `slackWebhookUrl` |
| `CUSTOM_WEBHOOK_URL` | Overrides `customWebhookUrl` |
| `CONFIG_FILE` | Path to config file (default: searches `~/.config/gee-bee/` then the working directory for `config.yml`/`config.yaml`) |
| `STATE_FILE` | Path to state file (default: `~/.config/gee-bee/state.json`) |

## Notification Channels

A notification is sent once per newly spotted aircraft, per run, to every channel that has a URL configured (or to the console, if `logPlanesToConsole` is enabled).

### Discord

Set `discordWebhookUrl` (or `DISCORD_WEBHOOK_URL`). Each spotted aircraft is rendered as a rich embed -- callsign, registration, speed, altitude, heading, and type/owner as fields, a color that shifts with altitude, and a thumbnail photo when one is found.

### Slack

Set `slackWebhookUrl` (or `SLACK_WEBHOOK_URL`). Uses Slack's Block Kit for a per-aircraft section (same fields as Discord) followed by an image block when a photo is available.

### Custom webhook

Set `customWebhookUrl` (or `CUSTOM_WEBHOOK_URL`) to receive a plain JSON POST:

```json
{
  "altitude": "4500ft | 1371m",
  "callsign": "N28000",
  "description": "Boeing 747-200",
  "heading": "270",
  "imageUrl": "https://cdn.planespotters.net/...",
  "ownerOperator": "United States Air Force",
  "speed": "310kn | 574km/h",
  "squawk": "1200",
  "tailNumber": "28000",
  "type": "B742",
  "url": "https://globe.adsbexchange.com/?icao=..."
}
```

Note: if more than one tracked aircraft is spotted in the same run, only the first is sent to the custom webhook. Discord, Slack, and the console all receive the full batch.

### Console

Enabled by default (`logPlanesToConsole: true`). Prints a formatted block per spotted aircraft to stdout, in addition to the standard log output.

## Scheduling

The binary is a one-shot job: it runs, checks for tail numbers, and exits.

### systemd timer (recommended)

systemd timers have proper log capture via `journalctl`, survive reboots cleanly with `Persistent=true`, run as a dedicated non-root user, and are standard on any modern Linux distro. Ready-to-use files are in `deploy/systemd/`.

**Quick install:**

```bash
# From a cloned repo
sudo bash deploy/systemd/install.sh

# Or run directly from GitHub (no clone needed)
curl -fsSL https://raw.githubusercontent.com/rorpage/gee-bee/main/deploy/systemd/install.sh | sudo bash
```

The script downloads the latest binary from GitHub Releases (or builds from source if Go is available), creates a `gee-bee` system user, sets up `/etc/gee-bee/` and `/var/lib/gee-bee/`, and enables the timer. The env file pre-configures `CONFIG_FILE` and `STATE_FILE` to those system paths. Then:

```bash
# Set your tail numbers and webhook URL(s)
sudo nano /etc/gee-bee/env

# Start
sudo systemctl start gee-bee.timer
```

**Useful commands:**

```bash
systemctl status gee-bee.timer     # next scheduled run
systemctl start gee-bee.service    # run immediately
journalctl -u gee-bee -f           # follow logs
```

**Changing the schedule:** Edit `/etc/systemd/system/gee-bee.timer`, update `OnCalendar`, then:

```bash
sudo systemctl daemon-reload && sudo systemctl restart gee-bee.timer
```

Common values:

```ini
OnCalendar=minutely  # every minute (default)
OnCalendar=*:0/5     # every 5 minutes
```

### cron

Add a line to your crontab with `crontab -e`:

```cron
* * * * * DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/... /usr/local/bin/gee-bee
```

Or if you prefer an env file:

```cron
* * * * * env $(cat ~/.config/gee-bee/env | xargs) /usr/local/bin/gee-bee
```

Logs go to syslog (`journalctl -t gee-bee` or `/var/log/syslog`).

## How It Works

1. Load config from `CONFIG_FILE` (default: search `~/.config/gee-bee/` then the working directory for `config.yml`/`config.yaml`)
2. Load the list of currently-spotted aircraft from `STATE_FILE`
3. Query ADS-B Exchange for the configured tail numbers
4. Compare the results against the persisted spotted list; aircraft no longer in range are dropped from it
5. For each newly spotted aircraft, look up a photo (best-effort) and send a notification to every enabled channel
6. Save the updated spotted list to `STATE_FILE`

As long as the state file persists across runs, an aircraft will only notify once per continuous sighting -- it can notify again the next time it appears after having left range.

## Development

This project uses [Task](https://taskfile.dev/) for common commands:

- `task build`: build the app as an executable
- `task format`: format all of the code to Golang standards (`gofmt -s -w .`)
- `task run`: run the app once, pulling config from `config.yml`, `config.yaml`, or environment variables

```bash
go build ./...              # compile
go vet ./...                # static analysis
```

## Versioning

Releases are created automatically on every push to `main`. The version follows CalVer: `YYYY.MM.DD.N` where N is the GitHub Actions run number (e.g. `2026.07.03.4`). No manual tagging required.

The version is embedded in the binary at build time:

```bash
gee-bee --version
# 2026.07.03.4
```

When running locally from source, `--version` prints `dev`.

The latest release is always available at `https://github.com/rorpage/gee-bee/releases/latest`.

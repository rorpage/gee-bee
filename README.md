# Gee Bee

An app that alerts you if it detects airplane tail numbers in flight!

Gee Bee polls [ADS-B Exchange](https://www.adsbexchange.com/) for a list of tail numbers you care about and fires a notification the moment one of them shows up on radar. It runs continuously as a Docker container, checking on a configurable interval, and can notify Slack, Discord, a custom webhook, and/or the console -- any combination at once.

### Quick notes

- This app uses quite a bit of code and logic from [Jetspotter](https://github.com/vvanouytsel/jetspotter). Many thanks to [Vincent](https://github.com/vvanouytsel) for his wonderful work there!
- There is an example config file included that you can modify for your own needs. Make a copy of it and name it `config.yml` or `config.yaml`. Sensible defaults are included in the app if you'd like to test drive it first.

## Features

- Watches any list of tail numbers/registrations, not tied to a specific airline or fleet
- Polls on a configurable interval (default 60s)
- Notification fan-out: Slack, Discord, a generic custom webhook, and console logging can all be enabled simultaneously
- Rich notifications include callsign, registration, speed, altitude, heading, aircraft type/owner, a live tracker link, and a photo (via [planespotters.net](https://www.planespotters.net/)) when available
- Config via YAML file or environment variables
- Ships as a small, dependency-free Docker image

## Installation

### Docker (recommended)

Pull the published image from GitHub Container Registry:

```bash
docker pull ghcr.io/rorpage/gee-bee:latest
```

Run it with your config baked in as environment variables:

```bash
docker run -d \
  --name gee-bee \
  -e TAIL_NUMBERS="28000,29000" \
  -e FETCH_INTERVAL=60 \
  -e DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..." \
  ghcr.io/rorpage/gee-bee:latest
```

Or mount a config file instead of setting individual env vars:

```bash
docker run -d \
  --name gee-bee \
  -v $(pwd)/config.yml:/root/config.yml:ro \
  ghcr.io/rorpage/gee-bee:latest
```

### Build from source

Requires Go 1.23+ and [Task](https://taskfile.dev/) (optional, but used by the tasks below).

```bash
git clone https://github.com/rorpage/gee-bee.git
cd gee-bee
task build     # or: go build -o gee-bee.exe cmd/geebee/*
```

## Configuration

Copy the example config and edit it:

```bash
cp config.example.yml config.yml
```

```yaml
tailNumbers: 28000,29000
fetchInterval: 60
logPlanesToConsole: true
discordWebhookUrl: https://discord.com/api/webhooks/12345/abcdef-ghijkl
slackWebhookUrl: https://hooks.slack.com/services/random/ids/andstrings
customWebhookUrl: https://your-custom-url.com/endpoint
```

The config file must be named `config.yml` or `config.yaml` and live in the working directory the app is run from. If no config file is found, the app runs with the defaults below.

### Config fields

| Field | Required | Default | Description |
|---|---|---|---|
| `tailNumbers` | No | `28000,29000` | Comma-separated list of tail numbers/registrations to watch for |
| `fetchInterval` | No | `60` | Seconds between polls. Values under 60 will log a rate-limit warning. |
| `logPlanesToConsole` | No | `true` | Print spotted-aircraft details to stdout in addition to the "spotted" log line |
| `discordWebhookUrl` | No | -- | Discord incoming webhook URL. Leave blank to disable Discord notifications. |
| `slackWebhookUrl` | No | -- | Slack incoming webhook URL. Leave blank to disable Slack notifications. |
| `customWebhookUrl` | No | -- | Any URL to receive a generic JSON POST. Leave blank to disable. |

At least one of `discordWebhookUrl`, `slackWebhookUrl`, or `customWebhookUrl` should be set to receive notifications outside the console -- all three, none, or any combination can be active at once.

## Environment Variables

Every config field can be set via environment variable instead of (or as an override for) the config file:

| Variable | Overrides |
|---|---|
| `TAIL_NUMBERS` | `tailNumbers` |
| `FETCH_INTERVAL` | `fetchInterval` |
| `LOG_PLANES_TO_CONSOLE` | `logPlanesToConsole` |
| `DISCORD_WEBHOOK_URL` | `discordWebhookUrl` |
| `SLACK_WEBHOOK_URL` | `slackWebhookUrl` |
| `CUSTOM_WEBHOOK_URL` | `customWebhookUrl` |

## Notification Channels

A notification is sent once per newly spotted aircraft, per poll, to every channel that has a URL configured (or to the console, if `logPlanesToConsole` is enabled).

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

Note: if more than one tracked aircraft is spotted in the same poll, only the first is sent to the custom webhook. Discord, Slack, and the console all receive the full batch.

### Console

Enabled by default (`logPlanesToConsole: true`). Prints a formatted block per spotted aircraft to stdout, in addition to the container logs.

## How It Works

1. Load config from `config.yml`/`config.yaml` in the working directory (or env vars/defaults if none is found)
2. Loop forever: every `fetchInterval` seconds, query ADS-B Exchange for the configured tail numbers
3. Compare the results against the in-memory list of already-spotted aircraft for this run
4. For each newly spotted aircraft, look up a photo (best-effort) and send a notification to every enabled channel
5. Update the in-memory spotted list and sleep until the next poll

Because the spotted list lives only in memory, restarting the container resets it -- a still-airborne tracked aircraft can notify again after a restart.

## Development

This project uses [Task](https://taskfile.dev/) for common commands:

- `task build`: build the app as an executable
- `task format`: format all of the code to Golang standards (`gofmt -s -w .`)
- `task run`: run the app, pulling config from `config.yml`, `config.yaml`, or environment variables

## Versioning

There are no tagged releases; the Docker image is the published artifact. `.github/workflows/docker-build.yml` builds and pushes to `ghcr.io/rorpage/gee-bee` on every push to `main`, tagged both `latest` and `YYYY.MM.DD-<short-sha>`.

# AGENTS.md

Context for AI agents (Claude Code, Copilot, etc.) working in this repository.

## What this project does

Gee Bee is a long-running Go service that polls the ADS-B Exchange (`api.adsb.one`) for a configured list of aircraft tail numbers and sends a notification the moment one of them is airborne and detected. It runs in a loop (not one-shot), fetching on a configurable interval, and is designed to run as a Docker container. It is a fork/derivative of [Jetspotter](https://github.com/vvanouytsel/jetspotter) by Vincent Van Ouytsel.

## Repository layout

```
cmd/geebee/main.go                        -- entry point; loop, dispatches to fetch + notification

internal/configuration/configuration.go   -- config types, loading from YAML file + env var overrides (Viper)
internal/geebee/geebee.go                 -- ADS-B fetch, in-memory dedup, output shaping, distance/bearing helpers
internal/geebee/types.go                  -- FlightData/Aircraft (raw API shape) and AircraftOutput (normalized shape)
internal/planespotter/planespotter.go     -- fetches an aircraft photo from planespotters.net by ICAO hex or registration
internal/notification/message.go          -- shared Notification type, HTTP POST + per-field formatting helpers
internal/notification/discord.go          -- Discord embed payload (bwmarrin/discordgo), colored by altitude
internal/notification/slack.go            -- Slack Block Kit payload
internal/notification/custom-url.go       -- generic JSON payload for any webhook URL
internal/notification/terminal.go         -- stdout/log output

go.mod                        -- module `geebee`; Viper for config, discordgo for Discord, jftuga/geodist for distance math
config.example.yml            -- copy to config.yml (gitignored) to run locally
Taskfile.yml                  -- build / format / run tasks

Dockerfile                    -- multi-stage build; alpine runtime image
.github/workflows/
  docker-build.yml            -- on push to main (or manual dispatch): builds and pushes image to GHCR
```

## Key design decisions

- **Continuous polling loop, not one-shot.** `HandleGeeBee` in `main.go` loops forever, sleeping `fetchInterval` seconds (default 60) between fetches. This is a different model from the `game-over-man`/`now-you-noaa` siblings, which are one-shot binaries scheduled externally by cron/systemd -- gee-bee owns its own scheduling and is meant to run as a long-lived container.
- **Docker-first deployment, no systemd/cron path.** The Dockerfile bakes in default env vars and the GitHub Action publishes to GHCR on every push to `main`. There is no `deploy/systemd` equivalent here; this app is not intended to run as a bare binary on a home server the way the sibling projects are.
- **In-memory dedup only, no state file.** `alreadySpottedAircraft` is a plain slice held in the process, not persisted to disk. Restarting the container (e.g. redeploying) resets the "already notified" set, so a still-airborne tracked aircraft can trigger a duplicate notification after a restart. This is a real difference from the siblings' `state.json` approach -- there is no idempotency across restarts.
- **Config-file-first, env-var override.** Viper reads `config.yml`/`config.yaml` from the working directory and binds each key to a `SCREAMING_SNAKE_CASE` env var (e.g. `TAIL_NUMBERS`, `FETCH_INTERVAL`) that overrides it. No `CONFIG_FILE`-style path override exists -- the file must be named `config.yml`/`config.yaml` and live in the process's working directory.
- **Multiple notification channels can fire independently per event.** Unlike a single `notificationType` selector, `sendNotifications` in `main.go` fires Custom URL, Discord, Slack, and terminal output independently -- whichever webhook URLs are non-empty all receive the notification for the same spotted aircraft.
- **Tail numbers are looked up directly via ADS-B Exchange's `/v2/reg/{tailNumbers}` endpoint** (comma-joined), not by geographic radius. Distance (`CalculateDistance`) and bearing (`CalculateBearing`) helpers exist in `geebee.go` but are currently unused by any caller -- they're carried over from Jetspotter's location-based spotting model and are not wired up.
- **Aircraft photos are best-effort.** `planespotter.GetImageFromAPI` tries planespotters.net by ICAO hex first, falls back to registration, and returns an empty `Image{}` (not an error) if neither matches. Notifications render fine with no image.
- **Custom URL notification sends only the first aircraft in a batch.** `buildCustomMessage` in `custom-url.go` indexes `aircraft[0]` -- if more than one tracked aircraft is newly spotted in the same poll, only one is delivered to the custom webhook. Discord, Slack, and terminal output all include the full batch.

## Config fields

| Field | YAML key | Env var | Default | Description |
|---|---|---|---|---|
| Tail numbers | `tailNumbers` | `TAIL_NUMBERS` | `28000,29000` | Comma-separated list of tail numbers/registrations to watch for |
| Fetch interval | `fetchInterval` | `FETCH_INTERVAL` | `60` | Seconds between polls. Values under 60 log a rate-limit warning. |
| Log planes to console | `logPlanesToConsole` | `LOG_PLANES_TO_CONSOLE` | `true` | Whether spotted-aircraft details are printed to stdout in addition to the "spotted" log line |
| Custom webhook URL | `customWebhookUrl` | `CUSTOM_WEBHOOK_URL` | `""` | Generic JSON POST target; empty disables this channel |
| Discord webhook URL | `discordWebhookUrl` | `DISCORD_WEBHOOK_URL` | `""` | Discord incoming webhook URL; empty disables this channel |
| Slack webhook URL | `slackWebhookUrl` | `SLACK_WEBHOOK_URL` | `""` | Slack incoming webhook URL; empty disables this channel |

Config file location: `config.yml` or `config.yaml` in the working directory. No config file is required -- the app runs with the built-in defaults above if none is found.

## Notification payload shapes

**Custom URL** (`internal/notification/custom-url.go`, first aircraft in batch only):

```go
type CustomUrlMessage struct {
    Altitude      string `json:"altitude"`
    Callsign      string `json:"callsign"`
    Description   string `json:"description"`
    Heading       string `json:"heading"`
    ImageUrl      string `json:"imageUrl"`
    OwnerOperator string `json:"ownerOperator"`
    Speed         string `json:"speed"`
    Squawk        string `json:"squawk"`
    TailNumber    string `json:"tailNumber"`
    Type          string `json:"type"`
    Url           string `json:"url"`
}
```

**Discord** -- one `discordgo.MessageEmbed` per aircraft, `Color` set by altitude band (see `getColorByAltitude` in `discord.go`), thumbnail set to the planespotters.net image if available.

**Slack** -- Block Kit message: one intro section, then per aircraft a `section` block (callsign/registration/speed/altitude/heading/type as `mrkdwn` fields) followed by an `image` block (if a photo was found) and a `divider`.

**Terminal** -- plain formatted text via `FormatAircraft`, gated by `logPlanesToConsole`.

## ADS-B API

Base URL: `https://api.adsb.one/v2`

Endpoint used: `GET /v2/reg/{tailNumbers}` where `{tailNumbers}` is a comma-joined list. Response is unmarshaled into `FlightData.AC` (`[]Aircraft`, see `types.go` for the full field set, sourced from the [ADS-B Exchange field explanations](https://www.adsbexchange.com/ads-b-data-field-explanations/)). `DbFlags == 1` marks an aircraft as military (`isAircraftMilitary`).

## Development workflow

```bash
go build ./...              # compile
go vet ./...                # static analysis
gofmt -s -l .                # check formatting (task `format` applies fixes with -w)

task build                  # build cmd/geebee -> gee-bee.exe (see Taskfile.yml)
task run                    # go run cmd/geebee/*, reads config.yml/config.yaml or env vars
```

To test locally without a real webhook, run a listener in another terminal and point `customWebhookUrl`/env at it:

```bash
python3 -m http.server 3001
```

## GitHub Actions

`.github/workflows/docker-build.yml` triggers on every push to `main` (and via manual `workflow_dispatch`). It computes a version string (`YYYY.MM.DD-<short-sha>`), builds the Docker image with Buildx, and pushes it to `ghcr.io/<repo>` tagged both `latest` and the computed version, with OCI source/revision/version labels. There is no separate CI-on-PR workflow and no GitHub Releases -- the container image is the only published artifact.

## Style conventions

- Standard Go idioms; run `go vet` and `gofmt -s` before committing
- Log lines are plain `log.Printf`/`log.Println` (no `[module]` prefix convention here, unlike the sibling projects)
- Do not write comments that explain what the code does -- only add one when the WHY is non-obvious
- No em dashes anywhere in code or documentation

## Files to keep updated

When making changes, keep README.md and AGENTS.md in sync:
- New config fields -> Configuration table in README.md and Config fields table in this file
- New notification channels -> README.md notification section and this file's payload shapes section
- Architectural changes -> Key design decisions section in this file

# AGENTS.md

Context for AI agents (Claude Code, Copilot, etc.) working in this repository.

## What this project does

Gee Bee is a one-shot Go binary for home servers. It queries ADS-B Exchange (`api.adsb.one`) for a configured list of aircraft tail numbers, finds aircraft newly detected in range, and sends a notification per aircraft to every enabled channel (Discord, Slack, custom webhook, console). Idempotency within a continuous sighting is maintained via a JSON state file. It runs directly on Linux or macOS -- no Docker, no runtime dependencies. It is a fork/derivative of [Jetspotter](https://github.com/vvanouytsel/jetspotter) by Vincent Van Ouytsel.

## Repository layout

```
cmd/geebee/main.go                        -- entry point; orchestrates config, fetch, notify, state

internal/configuration/configuration.go   -- config types, loading from YAML file + env var overrides (Viper)
internal/geebee/geebee.go                 -- ADS-B fetch, dedup logic, output shaping, distance/bearing helpers
internal/geebee/types.go                  -- FlightData/Aircraft (raw API shape) and AircraftOutput (normalized shape)
internal/geebee/state.go                  -- reads/writes the state file
internal/planespotter/planespotter.go     -- fetches an aircraft photo from planespotters.net by ICAO hex or registration
internal/notification/message.go          -- shared Notification type, HTTP POST + per-field formatting helpers
internal/notification/discord.go          -- Discord embed payload (bwmarrin/discordgo), colored by altitude
internal/notification/slack.go            -- Slack Block Kit payload
internal/notification/custom-url.go       -- generic JSON payload for any webhook URL
internal/notification/terminal.go         -- stdout/log output

go.mod                        -- module `geebee`; Viper for config, discordgo for Discord, jftuga/geodist for distance math
config.example.yml            -- copy to config.yml (gitignored) to run locally
Taskfile.yml                  -- build / format / run tasks

.github/workflows/
  ci.yml        -- on PR to main: vet + cross-compile for 5 platforms, upload as PR artifact
  publish.yml   -- on push to main: cross-compiles binaries for 5 platforms, creates a GitHub Release

deploy/
  systemd/
    gee-bee.service   -- oneshot unit; runs binary as gee-bee user
    gee-bee.timer     -- fires every minute, Persistent=true
    env.example       -- template for /etc/gee-bee/env (holds tail numbers + webhook URLs)
    install.sh        -- downloads binary (or builds from source), creates user/dirs, enables timer
```

## Key design decisions

- **Single static binary, no runtime dependencies.** CGO is disabled so the binary runs on any Linux/macOS without glibc or other shared libraries.
- **No Docker.** This project targets home server use, scheduled by cron or a systemd timer. Docker adds path, permission, and user complexity that the Go binary approach eliminates entirely.
- **One-shot execution.** The binary runs, checks for tail numbers currently in range, and exits. Scheduling is the caller's responsibility. There is no internal sleep loop.
- **State file preserves "currently in range" across runs, not a permanent log.** `state.json` (`internal/geebee/state.go`) stores the list of aircraft the previous run considered spotted. Each run replaces this list with whichever of those aircraft are still detected -- aircraft that drop out of range are removed from state and will notify again the next time they're detected. This differs from the `game-over-man`/`now-you-noaa` siblings, whose state files are an append-only, day-pruned log of one-time events (games, alerts) that never repeat; a given aircraft sighting here is expected to start and stop many times, so "still present vs. no longer present" is the correct dedup key, not calendar-day pruning.
- **Config-file-first, env-var override.** Viper reads `config.yml`/`config.yaml`, checking `~/.config/gee-bee/` then the working directory, unless `CONFIG_FILE` points at an explicit path. Every config key has a matching `SCREAMING_SNAKE_CASE` env var (`TAIL_NUMBERS`, `DISCORD_WEBHOOK_URL`, etc.) that overrides it. `STATE_FILE` overrides `stateFilePath` the same way.
- **Multiple notification channels fire independently per event.** Unlike a single `notificationType` selector, `sendNotifications` in `main.go` fires Custom URL, Discord, Slack, and terminal output independently -- whichever webhook URLs are non-empty all receive the notification for the same spotted aircraft.
- **Tail numbers are looked up directly via ADS-B Exchange's `/v2/reg/{tailNumbers}` endpoint** (comma-joined), not by geographic radius. Distance (`CalculateDistance`) and bearing (`CalculateBearing`) helpers exist in `geebee.go` but are currently unused by any caller -- they're carried over from Jetspotter's location-based spotting model and are not wired up.
- **Aircraft photos are best-effort.** `planespotter.GetImageFromAPI` tries planespotters.net by ICAO hex first, falls back to registration, and returns an empty `Image{}` (not an error) if neither matches. Notifications render fine with no image.
- **Custom URL notification sends only the first aircraft in a batch.** `buildCustomMessage` in `custom-url.go` indexes `aircraft[0]` -- if more than one tracked aircraft is newly spotted in the same run, only one is delivered to the custom webhook. Discord, Slack, and terminal output all include the full batch.

## Config fields

| Field | YAML key | Env var | Default | Description |
|---|---|---|---|---|
| Tail numbers | `tailNumbers` | `TAIL_NUMBERS` | `28000,29000` | Comma-separated list of tail numbers/registrations to watch for |
| Log planes to console | `logPlanesToConsole` | `LOG_PLANES_TO_CONSOLE` | `true` | Whether spotted-aircraft details are printed to stdout in addition to the "spotted" log line |
| Custom webhook URL | `customWebhookUrl` | `CUSTOM_WEBHOOK_URL` | `""` | Generic JSON POST target; empty disables this channel |
| Discord webhook URL | `discordWebhookUrl` | `DISCORD_WEBHOOK_URL` | `""` | Discord incoming webhook URL; empty disables this channel |
| Slack webhook URL | `slackWebhookUrl` | `SLACK_WEBHOOK_URL` | `""` | Slack incoming webhook URL; empty disables this channel |
| State file path | `stateFilePath` | `STATE_FILE` | `~/.config/gee-bee/state.json` | Where to persist the currently-spotted aircraft list |

## Default file paths

| Purpose | Default |
|---|---|
| Config | `~/.config/gee-bee/config.yml` (also checks the working directory) |
| State | `~/.config/gee-bee/state.json` |

Both are overridable via `CONFIG_FILE` and `STATE_FILE` respectively. The systemd deploy files point these at `/etc/gee-bee/config.yml` and `/var/lib/gee-bee/state.json` instead.

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

task build                  # build cmd/geebee -> gee-bee (see Taskfile.yml)
task run                    # go run ./cmd/geebee, reads config.yml/config.yaml or env vars, runs once
```

To test locally without a real webhook, run a listener in another terminal and point `customWebhookUrl`/`DISCORD_WEBHOOK_URL`/etc. at it:

```bash
python3 -m http.server 3001
```

## Versioning

Releases use CalVer: `YYYY.MM.DD.N` where N is `github.run_number`. The version is embedded at build time via `-ldflags "-X main.version=..."` into the `version` variable in `main.go`. Running from source prints `dev`. No manual tagging; every push to main creates a release.

## GitHub Actions

- `.github/workflows/ci.yml` triggers on pull requests to `main`. It runs `go vet`, cross-compiles for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and windows/amd64, and uploads the binaries as a PR artifact.
- `.github/workflows/publish.yml` triggers on every push to `main`. It computes the CalVer string, cross-compiles the same 5 platform targets (with the version embedded via ldflags), then creates a GitHub Release tagged with the CalVer string using `softprops/action-gh-release` with `make_latest: true`.

## Scheduling

Two scheduling options are documented in README.md:
- **systemd timer** -- recommended; dedicated user; logs via `journalctl`; deploy files in `deploy/systemd/`
- **cron** -- simpler; one crontab line; logs go to syslog

Default schedule is every minute (`OnCalendar=minutely` in `deploy/systemd/gee-bee.timer`). ADS-B Exchange has not published a documented hard rate limit for this endpoint, but avoid scheduling much more frequently than once a minute.

## Working with the developer

Push back when a suggestion would make the code worse -- don't just agree. If the developer proposes something that adds unnecessary complexity, breaks a design principle, or has a better alternative, say so directly and explain why. The goal is a dialog, not compliance. It is fine (expected, even) to advocate for a different approach before implementing anything.

## Style conventions

- Standard Go idioms; run `go vet` and `gofmt -s` before committing
- Do not write comments that explain what the code does -- only add one when the WHY is non-obvious
- No em dashes anywhere in code or documentation

## Files to keep updated

When making changes, keep README.md and AGENTS.md in sync:
- New config fields -> Config fields table in README.md and this file
- New env vars -> Environment Variables table in README.md and this file
- New notification channels -> README.md notification section and this file's payload shapes section
- Architectural changes -> Key design decisions section in this file

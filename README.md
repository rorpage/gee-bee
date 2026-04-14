# Gee Bee
An app that alerts you if it detects airplane tail numbers in flight!

### Quick notes
- This app uses quite a bit of code and logic from [Jetspotter](https://github.com/vvanouytsel/jetspotter). Many thanks to [Vincent](https://github.com/vvanouytsel) for his wonderful work there!
- There is an example config file included that you can modify for your own needs. Make a copy of it and name it `config.yml` or `config.yaml`. Sensible defaults are included the app if you'd like to testdrive it first.

# Taskfile tasks
- `build`: build the app as an executable
- `format`: format all of the code to Golang standards
- `run`: run the app

For instance, `task run` will run the app and pull config values from a `config.yml`, `config.yaml`, or environment variables.

# Prometheus metrics

Gee Bee exposes a Prometheus metrics endpoint at `http://localhost:9090/metrics` (port configurable via `metricsPort` / `METRICS_PORT`).

| Metric | Type | Description |
|---|---|---|
| `geebee_aircraft_spotted_total` | Counter | Total aircraft spotted, labelled by `tail_number` |
| `geebee_api_calls_total` | Counter | Total calls made to the upstream ADS-B API |
| `geebee_api_errors_total` | Counter | Total errors returned by the upstream ADS-B API |
| `geebee_notifications_sent_total` | Counter | Total notifications sent, labelled by `channel` (`discord`, `slack`, `custom_url`, `terminal`) |

# Grafana dashboard

A pre-built Grafana dashboard is included and can be imported manually or provisioned automatically.

## Automatic provisioning (Docker Compose)

```bash
docker compose up
```

This starts Gee Bee, Prometheus (port 9091), and Grafana (port 3000). The dashboard is automatically provisioned and available at [http://localhost:3000](http://localhost:3000).

## Manual import

Import `grafana/provisioning/dashboards/gee-bee.json` into your Grafana instance.

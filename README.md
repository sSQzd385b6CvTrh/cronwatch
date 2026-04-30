# cronwatch

Lightweight daemon that monitors cron job execution times and alerts on drift or missed runs.

## Installation

```bash
go install github.com/yourname/cronwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/cronwatch.git && cd cronwatch && go build ./...
```

## Usage

Define your monitored jobs in a `cronwatch.yaml` config file:

```yaml
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    tolerance: 5m
    alert: slack

  - name: hourly-sync
    schedule: "0 * * * *"
    tolerance: 2m
    alert: email
```

Start the daemon:

```bash
cronwatch --config cronwatch.yaml
```

Signal job completion from your cron scripts:

```bash
# Add to the end of your cron script
cronwatch ping daily-backup
```

cronwatch will alert you if a job exceeds its expected schedule window or fails to check in entirely.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `cronwatch.yaml` | Path to config file |
| `--log-level` | `info` | Log verbosity (debug, info, warn) |
| `--port` | `9107` | HTTP metrics port |

## License

MIT © yourname
# Running ComposeGuard with systemd

ComposeGuard can be executed automatically on a Linux server using systemd timer.

This is useful for VPS deployments where you want periodic checks for:

- Docker container status
- Disk usage
- HTTP endpoint health
- SSL certificate expiry
- Telegram alerts

## Recommended paths

| File     | Path                                       |
| -------- | ------------------------------------------ |
| Binary   | `/usr/local/bin/composeguard`              |
| Config   | `/etc/composeguard/composeguard.yaml`      |
| Env file | `/etc/composeguard/composeguard.env`       |
| Service  | `/etc/systemd/system/composeguard.service` |
| Timer    | `/etc/systemd/system/composeguard.timer`   |

## Install binary

Download the release binary and place it in `/usr/local/bin`.

Example:

```bash
curl -L -o composeguard https://github.com/IKHINtech/composeguard/releases/download/v0.2.0/composeguard-linux-amd64
chmod +x composeguard
sudo mv composeguard /usr/local/bin/composeguard
composeguard version
```

Or install from source:

```bash
go install github.com/IKHINtech/composeguard/cmd/composeguard@latest
```

If installed with `go install`, copy the binary to `/usr/local/bin`:

```bash
sudo cp "$(which composeguard)" /usr/local/bin/composeguard
```

## Run automatically with systemd

ComposeGuard can run periodically using systemd timer.

Install systemd files:

```bash
sudo composeguard install-systemd
```

this command creates

```bash
/etc/composeguard/composeguard.yaml
/etc/composeguard/composeguard.env
/etc/systemd/system/composeguard.service
/etc/systemd/system/composeguard.timer
```

## Edit Config

```bash
sudo nano /etc/composeguard/composeguard.yaml
```

Example

```yaml
project_name: "production-server"

docker:
  containers:
    - nginx
    - postgres
    - redis
    - api

disk:
  paths:
    - path: "/"
      warning_percent: 80
      critical_percent: 90
    - path: "/var/lib/docker"
      warning_percent: 80
      critical_percent: 90

http:
  endpoints:
    - name: "API Health"
      url: "https://api.example.com/health"
      expected_status: 200
      timeout_seconds: 5

ssl:
  domains:
    - "example.com"
    - "api.example.com"
  warning_days: 30
  critical_days: 7

notification:
  telegram:
    enabled: true
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
    only_on_problem: true
```

## Edit Telegram Env

```bash
sudo nano /etc/composeguard/composeguard.env
```

Example

```env
TELEGRAM_BOT_TOKEN=123456789:ABCDEF
TELEGRAM_CHAT_ID=123456789
```

## Test Manually

```bash
sudo /usr/local/bin/composeguard check --config /etc/composeguard/composeguard.yaml --notify telegram
```

## Start Service once

```bash
sudo systemctl daemon-reload
sudo systemctl start composeguard.service
sudo systemctl status composeguard.service
```

## Enable Timer

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now composeguard.timer
```

## Check Log

```bash
journalctl -u composeguard.service -n 100 --no-pager
```

## View status

```bash
systemctl status composeguard.timer
```

## Change Interval

```bash
sudo nano /etc/systemd/system/composeguard.timer

OnUnitActiveSec=5min
OnUnitActiveSec=15min
OnUnitActiveSec=1h
```

Tambahkan di config example:

```yaml
notification:
  telegram:
    enabled: true
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
    only_on_problem: true
    cooldown_minutes: 60

state:
  path: "/var/lib/composeguard/state.json"
```

## State file

ComposeGuard uses a state file to prevent repeated Telegram alerts for the same problem.

Default systemd path:

```text
/var/lib/composeguard/state.json
```

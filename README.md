# composeguard

A lightweight Go CLI to monitor Docker Compose services, disk usage, HTTP health
checks, and SSL expiry for small VPS deployments.

## Features

- Check Docker container status
- Check disk usage
- Check HTTP health endpoints
- Check SSL certificate expiry
- Simple YAML configuration
- Useful for small VPS, self-hosted apps, and Docker Compose deployments

## Install

```bash
go install github.com/IKHINtech/composeguard/cmd/composeguard@latest
```

## Commands

### Initialize config

```bash
composeguard init
```

### Config Example

```yaml
docker:
  containers:
    - nginx
    - postgres

  system_df:
    enabled: true

    images:
      warning_gb: 10
      critical_gb: 30

    containers:
      warning_gb: 5
      critical_gb: 10

    local_volumes:
      warning_gb: 10
      critical_gb: 30

    build_cache:
      warning_gb: 10
      critical_gb: 30
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
    - "api.example.com"
    - "example.com"
  warning_days: 30
  critical_days: 7

notification:
  telegram:
    enabled: true
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
    only_on_problem: true
```

### Run All Check

```bash
composeguard check
```

### Use custom config

```bash
composeguard check --config /etc/composeguard.yaml
```

### Output JSON

```bash
composeguard check --json
```

### Run specific check

```bash
composeguard check --only docker
composeguard check --only disk
composeguard check --only http
composeguard check --only ssl
```

### Run with telegram Notification

```bash
composeguard check --notify telegram

```

### release

```bash
git tag -a v0.1.0 -m "composeguard v0.1.0"
git push origin v0.1.0
```

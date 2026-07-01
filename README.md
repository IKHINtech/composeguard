# composeguard

A lightweight Go CLI to monitor Docker Compose services, disk usage, HTTP health checks, and SSL expiry for small VPS deployments.

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

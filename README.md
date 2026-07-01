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

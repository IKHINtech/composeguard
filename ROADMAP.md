# ComposeGuard Roadmap

ComposeGuard is a lightweight Go CLI for monitoring small VPS and Docker Compose based deployments. It checks Docker containers, disk usage, HTTP health endpoints, SSL certificate expiry, and sends alerts such as Telegram notifications.

This roadmap is written as an implementation guide. Each version contains the goal, feature scope, expected commands, configuration shape, implementation notes, acceptance criteria, and non-goals.

---

## Product Direction

ComposeGuard should stay small, reliable, and easy to operate.

The primary target users are:

- solo developers running apps on a VPS;
- small teams using Docker Compose instead of Kubernetes;
- self-hosted app maintainers;
- backend developers who need simple health checks without deploying a full monitoring stack.

Core principles:

- CLI first.
- Minimal dependencies.
- Production-safe defaults.
- Clear exit codes.
- Human-readable output by default.
- JSON output for automation.
- Configuration through YAML and environment variables.
- No heavy UI until the CLI and automation workflow are stable.

---

## Current Baseline: v0.1.x

### Status

Initial MVP has been implemented.

### Existing Features

- `composeguard init`
- `composeguard check`
- `composeguard check --config composeguard.yaml`
- `composeguard check --json`
- `composeguard check --only docker`
- `composeguard check --only disk`
- `composeguard check --only http`
- `composeguard check --only ssl`
- `composeguard check --notify telegram`
- `composeguard version`
- YAML configuration.
- Docker container status check through Docker CLI.
- Disk usage check.
- HTTP endpoint status check.
- SSL certificate expiry check.
- Telegram notification.
- GitHub Actions CI.
- GitHub Actions release build.

### Current Exit Codes

| Exit Code | Meaning |
|---:|---|
| `0` | All checks are OK |
| `1` | At least one warning exists |
| `2` | At least one critical issue exists |

### Current Example Config

```yaml
project_name: "my-server"

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

### Known Issues / Improvements Needed

- `go install ...@latest` may show `composeguard dev` if version is only injected through release `ldflags`.
- Windows release build may fail if disk check uses Unix-specific `syscall.Statfs`.
- Telegram notification has no cooldown, so scheduled runs may spam alerts.
- No config validation command yet.
- No official systemd documentation yet.
- Docker check requires manual container names.

---

## v0.2.0 — Systemd Timer and Production Installation Docs

### Goal

Make ComposeGuard usable as an automated VPS health checker that runs periodically and sends Telegram alerts when problems are detected.

This version focuses on documentation and operational readiness, not new complex runtime logic.

### Features

#### 1. Systemd Service Example

Add documentation and example file for running ComposeGuard as a one-shot systemd service.

Recommended path:

```text
/etc/composeguard/composeguard.yaml
/etc/composeguard/composeguard.env
/etc/systemd/system/composeguard.service
/etc/systemd/system/composeguard.timer
```

Example service:

```ini
[Unit]
Description=ComposeGuard health check
Wants=network-online.target docker.service
After=network-online.target docker.service

[Service]
Type=oneshot
EnvironmentFile=/etc/composeguard/composeguard.env
ExecStart=/usr/local/bin/composeguard check --config /etc/composeguard/composeguard.yaml --notify telegram
```

#### 2. Systemd Timer Example

Example timer:

```ini
[Unit]
Description=Run ComposeGuard every 10 minutes

[Timer]
OnBootSec=1min
OnUnitActiveSec=10min
Unit=composeguard.service

[Install]
WantedBy=timers.target
```

#### 3. Environment File Example

Example `/etc/composeguard/composeguard.env`:

```env
TELEGRAM_BOT_TOKEN=123456:example-token
TELEGRAM_CHAT_ID=123456789
```

The environment file must not be committed to Git.

#### 4. Installation Guide

Add `docs/systemd.md` with:

- how to download binary;
- how to install to `/usr/local/bin/composeguard`;
- how to create `/etc/composeguard`;
- how to copy config;
- how to create env file;
- how to install service and timer;
- how to start timer;
- how to check logs.

Commands:

```bash
sudo mkdir -p /etc/composeguard
sudo cp composeguard.yaml /etc/composeguard/composeguard.yaml
sudo nano /etc/composeguard/composeguard.env
sudo chmod 600 /etc/composeguard/composeguard.env
sudo systemctl daemon-reload
sudo systemctl enable --now composeguard.timer
systemctl list-timers composeguard.timer
journalctl -u composeguard.service -n 100 --no-pager
```

#### 5. README Update

README should include a short section:

```md
## Run with systemd

See docs/systemd.md.
```

### Acceptance Criteria

- A user can install ComposeGuard on a Linux VPS and run it automatically every 10 minutes.
- Telegram credentials are stored outside YAML through environment variables.
- Documentation includes copy-paste-ready systemd service and timer examples.
- Documentation explains how to inspect timer status and service logs.

### Non-goals

- No built-in daemon mode yet.
- No Prometheus exporter yet.
- No stateful alert cooldown yet.

---

## v0.3.0 — Docker System Disk Usage Check

### Goal

Detect Docker disk usage problems, especially large build cache, images, containers, and volumes.

This solves a common VPS issue: Docker silently consuming huge disk space.

### Features

#### 1. Docker System DF Check

Implement check based on:

```bash
docker system df --format '{{json .}}'
```

or another stable Docker CLI output format.

The check should report:

- images total size;
- containers total size;
- local volumes total size;
- build cache total size.

#### 2. Config

Add config under `docker.system_df`:

```yaml
docker:
  containers:
    - nginx
    - postgres

  system_df:
    enabled: true
    build_cache_warning_gb: 20
    build_cache_critical_gb: 50
    images_warning_gb: 20
    images_critical_gb: 50
    volumes_warning_gb: 20
    volumes_critical_gb: 50
```

#### 3. Output Example

```text
✓ OK       Docker: nginx                container nginx is running
⚠ WARNING  Docker Build Cache           32GB used, warning threshold 20GB
✓ OK       Docker Images                5GB used
✓ OK       Docker Volumes               2GB used
```

#### 4. JSON Output

JSON output should include Docker system usage results using the existing result format:

```json
{
  "name": "Docker Build Cache",
  "status": "WARNING",
  "message": "32GB used, warning threshold 20GB"
}
```

#### 5. Helper Parser

Implementation should include a parser that converts Docker size strings such as:

```text
12B
10kB
120MB
1.5GB
```

to bytes or gigabytes.

### Acceptance Criteria

- `composeguard check --only docker` includes Docker container checks and Docker system disk usage checks when enabled.
- If `system_df.enabled` is false or missing, behavior remains unchanged.
- If Docker CLI is unavailable, result is `CRITICAL` with a clear error.
- Size parser has unit tests.
- No panic on empty Docker output.

### Non-goals

- Do not automatically run `docker system prune`.
- Do not delete images, volumes, or cache.
- Do not require Docker SDK yet.

---

## v0.4.0 — Docker Container Discovery

### Goal

Reduce manual config work by allowing users to discover running or existing Docker containers and generate config snippets.

### Features

#### 1. New Command

```bash
composeguard discover
```

Default behavior should print a YAML snippet to stdout.

Example output:

```yaml
docker:
  containers:
    - obat-in-api
    - obat-in-postgres
    - obat-in-redis
```

#### 2. Include All Containers

Add flag:

```bash
composeguard discover --all
```

Behavior:

- default: discover running containers only;
- `--all`: include stopped/exited containers from `docker ps -a`.

#### 3. Write to File

Add flag:

```bash
composeguard discover --write
```

Possible behavior:

- if `composeguard.yaml` does not exist, create a new file;
- if `composeguard.yaml` exists, do not overwrite by default;
- require `--force` to overwrite.

Commands:

```bash
composeguard discover --write
composeguard discover --write --force
```

#### 4. Custom Config Path

Support:

```bash
composeguard discover --write --config /etc/composeguard/composeguard.yaml
```

#### 5. Docker Compose Label Detection

Docker Compose containers commonly have labels such as:

```text
com.docker.compose.project
com.docker.compose.service
```

Future-friendly discovery should capture or at least not block this use case.

For v0.4.0, minimum output can remain container names only.

### Acceptance Criteria

- `composeguard discover` prints detected container names.
- `composeguard discover --all` includes stopped containers.
- `composeguard discover --write` creates config if it does not exist.
- Existing config is not overwritten unless `--force` is provided.
- Command works without affecting existing `check`, `init`, `version` commands.

### Non-goals

- Do not implement complex YAML merging yet.
- Do not detect HTTP or SSL endpoints automatically yet.

---

## v0.5.0 — Notification Cooldown and Alert State

### Goal

Prevent Telegram spam when ComposeGuard runs periodically through systemd timer or cron.

This version introduces local state.

### Features

#### 1. Cooldown Config

Add:

```yaml
notification:
  telegram:
    enabled: true
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
    only_on_problem: true
    cooldown_minutes: 60
```

Behavior:

- If a problem appears for the first time, send alert immediately.
- If the same problem still exists within cooldown period, skip notification.
- If the same problem still exists after cooldown period, send reminder.
- If a new problem appears, send notification even within cooldown period.

#### 2. State File

Default state path:

```text
~/.composeguard/state.json
```

For systemd/root usage:

```text
/var/lib/composeguard/state.json
```

Add config:

```yaml
state:
  path: "/var/lib/composeguard/state.json"
```

State should store:

```json
{
  "last_run_at": "2026-07-01T10:00:00Z",
  "last_alert_at": "2026-07-01T10:00:00Z",
  "active_problems": {
    "HTTP: API Health": {
      "status": "CRITICAL",
      "message": "https://example.com/health returned 404, expected 200",
      "first_seen_at": "2026-07-01T10:00:00Z",
      "last_seen_at": "2026-07-01T10:10:00Z",
      "last_notified_at": "2026-07-01T10:00:00Z"
    }
  }
}
```

#### 3. Problem Identity

Problem identity should be stable.

Recommended key:

```text
result.Name + "|" + result.Status
```

Do not use full message as identity because messages may contain changing values.

Example:

```text
Disk: /|CRITICAL
HTTP: API Health|CRITICAL
SSL: api.example.com|WARNING
```

#### 4. Notification Message

If skipped due to cooldown, CLI should print:

```text
telegram notification skipped: active problems are still within cooldown period
```

If sent:

```text
telegram notification sent
```

### Acceptance Criteria

- Repeated scheduled runs do not spam Telegram.
- New problems still trigger notifications immediately.
- Cooldown logic has unit tests.
- Missing state file is handled gracefully.
- Corrupted state file returns a clear warning or resets safely.

### Non-goals

- No remote state storage.
- No database.
- No web UI.

---

## v0.6.0 — Recovery Notification

### Goal

Notify users when a previously failing check becomes healthy again.

### Features

#### 1. Recovery Detection

Using the state file from v0.5.0, detect when previous active problems no longer exist.

Example:

Previous state:

```text
HTTP: API Health|CRITICAL
Disk: /|CRITICAL
```

Current result:

```text
HTTP: API Health OK
Disk: / OK
```

Send recovery message.

#### 2. Config

```yaml
notification:
  telegram:
    enabled: true
    only_on_problem: true
    cooldown_minutes: 60
    send_recovery: true
```

#### 3. Recovery Message Example

```text
COMPOSEGUARD RECOVERY
=====================
Project: my-server
Time: 2026-07-01 12:00:00

✅ HTTP: API Health is now OK
✅ Disk: / is now OK
```

#### 4. Partial Recovery

If one problem recovers and another still fails:

- send recovery for resolved problem;
- send alert/reminder based on cooldown for active problem.

### Acceptance Criteria

- Recovery notification is sent once per resolved problem.
- Recovery notification is not repeated every run.
- Recovery notification can be disabled.
- State is updated after recovery.

### Non-goals

- No advanced incident lifecycle.
- No acknowledgement workflow.

---

## v0.7.0 — Config Validation

### Goal

Help users catch config mistakes before running checks in production.

### Features

#### 1. New Command

```bash
composeguard validate
composeguard validate --config composeguard.yaml
```

#### 2. Validation Rules

Validate top-level config:

- `project_name` should not be empty.

Validate disk config:

- path must not be empty;
- `warning_percent` must be lower than `critical_percent`;
- percent values must be between `1` and `100`.

Validate HTTP config:

- endpoint name must not be empty;
- URL must be valid;
- URL must start with `http://` or `https://`;
- expected status must be between `100` and `599`;
- timeout must be positive.

Validate SSL config:

- domain must not be empty;
- domain should not include path;
- warning days must be greater than critical days;
- critical days must be greater than or equal to `0`.

Validate Telegram config:

- if enabled, bot token must resolve to a non-empty value;
- if enabled, chat ID must resolve to a non-empty value;
- if value is `${ENV_NAME}`, verify environment variable exists.

Validate state config:

- state path must be writable if parent directory exists;
- if parent directory does not exist, return warning with recommendation.

#### 3. Output Example

Valid config:

```text
✓ config is valid
```

Invalid config:

```text
✗ config validation failed

- disk.paths[0].warning_percent must be lower than critical_percent
- http.endpoints[0].url must start with http:// or https://
- notification.telegram.bot_token references TELEGRAM_BOT_TOKEN but environment variable is not set
```

#### 4. JSON Output

Support:

```bash
composeguard validate --json
```

Output:

```json
{
  "valid": false,
  "errors": [
    {
      "field": "http.endpoints[0].url",
      "message": "must start with http:// or https://"
    }
  ],
  "warnings": []
}
```

### Acceptance Criteria

- Invalid config returns non-zero exit code.
- Valid config returns exit code `0`.
- Validation errors are clear and actionable.
- Validation does not run real Docker/HTTP/SSL checks.

### Non-goals

- No schema generation yet.
- No auto-fix yet.

---

## v0.8.0 — Prometheus Metrics Exporter

### Goal

Allow ComposeGuard to integrate with Prometheus and Grafana.

### Features

#### 1. New Command

```bash
composeguard serve --config composeguard.yaml --listen :9118
```

#### 2. Endpoints

```text
/healthz
/metrics
```

`/healthz` returns `200` if the exporter process is alive.

`/metrics` returns Prometheus metrics.

#### 3. Metrics

Recommended metrics:

```text
composeguard_check_status{name="Disk: /"} 2
composeguard_disk_used_percent{path="/"} 58
composeguard_ssl_days_left{domain="api.example.com"} 48
composeguard_http_status_code{name="API Health"} 200
composeguard_last_run_timestamp_seconds 1780000000
```

Status mapping:

| Status | Value |
|---|---:|
| OK | `0` |
| WARNING | `1` |
| CRITICAL | `2` |
| UNKNOWN | `3` |

#### 4. Config

```yaml
server:
  listen: ":9118"
  interval_seconds: 60
```

#### 5. Runtime Model

Two possible implementations:

1. Run checks on every `/metrics` request.
2. Run checks periodically in background and expose last known metrics.

Recommended for v0.8.0:

- use periodic background check;
- store latest result in memory;
- `/metrics` should be fast and stable.

### Acceptance Criteria

- `composeguard serve` starts an HTTP server.
- `/healthz` returns `200`.
- `/metrics` returns Prometheus-compatible metrics.
- Existing `check` command still works.
- Server handles check failures without crashing.

### Non-goals

- No authentication yet.
- No web dashboard yet.
- No persistent time-series storage.

---

## v0.9.0 — Better Docker Compose Awareness

### Goal

Improve support for Docker Compose stacks by understanding Compose projects and services, not only container names.

### Features

#### 1. Compose Project Check

Config:

```yaml
docker:
  compose_projects:
    - name: obat-in
      services:
        - api
        - postgres
        - redis
```

ComposeGuard should use Docker labels:

```text
com.docker.compose.project
com.docker.compose.service
```

to match containers.

#### 2. Service-Level Output

Instead of only:

```text
Docker: obat-in-api-1 running
```

Prefer:

```text
Docker Compose: obat-in/api running
```

#### 3. Restart Count Warning

Detect frequent restarts if possible.

Config:

```yaml
docker:
  restart_policy:
    warning_count: 3
    critical_count: 10
```

Output:

```text
⚠ WARNING Docker Compose: obat-in/api restarted 5 times
```

#### 4. Health Status

Docker containers can have health status when `HEALTHCHECK` is configured.

Report:

```text
✓ OK Docker Compose: obat-in/api healthy
✗ CRITICAL Docker Compose: obat-in/api unhealthy
```

### Acceptance Criteria

- ComposeGuard can monitor by Compose project/service names.
- Existing container-name based config remains supported.
- Docker labels are parsed safely.
- Containers without healthcheck do not fail automatically.

### Non-goals

- No Kubernetes support.
- No Docker Swarm support.

---

## v1.0.0 — Stable CLI Contract

### Goal

Release the first stable version with a reliable CLI contract, documentation, and production-ready behavior for small VPS users.

### Required Features Before v1.0.0

- `init`
- `validate`
- `check`
- `check --config`
- `check --json`
- `check --only`
- `check --notify telegram`
- `discover`
- systemd documentation
- Docker system disk usage check
- notification cooldown
- recovery notification
- GitHub release binaries
- installation documentation
- clear README
- examples directory

### Stable Commands

```bash
composeguard init
composeguard validate
composeguard check
composeguard check --config composeguard.yaml
composeguard check --json
composeguard check --only docker
composeguard check --only disk
composeguard check --only http
composeguard check --only ssl
composeguard check --notify telegram
composeguard discover
composeguard version
```

Optional if implemented before v1.0.0:

```bash
composeguard serve
```

### Documentation Required

- `README.md`
- `ROADMAP.md`
- `docs/systemd.md`
- `docs/configuration.md`
- `docs/telegram.md`
- `docs/install.md`
- `examples/composeguard.example.yaml`
- `examples/systemd/composeguard.service`
- `examples/systemd/composeguard.timer`

### Quality Requirements

- Unit tests for parsers and cooldown logic.
- GitHub Actions CI must pass.
- Release workflow must produce Linux and macOS binaries.
- No secrets committed.
- No panic on missing Docker, invalid config, empty config, or network failure.
- Error messages must be clear.
- Exit codes must remain documented and stable.

### Non-goals

- No enterprise SaaS dashboard.
- No Kubernetes support.
- No distributed monitoring agents.
- No auto-remediation by default.

---

## Future Ideas After v1.0.0

These ideas are intentionally kept outside the initial stable roadmap.

### Web Dashboard

A small local dashboard:

```bash
composeguard dashboard --listen :8088
```

Could show:

- latest check result;
- current active problems;
- SSL days left;
- disk usage;
- Docker container status.

### Multiple Notification Providers

Possible providers:

- Discord webhook;
- Slack webhook;
- generic webhook;
- email SMTP;
- WhatsApp provider through third-party API.

Config concept:

```yaml
notification:
  providers:
    - type: telegram
    - type: webhook
```

### Auto-remediation

Possible future command:

```bash
composeguard doctor
```

Examples:

- suggest `docker system prune` when build cache is too large;
- suggest renewing SSL if expiry is near;
- suggest checking systemd logs if container is restarting.

Important: auto-remediation should not delete or modify system resources without explicit user confirmation.

### Configuration Schema

Generate JSON Schema for `composeguard.yaml`, useful for editor autocomplete.

### Homebrew / APT Install

Distribution ideas:

- Homebrew tap;
- Debian package;
- install script.

---

## Version Planning Summary

| Version | Main Focus | Key Features |
|---|---|---|
| v0.1.x | MVP | Checks, JSON output, Telegram, GitHub release |
| v0.2.0 | Production scheduling | systemd service and timer docs |
| v0.3.0 | Docker disk visibility | Docker system df check |
| v0.4.0 | Easier onboarding | Docker container discovery |
| v0.5.0 | Alert quality | cooldown and local state |
| v0.6.0 | Recovery lifecycle | recovery notification |
| v0.7.0 | Reliability | config validation |
| v0.8.0 | Observability | Prometheus exporter |
| v0.9.0 | Compose awareness | project/service label support |
| v1.0.0 | Stable release | stable CLI, docs, tests, production-ready behavior |

---

## Recommended Immediate Next Tasks

Recommended order from current state:

1. Add this `ROADMAP.md` to the repository.
2. Add `docs/systemd.md`.
3. Add `examples/systemd/composeguard.service`.
4. Add `examples/systemd/composeguard.timer`.
5. Add `docs/configuration.md`.
6. Release `v0.2.0` after systemd documentation is complete.

Suggested commit:

```bash
git add ROADMAP.md
git commit -m "docs: add project roadmap"
```

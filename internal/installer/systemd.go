package installer

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultConfigDir   = "/etc/composeguard"
	defaultBinaryPath  = "/usr/local/bin/composeguard"
	defaultSystemdDir  = "/etc/systemd/system"
	defaultServiceName = "composeguard.service"
	defaultTimerName   = "composeguard.timer"
	defaultEnvName     = "composeguard.env"
	defaultConfigName  = "composeguard.yaml"
)

func InstallSystemd() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("install-systemd must be run as root. Try: sudo composeguard install-systemd")
	}

	if err := ensureDir(defaultConfigDir, 0755); err != nil {
		return err
	}

	if err := ensureDir(defaultSystemdDir, 0755); err != nil {
		return err
	}

	if err := installConfigIfMissing(); err != nil {
		return err
	}

	if err := installEnvIfMissing(); err != nil {
		return err
	}

	if err := writeFile(
		filepath.Join(defaultSystemdDir, defaultServiceName),
		[]byte(serviceTemplate()),
		0644,
	); err != nil {
		return err
	}

	if err := writeFile(
		filepath.Join(defaultSystemdDir, defaultTimerName),
		[]byte(timerTemplate()),
		0644,
	); err != nil {
		return err
	}

	printInstallSummary()

	return nil
}

func installConfigIfMissing() error {
	target := filepath.Join(defaultConfigDir, defaultConfigName)

	if fileExists(target) {
		fmt.Printf("config already exists: %s\n", target)
		return nil
	}

	return writeFile(target, []byte(defaultConfigTemplate()), 0644)
}

func installEnvIfMissing() error {
	target := filepath.Join(defaultConfigDir, defaultEnvName)

	if fileExists(target) {
		fmt.Printf("env already exists: %s\n", target)
		return nil
	}

	return writeFile(target, []byte(defaultEnvTemplate()), 0600)
}

func ensureDir(path string, perm os.FileMode) error {
	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}

	return nil
}

func writeFile(path string, content []byte, perm os.FileMode) error {
	if err := os.WriteFile(path, content, perm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	fmt.Printf("created/updated: %s\n", path)
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func serviceTemplate() string {
	return `[Unit]
Description=ComposeGuard health check
Documentation=https://github.com/IKHINtech/composeguard
Wants=network-online.target docker.service
After=network-online.target docker.service

[Service]
Type=oneshot
EnvironmentFile=-/etc/composeguard/composeguard.env
ExecStart=/usr/local/bin/composeguard check --config /etc/composeguard/composeguard.yaml --notify telegram

User=root
Group=root

Nice=5
IOSchedulingClass=best-effort

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=read-only

[Install]
WantedBy=multi-user.target
`
}

func timerTemplate() string {
	return `[Unit]
Description=Run ComposeGuard health check every 10 minutes

[Timer]
OnBootSec=1min
OnUnitActiveSec=10min
AccuracySec=30s
Unit=composeguard.service

[Install]
WantedBy=timers.target
`
}

func defaultEnvTemplate() string {
	return `TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
`
}

func defaultConfigTemplate() string {
	return `project_name: "my-server"

docker:
  containers: []

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
      url: "https://example.com/health"
      expected_status: 200
      timeout_seconds: 5

ssl:
  domains:
    - "example.com"
  warning_days: 30
  critical_days: 7

notification:
  telegram:
    enabled: true
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
    only_on_problem: true
`
}

func printInstallSummary() {
	fmt.Println()
	fmt.Println("ComposeGuard systemd files installed.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit config:")
	fmt.Println("     sudo nano /etc/composeguard/composeguard.yaml")
	fmt.Println()
	fmt.Println("  2. Edit Telegram env:")
	fmt.Println("     sudo nano /etc/composeguard/composeguard.env")
	fmt.Println()
	fmt.Println("  3. Reload systemd:")
	fmt.Println("     sudo systemctl daemon-reload")
	fmt.Println()
	fmt.Println("  4. Test service:")
	fmt.Println("     sudo systemctl start composeguard.service")
	fmt.Println("     sudo systemctl status composeguard.service")
	fmt.Println()
	fmt.Println("  5. Enable timer:")
	fmt.Println("     sudo systemctl enable --now composeguard.timer")
	fmt.Println()
	fmt.Println("  6. Check timer:")
	fmt.Println("     systemctl list-timers | grep composeguard")
	fmt.Println()
}

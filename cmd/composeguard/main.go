package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"slices"

	"github.com/IKHINtech/composeguard/internal/checker"
	"github.com/IKHINtech/composeguard/internal/config"
	"github.com/IKHINtech/composeguard/internal/diskcheck"
	"github.com/IKHINtech/composeguard/internal/dockercheck"
	"github.com/IKHINtech/composeguard/internal/httpcheck"
	"github.com/IKHINtech/composeguard/internal/installer"
	"github.com/IKHINtech/composeguard/internal/notifier"
	telegramnotifier "github.com/IKHINtech/composeguard/internal/notifier/telegram"
	"github.com/IKHINtech/composeguard/internal/sslcheck"
)

const defaultConfigPath = "composeguard.yaml"

var version = "dev"

func resolvedVersion() string {
	if version != "dev" && version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}

	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	return version
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "check":
		runCheck()
	case "version":
		fmt.Printf("composeguard %s \n", resolvedVersion())
	case "init":
		runInit()
	case "install-systemd":
		runInstallSystemd()
	default:
		printUsage()
		os.Exit(1)
	}
}

func runInstallSystemd() {
	if err := installer.InstallSystemd(); err != nil {
		fmt.Printf("failed to install systemd: %v\n", err)
		os.Exit(1)
	}
}

func runInit() {
	const target = "composeguard.yaml"
	if _, err := os.Stat(target); err == nil {
		fmt.Println("composeguard.yaml already exists")
		os.Exit(1)
	}

	content := `project_name: "my-server"

docker:
  containers: []

disk:
  paths:
    - path: "/"
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
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		fmt.Printf("failed to create composeguard.yaml: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("composeguard.yaml created")
}

func runCheck() {
	configPath := getArgValue("--config")
	if configPath == "" {
		configPath = defaultConfigPath
	}
	if len(os.Args) >= 4 && os.Args[2] == "--config" {
		configPath = os.Args[3]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	only := getArgValue("--only")
	if !isValidOnly(only) {
		fmt.Printf("invalid --only value: %s\n", only)
		fmt.Println("valid values: docker, disk, http, ssl")
		os.Exit(1)
	}

	results := make([]checker.Result, 0)

	if only == "" || only == "docker" {
		results = append(results, dockercheck.CheckContainers(cfg.Docker.Containers)...)
		results = append(results, dockercheck.CheckSystemDF(cfg.Docker.SystemDF)...)
	}

	if only == "" || only == "disk" {
		results = append(results, diskcheck.Check(cfg.Disk.Paths)...)
	}

	if only == "" || only == "http" {
		results = append(results, httpcheck.Check(cfg.HTTP.Endpoints)...)
	}

	if only == "" || only == "ssl" {
		results = append(results, sslcheck.Check(cfg.SSL)...)
	}

	if hasArg("--json") {
		printJSONReport(cfg.ProjectName, results)
	} else {
		printReport(cfg.ProjectName, results)
	}

	if getArgValue("--notify") == "telegram" {
		sendTelegramNotification(cfg, results)
	}
	if hasCritical(results) {
		os.Exit(2)
	}

	if hasWarning(results) {
		os.Exit(1)
	}
}

func sendTelegramNotification(cfg *config.Config, results []checker.Result) {
	telegramCfg := cfg.Notification.Telegram
	if !telegramCfg.Enabled {
		fmt.Println("telegram notifications are disabled")
		return
	}

	hasProblem := hasCritical(results) || hasWarning(results)
	if telegramCfg.OnlyOnProblem && !hasProblem {
		fmt.Println("telegram notification skipped: no problem found")
		return
	}

	message := notifier.BuidMessage(cfg.ProjectName, cleanResults(results))

	if err := telegramnotifier.Send(telegramCfg, message); err != nil {
		fmt.Printf("failed to send telegram notification: %v\n", err)
		return
	}

	fmt.Println("telegram notification sent")
}

func isValidOnly(value string) bool {
	switch value {
	case "", "docker", "disk", "http", "ssl":
		return true
	default:
		return false
	}
}

func printJSONReport(projectName string, results []checker.Result) {
	payload := struct {
		ProjectName string           `json:"project_name"`
		Results     []checker.Result `json:"results"`
	}{
		ProjectName: projectName,
		Results:     cleanResults(results),
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Printf("failed to render json: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(raw))
}

func cleanResults(results []checker.Result) []checker.Result {
	cleaned := make([]checker.Result, 0, len(results))
	for _, result := range results {
		if result.Name == "" && result.Message == "" && result.Status == "" {
			continue
		}

		if result.Status == "" {
			result.Status = checker.StatusUnknown
		}

		cleaned = append(cleaned, result)
	}
	return cleaned
}

func printReport(projectName string, results []checker.Result) {
	if projectName == "" {
		projectName = "composeguard"
	}

	fmt.Println()
	fmt.Printf("COMPOSEGUARD REPORT: %s\n", projectName)
	fmt.Println("================================")

	for _, result := range results {
		if result.Name == "" && result.Message == "" && result.Status == "" {
			continue
		}

		if result.Status == "" {
			result.Status = checker.StatusUnknown
		}
		icon := iconForStatus(result.Status)
		fmt.Printf("%s %-10s %-30s %s\n", icon, result.Status, result.Name, result.Message)
	}

	fmt.Println()
}

func iconForStatus(status checker.Status) string {
	switch status {
	case checker.StatusOK:
		return "✓"
	case checker.StatusWarning:
		return "⚠"
	case checker.StatusCritical:
		return "✗"
	default:
		return "?"
	}
}

func hasCritical(results []checker.Result) bool {
	for _, result := range results {
		if result.Status == checker.StatusCritical {
			return true
		}
	}

	return false
}

func hasWarning(results []checker.Result) bool {
	for _, result := range results {
		if result.Status == checker.StatusWarning {
			return true
		}
	}

	return false
}

func hasArg(name string) bool {
	return slices.Contains(os.Args, name)
}

func printUsage() {
	fmt.Println(`composeguard - lightweight Docker Compose server health monitor

Usage:
	composeguard init
  composeguard check
  composeguard check --config composeguard.yaml
	composeguard check --json
  composeguard check --only docker
  composeguard check --only disk
  composeguard check --only http
  composeguard check --only ssl
	composeguard check --notify telegram
	composeguard install-systemd
  composeguard version`)
}

func getArgValue(name string) string {
	for i, arg := range os.Args {
		if arg == name && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return ""
}

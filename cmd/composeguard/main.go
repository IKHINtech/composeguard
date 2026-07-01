package main

import (
	"fmt"
	"os"

	"github.com/IKHINtech/composeguard/internal/checker"
	"github.com/IKHINtech/composeguard/internal/config"
	"github.com/IKHINtech/composeguard/internal/diskcheck"
	"github.com/IKHINtech/composeguard/internal/dockercheck"
	"github.com/IKHINtech/composeguard/internal/httpcheck"
	"github.com/IKHINtech/composeguard/internal/sslcheck"
)

const defaultConfigPath = "composeguard.yaml"

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
		fmt.Println("composeguard v0.1.0")
	default:
		printUsage()
		os.Exit(1)
	}
}

func runCheck() {
	configPath := defaultConfigPath

	if len(os.Args) >= 4 && os.Args[2] == "--config" {
		configPath = os.Args[3]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	results := make([]checker.Result, 0)

	results = append(results, dockercheck.CheckContainers(cfg.Docker.Containers)...)
	results = append(results, diskcheck.Check(cfg.Disk.Paths)...)
	results = append(results, httpcheck.Check(cfg.HTTP.Endpoints)...)
	results = append(results, sslcheck.Check(cfg.SSL)...)

	printReport(cfg.ProjectName, results)

	if hasCritical(results) {
		os.Exit(2)
	}

	if hasWarning(results) {
		os.Exit(1)
	}
}

func printReport(projectName string, results []checker.Result) {
	if projectName == "" {
		projectName = "composeguard"
	}

	fmt.Println()
	fmt.Printf("COMPOSEGUARD REPORT: %s\n", projectName)
	fmt.Println("================================")

	for _, result := range results {
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

func printUsage() {
	fmt.Println(`composeguard - lightweight Docker Compose server health monitor

Usage:
  composeguard check
  composeguard check --config composeguard.yaml
  composeguard version`)
}

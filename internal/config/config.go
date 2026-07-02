// Package config...
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ProjectName  string             `yaml:"project_name"`
	Docker       DockerConfig       `yaml:"docker"`
	Disk         DiskConfig         `yaml:"disk"`
	HTTP         HTTPConfig         `yaml:"http"`
	SSL          SSLConfig          `yaml:"ssl"`
	Notification NotificationConfig `yaml:"notification"`
}

type DockerConfig struct {
	Containers []string             `yaml:"containers"`
	SystemDF   DockerSystemDFConfig `yaml:"system_df"`
}

type DiskConfig struct {
	Paths []DiskPath `yaml:"paths"`
}

type DiskPath struct {
	Path            string `yaml:"path"`
	WarningPercent  int64  `yaml:"warning_percent"`
	CriticalPercent int64  `yaml:"critical_percent"`
}

type HTTPConfig struct {
	Endpoints []HTTPEndpoint `yaml:"endpoints"`
}

type HTTPEndpoint struct {
	Name           string `yaml:"name"`
	URL            string `yaml:"url"`
	ExpectedStatus int    `yaml:"expected_status"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

type SSLConfig struct {
	Domains      []string `yaml:"domains"`
	WarningDays  int      `yaml:"warning_days"`
	CriticalDays int      `yaml:"critical_days"`
}

func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type NotificationConfig struct {
	Telegram TelegramConfig `yaml:"telegram"`
}

type TelegramConfig struct {
	Enabled       bool   `yaml:"enabled"`
	BotToken      string `yaml:"bot_token"`
	ChatID        string `yaml:"chat_id"`
	OnlyOnProblem bool   `yaml:"only_on_problem"`
}

type DockerSystemDFConfig struct {
	Enabled      bool            `yaml:"enabled"`
	Images       DockerSizeLimit `yaml:"images"`
	Containers   DockerSizeLimit `yaml:"containers"`
	LocalVolumes DockerSizeLimit `yaml:"local_volumes"`
	BuildCache   DockerSizeLimit `yaml:"build_cache"`
}

type DockerSizeLimit struct {
	WarningGB  float64 `yaml:"warning_gb"`
	CriticalGB float64 `yaml:"critical_gb"`
}

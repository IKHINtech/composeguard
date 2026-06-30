package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	ProjectName string       `yaml:"project_name"`
	Docker      DockerConfig `yaml:"docker"`
	Disk        DiskConfig   `yaml:"disk"`
	HTTP        HTTPConfig   `yaml:"http"`
	SSL         SSLConfig    `yaml:"ssl"`
}

type DockerConfig struct {
	Containers []string `yaml:"containers"`
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

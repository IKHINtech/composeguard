package discovery

import (
	"fmt"
	"os"

	"github.com/IKHINtech/composeguard/internal/config"
	"gopkg.in/yaml.v3"
)

func WriterContainersToConfig(configPath string, containerNames []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.Docker.Containers = containerNames
	raw, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, raw, 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

// Package discovery...
package discovery

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

type DockerContainer struct {
	Name   string
	State  string
	Status string
}

type dockerPSRow struct {
	Names  string `json:"Names"`
	State  string `json:"State"`
	Status string `json:"Status"`
}

func DiscoverDockerContainers(runningOnly bool) ([]DockerContainer, error) {
	args := []string{"ps", "-a", "--format", "{{json .}}"}
	if runningOnly {
		args = []string{
			"ps", "--format", "{{json .}}",
		}
	}
	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to  run docker %s: %w", strings.Join(args, " "), err)
	}
	raw := strings.TrimSpace(string(output))
	if raw == "" {
		return []DockerContainer{}, nil
	}

	lines := strings.Split(raw, "\n")
	containers := make([]DockerContainer, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var row dockerPSRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, fmt.Errorf("failed to parse docker ps output: %w", err)
		}

		name := normalizeContainerName(row.Names)
		if name == "" {
			continue
		}
		containers = append(containers, DockerContainer{
			Name:   name,
			State:  row.State,
			Status: row.Status,
		})
	}
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})
	return containers, nil
}

func ContainerNames(containers []DockerContainer) []string {
	names := make([]string, 0, len(containers))
	for _, container := range containers {
		if strings.TrimSpace(container.Name) == "" {
			continue
		}
		names = append(names, container.Name)
	}
	return names
}

func normalizeContainerName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "/")
	return name
}

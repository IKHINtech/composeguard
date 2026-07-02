// Package dockercheck...
package dockercheck

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/IKHINtech/composeguard/internal/checker"
)

type dockerContainer struct {
	Names  string `json:"Names"`
	Status string `json:"Status"`
	State  string `json:"State"`
}

var listContainersFn = listContainers

func CheckContainers(expected []string) []checker.Result {
	results := make([]checker.Result, 0)
	if len(expected) == 0 {
		return results
	}

	containers, err := listContainersFn()
	if err != nil {
		return []checker.Result{
			{
				Name:    "Docker",
				Status:  checker.StatusCritical,
				Message: err.Error(),
			},
		}
	}

	for _, name := range expected {
		name = normalizeName(name)
		found := false

		for _, container := range containers {
			if normalizeName(container.Names) == name {
				found = true
				status := checker.StatusOK
				message := fmt.Sprintf("Container %s is %s", name, container.State)

				if container.State != "running" {
					status = checker.StatusCritical
					message = fmt.Sprintf("Container %s is %s (%s)", name, container.State, container.Status)
				}

				results = append(results, checker.Result{
					Name:    "Docker: " + name,
					Status:  status,
					Message: message,
				})

				break
			}
		}

		if !found {
			results = append(results, checker.Result{
				Name:    "Docker: " + name,
				Status:  checker.StatusCritical,
				Message: "container not found",
			})
		}
	}

	return results
}

func listContainers() ([]dockerContainer, error) {
	cmd := exec.Command(
		"docker",
		"ps",
		"-a",
		"--format",
		"{{json .}}")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run docker ps: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	containers := make([]dockerContainer, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var c dockerContainer
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			return nil, fmt.Errorf("failed to unmarshal docker ps output: %w", err)
		}
		containers = append(containers, c)
	}

	return containers, nil
}

func normalizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "/")
	return name
}

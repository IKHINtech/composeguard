package discovery

import (
	"fmt"
	"strings"
)

func FormatContainersYAML(containers []DockerContainer) string {
	var builder strings.Builder

	builder.WriteString("docker:\n")
	builder.WriteString("  containers:\n")

	if len(containers) == 0 {
		builder.WriteString("    []\n")
		return builder.String()
	}

	for _, container := range containers {
		fmt.Fprintf(&builder, "    - %s\n", container.Name)
	}

	return builder.String()
}

func FormatContainersTable(containers []DockerContainer) string {
	var builder strings.Builder
	if len(containers) == 0 {
		return "No Docker containers found.\n"
	}
	builder.WriteString("Discovered Docker containers:\n")
	builder.WriteString("----------------------------\n")
	for _, container := range containers {
		fmt.Fprintf(&builder, "- %-30s %-12s %s\n", container.Name, container.State, container.Status)
	}
	return builder.String()
}

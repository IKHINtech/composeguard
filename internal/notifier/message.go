// Package notifier...
package notifier

import (
	"fmt"
	"strings"
	"time"

	"github.com/IKHINtech/composeguard/internal/checker"
)

func BuildMessage(projectName string, results []checker.Result) string {
	if projectName == "" {
		projectName = "composeguard"
	}
	var critialCount, warningCount, okCount int
	for _, result := range results {
		switch result.Status {
		case checker.StatusCritical:
			critialCount++
		case checker.StatusWarning:
			warningCount++
		case checker.StatusOK:
			okCount++
		}
	}

	var builder strings.Builder
	builder.WriteString("COMPOSEGUARD ALERT\n")
	builder.WriteString("==================\n")
	builder.WriteString("Project: " + projectName + "\n")
	fmt.Fprintf(&builder, "Time: %s\n", time.Now().Format("15:04:05 02-01-2006"))
	fmt.Fprintf(&builder, "Status: %d critical, %d warning, %d ok\n\n", critialCount, warningCount, okCount)
	for _, result := range results {
		if result.Name == "" && result.Message == "" && result.Status == "" {
			continue
		}

		if result.Status == checker.StatusOK {
			continue
		}

		icon := iconForStatus(result.Status)
		fmt.Fprintf(&builder, "%s %s\n", icon, result.Name)
		fmt.Fprintf(&builder, "   Status: %s\n", result.Status)
		fmt.Fprintf(&builder, "   Message: %s\n\n", result.Message)
	}
	return strings.TrimSpace(builder.String())
}

func iconForStatus(status checker.Status) string {
	switch status {
	case checker.StatusCritical:
		return "❌"
	case checker.StatusWarning:
		return "⚠️"
	case checker.StatusOK:
		return "✅"
	default:
		return "❔"
	}
}

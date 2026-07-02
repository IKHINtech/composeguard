package dockercheck

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/IKHINtech/composeguard/internal/checker"
	"github.com/IKHINtech/composeguard/internal/config"
)

const bytesInGB = 1024 * 1024 * 1024

type dockerSystemDFRow struct {
	Type        string `json:"Type"`
	TotalCount  string `json:"TotalCount"`
	Active      string `json:"Active"`
	Size        string `json:"Size"`
	Reclaimable string `json:"Reclaimable"`
}

func CheckSystemDF(cfg config.DockerSystemDFConfig) []checker.Result {
	if !cfg.Enabled {
		return nil
	}

	fmt.Println("masuk sini")
	rows, err := listDockerSystemDF()
	if err != nil {
		return []checker.Result{
			{
				Name:    "Docker: System DF",
				Status:  checker.StatusCritical,
				Message: err.Error(),
			},
		}
	}

	results := make([]checker.Result, 0, len(rows))

	for _, row := range rows {
		if strings.TrimSpace(row.Size) == "" || strings.TrimSpace(row.Size) == "<nil>" {
			results = append(results, checker.Result{
				Name:    "Docker: " + normalizeDockerSystemDFType(row.Type),
				Status:  checker.StatusUnknown,
				Message: "size is empty",
			})
			continue
		}
		sizeBytes, err := parseDockerSizeToBytes(row.Size)
		if err != nil {
			results = append(results, checker.Result{
				Name:    "Docker: " + row.Type,
				Status:  checker.StatusUnknown,
				Message: fmt.Sprintf("failed to parse size %q: %v", row.Size, err),
			})
			continue
		}

		limit, ok := limitForDockerSystemDFType(cfg, row.Type)
		if !ok {
			continue
		}

		sizeGB := float64(sizeBytes) / bytesInGB
		status := statusFromGB(sizeGB, limit)

		results = append(results, checker.Result{
			Name:    "Docker: " + normalizeDockerSystemDFType(row.Type),
			Status:  status,
			Message: fmt.Sprintf("%.2f GB used", sizeGB),
		})
	}

	return results
}

func listDockerSystemDF() ([]dockerSystemDFRow, error) {
	cmd := exec.Command("docker", "system", "df", "--format", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run docker system df: %w", err)
	}

	raw := strings.TrimSpace(string(output))
	if raw == "" {
		return nil, nil
	}

	// Docker commonly returns one JSON object per line for --format json.
	lines := strings.Split(raw, "\n")
	rows := make([]dockerSystemDFRow, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var row dockerSystemDFRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, fmt.Errorf("failed to parse docker system df output: %w", err)
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func limitForDockerSystemDFType(
	cfg config.DockerSystemDFConfig,
	rowType string,
) (config.DockerSizeLimit, bool) {
	switch strings.ToLower(strings.TrimSpace(rowType)) {
	case "images":
		return withDefaultLimit(cfg.Images, 10, 30), true
	case "containers":
		return withDefaultLimit(cfg.Containers, 5, 10), true
	case "local volumes":
		return withDefaultLimit(cfg.LocalVolumes, 10, 30), true
	case "build cache":
		return withDefaultLimit(cfg.BuildCache, 10, 30), true
	default:
		return config.DockerSizeLimit{}, false
	}
}

func withDefaultLimit(
	limit config.DockerSizeLimit,
	defaultWarningGB float64,
	defaultCriticalGB float64,
) config.DockerSizeLimit {
	if limit.WarningGB <= 0 {
		limit.WarningGB = defaultWarningGB
	}

	if limit.CriticalGB <= 0 {
		limit.CriticalGB = defaultCriticalGB
	}

	return limit
}

func statusFromGB(sizeGB float64, limit config.DockerSizeLimit) checker.Status {
	if sizeGB >= limit.CriticalGB {
		return checker.StatusCritical
	}

	if sizeGB >= limit.WarningGB {
		return checker.StatusWarning
	}

	return checker.StatusOK
}

func normalizeDockerSystemDFType(rowType string) string {
	switch strings.ToLower(strings.TrimSpace(rowType)) {
	case "images":
		return "Images"
	case "containers":
		return "Containers"
	case "local volumes":
		return "Local Volumes"
	case "build cache":
		return "Build Cache"
	default:
		return strings.TrimSpace(rowType)
	}
}

func parseDockerSizeToBytes(value string) (uint64, error) {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "")

	if value == "" {
		return 0, fmt.Errorf("empty size")
	}

	units := []struct {
		suffix     string
		multiplier float64
	}{
		{"KiB", 1024},
		{"MiB", 1024 * 1024},
		{"GiB", 1024 * 1024 * 1024},
		{"TiB", 1024 * 1024 * 1024 * 1024},
		{"kB", 1000},
		{"MB", 1000 * 1000},
		{"GB", 1000 * 1000 * 1000},
		{"TB", 1000 * 1000 * 1000 * 1000},
		{"B", 1},
	}

	for _, unit := range units {
		if strings.HasSuffix(value, unit.suffix) {
			numberPart := strings.TrimSuffix(value, unit.suffix)

			number, err := strconv.ParseFloat(numberPart, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number %q: %w", numberPart, err)
			}

			return uint64(number * unit.multiplier), nil
		}
	}

	// Docker sometimes can return plain number-like values in some formats.
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("unsupported size format %q", value)
	}

	return uint64(number), nil
}

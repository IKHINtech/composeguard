package diskcheck

import (
	"fmt"
	"syscall"

	"github.com/IKHINtech/composeguard/internal/checker"
	"github.com/IKHINtech/composeguard/internal/config"
)

func Check(paths []config.DiskPath) []checker.Result {
	results := make([]checker.Result, len(paths))
	for _, item := range paths {
		var stat syscall.Statfs_t

		if err := syscall.Statfs(item.Path, &stat); err != nil {
			results = append(results, checker.Result{
				Name:    "Disk",
				Status:  checker.StatusCritical,
				Message: err.Error(),
			})
			continue
		}

		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bavail * uint64(stat.Bsize)
		used := total - free
		usedPercent := int64(float64(used) / float64(total) * 100)
		status := checker.StatusOK
		if usedPercent >= item.CriticalPercent {
			status = checker.StatusCritical
		} else if usedPercent >= item.WarningPercent {
			status = checker.StatusWarning
		}

		results = append(results, checker.Result{
			Name:    "Disk: " + item.Path,
			Status:  status,
			Message: fmt.Sprintf("%d%% used", usedPercent),
		})
	}

	return results
}

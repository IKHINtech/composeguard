package httpcheck

import (
	"fmt"
	"net/http"
	"time"

	"github.com/IKHINtech/composeguard/internal/checker"
	"github.com/IKHINtech/composeguard/internal/config"
)

func Check(endpoints []config.HTTPEndpoint) []checker.Result {
	results := make([]checker.Result, 0, len(endpoints))

	for _, endpoint := range endpoints {
		timeout := endpoint.TimeoutSeconds
		if timeout <= 0 {
			timeout = 5
		}

		client := http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}

		resp, err := client.Get(endpoint.URL)
		if err != nil {
			results = append(results, checker.Result{
				Name:    "HTTP: " + endpoint.Name,
				Status:  checker.StatusCritical,
				Message: err.Error(),
			})
			continue
		}

		_ = resp.Body.Close()

		expected := endpoint.ExpectedStatus
		if expected == 0 {
			expected = http.StatusOK
		}

		status := checker.StatusOK
		if resp.StatusCode != expected {
			status = checker.StatusCritical
		}

		results = append(results, checker.Result{
			Name:    "HTTP: " + endpoint.Name,
			Status:  status,
			Message: fmt.Sprintf("%s returned %d, expected %d", endpoint.URL, resp.StatusCode, expected),
		})
	}

	return results
}

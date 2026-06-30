package sslcheck

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/IKHINtech/composeguard/internal/checker"
	"github.com/IKHINtech/composeguard/internal/config"
)

func Check(cfg config.SSLConfig) []checker.Result {
	results := make([]checker.Result, 0, len(cfg.Domains))

	warningDays := cfg.WarningDays
	if warningDays <= 0 {
		warningDays = 30
	}

	criticalDays := cfg.CriticalDays
	if criticalDays <= 0 {
		criticalDays = 7
	}

	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	for _, domain := range cfg.Domains {
		domain = normalizeDomain(domain)

		if domain == "" {
			results = append(results, checker.Result{
				Name:    "SSL",
				Status:  checker.StatusWarning,
				Message: "empty domain skipped",
			})
			continue
		}

		conn, err := tls.DialWithDialer(dialer, "tcp", domain+":443", &tls.Config{
			ServerName: domain,
			MinVersion: tls.VersionTLS12,
		})
		if err != nil {
			results = append(results, checker.Result{
				Name:    "SSL: " + domain,
				Status:  checker.StatusCritical,
				Message: err.Error(),
			})
			continue
		}

		certs := conn.ConnectionState().PeerCertificates
		_ = conn.Close()

		if len(certs) == 0 {
			results = append(results, checker.Result{
				Name:    "SSL: " + domain,
				Status:  checker.StatusCritical,
				Message: "no certificate found",
			})
			continue
		}

		expiry := certs[0].NotAfter
		daysLeft := int(time.Until(expiry).Hours() / 24)

		status := checker.StatusOK
		if daysLeft <= criticalDays {
			status = checker.StatusCritical
		} else if daysLeft <= warningDays {
			status = checker.StatusWarning
		}

		results = append(results, checker.Result{
			Name:    "SSL: " + domain,
			Status:  status,
			Message: fmt.Sprintf("expires in %d days on %s", daysLeft, expiry.Format("2006-01-02")),
		})
	}

	return results
}

func normalizeDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")

	return domain
}

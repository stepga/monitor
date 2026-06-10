package collector

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/reporter"
)

type CertInfo struct {
	Url    string
	Expiry *time.Time
	Error  error
	Took   time.Duration
}

type CertCollector struct{}

func certExpiry(rawURL string) (*time.Time, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "443"
	}

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		net.JoinHostPort(host, port),
		&tls.Config{
			ServerName: host,
			// InsecureSkipVerify: true,
		},
	)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	state := conn.ConnectionState()

	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("no certificate presented")
	}

	cert := state.PeerCertificates[0]
	return &cert.NotAfter, nil
}

func CheckCerts(urls []string) []CertInfo {
	var wg sync.WaitGroup
	outChan := make(chan CertInfo)

	for _, u := range urls {
		wg.Go(func() {
			result := CertInfo{Url: u}
			start := time.Now()
			expiry, err := certExpiry(u)
			result.Took = time.Since(start)
			if err != nil {
				result.Error = err
			} else {
				result.Expiry = expiry
			}
			outChan <- result
		})
	}
	go func() {
		wg.Wait()
		close(outChan)
	}()

	var results []CertInfo
	for result := range outChan {
		results = append(results, result)
	}
	return results
}

func (c *CertCollector) Init(cfg *config.Config, reporter reporter.Reporter) {
	go func() {
		threshold := time.Duration(cfg.Cert.MinimumDaysLeft*24) * time.Hour
		for {
			info := CheckCerts(cfg.Cert.Urls)
			for _, info := range info {
				if info.Error != nil {
					reporter.Report(fmt.Sprintf("%s (%dms): ERROR: %s",
						info.Url,
						info.Took.Milliseconds(),
						info.Error,
					))
					continue
				}
				remaining := time.Until(*info.Expiry)
				if remaining < threshold {
					reporter.Report(
						fmt.Sprintf(
							"%s (%dms): EXPIRES SOON %v remaining, expires %s",
							info.Url,
							info.Took.Milliseconds(),
							remaining,
							info.Expiry.Format(time.UnixDate),
						))
				} else {
					reporter.Report(
						fmt.Sprintf(
							"%s (%dms): OK %v remaining, expires %s",
							info.Url,
							info.Took.Milliseconds(),
							remaining,
							info.Expiry.Format(time.UnixDate),
						))
				}
			}

			time.Sleep(1 * time.Minute)
		}
	}()
}

func (c *CertCollector) Info() interface{} {
	return nil
}

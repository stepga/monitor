package subsystems

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
)

type CertCheck struct{}

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

func checkCerts(urls []string) []bus.CertInfo {
	var wg sync.WaitGroup
	outChan := make(chan bus.CertInfo)

	for _, u := range urls {
		wg.Go(func() {
			result := bus.CertInfo{
				Url:  u,
				Time: time.Now(),
			}
			start := time.Now()
			expiry, err := certExpiry(u)
			result.Took = time.Since(start)
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Expiry = *expiry
			}
			outChan <- result
		})
	}
	go func() {
		wg.Wait()
		close(outChan)
	}()

	var results []bus.CertInfo
	for result := range outChan {
		results = append(results, result)
	}
	return results
}

func (c *CertCheck) Init() error {
	go func() {
		for {
			infos := checkCerts(config.Cfg.Cert.Urls)
			threshold := time.Duration(config.Cfg.Cert.MinimumDaysLeft*24) * time.Hour
			for _, info := range infos {
				bus.Publish(info)
				remaining := time.Until(info.Expiry)
				if info.Error != "" {
					bus.Publish(bus.CertError{
						Url:   info.Url,
						Error: info.Error,
						Time:  info.Time,
					})
				} else if remaining < threshold {
					bus.Publish(bus.CertExpiresSoon{
						Url:       info.Url,
						Remaining: remaining,
						Expiry:    info.Expiry,
						Time:      info.Time,
					})
				} else {
					bus.Publish(bus.CertOk{
						Url:       info.Url,
						Remaining: remaining,
						Expiry:    info.Expiry,
						Time:      info.Time,
					})
				}
			}
			time.Sleep(config.Cfg.Cert.CheckIntervalInHours * time.Hour)
		}
	}()

	return nil
}

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

func CheckCerts(urls []string) []bus.CertInfo {
	var wg sync.WaitGroup
	outChan := make(chan bus.CertInfo)

	for _, u := range urls {
		wg.Go(func() {
			result := bus.CertInfo{Url: u}
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

	var results []bus.CertInfo
	for result := range outChan {
		results = append(results, result)
	}
	return results
}

func (c *CertCollector) Init() error {
	go func() {
		for {
			info := CheckCerts(config.Cfg.Cert.Urls)
			for _, info := range info {
				bus.Publish(info)
			}

			time.Sleep(1 * time.Minute)
		}
	}()

	return nil
}

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/stepga/monitor/collector"
	"github.com/stepga/monitor/collector/cert"
	"github.com/stepga/monitor/collector/listener"
	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/reporter"
	"github.com/stepga/monitor/reporter/stdout"
)

var AvailableCollectors = map[string]collector.Collector{
	"certs":    &cert.CertsCollector{},
	"listener": &listener.ListenerCollector{},
}

func LoadConfig(path string) (*config.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg config.Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func printCertInfo(info []cert.CertInfo, minimumDaysLeft int) {
	threshold := time.Duration(minimumDaysLeft*24) * time.Hour
	for _, info := range info {
		fmt.Printf("%s (%dms): ", info.Url, info.Took.Milliseconds())
		if info.Error != nil {
			fmt.Printf("ERROR: %s\n", info.Error)
			continue
		}
		remaining := time.Until(*info.Expiry)

		if remaining < threshold {
			fmt.Printf(
				"EXPIRES SOON %v remaining, expires %s\n",
				remaining.Round(time.Hour),
				info.Expiry.Format(time.UnixDate),
			)
		} else {
			fmt.Printf(
				"OK %v remaining, expires %s\n",
				remaining.Round(time.Hour),
				info.Expiry.Format(time.UnixDate),
			)
		}
	}
}

/*
   Main Reporte implementation
   Starts reporters from config
*/

type rep struct {
	activeReporters []reporter.Reporter
}

var AvailableReporters = map[string]reporter.Reporter{
	"stdout": &stdout.StdoutReporter{},
}

func (r *rep) Init(cfg *config.Config) {
	for name, reporter := range AvailableReporters {
		if slices.Contains(cfg.Reporter, name) {
			r.activeReporters = append(r.activeReporters, reporter)
			reporter.Init(cfg)
		}
	}
}

func (r *rep) Report(msg string) {
	for _, reporter := range r.activeReporters {
		reporter.Report(msg)
	}
}

func main() {
	config_file := flag.String("config", "config.json", "Path to config.json file")
	flag.Parse()

	cfg, err := LoadConfig(*config_file)
	if err != nil {
		panic(err)
	}

	reporter := rep{}
	reporter.Init(cfg)

	for name, collector := range AvailableCollectors {
		if slices.Contains(cfg.Collectors, name) {
			collector.Init(cfg, &reporter)
		}
	}

	// TODO: Use Ctrl-C/Signals or something
	fmt.Printf("Press enter to quit\n")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}

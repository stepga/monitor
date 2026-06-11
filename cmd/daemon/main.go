package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/collector"
	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/node"
	"github.com/stepga/monitor/reporter"
	"github.com/stepga/monitor/reporter/stdout"
)

var AvailableCollectors = map[string]collector.Collector{
	"cert":     &collector.CertCollector{},
	"listener": &collector.ListenerCollector{},
}

var AvailableReporters = map[string]reporter.Reporter{
	"stdout": &stdout.StdoutReporter{},
}

func main() {
	config_file := flag.String("config", "config.json", "Path to config.json file")
	flag.Parse()

	err := config.LoadConfig(*config_file)
	if err != nil {
		panic(err)
	}

	for name, reporter := range AvailableReporters {
		if slices.Contains(config.Cfg.Reporter, name) {
			reporter.Init()
		}
	}

	for name, collector := range AvailableCollectors {
		if slices.Contains(config.Cfg.Collectors, name) {
			collector.Init()
		}
	}

	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(reloadSignal, syscall.SIGUSR1)
	go func() {
		for {
			<-reloadSignal
			fmt.Println("Got USR1, reloading config")
			err := config.LoadConfig(*config_file)
			if err != nil {
				fmt.Printf("Failed to relaod config: %s", err)
			}
		}
	}()

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

}

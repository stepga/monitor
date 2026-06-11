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

type DiskGettingFull struct {
	Hostname string
	Disk     node.FileSystem
}

type DiskFineAgain struct {
	Hostname string
	Disk     node.FileSystem
}

func (d DiskGettingFull) Report() string {
	return fmt.Sprintf("Disk %s on %s is getting full: %s!", d.Disk.Source, d.Hostname, d.Disk.Capacity)
}

func (d DiskFineAgain) Report() string {
	return fmt.Sprintf("Disk %s on %s is is fine again: %s!", d.Disk.Source, d.Hostname, d.Disk.Capacity)
}

func DiskWarner() {
	ch := bus.Subscribe()
	defer bus.Unsubscribe(ch)
	disksReported := make(map[string]struct{})
	for m := range ch {
		switch msg := m.(type) {
		case node.NodeInfo:
			for _, fs := range msg.FileSystems {
				if fs.Source == "none" {
					continue
				}
				key := msg.HostName + ":" + fs.Source
				_, exists := disksReported[key]
				full := float64(fs.UsedBytes) > (float64(fs.AvailableBytes+fs.UsedBytes) * float64(config.Cfg.DiskThreshold))
				if full && !exists {
					bus.Publish(DiskGettingFull{
						Hostname: msg.HostName,
						Disk:     fs,
					})
					disksReported[key] = struct{}{}
				} else if !full && exists {
					bus.Publish(DiskFineAgain{
						Hostname: msg.HostName,
						Disk:     fs,
					})
					delete(disksReported, key)
				}
			}
		}
	}
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

	go DiskWarner()

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

}

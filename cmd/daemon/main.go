package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

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

type StoreInfo struct {
	Info     string
	LastSeen time.Time
	Name     string
}

/*
Store owner has to:
- TODO: go over store every X hour and check for missing nodes
- TODO: publish notifications
*/
func Store() {
	ch := bus.Subscribe()
	defer bus.Unsubscribe(ch)
	store := make(map[string]StoreInfo)
	for m := range ch {
		switch msg := m.(type) {
		case node.NodeInfo:
			if _, exists := store[msg.HostName]; exists {
				bus.Publish(fmt.Sprintf("message from known node: %s", msg.HostName))
			} else {
				bus.Publish(fmt.Sprintf("message from unknown node: %s", msg.HostName))
			}
			store[msg.HostName] = StoreInfo{
				Name:     msg.HostName,
				Info:     string(msg.String()),
				LastSeen: time.Now(),
			}
			fmt.Printf("Store:\n")
			for _, v := range store {
				fmt.Printf("%s (last seen before %s): %s", v.Name, time.Until(v.LastSeen).Round(time.Second), v.Info)
			}
			fmt.Printf("\n")
		default:
			// Not interested
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

	go Store()

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

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

}

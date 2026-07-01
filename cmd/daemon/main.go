package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/store"
	"github.com/stepga/monitor/subsystems"
)

var AvailableSubsystems = map[string]subsystems.Subsystem{
	"cert":      &subsystems.CertCheck{},
	"listener":  &subsystems.Listener{},
	"stdout":    &subsystems.Stdout{},
	"pushover":  &subsystems.Pushover{},
	"webui":     &subsystems.WebUi{},
	"heartbeat": &subsystems.Heartbeat{},
	"diskmon":   &subsystems.Diskmon{},
	"rebootmon": &subsystems.Rebootmon{},
	"redmine":   &subsystems.Redmine{},
	"dump":      &subsystems.Dump{},
}

func main() {
	configFile := flag.String("config", "config.json", "Path to config.json file")
	flag.Parse()

	err := config.LoadConfig(*configFile)
	if err != nil {
		panic(err)
	}

	store.Start()

	for name, subsystem := range AvailableSubsystems {
		if slices.Contains(config.Cfg.Subsystems, name) {
			err := subsystem.Init()
			if err == nil {
				fmt.Printf("Initialized %s subsystem\n", name)
			} else {
				panic(fmt.Errorf("subsystem %s init failed: %s", name, err))
			}
		}
	}

	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(reloadSignal, syscall.SIGUSR1)
	go func() {
		for {
			<-reloadSignal
			err := config.LoadConfig(*configFile)
			if err != nil {
				fmt.Printf("Failed to reload config: %s\n", err)
			}
			bus.Publish(bus.ConfigReloaded{})
		}
	}()

	dumpCriticalSignal := make(chan os.Signal, 1)
	signal.Notify(dumpCriticalSignal, syscall.SIGUSR2)
	go func() {
		for {
			<-dumpCriticalSignal
			fmt.Println("Dump Critical Messages (due to SIGUSR2):")
			for _, msg := range store.FetchCritical() {
				fmt.Printf("  %s\n", msg.Summary())
			}
		}
	}()

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}

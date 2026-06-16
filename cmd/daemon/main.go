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
	"github.com/stepga/monitor/webui"
)

var AvailableSubsystems = map[string]subsystems.Subsystem{
	"cert":      &subsystems.CertCheck{},
	"listener":  &subsystems.Listener{},
	"stdout":    &subsystems.Stdout{},
	"pushover":  &subsystems.Pushover{},
	"webui":     &webui.Server{},
	"heartbeat": &subsystems.Heartbeat{},
	"diskmon":   &subsystems.Diskmon{},
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
				fmt.Printf("Failed to relaod config: %s", err)
			}
			bus.Publish(bus.ConfigReloaded{})
		}
	}()

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}

package subsystems

import (
	"fmt"
	"time"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/store"
)

type Rebootmon struct{}

func (_ *Rebootmon) Init() error {
	if !config.AllSubsystemsConfigured([]string{"listener"}) {
		return fmt.Errorf("rebootmon requires listener subsystem to be configured")
	}

	go func() {
		ch := bus.Subscribe()
		defer bus.Unsubscribe(ch)
		for m := range ch {
			switch msg := m.(type) {
			case bus.NodeInfo:
				critical := bus.RebootRequired{
					Hostname: msg.Hostname,
					Time:     time.Now(),
				}
				exists := store.Exists(critical)
				if msg.RebootRequired && !exists {
					bus.Publish(critical)
				} else if !msg.RebootRequired && exists {
					bus.Publish(bus.Rebooted{
						Hostname: msg.Hostname,
						Time:     time.Now(),
					})
				}
			}
		}
	}()

	return nil
}

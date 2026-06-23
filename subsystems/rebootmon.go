package subsystems

import (
	"fmt"
	"time"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
)

type Rebootmon struct{}

func (_ *Rebootmon) Init() error {
	if !config.AllSubsystemsConfigured([]string{"listener"}) {
		return fmt.Errorf("rebootmon requires listener subsystem to be configured")
	}

	go func() {
		ch := bus.Subscribe()
		defer bus.Unsubscribe(ch)
		rebootsReported := make(map[string]struct{})
		for m := range ch {
			switch msg := m.(type) {
			case bus.NodeInfo:
				_, exists := rebootsReported[msg.Hostname]
				if msg.RebootRequired && !exists {
					bus.Publish(bus.RebootRequired{
						Hostname: msg.Hostname,
						Time:     time.Now(),
					})
					rebootsReported[msg.Hostname] = struct{}{}
				} else if !msg.RebootRequired && exists {
					bus.Publish(bus.Rebooted{
						Hostname: msg.Hostname,
						Time:     time.Now(),
					})
					delete(rebootsReported, msg.Hostname)
				}
			}
		}
	}()

	return nil
}

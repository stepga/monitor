package subsystems

import (
	"time"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/node"
)

type Heartbeat struct {
	reports map[string]time.Time
}

func (h *Heartbeat) nodePing(hostname string) {
	_, exists := h.reports[hostname]
	if !exists {
		bus.Publish(bus.NewNode{
			Hostname: hostname,
		})
	}
	h.reports[hostname] = time.Now()
}

func (h *Heartbeat) tick() {
	now := time.Now()
	timeout := config.Cfg.Heartbeat.NodeTimeoutInMinutes * time.Minute
	for hostname, lastSeen := range h.reports {
		if now.Sub(lastSeen) > timeout {
			bus.Publish(bus.NodeTimeout{
				Hostname: hostname,
				LastSeen: lastSeen,
			})
			delete(h.reports, hostname)
		}
	}
}

func (h *Heartbeat) Init() error {
	h.reports = make(map[string]time.Time)
	go func() {
		ch := bus.Subscribe()
		defer bus.Unsubscribe(ch)
		for {
			interval := config.Cfg.Heartbeat.CheckIntervalInMinutes * time.Minute
			ticker := time.NewTicker(interval)
			select {
			case <-ticker.C:
				h.tick()
			case m := <-ch:
				switch msg := m.(type) {
				case node.NodeInfo:
					h.nodePing(msg.HostName)
				}
			}
			ticker.Stop()
		}
	}()

	return nil
}

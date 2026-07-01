package subsystems

import (
	"fmt"
	"time"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
)

type Diskmon struct{}

func (_ *Diskmon) Init() error {
	if !config.AllSubsystemsConfigured([]string{"listener"}) {
		return fmt.Errorf("diskmon requires listener subsystem to be configured")
	}

	threshold := config.Cfg.Diskmon.Threshold

	go func() {
		ch := bus.Subscribe()
		defer bus.Unsubscribe(ch)
		disksReported := make(map[string]struct{})
		for m := range ch {
			switch msg := m.(type) {
			case bus.NodeInfo:
				for _, fs := range msg.FileSystems {
					if fs.Source == "none" {
						continue
					}
					key := msg.Hostname + ":" + fs.Source
					_, exists := disksReported[key]
					full := float64(fs.UsedBytes) > (float64(fs.AvailableBytes+fs.UsedBytes) * threshold)
					if full && !exists {
						bus.Publish(bus.DiskGettingFull{
							Hostname: msg.Hostname,
							Disk:     fs,
							Time:     time.Now(),
						})
						disksReported[key] = struct{}{}
					} else if !full && exists {
						bus.Publish(bus.DiskFineAgain{
							Hostname: msg.Hostname,
							Disk:     fs,
							Time:     time.Now(),
						})
						delete(disksReported, key)
					}
				}
			}
		}
	}()

	return nil
}

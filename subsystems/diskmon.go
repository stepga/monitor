package subsystems

import (
	"fmt"
	"time"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/store"
)

type Diskmon struct{}

func (_ *Diskmon) Init() error {
	if !config.AllSubsystemsConfigured([]string{"listener"}) {
		return fmt.Errorf("diskmon requires listener subsystem to be configured")
	}

	go func() {
		ch := bus.Subscribe()
		defer bus.Unsubscribe(ch)
		for m := range ch {
			switch msg := m.(type) {
			case bus.NodeInfo:
				for _, fs := range msg.FileSystems {
					if fs.Source == "none" {
						continue
					}
					critical := bus.DiskGettingFull{
						Hostname: msg.Hostname,
						Disk:     fs,
						Time:     time.Now(),
					}
					exists := store.Exists(critical)
					full := float64(fs.UsedBytes) > (float64(fs.AvailableBytes+fs.UsedBytes) * float64(config.Cfg.DiskThreshold))
					if full && !exists {
						bus.Publish(critical)
					} else if !full && exists {
						bus.Publish(bus.DiskFineAgain{
							Hostname: msg.Hostname,
							Disk:     fs,
							Time:     time.Now(),
						})
					}
				}
			}
		}
	}()

	return nil
}

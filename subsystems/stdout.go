package subsystems

import (
	"fmt"
	"github.com/stepga/monitor/bus"
)

type StdoutReporter struct{}

func (r *StdoutReporter) Init() error {
	ch := bus.Subscribe()
	go func() {
		defer bus.Unsubscribe(ch)
		for msg := range ch {
			switch m := msg.(type) {
			case bus.Important:
				fmt.Printf("IMPORTANT: %s\n", m.Summary())
			case bus.Info:
				fmt.Printf("%s\n", m.Summary())
			case string:
				fmt.Printf("Bus msg '%s'\n", m)
			case bus.CertInfo:
				// ignore
			default:
				fmt.Printf("stdout: Unknown message type: %T\n", msg)
			}

		}
	}()

	return nil
}

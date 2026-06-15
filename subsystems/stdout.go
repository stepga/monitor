package subsystems

import (
	"fmt"
	"github.com/stepga/monitor/bus"
	"os"
)

type StdoutReporter struct{}

func (r *StdoutReporter) Init() error {
	ch := bus.Subscribe()
	go func() {
		defer bus.Unsubscribe(ch)
		_, noColor := os.LookupEnv("NO_COLOR")
		for msg := range ch {
			switch m := msg.(type) {
			case bus.Important:
				if noColor {
					fmt.Printf("IMPORTANT: %s\n", m.Summary())
				} else {
					fmt.Printf("\033[31m%s\033[0m\n", m.Summary())
				}
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

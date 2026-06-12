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
			case Oneline:
				fmt.Printf("stdout: %s\n", m.Oneline())
			case string:
				fmt.Printf("stdout: Bus msg %s\n", m)
			case Report:
				fmt.Printf("stdout: Report: %s\n", m.Report())
			case bus.CertInfo:
				// ignore
			default:
				fmt.Printf("stdout: Unknown message type: %T\n", msg)
			}

		}
	}()

	return nil
}

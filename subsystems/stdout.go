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
			case string:
				fmt.Printf("stdout: Bus msg %s\n", m)
			case NodeTimeout:
				fmt.Printf("stdout: Node timeout: %s, last seen %v\n", m.Hostname, m.LastSeen)
			case NewNode:
				fmt.Printf("stdout: New Node: %s\n", m.Hostname)
			case Report:
				fmt.Printf("stdout: Report: %s\n", m.Report())
			default:
				fmt.Printf("stdout: Unknown message type: %T\n", msg)
			}

		}
	}()

	return nil
}

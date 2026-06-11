package stdout

import (
	"fmt"
	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/node"
	"github.com/stepga/monitor/reporter"
)

type StdoutReporter struct{}

func (r *StdoutReporter) Init() {
	fmt.Println("Initialized stdout reporter!")
	ch := bus.Subscribe()
	go func() {
		defer bus.Unsubscribe(ch)
		for msg := range ch {
			switch m := msg.(type) {
			case node.NodeInfo:
				fmt.Printf("stdout: NodeInfo: HostName %s\n", m.HostName)
			case string:
				fmt.Printf("stdout: Bus msg %s\n", m)
			case reporter.Report:
				fmt.Printf("stdout: Report: %s\n", m.Report())
			default:
				fmt.Printf("stdout: Unknown message type: %T\n", msg)
			}

		}
	}()
}

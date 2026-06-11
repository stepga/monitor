package subsystems

import (
	"fmt"

	"github.com/stepga/monitor/bus"
)

type WebUiReporter struct {
	RelevantMessages chan any
}

func (r *WebUiReporter) Init() {
	fmt.Println("Initialized webui reporter!")
	ch := bus.Subscribe()
	go func() {
		defer bus.Unsubscribe(ch)
		for msg := range ch {
			switch m := msg.(type) {
			case Report:
				fmt.Printf("webui: Report: %s\n", m.Report())
				r.RelevantMessages <- m
			default:
			}
		}
	}()
}

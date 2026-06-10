package stdout

import (
	"fmt"
	"github.com/stepga/monitor/config"
)

type StdoutReporter struct{}

func (r *StdoutReporter) Init(cfg *config.Config) {
	fmt.Println("Initialized stdout reporter!")
}

func (r *StdoutReporter) Report(msg string) {
	fmt.Println(msg)
}

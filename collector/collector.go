package collector

import (
	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/reporter"
)

type Collector interface {
	Init(*config.Config, reporter.Reporter)
}

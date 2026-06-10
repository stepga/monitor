package reporter

import (
	"github.com/stepga/monitor/config"
)

type Reporter interface {
	Init(*config.Config)
	Report(string)
}

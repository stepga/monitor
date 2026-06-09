package main

import (
	"log/slog"
	"os"

	"github.com/stepga/monitor/nodeinfo"
)

func main() {
	info, err := nodeinfo.CreateInfo()
	if err != nil {
		slog.Error("CreateInfo() failed", "info", info)
		os.Exit(1)
	}
	slog.Info("CreateInfo() succeeded", "info", info)
}

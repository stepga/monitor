package main

import (
	"log/slog"
	"os"

	"github.com/stepga/monitor/uname"
)

type FileSystem struct {
	MountPoint string `json:"mount_point"`
	UsedBytes  uint64 `json:"used"`
	TotalBytes uint64 `json:"total"`
}

type Info struct {
	// as reported by `uname -n`
	Hostname string `json:"hostname"`
	// as reported by `uname -s`
	OperatingSystemName string `json:"operating_system_name"`
	// as reported by `uname -r`
	OperatingSystemVersion string `json:"operating_system_version"`
	// e.g. due to linux system upgrade and indication via
	// /var/run/reboot-required
	RebootRequired bool `json:"reboot_required"`
	// an array of the mounted filesystems and their respective
	// used and total sizes in bytes
	FileSystems []FileSystem `json:"filesystems"`
}

func createInfo() (*Info, error) {
	var err error
	info := &Info{}

	info.Hostname, err = uname.Hostname()
	if err != nil {
		return nil, err
	}
	info.OperatingSystemName = uname.OperatingSystemName()
	info.OperatingSystemVersion, err = uname.OperatingSystemVersion()
	if err != nil {
		return nil, err
	}

	return info, nil
}

func main() {
	info, err := createInfo()
	if err != nil {
		slog.Error("createInfo() failed", "info", info)
		os.Exit(1)
	}
	slog.Info("createInfo() succeeded", "info", info)
}

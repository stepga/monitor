package node

import (
	"fmt"
	"os"
	"runtime"
)

// HostName returns (network node) hostname as reported by `uname -n`.
func HostName() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("HostName(): %w", err)
	}
	return hostname, nil
}

// OperatingSystemName returns the kernel name as reported by `uname -s`.
func OperatingSystemName() string {
	return runtime.GOOS
}

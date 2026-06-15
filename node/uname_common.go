package node

import (
	"fmt"
	"os"
	"runtime"
)

// Hostname returns (network node) hostname as reported by `uname -n`.
func Hostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("Hostname(): %w", err)
	}
	return hostname, nil
}

// OperatingSystemName returns the kernel name as reported by `uname -s`.
func OperatingSystemName() string {
	return runtime.GOOS
}

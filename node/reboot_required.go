package node

import (
	"errors"
	"os"
	"runtime"
)

func rebootRequiredLinux() (bool, error) {
	_, err := os.Stat("/var/run/reboot-required")
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func RebootRequired() (bool, error) {
	switch runtime.GOOS {
	case "openbsd":
		// No standard reboot-required file.
		// Ideas for different mechanisms:
		// - track whether syspatch installed a new kernel
		// - compare running kernel vs installed kernel
		return false, nil
	default:
		// Debian/Ubuntu
		return rebootRequiredLinux()
	}
}

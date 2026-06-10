//go:build openbsd

package uname

import (
	"fmt"
	"syscall"
)

// OperatingSystemVersion returns the kernel release version as reported by
// `uname -r`.
//
// As OpenBSD does not implement syscall.Uname, the kernel state is gathered
// via sysctl.
func OperatingSystemVersion() (string, error) {
	version, err := syscall.Sysctl("kern.osrelease")
	if err != nil {
		return "", fmt.Errorf("OperatingSystemVersion(): %w", err)
	}
	return version, nil
}

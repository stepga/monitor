//go:build linux

package uname

import (
	"fmt"
	"os"
)

// OperatingSystemVersion returns the kernel release version as reported by
// `uname -r`.
//
// No convenient function exists in "os", and instead of parsing
// syscall.Utsname.Release, make use of the linux procfs.
func OperatingSystemVersion() (string, error) {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return "", fmt.Errorf("OperatingSystemVersion(): %w", err)
	}
	return string(data), nil
}

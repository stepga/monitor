//go:build linux

package node

import (
	"fmt"
	"os"
	"strings"
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
	return strings.Trim(string(data), " \n"), nil
}

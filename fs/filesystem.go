// The fs package provides information about the mounted filesystems, basically
// by parsing the output of `df -P`. `df` is part of the OpenBSD base system and
// the GNU coreutils, and should be present everywhere.
package fs

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type FileSystem struct {
	Source         string `json:"source"`
	UsedBytes      int64  `json:"used_bytes"`
	AvailableBytes int64  `json:"available_bytes"`
	Capacity       string `json:"capacity"`
	MountPoint     string `json:"mount_point"`
}

func FileSystems() ([]FileSystem, error) {
	scanner, err := runDF()
	if err != nil {
		return nil, err
	}

	return parseDF(scanner)
}

func runDF() (*bufio.Scanner, error) {
	cmd := exec.Command("df", "-P")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// wrapping the unbuffered string with a buffered scanner gives us a convenient `Scan`
	// method that advances the scanner to the next token (default: next line)
	return bufio.NewScanner(strings.NewReader(string(out))), nil
}

func parseDF(scanner *bufio.Scanner) ([]FileSystem, error) {
	var entries []FileSystem

	first := true
	for scanner.Scan() {
		// skip header line "Filesystem ..."
		if first {
			first = false
			continue
		}

		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			return nil, fmt.Errorf("unexpected format: %q", line)
		}

		// skip tmpfs and efivarfs (linux-only) filesystems
		re := regexp.MustCompile(`tmpfs|efivarfs`)
		if re.MatchString(fields[0]) {
			continue
		}

		used, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return nil, err
		}

		avail, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			return nil, err
		}

		entry := FileSystem{
			Source:         fields[0],
			UsedBytes:      used,
			AvailableBytes: avail,
			Capacity:       fields[4],
			MountPoint:     strings.Join(fields[5:], " "),
		}

		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}

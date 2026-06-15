package bus

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/stepga/monitor/node"
)

const BusMsgSize = 16

var (
	globalBus = &Bus{
		subscribers: make(map[chan any]struct{}),
	}
)

// List of Bus Messages

// Reported for every checked certificate
type CertInfo struct {
	Url    string
	Expiry time.Time
	Error  error
	Took   time.Duration
}

// Reported when looking up the certificate failed
type CertError struct {
	Url   string
	Error error
}

// Reported when a cert expires soon
type CertExpiresSoon struct {
	Url       string
	Remaining time.Duration
	Expiry    time.Time
}

// Reported when a node stopped reporting
type NodeTimeout struct {
	Hostname string
	LastSeen time.Time
}

// Reported when a new node started
type NewNode struct {
	Hostname string
}

// Reported when a disk is full
type DiskGettingFull struct {
	Hostname string
	Disk     node.FileSystem
}

// Report when a disk is fine again
type DiskFineAgain struct {
	Hostname string
	Disk     node.FileSystem
}

// Reported when config.Cfg was reloaded
type ConfigReloaded struct{}

// Reported when a node sends a message to the daemon
type NodeInfo struct {
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
	FileSystems []node.FileSystem `json:"file_systems"`
}

// Bus messages interaces and Reporing
// Any message that implements the Info interface will get reportet,
// e.g. displayed on the gui, or written to a log file. It can also
// implement the Important interace, which should be used for messages
// that require attention, e.g. a disk is getting full or a
// certificate is nearing end of life.

type Summary interface {
	// Single line string describing the message
	Summary() string
}

// TODO: Remove, replaced by Info and Important
type Report interface {
	Summary
	Report() string
}

type Info interface {
	Summary
	_info()
}

type Important interface {
	Info
	_important()
}

// Bus Message interface implementations

func (info CertError) Report() string {
	return fmt.Sprintf("%s: ERROR: %s",
		info.Url,
		info.Error,
	)
}

func (info CertExpiresSoon) Report() string {
	return fmt.Sprintf(
		"%s: EXPIRES SOON %v remaining, expires %s\n",
		info.Url,
		info.Remaining,
		info.Expiry.Format(time.UnixDate),
	)
}

func (d DiskGettingFull) Report() string {
	return fmt.Sprintf("Disk %s on %s is getting full: %s!", d.Disk.Source, d.Hostname, d.Disk.Capacity)
}

func (d DiskFineAgain) Report() string {
	return fmt.Sprintf("Disk %s on %s is is fine again: %s!", d.Disk.Source, d.Hostname, d.Disk.Capacity)
}

func (n NodeInfo) Report() string {
	data, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

// Summary impl for bus messages

func (d DiskGettingFull) Summary() string { return d.Report() }
func (d DiskFineAgain) Summary() string   { return d.Report() }
func (c CertError) Summary() string       { return c.Report() }
func (c CertExpiresSoon) Summary() string { return c.Report() }
func (c ConfigReloaded) Summary() string  { return "Configuration reloaded" }
func (n NewNode) Summary() string         { return fmt.Sprintf("New Node: %s", n.Hostname) }
func (n NodeTimeout) Summary() string     { return fmt.Sprintf("NodeTimeout: %s", n.Hostname) }
func (n NodeInfo) Summary() string        { return fmt.Sprintf("Node message from %s", n.Hostname) }

func (DiskGettingFull) _info() {}
func (DiskFineAgain) _info()   {}
func (CertError) _info()       {}
func (CertExpiresSoon) _info() {}
func (ConfigReloaded) _info()  {}
func (NewNode) _info()         {}
func (NodeTimeout) _info()     {}
func (NodeInfo) _info()        {}

func (DiskGettingFull) _important() {}
func (CertError) _important()       {}
func (CertExpiresSoon) _important() {}
func (NodeTimeout) _important()     {}

// Bus Implementaiton

type Bus struct {
	subscribers map[chan any]struct{}
	lock        sync.RWMutex
}

func (b *Bus) subscribe() chan any {
	ch := make(chan any, BusMsgSize)
	b.lock.Lock()
	b.subscribers[ch] = struct{}{}
	b.lock.Unlock()
	return ch
}

func (b *Bus) publish(msg any) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- msg:
		default:
			fmt.Printf("Could not deliver msg\n")
		}
	}
}

func (b *Bus) unsubscribe(ch chan any) {
	b.lock.Lock()
	delete(b.subscribers, ch)
	b.lock.Unlock()
	close(ch)
}

func Subscribe() chan any {
	return globalBus.subscribe()
}

func Publish(msg any) {
	globalBus.publish(msg)
}

func Unsubscribe(ch chan any) {
	globalBus.unsubscribe(ch)
}

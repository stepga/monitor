package bus

import (
	"fmt"
	"sync"
	"time"

	"github.com/stepga/monitor/node"
)

const BusMsgSize = 16

var globalBus = &Bus{
	subscribers: make(map[chan any]struct{}),
}

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

// Reported when a cert is ok
type CertOk struct {
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

// Reported when a Node requires reboot
type RebootRequired struct {
	Hostname string
}

// Reported when a Node was rebooted
type Rebooted struct {
	Hostname string
}

// Bus messages interfaces and Reporting
//
// Any message that implements the Info interface will get reported,
// e.g. displayed on the gui, or written to a log file.
// A message can also implement the Important interface, which should
// be used for messages that require attention, e.g. a disk is getting
// full or a certificate is nearing end of life.

type Summary interface {
	// Single line string describing the message
	Summary() string
}

type Info interface {
	Summary
	ID() any
	_info()
}

// Sticky messages are alerts that stay active until a non-sticky
// message with the same ID is published which clears it again.
type Sticky interface {
	Info
	_sticky()
}

type Important interface {
	Info
	_important()
}

// Bus Message interface implementations

func (info CertError) Summary() string {
	return fmt.Sprintf("%s: %s",
		info.Url,
		info.Error,
	)
}

func (info CertExpiresSoon) Summary() string {
	return fmt.Sprintf(
		"%s: expires soon, %v remaining, expires at %s",
		info.Url,
		info.Remaining,
		info.Expiry.Format(time.UnixDate),
	)
}

func (info CertOk) Summary() string {
	return fmt.Sprintf(
		"%s: ok, %v remaining, expires at %s",
		info.Url,
		info.Remaining,
		info.Expiry.Format(time.UnixDate),
	)
}

func (d DiskGettingFull) Summary() string {
	return fmt.Sprintf("Disk %s on %s is getting full: %s!", d.Disk.Source, d.Hostname, d.Disk.Capacity)
}

func (d DiskFineAgain) Summary() string {
	return fmt.Sprintf("Disk %s on %s is is fine again: %s!", d.Disk.Source, d.Hostname, d.Disk.Capacity)
}

func (ConfigReloaded) Summary() string { return "Configuration reloaded" }
func (n NewNode) Summary() string      { return fmt.Sprintf("New Node: %s", n.Hostname) }
func (n NodeTimeout) Summary() string  { return fmt.Sprintf("NodeTimeout: %s", n.Hostname) }
func (n NodeInfo) Summary() string     { return fmt.Sprintf("Node message from %s", n.Hostname) }
func (n RebootRequired) Summary() string {
	return fmt.Sprintf("Node %s needs to be rebooted", n.Hostname)
}
func (n Rebooted) Summary() string {
	return fmt.Sprintf("Node %s was rebooted", n.Hostname)
}

func (d DiskGettingFull) ID() any { return d.Hostname + ":" + d.Disk.Source }
func (d DiskFineAgain) ID() any   { return d.Hostname + ":" + d.Disk.Source }
func (c CertError) ID() any       { return c.Url }
func (c CertExpiresSoon) ID() any { return c.Url }
func (c CertOk) ID() any          { return c.Url }
func (c ConfigReloaded) ID() any  { return "ConfigReloaded" }
func (n NewNode) ID() any         { return n.Hostname }
func (n NodeTimeout) ID() any     { return n.Hostname }
func (n NodeInfo) ID() any        { return n.Hostname }
func (n RebootRequired) ID() any  { return "Reboot:" + n.Hostname }
func (n Rebooted) ID() any        { return "Reboot:" + n.Hostname }

func (DiskGettingFull) _info() {}
func (DiskFineAgain) _info()   {}
func (CertError) _info()       {}
func (CertExpiresSoon) _info() {}
func (CertOk) _info()          {}
func (ConfigReloaded) _info()  {}
func (NewNode) _info()         {}
func (NodeTimeout) _info()     {}
func (NodeInfo) _info()        {}
func (RebootRequired) _info()  {}
func (Rebooted) _info()        {}

func (DiskGettingFull) _important() {}
func (CertError) _important()       {}
func (CertExpiresSoon) _important() {}
func (NodeTimeout) _important()     {}
func (RebootRequired) _important()  {}

func (DiskGettingFull) _sticky() {}
func (CertError) _sticky()       {}
func (CertExpiresSoon) _sticky() {}
func (NodeTimeout) _sticky()     {}
func (RebootRequired) _sticky()  {}

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

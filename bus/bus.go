package bus

import (
	"fmt"
	"sync"
	"time"

	"github.com/stepga/monitor/node"
)

const subscriberChannelBufferSize = 16

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

// Reported when the list of critical messages in store/store.go changes
type CriticalListChanged struct{}

// Bus messages interfaces and Reporting
//
// Any message that implements the Info interface will get reported,
// e.g. displayed on the gui, or written to a log file.

type Info interface {
	Summary() string
	Identifier() any
}

// Critical messages are alerts that stay active until a non-critical
// message with the same Identifier is published and clears it again.
type Critical interface {
	Info
	_critical()
}

// Bus Message interface implementations

func (info CertError) Summary() string {
	return fmt.Sprintf("Could not get cert from '%s': %s",
		info.Url,
		info.Error,
	)
}

func (info CertExpiresSoon) Summary() string {
	return fmt.Sprintf(
		"Certificate of '%s' expires soon, %d days remaining, expires at %s",
		info.Url,
		int(info.Remaining.Hours()/24.0),
		info.Expiry.Format(time.DateTime),
	)
}

func (info CertOk) Summary() string {
	return fmt.Sprintf(
		"Certificate of '%s' is ok, %d days remaining, expires at %s",
		info.Url,
		int(info.Remaining.Hours()/24.0),
		info.Expiry.Format(time.DateTime),
	)
}

func (d DiskGettingFull) Summary() string {
	return fmt.Sprintf("Disk %s on %s is getting full: %s", d.Disk.Source, d.Hostname, d.Disk.Capacity)
}

func (d DiskFineAgain) Summary() string {
	return fmt.Sprintf("Disk %s on %s is is fine again: %s", d.Disk.Source, d.Hostname, d.Disk.Capacity)
}

func (n NewNode) Summary() string     { return fmt.Sprintf("New Node connected: %s", n.Hostname) }
func (n NodeTimeout) Summary() string { return fmt.Sprintf("Node %s timed out", n.Hostname) }
func (n NodeInfo) Summary() string    { return fmt.Sprintf("Received message from node %s", n.Hostname) }
func (n RebootRequired) Summary() string {
	return fmt.Sprintf("Node %s needs to be rebooted", n.Hostname)
}
func (n Rebooted) Summary() string {
	return fmt.Sprintf("Node %s was rebooted", n.Hostname)
}

func (d DiskGettingFull) Identifier() any { return "DiskUsage:" + d.Hostname + ":" + d.Disk.Source }
func (d DiskFineAgain) Identifier() any   { return "DiskUsage:" + d.Hostname + ":" + d.Disk.Source }
func (c CertError) Identifier() any       { return "Cert:" + c.Url }
func (c CertExpiresSoon) Identifier() any { return "Cert:" + c.Url }
func (c CertOk) Identifier() any          { return "Cert:" + c.Url }
func (n NewNode) Identifier() any         { return "Node:" + n.Hostname }
func (n NodeTimeout) Identifier() any     { return "Node:" + n.Hostname }
func (n NodeInfo) Identifier() any        { return "Node:" + n.Hostname }
func (n RebootRequired) Identifier() any  { return "Reboot:" + n.Hostname }
func (n Rebooted) Identifier() any        { return "Reboot:" + n.Hostname }

func (DiskGettingFull) _critical() {}
func (CertError) _critical()       {}
func (CertExpiresSoon) _critical() {}
func (NodeTimeout) _critical()     {}
func (RebootRequired) _critical()  {}

// Bus Implementaiton

type Bus struct {
	subscribers map[chan any]struct{}
	lock        sync.RWMutex
}

func (b *Bus) subscribe() chan any {
	ch := make(chan any, subscriberChannelBufferSize)
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
		case <-time.After(5 * time.Second):
			fmt.Printf("Error: Could not deliver msg\n")
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

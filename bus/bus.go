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
	Time   time.Time
}

// Reported when looking up the certificate failed
type CertError struct {
	Url   string
	Error error
	Time  time.Time
}

// Reported when a cert expires soon
type CertExpiresSoon struct {
	Url       string
	Remaining time.Duration
	Expiry    time.Time
	Time      time.Time
}

// Reported when a cert is ok
type CertOk struct {
	Url       string
	Remaining time.Duration
	Expiry    time.Time
	Time      time.Time
}

// Reported when a node stopped reporting
type NodeTimeout struct {
	Hostname string
	LastSeen time.Time
	Time     time.Time
}

// Reported when a new node started
type NewNode struct {
	Hostname string
	Time     time.Time
}

// Reported when a disk is full
type DiskGettingFull struct {
	Hostname string
	Disk     node.FileSystem
	Time     time.Time
}

// Report when a disk is fine again
type DiskFineAgain struct {
	Hostname string
	Disk     node.FileSystem
	Time     time.Time
}

// Reported when config.Cfg was reloaded
type ConfigReloaded struct{}

// Reported when a node sends a message to the daemon
type NodeInfo struct {
	Time time.Time
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
	Time     time.Time
}

// Reported when a Node was rebooted
type Rebooted struct {
	Hostname string
	Time     time.Time
}

// Reported when the list of critical messages in store/store.go changes
type CriticalListChanged struct{}

// Reported as a fake success notification e.g. due to manually delete critical
// notifications via web interface.
type StoreInfoDelete struct {
	Id   string
	Time time.Time
}

// Bus messages interfaces and Reporting
//
// Any message that implements the Info interface will get reported,
// e.g. displayed on the gui, or written to a log file.

type Info interface {
	Summary() string
	Identifier() string
	Details() string
	Timestamp() string
}

// Critical messages are alerts that stay active until a non-critical
// message with the same Identifier is published and clears it again.
type Critical interface {
	Info
	Critical()
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

func (i StoreInfoDelete) Summary() string {
	return fmt.Sprintf("Enforce deletion of critical infos with Identifier: %s", i.Id)
}

func (d DiskGettingFull) Details() string {
	return fmt.Sprintf(
		`Hostname: %s
Source: %s
Mount point: %s
Available bytes: %d
Used bytes: %d
Capacity: %s`,
		d.Hostname,
		d.Disk.Source,
		d.Disk.MountPoint,
		d.Disk.AvailableBytes,
		d.Disk.UsedBytes,
		d.Disk.Capacity,
	)
}

func (d DiskFineAgain) Details() string {
	return fmt.Sprintf(
		`Hostname: %s
Source: %s
Mount point: %s
Available bytes: %d
Used bytes: %d
Capacity: %s`,
		d.Hostname,
		d.Disk.Source,
		d.Disk.MountPoint,
		d.Disk.AvailableBytes,
		d.Disk.UsedBytes,
		d.Disk.Capacity,
	)
}

func (c CertError) Details() string {
	return fmt.Sprintf(
		`Url: %s
Error: %s`,
		c.Url,
		c.Error,
	)
}

func (c CertExpiresSoon) Details() string {
	return fmt.Sprintf(
		`Url: %s
Valid until: %s
Remaining: %d days`,
		c.Url,
		c.Expiry.Format(time.DateTime),
		int(c.Remaining.Hours()/24.0),
	)
}

func (c CertOk) Details() string {
	return fmt.Sprintf(
		`Url: %s
Valid until: %s
Remaining: %d days`,
		c.Url,
		c.Expiry.Format(time.DateTime),
		int(c.Remaining.Hours()/24.0),
	)
}

func (n NewNode) Details() string {
	return fmt.Sprintf(
		`Hostname: %s`,
		n.Hostname,
	)
}

func (n NodeTimeout) Details() string {
	return fmt.Sprintf(
		`Hostname: %s
Last Seen: %s`,
		n.Hostname,
		n.LastSeen.Format(time.DateTime),
	)
}

func (n NodeInfo) Details() string {
	str := fmt.Sprintf(
		`Hostname: %s
Operating System: %s
Operating System Version: %s
Reboot required: %t`,
		n.Hostname,
		n.OperatingSystemName,
		n.OperatingSystemVersion,
		n.RebootRequired,
	)
	for _, fs := range n.FileSystems {
		str = str + fmt.Sprintf(
			`
  %s
    Capacity: %s
    Source: %s
    Available Bytes: %d
    Used Bytes: %d`,
			fs.MountPoint,
			fs.Capacity,
			fs.Source,
			fs.AvailableBytes,
			fs.UsedBytes,
		)
	}
	return str
}

func (n RebootRequired) Details() string {
	return fmt.Sprintf(
		`Hostname: %s`,
		n.Hostname,
	)
}

func (n Rebooted) Details() string {
	return fmt.Sprintf(
		`Hostname: %s`,
		n.Hostname,
	)
}

func (i StoreInfoDelete) Details() string { return "" }

func getTimestamp(t *time.Time) string {
	ret := time.Now()
	if t != nil && !t.IsZero() {
		ret = *t
	}
	return ret.Format(time.DateTime)
}
func (c CertError) Timestamp() string       { return getTimestamp(&c.Time) }
func (c CertExpiresSoon) Timestamp() string { return getTimestamp(&c.Time) }
func (c CertInfo) Timestamp() string        { return getTimestamp(&c.Time) }
func (c CertOk) Timestamp() string          { return getTimestamp(&c.Time) }
func (d DiskFineAgain) Timestamp() string   { return getTimestamp(&d.Time) }
func (d DiskGettingFull) Timestamp() string { return getTimestamp(&d.Time) }
func (n NewNode) Timestamp() string         { return getTimestamp(&n.Time) }
func (n NodeInfo) Timestamp() string        { return getTimestamp(&n.Time) }
func (n NodeTimeout) Timestamp() string     { return getTimestamp(&n.Time) }
func (r RebootRequired) Timestamp() string  { return getTimestamp(&r.Time) }
func (r Rebooted) Timestamp() string        { return getTimestamp(&r.Time) }
func (i StoreInfoDelete) Timestamp() string { return getTimestamp(&i.Time) }

func (c CertError) Identifier() string       { return "Cert:" + c.Url }
func (c CertExpiresSoon) Identifier() string { return "Cert:" + c.Url }
func (c CertOk) Identifier() string          { return "Cert:" + c.Url }
func (d DiskFineAgain) Identifier() string   { return "DiskUsage:" + d.Hostname + ":" + d.Disk.Source }
func (d DiskGettingFull) Identifier() string { return "DiskUsage:" + d.Hostname + ":" + d.Disk.Source }
func (n NewNode) Identifier() string         { return "Node:" + n.Hostname }
func (n NodeInfo) Identifier() string        { return "Node:" + n.Hostname }
func (n NodeTimeout) Identifier() string     { return "Node:" + n.Hostname }
func (n RebootRequired) Identifier() string  { return "Reboot:" + n.Hostname }
func (n Rebooted) Identifier() string        { return "Reboot:" + n.Hostname }
func (i StoreInfoDelete) Identifier() string { return i.Id }

func (DiskGettingFull) Critical() {}
func (CertError) Critical()       {}
func (CertExpiresSoon) Critical() {}
func (NodeTimeout) Critical()     {}
func (RebootRequired) Critical()  {}

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

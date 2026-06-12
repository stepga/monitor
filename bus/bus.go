package bus

import (
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

// Oneline impl for bus messages

func (d DiskGettingFull) Oneline() string { return d.Report() }
func (d DiskFineAgain) Oneline() string   { return d.Report() }
func (c CertError) Oneline() string       { return c.Report() }
func (c CertExpiresSoon) Oneline() string { return c.Report() }
func (c ConfigReloaded) Oneline() string  { return "Configuration reloaded" }
func (n NewNode) Oneline() string         { return fmt.Sprintf("New Node: %s", n.Hostname) }
func (n NodeTimeout) Oneline() string     { return fmt.Sprintf("NodeTimeout: %s", n.Hostname) }

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

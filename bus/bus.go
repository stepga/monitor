package bus

import (
	"fmt"
	"sync"
	"time"

	"github.com/stepga/monitor/config"
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
	Expiry *time.Time
	Error  error
	Took   time.Duration
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

// Bus Message interface implementations

func (info CertInfo) Report() string {
	threshold := time.Duration(config.Cfg.Cert.MinimumDaysLeft*24) * time.Hour

	if info.Error != nil {
		return fmt.Sprintf("%s (%dms): ERROR: %s",
			info.Url,
			info.Took.Milliseconds(),
			info.Error,
		)
	}
	remaining := time.Until(*info.Expiry)
	if remaining < threshold {
		return fmt.Sprintf(
			"%s (%dms): EXPIRES SOON %v remaining, expires %s",
			info.Url,
			info.Took.Milliseconds(),
			remaining,
			info.Expiry.Format(time.UnixDate),
		)
	} else {
		return fmt.Sprintf(
			"%s (%dms): OK %v remaining, expires %s",
			info.Url,
			info.Took.Milliseconds(),
			remaining,
			info.Expiry.Format(time.UnixDate),
		)
	}
}

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

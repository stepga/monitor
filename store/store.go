package store

import (
	"sync"

	"github.com/stepga/monitor/bus"
)

type Store struct {
	sticky map[any]bus.Info
	lock   sync.RWMutex
}

var store = &Store{
	sticky: make(map[any]bus.Info),
}

func Start() {
	go func() {
		ch := bus.Subscribe()
		defer bus.Unsubscribe(ch)
		for m := range ch {
			changed := false
			switch msg := m.(type) {
			case bus.Sticky:
				store.lock.Lock()
				_, exists := store.sticky[msg.Identifier()]
				if !exists {
					store.sticky[msg.Identifier()] = msg
					changed = true
				}
				store.lock.Unlock()
			case bus.Info:
				store.lock.Lock()
				_, exists := store.sticky[msg.Identifier()]
				if exists {
					delete(store.sticky, msg.Identifier())
					changed = true
				}
				store.lock.Unlock()
			}
			if changed {
				bus.Publish(bus.StickyListChanged{})
			}
		}
	}()
}

func FetchSticky() []bus.Info {
	ret := make([]bus.Info, 0, len(store.sticky))

	store.lock.RLock()
	for _, v := range store.sticky {
		ret = append(ret, v)
	}
	store.lock.RUnlock()

	return ret
}

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
			switch msg := m.(type) {
			case bus.Sticky:
				store.lock.Lock()
				store.sticky[msg.ID()] = msg
				store.lock.Unlock()
			case bus.Info:
				store.lock.Lock()
				delete(store.sticky, msg.ID())
				store.lock.Unlock()
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

package store

import (
	"sync"

	"github.com/stepga/monitor/bus"
)

type Store struct {
	critical map[string]bus.Critical
	lock     sync.RWMutex
}

var store = &Store{
	critical: make(map[string]bus.Critical),
}

func (s *Store) delete(identifier string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, exists := store.critical[identifier]
	if exists {
		delete(s.critical, identifier)
		return true
	}
	return false
}

func Start() {
	go func() {
		ch := bus.Subscribe()
		defer bus.Unsubscribe(ch)
		for m := range ch {
			changed := false
			switch msg := m.(type) {
			case bus.Critical:
				store.lock.Lock()
				_, exists := store.critical[msg.Identifier()]
				if !exists {
					store.critical[msg.Identifier()] = msg
					changed = true
				} else {
				}
				store.lock.Unlock()
			case bus.StoreInfoDelete:
				changed = store.delete(msg.Identifier())
			case bus.Info:
				changed = store.delete(msg.Identifier())
			}
			if changed {
				bus.Publish(bus.CriticalListChanged{})
			}
		}
	}()
}

func FetchCritical() []bus.Critical {
	ret := make([]bus.Critical, 0, len(store.critical))

	store.lock.RLock()
	for _, v := range store.critical {
		ret = append(ret, v)
	}
	store.lock.RUnlock()

	return ret
}

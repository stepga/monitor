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

func (store *Store) Add(msg bus.Critical) bool {
	store.lock.Lock()
	defer store.lock.Unlock()

	_, exists := store.critical[msg.Identifier()]
	if !exists {
		store.critical[msg.Identifier()] = msg
		return true
	}
	return false
}

func (store *Store) Exists(msg bus.Critical) bool {
	store.lock.Lock()
	defer store.lock.Unlock()

	_, exists := store.critical[msg.Identifier()]
	return exists
}

func (store *Store) Delete(msg bus.Info) bool {
	store.lock.Lock()
	defer store.lock.Unlock()

	_, exists := store.critical[msg.Identifier()]
	if exists {
		delete(store.critical, msg.Identifier())
		return true
	}
	return false
}

func Add(msg bus.Critical) bool {
	return store.Add(msg)
}

func Exists(msg bus.Critical) bool {
	return store.Exists(msg)
}
func Start() {
	go func() {
		ch := bus.Subscribe()
		defer bus.Unsubscribe(ch)
		for m := range ch {
			changed := false
			switch msg := m.(type) {
			case bus.Critical:
				changed = store.Add(msg)
			case bus.Info:
				changed = store.Delete(msg)
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

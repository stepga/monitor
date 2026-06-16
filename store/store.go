package store

import (
	"github.com/stepga/monitor/bus"
)

type Store struct {
	sticky map[any]bus.Info
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
				store.sticky[msg.ID()] = msg
			case bus.Info:
				delete(store.sticky, msg.ID())
			}
		}
	}()
}

func FetchSticky() []bus.Info {
	ret := make([]bus.Info, 0, len(store.sticky))

	// TODO: Is range on a map threadsafe?
	for _, v := range store.sticky {
		ret = append(ret, v)
	}

	return ret
}

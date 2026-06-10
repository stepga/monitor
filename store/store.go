package store

import (
	"fmt"
	"time"

	"github.com/stepga/monitor/nodeinfo"
)

type StoreInfo struct {
	Info     string
	LastSeen time.Time
	Name     string
}

/*
Store owner has to:
- TODO: go over store every X hour and check for missing nodes
- TODO: publish notifications
*/
func Start(storeMsgChannel <-chan nodeinfo.NodeInfo) {
	store := make(map[string]StoreInfo)
	for msg := range storeMsgChannel {
		if _, exists := store[msg.HostName]; exists {
			fmt.Printf("message from known node: %s\n", msg.HostName)
		} else {
			fmt.Printf("message from unknown node: %s\n", msg.HostName)
		}
		store[msg.HostName] = StoreInfo{
			Name:     msg.HostName,
			Info:     string(msg.String()),
			LastSeen: time.Now(),
		}
		fmt.Printf("Store:\n")
		for _, v := range store {
			fmt.Printf("%s (last seen before %s): %s\n", v.Name, time.Until(v.LastSeen).Round(time.Second), v.Info)
		}
		fmt.Printf("\n")
	}
}

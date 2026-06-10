package store

import (
	"fmt"
	"time"

	"github.com/stepga/monitor/listener"
)

type NodeInfo struct {
	Name     string
	LastSeen time.Time
}

/*
Store owner has to:
- go over store every X hour and check for missing nodes
- accept new node messages and put them in the store
- publish notifications
*/
func Start(storeMsgChannel <-chan listener.NodeMsg) {
	store := make(map[string]NodeInfo)
	for msg := range storeMsgChannel {
		info, exists := store[msg.Name]
		if exists {
			fmt.Printf("Client ping: %s\n", msg.Name)
			info.LastSeen = time.Now()
			store[msg.Name] = info
		} else {
			fmt.Printf("New client: %s\n", msg.Name)
			store[msg.Name] = NodeInfo{
				Name:     msg.Name,
				LastSeen: time.Now(),
			}
		}
		fmt.Printf("Store:\n")
		for _, v := range store {
			fmt.Printf("  %s: %s\n", v.Name, time.Until(v.LastSeen).Round(time.Second))
		}
	}
}

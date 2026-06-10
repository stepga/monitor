package listener

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type NodeMsg struct {
	Name string `json:"name"`
}

type NodeInfo struct {
	Name     string
	LastSeen time.Time
}

func parseNodeMsg(conn net.Conn, out chan<- NodeMsg) {
	defer conn.Close()

	const maxMsgSize = 5 * 1024 * 1024 // 5 MB
	limited := io.LimitReader(conn, maxMsgSize+1)

	data, err := io.ReadAll(limited)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(data) > maxMsgSize {
		fmt.Println(errors.New("JSON payload exceeds 5 MB limit"))
		return
	}

	var msg NodeMsg
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&msg); err != nil {
		fmt.Printf("Failed to decode: %s\n", err)
		return
	}
	out <- msg
}

func Thing(store map[string]NodeInfo, in <-chan NodeMsg) {
	for msg := range in {
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
		fmt.Printf("Store: %v\n", store)
	}
}

func Start(address string) (net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listener: %s\n", err)
	}

	// TODO: Give thing a good name
	thingChannel := make(chan NodeMsg)
	go Thing(make(map[string]NodeInfo), thingChannel)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				} else {
					fmt.Println("Error accepting conn:", err)
					continue
				}
			}
			go parseNodeMsg(conn, thingChannel)
		}
	}()
	return listener, nil
}

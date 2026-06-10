package collector

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/stepga/monitor/config"
	"github.com/stepga/monitor/nodeinfo"
	ni "github.com/stepga/monitor/nodeinfo"
	"github.com/stepga/monitor/reporter"
)

type ListenerCollector struct{}

func parseNodeMsg(conn net.Conn, out chan<- ni.NodeInfo) {
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

	var msg ni.NodeInfo
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&msg); err != nil {
		fmt.Printf("Failed to decode: %s\n", err)
		return
	}
	out <- msg
}

func Start(address string, storeMsgChannel chan<- ni.NodeInfo) (net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listener: %s\n", err)
	}

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
			go parseNodeMsg(conn, storeMsgChannel)
		}
	}()
	return listener, nil
}

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
func StartStore(storeMsgChannel <-chan nodeinfo.NodeInfo, reporter reporter.Reporter) {
	store := make(map[string]StoreInfo)
	for msg := range storeMsgChannel {
		if _, exists := store[msg.HostName]; exists {
			reporter.Report(fmt.Sprintf("message from known node: %s\n", msg.HostName))
		} else {
			reporter.Report(fmt.Sprintf("message from unknown node: %s\n", msg.HostName))
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

func (c *ListenerCollector) Init(cfg *config.Config, reporter reporter.Reporter) {
	storeMsgChannel := make(chan ni.NodeInfo)
	go StartStore(storeMsgChannel, reporter)

	l, err := Start(cfg.Listener.Address, storeMsgChannel)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	fmt.Printf("Listening on %s\n", l.Addr())
}

func (c *ListenerCollector) Info() interface{} {
	return nil
}

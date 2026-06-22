package subsystems

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/stepga/monitor/bus"
	"github.com/stepga/monitor/config"
)

type Listener struct{}

func decodeNodeInfo(conn net.Conn) {
	defer conn.Close()

	msgMax := config.Cfg.Listener.MaxMsgSizeMB * 1024 * 1024
	limited := io.LimitReader(conn, int64(msgMax+1))

	data, err := io.ReadAll(limited)
	if err != nil {
		fmt.Printf("listener: ReadAll failed: %s\n", err)
		return
	}

	if len(data) > msgMax {
		fmt.Printf("listener: payload exceeds max size: %d\n", len(data))
		return
	}

	var info bus.NodeInfo
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&info); err != nil {
		fmt.Printf("listener: json decode failed: %s\n", err)
		return
	}
	info.Time = time.Now()
	bus.Publish(info)
}

// Starts a goroutine that accepts connections from listener and
// returns a channel which will get filled with accepted connections.
// Goroutine stops when the listener is closed.
func acceptor(listener net.Listener) <-chan net.Conn {
	out := make(chan net.Conn)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					close(out)
					return
				} else {
					fmt.Printf("listener: Accept failed: %s\n", err)
					continue
				}
			}
			out <- conn
		}
	}()
	return out
}

func (l *Listener) Init() error {
	listener, err := net.Listen("tcp", config.Cfg.Listener.Address)
	if err != nil {
		return fmt.Errorf("listener: %s", err)
	}
	out := acceptor(listener)

	go func() {
		ch := bus.Subscribe()
		defer bus.Unsubscribe(ch)
		for {
			select {
			case conn, ok := <-out:
				if ok {
					go decodeNodeInfo(conn)
				} else {
					l.Init()
					return
				}
			case m := <-ch:
				switch m.(type) {
				case bus.ConfigReloaded:
					listener.Close()
				}
			}
		}
	}()

	fmt.Printf("listener: listening on %s\n", listener.Addr())

	return nil
}

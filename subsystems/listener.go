package subsystems

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

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
		fmt.Println(err)
		return
	}

	if len(data) > msgMax {
		slog.Error("decodeNodeInfo: payload exceeds max. size",
			"MaxListenerMessageSize", msgMax,
			"size", len(data))
		return
	}

	var msg bus.NodeInfo
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&msg); err != nil {
		slog.Error("decodeNodeInfo failed", "error", err)
		return
	}
	bus.Publish(msg)
}

func (c *Listener) Init() error {
	listener, err := net.Listen("tcp", config.Cfg.Listener.Address)
	if err != nil {
		return fmt.Errorf("listener: %s", err)
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
			go decodeNodeInfo(conn)
		}
	}()
	slog.Info("node listener listens on", "address", listener.Addr())

	return nil
}

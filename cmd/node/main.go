package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"

	"github.com/stepga/monitor/nodeinfo"
)

type NodeArguments struct {
	DaemonHost string
	DaemonPort uint16
}

func (args NodeArguments) address() string {
	return fmt.Sprintf("%s:%d", args.DaemonHost, args.DaemonPort)
}

func parseNodeArguments() (*NodeArguments, error) {
	daemon := flag.String("daemon", "127.0.0.1:5566",
		"daemon address in the form of host:port")
	flag.Parse()

	host, portStr, err := net.SplitHostPort(*daemon)
	if err != nil {
		return nil, fmt.Errorf("invalid daemon address: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid port: %s", portStr)
	}

	return &NodeArguments{
		DaemonHost: host,
		DaemonPort: uint16(port),
	}, nil
}

func sendData(address string, data []byte) (int, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	return conn.Write(data)
}

func main() {
	args, err := parseNodeArguments()
	if err != nil {
		slog.Error("invalid arguments", "error", err)
		os.Exit(1)
	}

	slog.Info("print args demo", "args", args)

	info, err := nodeinfo.CreateInfo()
	if err != nil {
		slog.Error("CreateInfo() failed", "error", err)
		os.Exit(1)
	}

	info_bytes, err := info.Marshal()
	if err != nil {
		slog.Error("json.Marshal() failed", "error", err)
		os.Exit(1)
	}
	written_bytes, err := sendData(args.address(), info_bytes)
	if err != nil {
		slog.Error("sendData() failed", "error", err)
		os.Exit(1)
	}

	slog.Info("sendData() succeeded", "daemon address", args.address(), "sent bytes", written_bytes)
}

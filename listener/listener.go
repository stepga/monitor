package listener

import (
	"errors"
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Printf("Got a client: %v\n", conn)
}

func Start(address string) (net.Listener, error) {
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
			go handleConnection(conn)
		}
	}()
	return listener, nil
}

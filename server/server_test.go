package server

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestBasicServerClient(t *testing.T) {
	config := &Config{
		raft_addr: &Node{host: "localhost", port: 8001},
	}

	srv := New(config)
	srv.startServer()
	time.Sleep(time.Millisecond * 100)

	go func() {
		conn, err := net.Dial("tcp", "localhost:8001")
		if err != nil {
			fmt.Println("dial error:", err)
			return
		}
		defer conn.Close()

		peerA := NewSocketPeer(conn)

		msg := Hello{
			Id:      10,
			Details: map[string]interface{}{"foo": "bar"},
		}
		peerA.Send(msg)

		/*		m, err := peerA.Receive()
				fmt.Println("Message is ", m, "err", err)
				if m.MessageType() != 0 {
					t.Error("Error on received Hello Message ", m)
				}*/

	}()

	time.Sleep(time.Millisecond * 100)
}

package server

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestBasicServerClient(t *testing.T) {
	config := &Config{
		addr: &Node{host: "localhost", port: 8001},
	}

	srv := New(config)
	srv.startServer()
	time.Sleep(time.Millisecond * 100)

	done := make(chan bool)
	go func(d chan bool) {
		conn, err := net.Dial("tcp", "localhost:8001")
		if err != nil {
			fmt.Println("dial error:", err)
			return
		}
		defer conn.Close()

		peerA := NewJSONSocketPeer(conn)

		msg := Hello{
			Id:      10,
			Details: map[string]interface{}{"foo": "bar"},
		}
		peerA.Send(msg)
		done <- true
	}(done)

	<-done

	/*	peers := srv.PeerHandler.Peers()
		if len(peers) != 1 {
			t.Error("Unexpected Registered Peers size ", peers)
		}
		fmt.Println("Total Peer List is ", len(peers), peers)
	*/
}

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
	//Router Addition get in charge from ReadMessages Channel...so commented and wait!
	/*
		for _, p := range peers {
			timeout := time.NewTimer(time.Second * 1)
			select {
			case m := <-p.ReadMessage():
				if m == nil {
					t.Error("Unexpected receive result")
					r := <-p.ReadMessage()
					fmt.Println("r is ", r)
					t.Fail()
				}
				fmt.Println("Message is ", m)
				if m.MessageType() != 0 {
					t.Error("Error on received Hello Message ", m)
				}
			case <-timeout.C:
				t.Error("Timeout receiving expected response")
			}
	*/
}

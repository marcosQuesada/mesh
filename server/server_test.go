package server

import (
	"fmt"
	"github.com/marcosQuesada/mesh/config"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"net"
	"testing"
	"time"
)

func TestBasicServerClient(t *testing.T) {
	org := node.Node{Host: "localhost", Port: 8001}
	config := &config.Config{
		Addr: org,
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

		linkA := peer.NewJSONSocketLink(conn)

		msg := message.Hello{
			Id:      10,
			Details: map[string]interface{}{"foo": "bar"},
		}
		linkA.Send(msg)
		time.Sleep(time.Second)
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

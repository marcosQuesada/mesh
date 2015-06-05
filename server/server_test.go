package server

import (
	"fmt"
	"github.com/marcosQuesada/mesh/client"
	"github.com/marcosQuesada/mesh/cluster"
	"github.com/marcosQuesada/mesh/config"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"net"
	"testing"
	"time"
)

func TestBasicServerClient(t *testing.T) {
	config := &config.Config{
		Addr: node.Node{Host: "localhost", Port: 8001},
	}

	srv := New(config)
	srv.startServer(&cluster.Orchestrator{})
	time.Sleep(time.Millisecond * 100)

	done := make(chan bool)
	go func(d chan bool) {
		conn, err := net.Dial("tcp", "localhost:8001")
		if err != nil {
			fmt.Println("dial error:", err)
			return
		}
		defer conn.Close()

		peerA := client.NewJSONSocketPeer(conn)

		msg := message.Hello{
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

package cluster

import (
	//"fmt"
	//"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	//"net"
	"testing"
	"time"
)

var o *Orchestrator

/*
var done chan struct{} = make(chan struct{}, 0)

func TestBasicOrchestrator(t *testing.T) {
	go startBasicTestServer()
	from := node.Node{Host: "localhost", Port: 9000}
	n := node.Node{Host: "localhost", Port: 9011}
	members := make(map[string]node.Node, 2)
	members[n.String()] = n
	members[from.String()] = from // as fake local node

	o = StartOrchestrator(from, members, peer.DefaultPeerHandler(from))
	go o.Run()

	time.Sleep(time.Millisecond * 100)

	if !o.State() {
		t.Error("Expected estatus completed")
	}

	clients := o.peerHandler.Peers()
	if len(clients) != 1 {
		t.Error("Unexpected client registered size", clients)
	}

	_, ok := clients["localhost:9011"]
	if !ok {
		t.Error("Unexpected client registered", clients)
	}
	time.Sleep(time.Millisecond * 100)

	fmt.Println("Fired Ping")
	o.Exit()
}

func startBasicTestServer() error {
	fmt.Println("Started Server")
	port := ":9011"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error starting Socket Server: ", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		defer listener.Close()

		c := peer.NewAcceptor(conn, node.Node{})
		go c.Run()

		select {
		case m := <-c.ReceiveChan():
			fmt.Println("Server received m ", m)
			msg := &message.Welcome{
				Id:      m.(*message.Hello).Id,
				Details: m.(*message.Hello).Details,
			}
			err = c.Send(msg)
			if err != nil {
				fmt.Println("Error Receiving: ", err)
			}
			fmt.Println("XX Welcome Send")

			time.Sleep(time.Millisecond * 100)
			//say ping to test what happen
			pingMsg := &message.Ping{
				Id:      123,
				From:    node.Node{},
				Details: map[string]interface{}{"foo": "bar"},
			}
			c.Send(pingMsg)
			fmt.Println("XX Fired Ping")

			return nil
		}
	}
}*/

func TestForwardingChannel(t *testing.T) {
	from := node.Node{Host: "localhost", Port: 9000}
	n := node.Node{Host: "localhost", Port: 9011}
	members := make(map[string]node.Node, 2)
	members[n.String()] = n
	members[from.String()] = from // as fake local node

	o = StartOrchestrator(from, members, peer.DefaultPeerHandler(from))
	go o.Run()
	time.Sleep(time.Second)
	o.Exit()
}

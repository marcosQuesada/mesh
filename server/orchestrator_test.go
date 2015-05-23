package server

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestBasicOrchestrator(t *testing.T) {
	go startBasicTestServer()
	node := &Node{host: "localhost", port: 9011}
	members := make(map[*Node]bool, 1)
	members[node] = false

	o := StartOrchestrator(members)
	go o.Run()

	time.Sleep(time.Second * 1)

	if !o.State() {
		t.Error("Expected estatus completed")
	}
}

func startBasicTestServer() error {
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

		c := StartAcceptClient(conn)
		go c.Run()

		select {
		case m := <-c.ReceiveChan():
			fmt.Println("received m ", m)
			msg := &Welcome{
				Id:      m.(*Hello).Id,
				Details: m.(*Hello).Details,
			}
			err = c.Send(msg)
			if err != nil {
				fmt.Println("Error Receiving: ", err)
			}
		}
	}
}

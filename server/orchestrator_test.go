package server

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"testing"
	"time"
)

var o *Orchestrator

var done chan struct{} = make(chan struct{}, 0)

func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	go startBasicTestServer()

	os.Exit(m.Run())
}

func TestBasicOrchestrator(t *testing.T) {
	node := &Node{host: "localhost", port: 9011}
	members := make(map[*Node]bool, 1)
	members[node] = false

	o = StartOrchestrator(members)
	go o.Run()

	time.Sleep(time.Millisecond * 100)

	if !o.State() {
		t.Error("Expected estatus completed")
	}

	clients := o.clientHandler.Clients()
	if len(clients) != 1 {
		t.Error("Unexpected client registered size", clients)
	}
	fmt.Println("clients", clients)

	_, ok := clients["localhost:9011"]
	if !ok {
		t.Error("Unexpected client registered", clients)
	}
	time.Sleep(time.Second * 1)

	fmt.Println("Fired Ping")
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
			fmt.Println("Server received m ", m)
			msg := &Welcome{
				Id:      m.(*Hello).Id,
				Details: m.(*Hello).Details,
			}
			err = c.Send(msg)
			if err != nil {
				fmt.Println("Error Receiving: ", err)
			}
			fmt.Println("XX Welcome Send")

			time.Sleep(time.Millisecond * 100)
			//say ping to test what happen
			pingMsg := &Ping{
				Id:      123,
				From:    ID(9),
				Details: map[string]interface{}{"foo": "bar"},
			}
			c.Send(pingMsg)
			fmt.Println("XX Fired Ping")

			return nil
		}
	}
}

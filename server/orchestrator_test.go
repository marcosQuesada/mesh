package server

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestBasicOrchestrator(t *testing.T) {

	startBasicTestServer()
	node := &Node{host: "localhost", port: 8005}

	members := make(map[*Node]bool, 1)
	members[node] = false

	o := StartOrchestrator(members)
	go o.Run()

	time.Sleep(time.Second * 1)

	if !o.State() {
		t.Error("Expected estatus completed")
	}
}

func startBasicTestServer() {
	port := ":8005"
	fmt.Println("Starting server: ", port)
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error starting Socket Server: ", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err)
			continue
		}
		peer := NewJSONSocketPeer(conn)
		go handleBasicConnection(peer)
	}
}

func handleBasicConnection(peer *SocketPeer) {
	fmt.Printf("Client %v connected.", peer.Conn.RemoteAddr(), "\n")
	for {
		m, err := peer.Receive()
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error Receiving: ", err)
			}
			peer.Conn.Close()
			break
		}
		err = peer.Send(m)
		if err != nil {
			fmt.Println("Error Receiving: ", err)
		}
	}

	fmt.Println("Connection from %v closed.", peer.Conn.RemoteAddr())
}

package server

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestBasicPeersOnServerClient(t *testing.T) {
	startTestServer()
	time.Sleep(time.Second * 1)

	conn, err := net.Dial("tcp", "localhost:8000")
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

	m, err := peerA.Receive()
	fmt.Println("Message is ", m, "err", err)
	if m.MessageType() != 0 {
		t.Error("Error on received Hello Message ", m)
	}

	newMsg := m.(*Hello)
	if newMsg.Id != 10 {
		t.Error("Error on received Hello Message ", m)
	}

	if newMsg.Details["foo"].(string) != "bar" {
		t.Error("Error on received Hello Message ", m)
	}
}

func startTestServer() {
	port := ":8000"
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
		peer := NewSocketPeer(conn)
		go handleConnection(peer)
	}
}

func handleConnection(peer *SocketPeer) {
	fmt.Println("Client %v connected.", peer.Conn.RemoteAddr())
	for {
		m, err := peer.Receive()
		if err != nil {
			fmt.Println("Error Receiving: ", err)
			peer.Conn.Close()
			break
		}
		err = peer.Send(m)
		if err != nil {

		}
	}

	fmt.Println("Connection from %v closed.", peer.Conn.RemoteAddr())
}

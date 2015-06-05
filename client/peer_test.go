package client

import (
	"fmt"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"io"
	"net"
	"testing"
	"time"
)

type testPeerHandler struct {
	lastMsg *message.Hello
	outChan chan *message.Hello
	exit    bool
}

func TestPeersOnPipes(t *testing.T) {
	a, b := net.Pipe()

	peerA := NewJSONSocketPeer(a)
	peerB := NewJSONSocketPeer(b)
	defer peerA.Conn.Close()
	defer peerB.Conn.Close()

	tp := &testPeerHandler{
		outChan: make(chan *message.Hello, 1),
	}
	go tp.handlePeer(peerA, node.Node{Host: "localhost", Port: 5000})
	go tp.handlePeer(peerB, node.Node{Host: "localhost", Port: 5005})

	//first message
	msg := message.Hello{
		Id:      0,
		From:    node.Node{Host: "localhost", Port: 5000},
		Details: map[string]interface{}{"foo": "bar"},
	}
	peerA.Send(msg)

	time.Sleep(time.Millisecond * 300)
	tp.exit = true
	lastMessage := <-tp.outChan
	fmt.Println("lastMessage A", lastMessage)

	if int(lastMessage.Id) == 0 {
		t.Error("LastMessage must be bigger ")
	}

	if lastMessage.Details["foo"].(string) != "PING" && lastMessage.Details["foo"].(string) != "PONG" {
		t.Error("LastMessageDetails has not changed! ")
	}
}

func (ph *testPeerHandler) handlePeer(p *SocketPeer, from node.Node) {
	defer func() {
		ph.outChan <- ph.lastMsg
	}()

	for !ph.exit {
		msg, err := p.Receive()
		if err != nil {
			fmt.Println("Error Receiving on server, err ", err)
			return
		}

		if msg.MessageType() != 0 {
			fmt.Println("Error on received Hello Message ", msg)
			return
		}

		newMsg := msg.(*message.Hello)
		if newMsg.Details["foo"].(string) == "PONG" {
			newMsg.Details["foo"] = "PING"
		} else {
			newMsg.Details["foo"] = "PONG"
		}
		newMsg.Id++
		newMsg.From = from

		err = p.Send(newMsg)
		if err != nil {
			//fmt.Println("Error sending: ", err)
			return
		}
		ph.lastMsg = newMsg
	}
}

func TestBasicPeersOnServerClient(t *testing.T) {
	go startTestServer()
	time.Sleep(time.Second * 1)

	conn, err := net.Dial("tcp", "localhost:8002")
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}
	defer conn.Close()

	peerA := NewJSONSocketPeer(conn)

	msg := message.Hello{
		Id:      10,
		Details: map[string]interface{}{"foo": "bar"},
	}
	peerA.Send(msg)

	msgA, err := peerA.Receive()
	if err != nil {
		t.Error("Error on received Hello Message ", err)
	}

	fmt.Println("Message is ", msgA, "err", err)
	if msgA.MessageType() != 0 {
		t.Error("Error on received Hello Message ", msgA)
	}

	newMsg := msgA.(*message.Hello)
	if newMsg.Id != 10 {
		t.Error("Error on received Hello Message ", msgA)
	}

	if newMsg.Details["foo"].(string) != "bar" {
		t.Error("Error on received Hello Message ", msgA)
	}
}

func startTestServer() {
	port := ":8002"
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
		defer listener.Close()
		peer := NewJSONSocketPeer(conn)
		go handleConnection(peer)
	}
}

func handleConnection(peer *SocketPeer) {
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
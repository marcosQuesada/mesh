package peer_handler

import (
	//"fmt"
	"net"
	"testing"

	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	//"time"
)

var c1 *peer.NopPeer
var c2 *peer.NopPeer
var clh *defaultPeerHandler

func TestBasicPeerHandler(t *testing.T) {
	c1 = &peer.NopPeer{Host: "foo", Port: 1234}
	c2 = &peer.NopPeer{Host: "bar", Port: 1234}

	clh = DefaultPeerHandler(node.Node{})
	err := clh.accept(c1)
	if err != nil {
		t.Error("Unexpected Error Accepting PeerNode")
	}

	err = clh.accept(c2)
	if err != nil {
		t.Error("Unexpected Error Accepting PeerNode")
	}

	if len(clh.peers) != 2 {
		t.Error("Unexpected client size ", clh.peers)
	}
}
func TestErrorOnAddTwiceSameClient(t *testing.T) {
	err := clh.accept(c2)
	if err == nil {
		t.Error("Unexpected Error Accepting PeerNode", err)
	}
	if err.Error() != "Peer: bar:1234 Already registered" {
		t.Error("Unexpected error message", err.Error())
	}
}

func TestToRemoveCLientFromPeerHandler(t *testing.T) {
	err := clh.Remove(c2)
	if err != nil {
		t.Error("Unexpected Error Removing PeerNode", err)
	}

	if len(clh.peers) != 1 {
		t.Error("Unexpected client size ", clh.peers)
	}
}

func TestErrorOnRemoveInexistentCLient(t *testing.T) {
	err := clh.Remove(c2)
	if err == nil {
		t.Error("Unexpected Error Remove PeerNode", err)
	}
	if err.Error() != "Peer Not found" {
		t.Error("Unexpected error message", err.Error())
	}
}

func TestHandlePeerUsingPipes(t *testing.T) {
	nodeA := node.Node{Host: "A", Port: 1}
	nodeB := node.Node{Host: "B", Port: 2}
	a, b := net.Pipe()

	c1 := peer.NewAcceptor(a, nodeA)
	c1.Identify(nodeB)
	go c1.Run()

	c1Mirror := peer.NewAcceptor(b, nodeB)
	c1Mirror.Identify(nodeA)
	go c1Mirror.Run()

	peerHandler := &defaultPeerHandler{
		watcher:   &fakeWatch{},
		peers:     make(map[string]peer.NodePeer),
		from:      node.Node{Host: "Server", Port: 1},
		eventChan: make(chan dispatcher.Event, 10),
		peerChan:  make(chan message.Message, 10),
	}

	c1Mirror.SayHello()
	peerHandler.Handle(c1)

	select {
	case msg, open := <-c1Mirror.ReceiveChan():
		if !open {
			t.Error("Closed Receive Chan")
		}
		if msg != nil {
			if msg.MessageType() != 1 {
				t.Error("Unexpected MessageType")
			}
		}
	}
	event := <-peerHandler.Events()
	if event.GetEventType() != "OnPeerConnectedEvent" {
		t.Error("Unexpected Event Type")
	}

	if len(peerHandler.peers) != 1 {
		t.Error("Unexpected Peer Handler list size ")
	}

	//@TODO : WRONG ORIGIN???
	c1Node := c1.Node()
	if _, ok := peerHandler.peers[c1Node.String()]; !ok {
		t.Error("Peer has not been accepted")
	}

	/*	c1Mirror.SayHello()
		time.Sleep(time.Second)
		peerHandler.Handle(c1)
		select {
		case msg, open := <-c1Mirror.ReceiveChan():
			if !open {
				t.Error("Closed Receive Chan")
			}
			if msg != nil {
				if msg.MessageType() != 2 {
					t.Error("Unexpected MessageType")
				}
			}
		}
	*/
	c1.Exit()
	c1Mirror.Exit()
	close(peerHandler.eventChan)
	close(peerHandler.peerChan)
}

type fakeWatch struct {
}

func (f *fakeWatch) Watch(peer.NodePeer) {

}

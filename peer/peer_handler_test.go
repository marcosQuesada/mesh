package peer

import (
	n "github.com/marcosQuesada/mesh/node"
	"testing"
)

var c1 *NopPeer
var c2 *NopPeer
var clh *defaultPeerHandler

func TestBasicPeerHandler(t *testing.T) {
	c1 = &NopPeer{host: "foo", port: 1234}
	c2 = &NopPeer{host: "bar", port: 1234}

	clh = DefaultPeerHandler(n.Node{})
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
	err := clh.remove(c2)
	if err != nil {
		t.Error("Unexpected Error Removing PeerNode", err)
	}

	if len(clh.peers) != 1 {
		t.Error("Unexpected client size ", clh.peers)
	}
}

func TestErrorOnRemoveInexistentCLient(t *testing.T) {
	err := clh.remove(c2)
	if err == nil {
		t.Error("Unexpected Error Remove PeerNode", err)
	}
	if err.Error() != "Peer Not found" {
		t.Error("Unexpected error message", err.Error())
	}
}

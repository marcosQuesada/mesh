package peer

import (
	"testing"
)

var c1 *NopPeer
var c2 *NopPeer
var clh PeerHandler

func TestBasicPeerHandler(t *testing.T) {
	c1 = &NopPeer{host: "foo", port: 1234}
	c2 = &NopPeer{host: "bar", port: 1234}

	clh = DefaultPeerHandler()
	err := clh.Accept(c1)
	if err != nil {
		t.Error("Unexpected Error Accepting PeerNode")
	}

	err = clh.Accept(c2)
	if err != nil {
		t.Error("Unexpected Error Accepting PeerNode")
	}

	peers := clh.Peers()
	if len(peers) != 2 {
		t.Error("Unexpected client size ", clh.Peers())
	}
}
func TestErrorOnAddTwiceSameClient(t *testing.T) {
	err := clh.Accept(c2)
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

	peers := clh.Peers()
	if len(peers) != 1 {
		t.Error("Unexpected client size ", clh.Peers())
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

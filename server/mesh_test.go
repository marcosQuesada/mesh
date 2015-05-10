package server

import (
	"testing"
)

type testCase []*Node

var ts = testCase{
	&Node{host: "127.0.0.1", port: 123},
	&Node{host: "127.0.0.2", port: 123},
	&Node{host: "127.0.0.3", port: 123},
}

var nodes = []address{address("127.0.0.1:123"), address("127.0.0.2:123"), address("127.0.0.3:123")}
var m = InitMesh(nodes)

func TestBasicMesh(t *testing.T) {

	if m.Completed() {
		t.Error("Unexpected complete")
	}

	for _, node := range ts {
		peer := &Peer{Link: NewDialLink()}
		peer.Identify(node)

		if !m.JoinRequest(peer) {
			t.Error("Unexpected Join Request")
		}
	}
}

func TestMeshIsCompleted(t *testing.T) {
	if !m.Completed() {
		t.Error("Unexpected uncomplete")
	}
}

func TestRejectOnExistentNode(t *testing.T) {
	node := &Node{host: "127.0.0.1", port: 123}
	peer := &Peer{Link: NewDialLink()}
	peer.Identify(node)

	if m.JoinRequest(peer) {
		t.Error("Unexpected Join Request")
	}
}

func TestRemovePeerFromMesh(t *testing.T) {
	node := &Node{host: "127.0.0.1", port: 123}
	peer := &Peer{Link: NewDialLink()}
	peer.Identify(node)

	if !m.RemoveRequest(peer) {
		t.Error("Unexpected Remove Request")
	}

	if m.RemoveRequest(peer) {
		t.Error("Unexpected Remove Request")
	}
}

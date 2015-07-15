package peer

import (
	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
)

type OnPeerConnectedEvent struct {
	Node  node.Node
	Event message.Status
	//Peer  NodePeer
}

func (m OnPeerConnectedEvent) GetEventType() dispatcher.EventType {
	return "OnPeerConnectedEvent"
}

type OnPeerAbortedEvent struct {
	Node  node.Node
	Event message.Status
}

func (m OnPeerAbortedEvent) GetEventType() dispatcher.EventType {
	return "OnPeerAbortedEvent"
}

type OnPeerErroredEvent struct {
	Node  node.Node
	Event message.Status
	Error error
}

func (m OnPeerErroredEvent) GetEventType() dispatcher.EventType {
	return "OnPeerErroredEvent"
}

type OnPeerDisconnectedEvent struct {
	Node  node.Node
	Event message.Status
	//Peer  NodePeer
}

func (m OnPeerDisconnectedEvent) GetEventType() dispatcher.EventType {
	return "OnPeerDisconnectedEvent"
}

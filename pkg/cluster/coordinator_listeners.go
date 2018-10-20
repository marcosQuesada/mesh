package cluster

import (
	"github.com/marcosQuesada/mesh/pkg/dispatcher"
	"github.com/marcosQuesada/mesh/pkg/peer"
	"log"
)

func (c *Coordinator) OnPeerConnectedEvent(e dispatcher.Event) {
	event := e.(*peer.OnPeerConnectedEvent)
	c.members[event.Node.String()] = event.Node
	c.connected[event.Node.String()] = true
	log.Println("OnPeerConnectedEvent, adding peer", event.Node.String(), "mode:", event.Mode)
}

func (c *Coordinator) OnPeerDisconnected(e dispatcher.Event) {
	event := e.(*peer.OnPeerDisconnectedEvent)
	c.members[event.Node.String()] = event.Node

	c.connected[event.Node.String()] = false
	log.Println("OnPeerDisconnectedEvent, removing peer", event.Node.String())
}

func (c *Coordinator) OnPeerAborted(e dispatcher.Event) {
	n := e.(*peer.OnPeerAbortedEvent)
	log.Println("OnPeerAbortedEvent", n.Node.String())
}

func (c *Coordinator) OnPeerErrored(e dispatcher.Event) {
	n := e.(*peer.OnPeerErroredEvent)
	c.connected[n.Node.String()] = false
	log.Println("OnPeerErroredEvent", n.Node.String())
}

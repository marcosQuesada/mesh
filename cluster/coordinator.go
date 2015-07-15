package cluster

import (
	"log"

	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
)

// Coordinator takes cares on all cluster related tasks
// Peer election & registration (could be delegated to PeerHandler)
// Cluster member definition & cluster status
// Leader election
// Execute Pool mechanisms and consesus resolution

const (
	ClusterStatusStarting  = message.Status("starting")
	ClusterStatusInService = message.Status("in service")
	ClusterStatusDegraded  = message.Status("degraded")
	ClusterStatusExit      = message.Status("exit")
)

type Coordinator struct {
	//peerHandler peer_handler.PeerHandler
	from     n.Node
	members  map[string]n.Node
	exitChan chan bool
}

func StartCoordinator(from n.Node, members map[string]n.Node) *Coordinator {
	return &Coordinator{
		from:     from,
		members:  members,
		exitChan: make(chan bool, 0),
	}
}

func (c *Coordinator) Run() {
	for {
		select {
		case <-c.exitChan:
			return
			/*		case m := <-c.peerHandler.AggregatedChan():
					log.Println("SERVER: Received Message on Main Channel ", m)*/
		}
	}
}

func (c *Coordinator) Exit() {
	c.exitChan <- true
}

func (c *Coordinator) OnPeerConnectedEvent(e dispatcher.Event) {
	event := e.(*peer.OnPeerConnectedEvent)
	c.members[event.Node.String()] = event.Node
	log.Println("OnPeerConnectedEvent, adding peer", event.Node.String()) //, "mode:", event.Peer.Mode()
}

func (c *Coordinator) OnPeerDisconnected(e dispatcher.Event) {
	event := e.(*peer.OnPeerDisconnectedEvent)
	c.members[event.Node.String()] = event.Node

	log.Println("OnPeerDisconnectedEvent, removing peer", event.Node.String())

	//c.peerHandler.Remove(event.Peer)
	//restart Dial Peer and try again
	//go c.peerHandler.InitDialClient(event.Node)
}

func (c *Coordinator) OnPeerAborted(e dispatcher.Event) {
	n := e.(*peer.OnPeerAbortedEvent)
	log.Println("OnPeerAbortedEvent", n.Node.String())
}

func (c *Coordinator) OnPeerErrored(e dispatcher.Event) {
	n := e.(*peer.OnPeerErroredEvent)
	log.Println("OnPeerErroredEvent", n.Node.String())
}

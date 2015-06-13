package cluster

import (
	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/peer_handler"
	"log"
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
	peerHandler peer_handler.PeerHandler
	from        n.Node
	members     map[string]n.Node
	exitChan    chan bool
}

func StartCoordinator(from n.Node, members map[string]n.Node, clh peer_handler.PeerHandler) *Coordinator {
	return &Coordinator{
		peerHandler: clh,
		from:        from,
		members:     members,
		exitChan:    make(chan bool, 0),
	}
}

func (c *Coordinator) Run() {
	go func() {
		for {
			select {
			case <-c.exitChan:
				return
			case m := <-c.peerHandler.AggregatedChan():
				log.Println("SERVER: Received Message on Main Channel ", m)
			}
		}
	}()
}

func (c *Coordinator) Exit() {
	c.exitChan <- true
}

func (c *Coordinator) OnPeerConnectedEvent(e dispatcher.Event) {
	n := e.(*peer.OnPeerConnectedEvent)
	c.members[n.Node.String()] = n.Node
	log.Println("Called Coordinator OnPeerConnectedEvent, adding peer", n.Node.String())
}

func (c *Coordinator) OnPeerDisconnected(e dispatcher.Event) {
	n := e.(*peer.OnPeerDisconnectedEvent)
	c.members[n.Node.String()] = n.Node
	log.Println("Called Coordinator OnPeerDisconnectedEvent, removing peer", n.Node.String())
}

func (c *Coordinator) OnPeerAborted(e dispatcher.Event) {
	n := e.(*peer.OnPeerAbortedEvent)
	log.Println("Called Coordinator OnPeerAbortedEvent", n.Node.String())
}

func (c *Coordinator) OnPeerErrored(e dispatcher.Event) {
	n := e.(*peer.OnPeerErroredEvent)
	log.Println("Called Coordinator OnPeerErroredEvent", n.Node.String())
}

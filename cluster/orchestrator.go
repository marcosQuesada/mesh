package cluster

import (
	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"log"
)

// Orchestrator takes cares on all cluster related tasks
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

type Orchestrator struct {
	peerHandler peer.PeerHandler
	from        n.Node
	members     map[string]n.Node
	exitChan    chan bool
}

func StartOrchestrator(from n.Node, members map[string]n.Node, clh peer.PeerHandler) *Orchestrator {
	return &Orchestrator{
		peerHandler: clh,
		from:        from,
		members:     members,
		exitChan:    make(chan bool, 0),
	}
}

func (o *Orchestrator) Run() {
	defer log.Println("Exiting Orchestrator Run")
	defer close(o.exitChan)

	//Start Dial Peers
	for _, node := range o.members {
		//avoid local connexion
		if node.String() == o.from.String() {
			continue
		}

		go func(d n.Node) {
			log.Println("Starting Dial Client on Node ", o.from.String(), "destination: ", node.String())
			//Blocking call, wait until connection success
			c := peer.NewDialer(o.from, d)
			c.Run()
			//Say Hello and wait response
			c.SayHello()
			r := o.peerHandler.Handle(c)
			log.Println("Dial link to to:", c.Node(), "result: ", r)
		}(node)
	}
}

func (o *Orchestrator) Exit() {
	o.exitChan <- true
}

func (o *Orchestrator) OnPeerConnectedEvent(e dispatcher.Event) {
	n := e.(*peer.OnPeerConnectedEvent)
	o.members[n.Node.String()] = n.Node
	log.Println("Called Orchestrator OnPeerConnectedEvent", e)
}

func (o *Orchestrator) OnPeerDisconnected(e dispatcher.Event) {
	n := e.(*peer.OnPeerDisconnectedEvent)
	//o.members[msg.Node.String()] = msg.Node
	log.Println("Called Orchestrator OnPeerDisconnectedEvent", n)
}

func (o *Orchestrator) OnPeerAborted(e dispatcher.Event) {
	n := e.(*peer.OnPeerAbortedEvent)
	//o.members[msg.Node.String()] = msg.Node
	log.Println("Called Orchestrator OnPeerAbortedEvent", n)
}

func (o *Orchestrator) OnPeerErrored(e dispatcher.Event) {
	n := e.(*peer.OnPeerErroredEvent)
	//o.members[msg.Node.String()] = msg.Node
	log.Println("Called Orchestrator OnPeerErroredEvent", n)
}

package cluster

import (
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"log"
	//"time"
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

type memberUpdate struct {
	node  n.Node
	event message.Status
}

type Orchestrator struct {
	peerHandler peer.PeerHandler
	from        n.Node
	members     map[string]n.Node
	MainChan    chan message.Message // Used as aggregated channel from Client Peers
	fwdChan     chan message.Message //Channel that gets routed to destination client
	exitChan    chan bool
}

func StartOrchestrator(from n.Node, members map[string]n.Node, clh peer.PeerHandler) *Orchestrator {
	return &Orchestrator{
		peerHandler: clh,
		from:        from,
		members:     members,
		MainChan:    make(chan message.Message, 0),
		fwdChan:     make(chan message.Message, 0),
		exitChan:    make(chan bool, 0),
	}
}

func (o *Orchestrator) Run() {
	defer log.Println("Exiting Orchestrator Run")
	defer close(o.exitChan)

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
			log.Println("Result from Dial is ", r)
		}(node)
	}

	//Member Updates to Coordinator?  handle updates from members
	for {
		select {
		case msg := <-o.peerHandler.Updates():
			switch msg.Event {
			case peer.PeerStatusConnected:
				log.Println("XXX  Peer Status Connected")
				//aggregate receiveChan to mainChan
				o.aggregate(msg.Peer.ReceiveChan())
				o.members[msg.Node.String()] = msg.Node
				if o.State() {
					log.Println("Cluster Completed!", "Total Members:", len(o.members))
				}

			case peer.PeerStatusError:
				log.Println("Client Exitting", msg.Node)
			}
		case <-o.exitChan:
			return
		}
	}
}

func (o *Orchestrator) State() bool {
	return o.peerHandler.Len() == (len(o.members) - 1) //@TODO: BAD APPROACH!
}

func (o *Orchestrator) Exit() {
	o.exitChan <- true
}

func (o *Orchestrator) aggregate(c chan message.Message) {
	log.Println("Aggregate")
	go func() {
		for {
			select {
			case m := <-c:
				o.MainChan <- m
			}
		}
	}()
}

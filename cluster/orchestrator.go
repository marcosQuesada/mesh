package cluster

import (
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"log"
	"time"
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
	clients     map[string]*peer.Peer
	inChan      chan memberUpdate
	MainChan    chan message.Message // Used as aggregated channel from Client Peers
	fwdChan     chan message.Message //Channel that gets routed to destination client
	exitChan    chan bool
}

func StartOrchestrator(from n.Node, members map[string]n.Node, clh peer.PeerHandler) *Orchestrator {
	return &Orchestrator{
		peerHandler: clh,
		from:        from,
		members:     members,
		clients:     make(map[string]*peer.Peer, 0),
		inChan:      make(chan memberUpdate, 0),
		MainChan:    make(chan message.Message, 0),
		fwdChan:     make(chan message.Message, 0),
		exitChan:    make(chan bool, 0),
	}
}

func (o *Orchestrator) Run() {
	defer log.Println("Exiting Orchestrator Run")
	defer close(o.exitChan)
	defer close(o.inChan)

	for _, node := range o.members {
		//avoid local connexion
		if node.String() == o.from.String() {
			continue
		}

		go o.bootClient(node)
	}

	//Member Updates to Coordinator?  handle updates from members
	for {
		select {
		case msg := <-o.inChan:
			switch msg.event {
			case peer.PeerStatusConnected:
				o.members[msg.node.String()] = msg.node
				if o.State() {
					log.Println("Cluster Completed!", "Total Members:", len(o.members), "Total Clients:", len(o.clients))
				}
			case peer.PeerStatusError:
				log.Println("Client Exitting", msg.node)
			}
		case <-o.exitChan:
			return
		}
	}

}

func (o *Orchestrator) State() bool {
	log.Println("len is ", len(o.clients))
	return o.peerHandler.Len() == (len(o.members) - 1) //@TODO: BAD APPROACH!
}

func (o *Orchestrator) Accept(p peer.NodePeer) (response message.Status) {
	select {
	case msg := <-p.ReceiveChan():
		log.Println("Msg Received ", msg)

		switch msg.(type) {
		case *message.Hello:
			p.Identify(msg.(*message.Hello).From)
			err := o.peerHandler.Accept(p)
			if err != nil {
				p.Send(&message.Abort{Id: msg.(*message.Hello).Id, From: o.from, Details: map[string]interface{}{"foo_bar": 1231}})
				p.Exit()

				response = peer.PeerStatusAbort
			}
			p.Send(&message.Welcome{Id: msg.(*message.Hello).Id, From: o.from, Details: map[string]interface{}{"foo_bar": 1231}})

			response = peer.PeerStatusConnected
		case *message.Ping:
			log.Println("Router Ping: ", msg.(*message.Ping))
			p.Send(&message.Pong{Id: msg.(*message.Ping).Id, From: o.from, Details: map[string]interface{}{}})
		}
	}

	return
}

func (o *Orchestrator) Exit() {
	o.exitChan <- true
}

func (o *Orchestrator) bootClient(node n.Node) {
	log.Println("Orchestrartor boot client", node)

	var c *peer.Peer
	log.Println("Starting Dial Client on Node ", o.from.String(), "destination: ", node.String())
	//Blocking call, wait until connection success
	c = peer.NewDialer(o.from, node)
	c.Run()

	//Say Hello and wait response
	c.SayHello()

	select {
	case <-time.NewTimer(time.Second).C:
		log.Println("Client has not receive response, Timeout")
		return
	case rsp := <-c.ReceiveChan():
		switch rsp.(type) {
		case *message.Welcome:
			log.Println("Client has received Welcome from", node.String(), rsp.(*message.Welcome))
			err := o.peerHandler.Accept(c)
			if err != nil {
				log.Println("Error Accepting Peer, Peer dies! ", err)
				o.inChan <- memberUpdate{
					node:  node,
					event: peer.PeerStatusError,
				}
				return
			} else {
				o.clients[node.String()] = c
				o.inChan <- memberUpdate{
					node:  node,
					event: peer.PeerStatusConnected,
				}

				//aggregate receiveChan to mainChan
				o.aggregate(c.ReceiveChan())
				log.Println("Client Achieved: ", node, "Cluster Status ", o.State())
			}
		case *message.Abort:
			log.Println("Response Abort ", rsp.(*message.Abort), " remote node:", node.String())
			o.inChan <- memberUpdate{
				node:  node,
				event: peer.PeerStatusError,
			}
		default:
			log.Println("Unexpected type On response ")
		}
	}

}

func (o *Orchestrator) aggregate(c chan message.Message) {
	go func() {
		for {
			select {
			case m := <-c:
				o.MainChan <- m
			}
		}
	}()
}

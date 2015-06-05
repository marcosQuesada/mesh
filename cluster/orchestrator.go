package cluster

import (
	"github.com/marcosQuesada/mesh/client"
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
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

type memberUpdate struct {
	node  n.Node
	event message.Status // ??
}

type Orchestrator struct {
	clientHandler client.ClientHandler
	from          n.Node
	members       map[string]n.Node
	clients       map[string]*client.Client
	inChan        chan memberUpdate
	MainChan      chan message.Message // Used as aggregated channel from Client Peers
	fwdChan       chan message.Message //Channel that gets routed to destination client
	exitChan      chan bool
}

func StartOrchestrator(from n.Node, members map[string]n.Node, clh client.ClientHandler) *Orchestrator {
	return &Orchestrator{
		clientHandler: clh,
		from:          from,
		members:       members,
		clients:       make(map[string]*client.Client, 0),
		inChan:        make(chan memberUpdate, 0),
		MainChan:      make(chan message.Message, 0),
		fwdChan:       make(chan message.Message, 0),
		exitChan:      make(chan bool, 0),
	}
}

func (o *Orchestrator) Run() {
	defer log.Println("Exiting Orchestrator Run")
	defer close(o.exitChan)
	defer close(o.inChan)

	o.bootClients()
	//handle updates from members
	for {
		select {
		case msg := <-o.inChan:
			//log.Println("Handle Update ", m, "State is: ", o.State())
			switch msg.event {
			case client.ClientStatusConnected:
				o.members[msg.node.String()] = msg.node
				if o.State() {
					log.Println("Cluster Completed!", "Total Members:", len(o.members), "Total Clients:", len(o.clients))
				}
			case client.ClientStatusError:
				log.Println("Client Exitting", msg.node)
			}
		case <-o.exitChan:
			return
		}
	}

}

func (o *Orchestrator) Exit() {
	o.exitChan <- true
}

func (o *Orchestrator) State() bool {
	log.Println("len is ", len(o.clients))
	return o.clientHandler.Len() == (len(o.members) - 1) //@TODO: BAD APPROACH!
}

func (o *Orchestrator) bootClients() {
	log.Println("ORchestrartor boot clients", len(o.members), o.members)
	for _, node := range o.members {
		//avoid local connexion
		if node.String() == o.from.String() {
			continue
		}

		var c *client.Client
		go func(n n.Node) {
			log.Println("Starting Dial Client on Node ", o.from.String(), "destination: ", n.String())
			//Blocking call, wait until connection success
			c = client.StartDial(o.from, n)
			c.Run()

			//Say Hello and wait response
			c.SayHello()

			//Handle Response
			//Blocking call again until response
			rsp := <-c.ReceiveChan()
			switch rsp.(type) {
			case *message.Welcome:
				log.Println("Client has received Welcome from", n.String(), rsp.(*message.Welcome))
				err := o.clientHandler.Accept(c)
				if err != nil {
					log.Println("Error Accepting Peer, Peer dies! ", err)
					o.inChan <- memberUpdate{
						node:  n,
						event: client.ClientStatusError,
					}
					return
				} else {
					o.clients[n.String()] = c
					o.inChan <- memberUpdate{
						node:  n,
						event: client.ClientStatusConnected,
					}

					//aggregate receiveChan to mainChan
					o.aggregate(c.ReceiveChan())
					log.Println("Client Achieved: ", n, "Cluster Status ", o.State())
				}
			case *message.Abort:
				log.Println("Response Abort ", rsp.(*message.Abort), " remote node:", n.String())
				o.inChan <- memberUpdate{
					node:  n,
					event: client.ClientStatusError,
				}
			default:
				log.Println("Unexpected type On response ")
			}
		}(node)
	}
}

func (o *Orchestrator) Accept(p client.PeerClient) (response message.Status) {
	select {
	case msg := <-p.ReceiveChan():
		log.Println("Msg Received ", msg)

		switch msg.(type) {
		case *message.Hello:
			p.Identify(msg.(*message.Hello).From)
			err := o.clientHandler.Accept(p)
			if err != nil {
				p.Send(&message.Abort{Id: msg.(*message.Hello).Id, From: o.from, Details: map[string]interface{}{"foo_bar": 1231}})
				p.Exit()

				response = client.ClientStatusAbort
			}
			p.Send(&message.Welcome{Id: msg.(*message.Hello).Id, From: o.from, Details: map[string]interface{}{"foo_bar": 1231}})

			response = client.ClientStatusConnected
			/*		case *Ping:
						log.Println("Router Ping: ", msg.(*Ping))
						p.Send(&Pong{Id: msg.(*Ping).Id, From: o.from, Details: map[string]interface{}{}})

					case *Pong:
						log.Println("Router Pong: ", msg.(*Pong))
						pong := msg.(*Pong)
						p.Send(&Error{Id: pong.Id, Details: pong.Details})
			*/
		}
	}

	return
}

func (o *Orchestrator) Route(m message.Message) {

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

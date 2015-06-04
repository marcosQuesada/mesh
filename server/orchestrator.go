package server

import (
	"log"
)

// Orchestrator takes cares on all cluster related tasks
// Peer election & registration (could be delegated to PeerHandler)
// Cluster member definition & cluster status
// Leader election
// Execute Pool mechanisms and consesus resolution

const (
	ClusterStatusStarting  = Status("starting")
	ClusterStatusInService = Status("in service")
	ClusterStatusDegraded  = Status("degraded")
)

type memberUpdate struct {
	node  Node
	event Status // ??
}

type Orchestrator struct {
	clientHandler ClientHandler
	from          Node
	members       map[string]Node
	clients       map[string]*Client
	inChan        chan memberUpdate
	mainChan      chan Message // Used as aggregated channel from Client Peers
	fwdChan       chan Message //Channel that gets routed to destination client
	exitChan      chan bool
}

func StartOrchestrator(from Node, members map[string]Node, clh ClientHandler) *Orchestrator {
	return &Orchestrator{
		clientHandler: clh,
		from:          from,
		members:       members,
		clients:       make(map[string]*Client, 0),
		inChan:        make(chan memberUpdate, 0),
		mainChan:      make(chan Message, 0),
		fwdChan:       make(chan Message, 0),
		exitChan:      make(chan bool, 0),
	}
}

func (o *Orchestrator) Run() {
	defer log.Println("Exiting Orchestrator Run")
	defer close(o.exitChan)
	defer close(o.inChan)
	go o.consumeMainChannel()
	o.bootClients()
	//handle updates from members
	for {
		select {
		case m := <-o.inChan:
			//log.Println("Handle Update ", m, "State is: ", o.State())
			switch m.event {
			case ClientStatusConnected:
				o.members[m.node.String()] = m.node
				if o.State() {
					log.Println("Cluster Completed!", "Total Members:", len(o.members), "Total Clients:", len(o.clients))
				}
			case ClientStatusError:
				log.Println("Client Exitting", m.node)
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
	log.Println("ORchestrartor boot clients", o.members)
	for _, node := range o.members {
		//avoid local connexion
		if node.String() == o.from.String() {
			continue
		}

		var c *Client
		go func(n Node) {
			log.Println("Starting Dial Client on Node ", o.from.String(), "destination: ", n.String())
			//Blocking call, wait until connection success
			c = StartDialClient(o.from, n)
			go c.Run()

			//Say Hello and wait response
			c.SayHello()

			//Blocking call again until response
			rsp := <-c.ReceiveChan()
			switch rsp.(type) {
			case *Welcome:
				log.Println("Client has received Welcome from", n.String(), rsp.(*Welcome))
				err := o.clientHandler.Accept(c)
				if err != nil {
					log.Println("Error Accepting Peer, Peer dies! ", err)
					o.inChan <- memberUpdate{
						node:  n,
						event: ClientStatusError,
					}
					return
				} else {
					o.clients[n.String()] = c
					o.inChan <- memberUpdate{
						node:  n,
						event: ClientStatusConnected,
					}

					//aggregate receiveChan to mainChan
					o.aggregate(c.ReceiveChan())
					log.Println("Client Achieved: ", n, "Cluster Status ", o.State())
				}
			case *Abort:
				log.Println("Response Abort ", rsp.(*Abort), " remote node:", n.String())
				o.inChan <- memberUpdate{
					node:  n,
					event: ClientStatusError,
				}
			default:
				log.Println("Unexpected type On response ")
			}
		}(node)
	}
}

func (o *Orchestrator) consumeMainChannel() {
	for {
		select {
		case m := <-o.mainChan:
			log.Println("Received Message on Main Channel ", m)
			//
			// Consumers may be exported
		}
	}
}

func (o *Orchestrator) aggregate(c chan Message) {
	go func() {
		for {
			select {
			case m := <-c:
				o.mainChan <- m
			}
		}
	}()
}

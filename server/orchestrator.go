package server

import (
	"log"
	"sync"
	//"time"
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
	from          Node
	clientHandler ClientHandler
	members       map[string]Node
	clients       map[string]*Client
	inChan        chan memberUpdate
	mainChan      chan Message // Used as aggregated channel from Client Peers
	exitChan      chan bool
}

func StartOrchestrator(from Node, members map[string]Node, clh ClientHandler) *Orchestrator {
	return &Orchestrator{
		clientHandler: clh,
		members:       members,
		clients:       make(map[string]*Client, 0),
		inChan:        make(chan memberUpdate, 0),
		exitChan:      make(chan bool, 0),
		from:          from,
		mainChan:      make(chan Message, 0),
	}
}

func (o *Orchestrator) Run() {
	defer log.Println("Exiting Orchestrator Run")
	go o.consumeMainChannel()
	o.bootClients()
	//handle updates from members
	for {
		select {
		case m := <-o.inChan:
			log.Println("Handle Update ", m, "State is: ", o.State())
			switch m.event {
			case ClientStatusConnected:
				o.members[m.node.String()] = m.node
				if o.State() {
					log.Println("Cluster Completed!", "Total Members:", len(o.members), "Total Clients:", len(o.clients))
				}
			case ClientStatusError:
				log.Println("Client Exitting")
				/*				o.members[m.node.String()] = nil
								o.clients[m.node.String()] = nil*/
			}
		case <-o.exitChan:
			return
		}
	}

}

func (o *Orchestrator) State() bool {
	for _, s := range o.clients {
		if s == nil {
			return false
		}
	}

	return true
}

func (o *Orchestrator) bootClients() {
	var done sync.WaitGroup
	for _, node := range o.members {
		/*		if v != nil {*/
		var c *Client
		done.Add(1)
		go func(n Node) {
			//Blocking call, wait until connection success
			log.Println("Starting Dial Client on Node ", o.from.String(), "destination: ", n.String())
			c = StartDialClient(o.from, n)
			go c.Run()
			done.Done()

			//Say Hello and wait response
			log.Println("Say Hello from ", n.String(), o.from.String())
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
					log.Println("Assign Client node:", n.String())
					o.clients[n.String()] = c
					o.inChan <- memberUpdate{
						node:  n,
						event: ClientStatusConnected,
					}

					//aggregate receiveChan to mainChan
					o.aggregate(c.ReceiveChan())
					log.Println("Client Achieved: ", n)
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
		//}
	}
	done.Wait()
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

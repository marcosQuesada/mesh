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
	node  *Node
	event Status // ??
}

type Orchestrator struct {
	clientHandler ClientHandler
	members       map[*Node]bool
	clients       map[*Node]*Client //NOOOOOOO
	inChan        chan memberUpdate
	mainChan      chan Message // Used as aggregated channel from Client Peers
	exitChan      chan bool
}

func StartOrchestrator(members map[*Node]bool) *Orchestrator {
	return &Orchestrator{
		clientHandler: DefaultClientHandler(),
		members:       members,
		clients:       make(map[*Node]*Client, 0),
		inChan:        make(chan memberUpdate, 0),
		exitChan:      make(chan bool, 0),

		mainChan: make(chan Message, 0),
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
				o.members[m.node] = true
				if o.State() {
					log.Println("Cluster Completed!")
				}
			case ClientStatusError:
				o.members[m.node] = false
				o.clients[m.node] = nil
			}
		case <-o.exitChan:
			return
		}
	}

}

func (o *Orchestrator) State() bool {
	for _, s := range o.members {
		if !s {
			return false
		}
	}

	return true
}

func (o *Orchestrator) bootClients() {
	var done sync.WaitGroup
	for node, inService := range o.members {
		if !inService {
			var c *Client
			done.Add(1)
			go func(n *Node) {
				//Blocking call, wait until connection success
				log.Println("Starting Dial Client on Node ", n)
				c = StartDialClient(n)
				go c.Run()
				o.clients[n] = c
				done.Done()

				//Say Hello and wait response
				c.SayHello()
				//Blocking call again until response
				rsp := <-c.ReceiveChan()
				switch rsp.(type) {
				case *Welcome:
					o.clientHandler.Accept(c)
					log.Println("Client has received Welcome from", n, rsp.(*Welcome))
					o.inChan <- memberUpdate{
						node:  n,
						event: ClientStatusConnected,
					}

					//aggregate receiveChan to mainChan
					o.aggregate(c.ReceiveChan())
					log.Println("Client Achieved: ", n)
				case *Abort:
					log.Println("response Abort ", rsp.(*Abort))
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

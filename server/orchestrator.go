package server

import (
	"fmt"
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
	clients       map[*Node]*Client
	inChan        chan memberUpdate
	exitChan      chan bool

	mainChan chan Message // Used as aggregates channel
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
	defer fmt.Println("Exiting Orchestrator Run")
	go o.consumeMainChannel()
	o.bootClients()
	//handle updates from members
	for {
		select {
		case m := <-o.inChan:
			fmt.Println("Handle Update ", m)
			switch m.event {
			case ClientStatusConnected:
				fmt.Println("Event Status Connected ", m)
				o.members[m.node] = true
			case ClientStatusError:
				fmt.Println("Event Status Error ", m)
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
			go func() {
				//Blocking call, wait until connection success
				c = StartDialClient(node)
				go c.Run()
				o.clients[node] = c
				done.Done()

				//Say Hello and wait response
				c.SayHello()
				//Blocking call again until response
				rsp := <-c.ReceiveChan()
				switch rsp.(type) {
				case *Welcome:
					o.clientHandler.Accept(c)
					fmt.Println("Client has received Welcome", rsp.(*Welcome))
					o.inChan <- memberUpdate{
						node:  node,
						event: ClientStatusConnected,
					}

					//aggregate receiveChan to mainChan
					o.aggregate(c.ReceiveChan())
					fmt.Println("Client Achieved: ", node)
				case *Abort:
					fmt.Println("response Abort ", rsp.(*Abort))
					o.inChan <- memberUpdate{
						node:  node,
						event: ClientStatusError,
					}
				default:
					fmt.Println("Unexpected type On response ")
				}
			}()
		}
	}
	done.Wait()

	fmt.Println("Done Initialicing")
}

func (o *Orchestrator) consumeMainChannel() {
	fmt.Println("Start consume Main Channel")
	for {
		select {
		case m := <-o.mainChan:
			fmt.Println("Received Message on Main Channel ", m)
		}
	}
}

func (o *Orchestrator) aggregate(c chan Message) {
	fmt.Println("Aggregating receive chan")
	go func() {
		for {
			select {
			case m := <-c:
				o.mainChan <- m
			}
		}
		fmt.Println("Aggregating receive chan Done")
	}()
}

//BIND MANY CHANNELS TO ONE AND HANDLE all

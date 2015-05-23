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
}

func StartOrchestrator(members map[*Node]bool) *Orchestrator {
	return &Orchestrator{
		clientHandler: DefaultClientHandler(),
		members:       members,
		clients:       make(map[*Node]*Client, 0),
		inChan:        make(chan memberUpdate, 0),
		exitChan:      make(chan bool, 0),
	}
}

func (o *Orchestrator) Run() {
	o.bootClients()
	for {
		select {
		case m := <-o.inChan:
			//handle updates from members
			fmt.Print("Handle Update ", m)
			switch m.event {
			case ClientStatusConnected:
				fmt.Print("Event Status Connected ", m)
				o.members[m.node] = true
			case ClientStatusError:
				fmt.Print("Event Status Error ", m)
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
				c = StartDialClient(node)
				go c.Run()
				o.clients[node] = c
				done.Done()
				m := memberUpdate{
					node:  node,
					event: ClientStatusConnected,
				}
				o.inChan <- m
				o.clientHandler.Accept(c)
				fmt.Println("Client Achieved: ", node)
			}()
		}

	}
	done.Wait()

	fmt.Println("Are: ", o.clients)
}

//BIND MANY CHANNELS TO ONE AND HANDLE all

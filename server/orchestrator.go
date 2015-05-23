package server

import (
	"fmt"
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
	clientHandler ClientHandler
	members       map[Node]bool
	clients       map[Node]*Client
	inChan        chan memberUpdate
	exitChan      chan bool
}

func StartOrchestrator(members map[Node]bool) *Orchestrator {
	return &Orchestrator{
		clientHandler: DefaultClientHandler(),
		members:       members,
		inChan:        make(chan memberUpdate, 0),
		exitChan:      make(chan bool, 0),
	}
}

func (o *Orchestrator) Run() {
	o.initClients()
	for {
		select {
		case m := <-o.inChan:
			//handle updates from members
			fmt.Print("Handle Update ", m)
			switch m.event {
			case ClientStatusConnected:
				o.members[m.node] = true
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

func (o *Orchestrator) initClients() {
	for n, _ := range o.members {
		fmt.Println("Starting client Peer ", n.String())

		//THIS GOES TO CLIENT!!!
		go func() {
			h := &handleClient{node: n, exit: make(chan bool, 0)}
			c := h.connect()

			//To PeerHandler!!!!
			o.clients[n] = c
			fmt.Println("Client Achieved: ", n)

			m := memberUpdate{
				node:  n,
				event: ClientStatusConnected,
			}
			o.inChan <- m

			return
		}()
	}
}

type handleClient struct {
	node Node
	exit chan bool
}

func (h *handleClient) connect() *Client {
	result := make(chan *Client, 0)
	go func() {
		for {
			/*			c, err := StartDialClient(h.node)
						if err == nil {
							result <- c
						}

						fmt.Println("Error ", err)
						time.Sleep(time.Second * 1)*/
		}
	}()

	r := <-result
	return r
}

//BIND MANY CHANNELS TO ONE AND HANDLE all

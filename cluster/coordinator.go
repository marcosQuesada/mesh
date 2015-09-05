package cluster

import (
	"log"

	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"time"
)

// Coordinator takes cares on all cluster related tasks
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

type Coordinator struct {
	from      n.Node
	members   map[string]n.Node
	connected map[string]bool
	exitChan  chan bool
	status    message.Status
}

func StartCoordinator(from n.Node, members map[string]n.Node) *Coordinator {
	return &Coordinator{
		from:      from,
		members:   members,
		connected: make(map[string]bool, len(members)-1),
		exitChan:  make(chan bool, 0),
		status:    ClusterStatusStarting,
	}
}

func (c *Coordinator) Run() {
	for {
		select {
		case <-c.exitChan:
			return
		}
	}
}

func (c *Coordinator) RunStatus() {
	for {
		select {
		case <- c.exitChan:
			return
		default:
			time.Sleep(time.Second * 1)
			complete := true

			if len(c.connected) == 0 {
				complete = false
			}

			for _, v := range c.connected {
				if !v {
					complete = false
				}
			}

			if complete {
				if c.status != ClusterStatusInService {
					c.status = ClusterStatusInService
					log.Println("+++++++++++++++++++++Cluster Complete!!!")
				}
			}

			if c.status == ClusterStatusInService && !complete{
				c.status = ClusterStatusDegraded
				log.Println("+++++++++++++++++++++Cluster Degraded!!!")
			}

		}
	}
}

func (c *Coordinator) Exit() {
	close(c.exitChan)
}

func (c *Coordinator) OnPeerConnectedEvent(e dispatcher.Event) {
	event := e.(*peer.OnPeerConnectedEvent)
	c.members[event.Node.String()] = event.Node
	c.connected[event.Node.String()] = true
	log.Println("OnPeerConnectedEvent, adding peer", event.Node.String(), "mode:", event.Mode)
}

func (c *Coordinator) OnPeerDisconnected(e dispatcher.Event) {
	event := e.(*peer.OnPeerDisconnectedEvent)
	c.members[event.Node.String()] = event.Node

	c.connected[event.Node.String()] = false
	log.Println("OnPeerDisconnectedEvent, removing peer", event.Node.String())
}

func (c *Coordinator) OnPeerAborted(e dispatcher.Event) {
	n := e.(*peer.OnPeerAbortedEvent)
	log.Println("OnPeerAbortedEvent", n.Node.String())
}

func (c *Coordinator) OnPeerErrored(e dispatcher.Event) {
	n := e.(*peer.OnPeerErroredEvent)
	c.connected[n.Node.String()] = false
	log.Println("OnPeerErroredEvent", n.Node.String())
}

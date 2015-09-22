package cluster

import (
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/router/handler"
	"log"
	"reflect"
	"time"
)

// Coordinator takes cares on all cluster related tasks
// Peer election & registration (could be delegated to PeerHandler)
// Cluster member definition & cluster status
// Leader election
// Execute Pool mechanisms and consesus resolution


type Coordinator struct {
	from      n.Node
	members   map[string]n.Node
	connected map[string]bool
	exitChan  chan bool
	sndChan   chan handler.Request
	status    message.Status
}

func StartCoordinator(from n.Node, members map[string]n.Node) *Coordinator {
	return &Coordinator{
		from:      from,
		members:   members,
		connected: make(map[string]bool, len(members)-1),
		exitChan:  make(chan bool, 0),
		sndChan:   make(chan handler.Request, 1),
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
		case <-c.exitChan:
			return
		default:
			time.Sleep(time.Second * 1)

			if c.isComplete() {
				if c.status != ClusterStatusInService {
					c.status = ClusterStatusInService
					log.Println("+++++++++++++++++++++Cluster Complete!!!")
					result := c.PoolRequest()
					for node, res := range result {
						log.Println("XXXXXX Result from ", node, reflect.TypeOf(res).String())
					}
				}
			}

			if c.status == ClusterStatusInService && !c.isComplete() {
				c.status = ClusterStatusDegraded
				log.Println("+++++++++++++++++++++Cluster Degraded!!!")
			}

		}
	}
}

func (c *Coordinator) PoolRequest() map[string] message.Message{
	time.Sleep(time.Second * 1)
	responseChn := make(chan message.Message)
	response := make(map[string]message.Message, len(c.connected))
	for nodeString, _ := range c.connected {
		node := c.members[nodeString]
		msg := &message.Command{Id: message.NewId(), From: c.from, To: node}

		//fire request to router
		c.sndChan <- handler.Request{responseChn, msg}

		//store response
		response[nodeString] = <-responseChn
	}

	return response
}

func (c *Coordinator) Exit() {
	close(c.exitChan)
}

func (c *Coordinator) SndChan() chan handler.Request {
	return c.sndChan
}

func(c *Coordinator) isComplete() bool {
	if len(c.connected) != len(c.members) -1 {
		return false
	}

	for _, v := range c.connected {
		if !v {
			return false
		}
	}

	return true
}
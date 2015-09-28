package cluster

import (
	"github.com/marcosQuesada/mesh/cluster/raft"
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
const (
	ClusterStatusStarting  = message.Status("starting")
	ClusterStatusInService = message.Status("in service")
	ClusterStatusDegraded  = message.Status("degraded")
	ClusterStatusExit      = message.Status("exit")
)

type Manager interface {
	Run()
	Ready()chan bool
	Request() chan interface{}
	Response() chan interface{}
}

type Coordinator struct {
	manager   Manager
	from      n.Node
	members   map[string]n.Node
	connected map[string]bool
	sndChan   chan handler.Request
	exitChan  chan bool
	status    message.Status
}

func Start(from n.Node, members map[string]n.Node) *Coordinator {
	log.Println("Starting coordinator on Node %s members: ", from.String(), members)

	//@TODO: Remove this!!!
	mates := make([]n.Node, len(members))
	for _, v := range members {
		mates = append(mates, v)
	}

	r := raft.New(from, mates)
	c := &Coordinator{
		manager:   r,
		from:      from,
		members:   members,
		connected: make(map[string]bool, len(members)-1),
		exitChan:  make(chan bool, 0),
		sndChan:   make(chan handler.Request, 1),
		status:    ClusterStatusStarting,
	}

	c.addSender(r.Request(), r.Response())
	go c.Run()

	return c
}

func (c *Coordinator) Run() {
	complete := make(chan bool,0)
	for {
		select {
		case <-c.exitChan:
			return
		case <- complete:
			log.Println("XXXXXX Complete, boot Manageer")
			c.manager.Run()

		case <- c.manager.Ready():
			log.Println("XXXXXX Manager Ready")

			// Example transaction
			result := c.PoolRequest("fooCommand")
			for node, res := range result {
				log.Println("XXXXXX Result from ", node, reflect.TypeOf(res).String())
			}

		default:
			go c.waitUntilComplete(complete)
		}
	}
}

func (c *Coordinator) addSender(s chan interface{}, r chan interface{}) {
	go func(){
		for{
			select{
			case msg := <-s:
				log.Println("On Add sender msg ", msg)
				r <- c.PoolRequest(msg)
			}
		}
	}()
}

func (c *Coordinator) waitUntilComplete(done chan bool) {
	for {
		time.Sleep(time.Second * 1)

		if c.isComplete() {
			if c.status != ClusterStatusInService {
				c.status = ClusterStatusInService
				log.Println("+++++++++++++++++++++Cluster Complete!!!")

				done <- true

				return
			}
		}
	}
}

func (c *Coordinator) PoolRequest(cmd interface{}) map[string]message.Message {
	time.Sleep(time.Second * 1)
	response := make(map[string]message.Message, len(c.connected))
	for nodeString, _ := range c.connected {
		node := c.members[nodeString]
		msg := &message.Command{Id: message.NewId(), From: c.from, To: node, Command: cmd}

		//fire request to router && store response
		response[nodeString] = <-c.sendRequest(msg)
	}

	return response
}

func (c *Coordinator) sendRequest(msg message.Message) chan message.Message {
	responseChn := make(chan message.Message)
	c.sndChan <- handler.Request{responseChn, msg}

	return responseChn
}

func (c *Coordinator) Exit() {
	close(c.exitChan)
}

func (c *Coordinator) SndChan() chan handler.Request {
	return c.sndChan
}

func (c *Coordinator) isComplete() bool {
	if len(c.connected) != len(c.members)-1 {
		return false
	}

	for _, v := range c.connected {
		if !v {
			return false
		}
	}

	return true
}

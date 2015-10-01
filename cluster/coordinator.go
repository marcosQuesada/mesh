package cluster

import (
	"github.com/marcosQuesada/mesh/cluster/raft"
	"github.com/marcosQuesada/mesh/message"
	n "github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/router/handler"
	"log"
	"reflect"
	"sync"
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
	Ready() chan bool
	Request() chan interface{}
	Response() chan interface{}
	Exit()
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

	r := raft.New(from, members)
	c := &Coordinator{
		manager:   r,
		from:      from,
		members:   members,
		connected: make(map[string]bool, len(members)-1),
		exitChan:  make(chan bool, 0),
		sndChan:   make(chan handler.Request, 10),
		status:    ClusterStatusStarting,
	}

	//enable manager to send and receive requests
	c.addSender(r.Request(), r.Response())
	go c.Run()

	return c
}

func (c *Coordinator) Run() {
	var runOnce sync.Once
	complete := make(chan bool, 0)
	go c.waitUntilComplete(complete)

	for {
		select {
		case <-c.exitChan:
			log.Println("Exit")
			return
		case <-complete:
			runOnce.Do(func() {
				log.Println("Boot manager just once")
				go c.manager.Run()
			})
		case <-c.manager.Ready():
			log.Println("XXXXXXXXXXXXXXXXXXXXXXX CLUSTER IN SERVICE! XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
		}
	}
}

func (c *Coordinator) Exit() {
	close(c.exitChan)
}

func (c *Coordinator) SndChan() chan handler.Request {
	return c.sndChan
}

func (c *Coordinator) addSender(sendChan chan interface{}, rcvChan chan interface{}) {
	go func() {
		for {
			select {
			case msg := <-sendChan:
				go func() {
					rcvChan <- c.poolRequest(msg)
				}()
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

func (c *Coordinator) poolRequest(cmd interface{}) raft.PoolResult {
	response := make(raft.PoolResult, len(c.members))
	result := make(chan message.Message, len(c.members))

	var wg sync.WaitGroup
	for nodeString, connected := range c.connected {
		if connected {
			wg.Add(1)
			go func(n string) {
				node := c.members[n]
				msg := message.Command{Id: message.NewId(), From: c.from, To: node, Command: cmd}
				result <- <-c.sendRequest(msg)
				wg.Done()
			}(nodeString)
		}
	}

	wg.Wait()
	close(result)

	for item := range result {
		rsp, ok := item.(*message.Response)
		if !ok {
			log.Println("------------------ PoolRequest unexpected type ", reflect.TypeOf(rsp).String())
			continue
		}
		response[rsp.From.String()] = rsp.Result.(string)
	}

	return response
}

func (c *Coordinator) sendRequest(msg message.Message) chan message.Message {
	responseChn := make(chan message.Message, 1)
	c.sndChan <- handler.Request{responseChn, msg}

	return responseChn
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

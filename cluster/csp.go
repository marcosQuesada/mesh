package cluster

import (
	"fmt"
	m "github.com/marcosQuesada/mesh/message"
	"log"
	"time"
)

type csp struct {
	rcvChan chan m.Message
	sndChan chan m.Message
	stop    bool
}

func Start() *csp {
	csp := &csp{
		rcvChan: make(chan m.Message),
		sndChan: make(chan m.Message),
		stop:    false,
	}

	go csp.run()

	return csp
}

func (c *csp) Send(msg m.Message) {
	c.rcvChan <- msg
}

func (c *csp) Receive() m.Message {
	timeout := time.NewTimer(time.Second * 1)
	select {
	case response := <-c.sndChan:
		return response
	case <-timeout.C:
		fmt.Printf("Receive Timeout \n")
		return nil
	}
}

func (c *csp) run() {
	for !c.stop {
		select {
		case msg := <-c.rcvChan:
			switch msg.MessageType() {
			case m.HELLO:
				cmd := msg.(*m.Hello)
				log.Println("HELLO Message", cmd.Id)
			case m.WELCOME:
				cmd := msg.(*m.Welcome)
				log.Println("Welcome Message", cmd.Id)
			case m.ABORT:
				cmd := msg.(*m.Abort)
				log.Println("Abort Message", cmd.Id)
			}
		}
	}
}

package server

import (
	"fmt"
	"time"
)

type coordinator struct {
	rcvChan chan Message
	sndChan chan Message
	stop    bool
}

func Start() *coordinator {
	coordinator := &coordinator{
		rcvChan: make(chan Message),
		sndChan: make(chan Message),
		stop:    false,
	}

	go coordinator.run()

	return coordinator
}

func (c *coordinator) Send(m Message) {
	c.rcvChan <- m
}

func (c *coordinator) Receive() Message {
	timeout := time.NewTimer(time.Second * 1)
	select {
	case response := <-c.sndChan:
		return response
	case <-timeout.C:
		fmt.Printf("Receive Timeout \n")
		return nil
	}
}

func (c *coordinator) run() {
	for !c.stop {
		select {
		case msg := <-c.rcvChan:
			switch msg.MessageType() {
			case HELLO:
				cmd := msg.(*Hello)
				fmt.Println("HELLO Message", cmd.Id)
			case WELCOME:
				cmd := msg.(*Welcome)
				fmt.Println("Welcome Message", cmd.Id)
			case ABORT:
				cmd := msg.(*Abort)
				fmt.Println("Abort Message", cmd.Id)
			}
		}
	}
}

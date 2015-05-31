package server

import (
	"net"
	"testing"
)

func TestClientMessagingUnderPipes(t *testing.T) {
	a, b := net.Pipe()

	c1 := &Client{
		Peer:     NewJSONSocketPeer(a),
		from:     Node{Host: "192.168.1.1", Port: 8000},
		node:     Node{Host: "foo", Port: 5678},
		message:  make(chan Message, 0),
		exitChan: make(chan bool),
	}
	go c1.Run()

	c2 := &Client{
		Peer:     NewJSONSocketPeer(b),
		from:     Node{Host: "192.168.1.10", Port: 8000},
		node:     Node{Host: "bar", Port: 5678},
		message:  make(chan Message, 0),
		exitChan: make(chan bool),
	}
	go c2.Run()

	resChan := make(chan Message, 2)
	doneChan := make(chan struct{})
	go func() {
		for {
			select {
			case m := <-c1.ReceiveChan():
				msg := m.(*Hello)
				msg.Id = 1
				resChan <- msg
			case m := <-c2.ReceiveChan():
				msg := m.(*Hello)
				msg.Id = 2
				resChan <- msg
			case <-doneChan:
				close(resChan)
				return
			}
		}
		return
	}()

	c1.SayHello()
	c2.SayHello()

	close(doneChan)

	r := make([]Message, 0)
	for k := range resChan {
		r = append(r, k)
	}

	if len(r) != 2 {
		t.Error("Unexpected response size", r)
	}

	h1 := r[0].(*Hello)
	if h1.Id != 2 {
		t.Error("Unexpected First Id received ", h1)
	}
	h2 := r[1].(*Hello)
	if h2.Id != 1 {
		t.Error("Unexpected First Id received ", h2)
	}
	c1.Exit()
	c2.Exit()
}

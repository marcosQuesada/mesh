package server

import (
	"fmt"
	"net"
	"testing"
)

func TestClientMessagingUnderPipes(t *testing.T) {
	a, b := net.Pipe()

	n1 := &Node{host: "foo", port: 5678}
	c1 := &Client{
		Peer:     NewJSONSocketPeer(a),
		node:     n1,
		message:  make(chan Message, 0),
		exitChan: make(chan bool),
	}
	go c1.Run()

	n2 := &Node{host: "bar", port: 5678}
	c2 := &Client{
		Peer:     NewJSONSocketPeer(b),
		node:     n2,
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

	msg := Hello{
		Id:      0,
		Details: map[string]interface{}{"foo": "bar"},
	}
	c1.Send(msg)
	c2.Send(msg)

	c1.Exit()
	c2.Exit()
	close(doneChan)

	r := make([]Message, 0)
	for k := range resChan {
		r = append(r, k)
		fmt.Println("k ", k)
	}

	if len(r) != 2 {
		t.Error("Unexpected response size", r)
	}

	h1 := r[0].(*Hello)
	if h1.Id != 2 {
		t.Error("Unexpected First Id received ", h1.Id)
	}
	h2 := r[1].(*Hello)
	if h2.Id != 1 {
		t.Error("Unexpected First Id received ", h1.Id)
	}
}

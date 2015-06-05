package client

import (
	m "github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"net"
	"testing"
	"time"
)

func TestClientMessagingUnderPipes(t *testing.T) {
	a, b := net.Pipe()

	c1 := &Client{
		Peer:        NewJSONSocketPeer(a),
		from:        node.Node{Host: "192.168.1.1", Port: 8000},
		node:        node.Node{Host: "foo", Port: 5678},
		messageChan: make(chan m.Message, 0),
		exitChan:    make(chan bool),
	}
	c1.Run()

	c2 := &Client{
		Peer:        NewJSONSocketPeer(b),
		from:        node.Node{Host: "192.168.1.10", Port: 8000},
		node:        node.Node{Host: "bar", Port: 5678},
		messageChan: make(chan m.Message, 0),
		exitChan:    make(chan bool),
	}
	c2.Run()

	resChan := make(chan m.Message, 2)
	doneChan := make(chan struct{})
	go func() {
		for {
			select {
			case r := <-c1.ReceiveChan():
				msg := r.(*m.Hello)
				msg.Id = 1
				resChan <- msg
			case r := <-c2.ReceiveChan():
				msg := r.(*m.Hello)
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

	time.Sleep(time.Millisecond * 100)

	close(doneChan)

	r := make([]m.Message, 0)
	for k := range resChan {
		r = append(r, k)
	}

	if len(r) != 2 {
		t.Error("Unexpected response size", r)
		t.Fail()
	}

	h1, ok := r[0].(*m.Hello)
	if !ok {
		t.Error("Error Casting to Hello ", h1)
	}

	if h1.Id != 2 {
		t.Error("Unexpected First Id received ", h1)
	}

	h2 := r[1].(*m.Hello)
	if h2.Id != 1 {
		t.Error("Unexpected First Id received ", h2)
	}

	c1.Exit()
	c2.Exit()
}

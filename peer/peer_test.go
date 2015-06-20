package peer

import (
	"fmt"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestPeerMessagingUnderPipes(t *testing.T) {
	a, b := net.Pipe()

	c1 := &Peer{
		Link:        NewJSONSocketLink(a),
		from:        node.Node{Host: "foo"},
		to:          node.Node{Host: "bar"},
		dataChan:    make(chan message.Message, 0),
		messageChan: make(chan message.Message, 0),
		exitChan:    make(chan bool),
		doneChan:    make(chan bool),
		pingChan:    make(chan message.Message, 0),
		pongChan:    make(chan message.Message, 0),
		mode:        "pipe",
	}
	go c1.Run()

	c2 := &Peer{
		Link:        NewJSONSocketLink(b),
		from:        node.Node{Host: "bar"},
		to:          node.Node{Host: "foo"},
		dataChan:    make(chan message.Message, 0),
		messageChan: make(chan message.Message, 0),
		exitChan:    make(chan bool),
		doneChan:    make(chan bool),
		pingChan:    make(chan message.Message, 0),
		pongChan:    make(chan message.Message, 0),
		mode:        "pipe",
	}
	go c2.Run()

	resChan := make(chan message.Message, 2)
	doneChan := make(chan struct{})
	go func() {
		for {
			select {
			case r := <-c1.ReceiveChan():
				msg := r.(*message.Hello)
				msg.Id = 1
				resChan <- msg
			case r := <-c2.ReceiveChan():
				msg := r.(*message.Hello)
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

	r := make([]message.Message, 0)
	for k := range resChan {
		r = append(r, k)
	}
	if len(r) != 2 {
		t.Error("Unexpected response size", r)
		t.Fail()
	}
	h1, ok := r[0].(*message.Hello)
	if !ok {
		t.Error("Error Casting to Hello ", h1)
	}

	if h1.Id != 2 {
		t.Error("Unexpected First Id received ", h1)
	}

	h2 := r[1].(*message.Hello)
	if h2.Id != 1 {
		t.Error("Unexpected First Id received ", h2)
	}

	c1.Exit()
	c2.Exit()
}

func TestBasicNopPeerTest(t *testing.T) {
	ch := make(chan message.Message, 10)
	pCh := make(chan message.Message, 10)
	fkc := &NopPeer{"localhost", 9000, ch, pCh}

	msg := message.Hello{
		Id:      999,
		From:    node.Node{"localhost", 9000},
		Details: map[string]interface{}{"foo": "bar"},
	}

	fkc.Send(msg)

	rcvMsg := <-fkc.ReceiveChan()
	if !reflect.DeepEqual(msg, rcvMsg) {
		t.Errorf("Expected %s, got %s", msg, rcvMsg)
	}
}

func TestBasicPingPongChannel(t *testing.T) {
	a, b := net.Pipe()

	c1 := NewAcceptor(a, node.Node{})
	go c1.Run()

	c2 := NewAcceptor(b, node.Node{})
	go c2.Run()

	resChan := make(chan message.Message, 4)
	doneChan := make(chan struct{})
	go func() {
		for {
			select {
			case r := <-c1.PingChan():
				msg := r.(*message.Ping)
				resChan <- msg
			case r := <-c1.PongChan():
				msg := r.(*message.Pong)
				resChan <- msg
			case r := <-c2.PingChan():
				msg := r.(*message.Ping)
				resChan <- msg
			case r := <-c2.PongChan():
				msg := r.(*message.Pong)
				resChan <- msg
			case <-doneChan:
				close(resChan)
				return
			}
		}
		return
	}()

	c1.Send(message.Ping{})
	c2.Send(message.Ping{})
	c1.Send(message.Pong{})
	c2.Send(message.Pong{})
	time.Sleep(time.Millisecond * 100)

	close(doneChan)
	if len(resChan) != 4 {
		t.Error("Unexpected result chan", len(resChan))
	}
	for k := range resChan {
		fmt.Println("Result Channel data: ", k)
	}

	c1.Exit()
	c2.Exit()
}

func TestHandlePeerUsingPipes(t *testing.T) {
	nodeA := node.Node{Host: "A", Port: 1}
	nodeB := node.Node{Host: "B", Port: 2}
	a, b := net.Pipe()

	c1 := NewAcceptor(a, nodeA)
	c1.Identify(nodeB)
	go c1.Run()

	c1Mirror := NewAcceptor(b, nodeB)
	c1Mirror.Identify(nodeA)
	go c1Mirror.Run()

	fmt.Println("Done")
	go func() {
		for {
			select {
			case msg, open := <-c1.ReceiveChan():
				if !open {
					return
				}

				if msg.MessageType() != 0 {
					t.Error("Unexpected message type")
				}
			}
		}
	}()

	c1Mirror.SayHello()

	c1.Exit()
	c1Mirror.Exit()
}

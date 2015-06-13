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
		from:        node.Node{},
		to:          node.Node{},
		messageChan: make(chan message.Message, 0),
		exitChan:    make(chan bool),
		mode:        "pipe",
	}
	c1.Run()

	c2 := &Peer{
		Link:        NewJSONSocketLink(b),
		from:        node.Node{},
		to:          node.Node{},
		messageChan: make(chan message.Message, 0),
		exitChan:    make(chan bool),
		mode:        "pipe",
	}
	c2.Run()

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

	c1 := &Peer{
		Link:        NewJSONSocketLink(a),
		from:        node.Node{},
		to:          node.Node{},
		messageChan: make(chan message.Message, 0),
		exitChan:    make(chan bool),
		mode:        "pipe",
	}
	c1.Run()

	c2 := &Peer{
		Link:        NewJSONSocketLink(b),
		from:        node.Node{},
		to:          node.Node{},
		messageChan: make(chan message.Message, 0),
		exitChan:    make(chan bool),
		mode:        "pipe",
	}
	c2.Run()

	resChan := make(chan message.Message, 2)
	doneChan := make(chan struct{})
	go func() {
		for {
			select {
			case r := <-c1.ReceiveChan():
				msg := r.(*message.Ping)
				msg.Id = 1
				resChan <- msg
			case r := <-c2.ReceiveChan():
				msg := r.(*message.Ping)
				msg.Id = 2
				resChan <- msg
			case <-doneChan:
				close(resChan)
				return
			}
		}
		return
	}()

	ping := &message.Ping{}
	c2.Send(ping)

	time.Sleep(time.Millisecond * 100)

	close(doneChan)
	r := <-resChan

	fmt.Println("Reschan is ", r)

	c1.Exit()
	c2.Exit()
}

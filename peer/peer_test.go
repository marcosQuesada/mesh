package peer

import (
	"fmt"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
)

func TestPeerMessagingUnderPipes(t *testing.T) {
	a, b := net.Pipe()

	c1 := &Peer{
		Link:         NewJSONSocketLink(a),
		from:         node.Node{Host: "foo"},
		to:           node.Node{Host: "bar"},
		dataChan:     make(chan message.Message, 10),
		sendChan:     make(chan message.Message, 10),
		messageChan:  make(chan message.Message, 10),
		exitChan:     make(chan bool),
		doneChan:     make(chan bool),
		rstWatchChan: make(chan bool, 100),
		mode:         "pipe",
	}
	go c1.Run()

	c1Mirror := &Peer{
		Link:         NewJSONSocketLink(b),
		from:         node.Node{Host: "bar"},
		to:           node.Node{Host: "foo"},
		dataChan:     make(chan message.Message, 10),
		sendChan:     make(chan message.Message, 10),
		messageChan:  make(chan message.Message, 10),
		exitChan:     make(chan bool),
		doneChan:     make(chan bool),
		rstWatchChan: make(chan bool, 100),
		mode:         "pipe",
	}
	go c1Mirror.Run()

	var wg sync.WaitGroup

	resChan := make(chan message.Message, 2)
	doneChan := make(chan struct{})
	wg.Add(1)
	go func() {
		for {
			select {
			case r := <-c1.ReceiveChan():
				msg := r.(*message.Hello)
				msg.Id = message.NewId()
				resChan <- msg
			case r := <-c1Mirror.ReceiveChan():
				msg := r.(*message.Hello)
				msg.Id = message.NewId()
				resChan <- msg
			case <-c1.ResetWatcherChan():
				continue
			case <-c1Mirror.ResetWatcherChan():
				continue
			case <-doneChan:
				close(resChan)
				wg.Done()
				return
			}
		}
		return
	}()

	c1.SayHello()
	c1Mirror.SayHello()

	time.Sleep(time.Millisecond * 100)

	close(doneChan)
	wg.Wait()
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

	h2, ok := r[1].(*message.Hello)
	if !ok {
		t.Error("Error Casting to Hello ", h2)
	}

	c1.Exit()
	c1Mirror.Exit()
}

func TestBasicNopPeerTest(t *testing.T) {
	ch := make(chan message.Message, 10)
	fkc := &NopPeer{"localhost", 9000, ch}

	msg := message.Hello{
		Id:      message.NewId(),
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

	c1Mirror := NewAcceptor(b, node.Node{})
	go c1Mirror.Run()

	resChan := make(chan message.Message, 6)
	doneChan := make(chan bool, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer close(resChan)
		for {
			select {
			case msg, open := <-c1.ReceiveChan():
				if !open {
					fmt.Println("closed Chan")
					return
				}
				resChan <- msg

			case msg, open := <-c1Mirror.ReceiveChan():
				if !open {
					fmt.Println("closed Chan")
					return
				}
				resChan <- msg
				c1Mirror.Commit(msg)
			case <-doneChan:

				wg.Done()

				return
			case <-c1.ResetWatcherChan():
				continue
			case <-c1Mirror.ResetWatcherChan():
				continue
			}

		}
		return
	}()

	c1.Commit(message.Ping{})
	c1Mirror.Commit(message.Ping{})
	c1.Commit(message.Pong{})
	c1Mirror.Commit(message.Pong{})
	time.Sleep(time.Millisecond * 100)

	close(doneChan)
	wg.Wait()

	if len(resChan) != 6 {
		t.Error("Unexpected result chan", len(resChan))
	}
	if cap(resChan) != 6 {
		t.Error("Unexpected result chan", len(resChan))
	}

	total := 0
	for _ = range resChan {
		total++
	}

	if total != 6 {
		t.Error("Wrong size", total)
	}

	c1.Exit()
	c1Mirror.Exit()
}

func TestPeersUsingPipes(t *testing.T) {
	nodeA := node.Node{Host: "A", Port: 1}
	nodeB := node.Node{Host: "B", Port: 2}
	a, b := net.Pipe()

	c1 := NewAcceptor(a, nodeA)
	c1.Identify(nodeB)
	go c1.Run()

	c1Mirror := NewAcceptor(b, nodeB)
	c1Mirror.Identify(nodeA)
	go c1Mirror.Run()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		total := 0
		for {
			select {
			case msg, open := <-c1.ReceiveChan():
				if !open {
					return
				}

				if msg.MessageType() != 0 {
					t.Error("Unexpected message type")
				}
				c1.Commit(&message.Abort{Id: msg.(*message.Hello).Id, From: msg.(*message.Hello).From})

				if total == 100 {
					wg.Done()
					return
				}
				total++
			case msg, open := <-c1Mirror.ReceiveChan():
				if !open {
					return
				}

				if msg.MessageType() != 2 {
					t.Error("Unexpected message type")
				}
				c1Mirror.SayHello()

				if total == 100 {
					wg.Done()
					return
				}
				total++
			case <-c1.ResetWatcherChan():
				continue
			case <-c1Mirror.ResetWatcherChan():
				continue
			}
		}
	}()

	c1Mirror.SayHello()

	wg.Wait()
	c1.Exit()
	c1Mirror.Exit()
}

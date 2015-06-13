package watch

import (
	"fmt"
	"github.com/marcosQuesada/mesh/dispatcher"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"net"
	"reflect"
	"testing"
	//"time"
	"sync"
)

func TestBasicPingPingOverFakeNopPeers(t *testing.T) {
	fakePeerA := &FakePeerA{
		peer.NopPeer{},
	}
	fakePeerB := &FakePeerA{
		peer.NopPeer{},
	}

	msg := message.Ping{
		Id:   999,
		From: node.Node{"localhost", 9000},
	}

	fakePeerA.Send(msg)

	rcvMsg := <-fakePeerB.PingChan()
	if !reflect.DeepEqual(msg, rcvMsg) {
		t.Errorf("Expected %s, got %s", msg, rcvMsg)
	}

	msgPong := message.Pong{
		Id:   999,
		From: node.Node{"localhost", 9000},
	}
	fakePeerB.Send(msgPong)

	rcvMsg = <-fakePeerA.PingChan()
	if !reflect.DeepEqual(msgPong, rcvMsg) {
		t.Errorf("Expected %s, got %s", msg, rcvMsg)
	}
}

func TestBasicWatchOverFakeNopPeers(t *testing.T) {
	fakePeerA := &FakePeerA{
		peer.NopPeer{},
	}

	fakePeerB := &FakePeerA{
		peer.NopPeer{},
	}

	fakePeerA.Run()
	fakePeerB.Run()

	evCh := make(chan dispatcher.Event, 0)
	w := New(evCh, 1)

	w.Watch(fakePeerA)
	w.Watch(fakePeerB)
	close(evCh)

	/*	for {
		select {
		case m := <-fakePeerA.PingChan():
		case m := <-fakePeerB.PingChan():

		case <-exit:
			return
		}
	}*/
	w.Exit()
}

var fkCh1 = make(chan message.Message, 10)
var fkCh2 = make(chan message.Message, 10)

type FakePeerA struct {
	peer.NopPeer
}

func (f *FakePeerA) Send(m message.Message) error {
	fkCh1 <- m
	return nil
}
func (f *FakePeerA) PingChan() chan message.Message {
	return fkCh2
}
func (f *FakePeerA) PongChan() chan message.Message {
	return fkCh2
}

func TestBasicPingPongOverPipesChannel(t *testing.T) {
	nodeA := node.Node{Host: "testA", Port: 1}
	nodeB := node.Node{Host: "testB", Port: 2}
	a, b := net.Pipe()

	c1 := peer.NewAcceptor(a, nodeA)
	c1.Identify(nodeB)
	c1.Run()

	c2 := peer.NewAcceptor(b, nodeB)
	c2.Identify(nodeA)
	c2.Run()

	var wg sync.WaitGroup

	total := 10
	last := 0
	doneChan := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case r := <-c2.PingChan():
				ping := r.(*message.Ping)
				pong := &message.Pong{Id: ping.Id, From: ping.To, To: ping.From}
				c2.Send(pong)
				last = pong.Id
				if last == total {
					fmt.Println("Ended")
					return
				}
			case <-doneChan:
				return
			}
		}
		return
	}()

	evCh := make(chan dispatcher.Event, 0)
	defer close(evCh)

	w := New(evCh, 1)
	w.Watch(c1)

	wg.Wait()
	close(doneChan)

	if last != total {
		t.Error("Unexpected last sample", last, "as total ", total)
	}

	w.Exit()
	c1.Exit()
	c2.Exit()
}

package watch

import (
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"reflect"
	"testing"
)

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

type FakePeerB struct {
	peer.NopPeer
}

func (f *FakePeerB) Send(m message.Message) error {
	fkCh2 <- m
	return nil
}
func (f *FakePeerB) PingChan() chan message.Message {
	return fkCh1
}
func TestBasicPingPingOverFakeNopPeers(t *testing.T) {
	fakePeerA := &FakePeerA{
		peer.NopPeer{},
	}
	fakePeerB := &FakePeerB{
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

	fakePeerB := &FakePeerB{
		peer.NopPeer{},
	}

	fakePeerA.Run()
	fakePeerB.Run()

	w := New(1)

	w.Watch(fakePeerA)
	w.Watch(fakePeerB)
}

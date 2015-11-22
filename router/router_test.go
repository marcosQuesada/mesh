package router

import (
	"net"
	"testing"

	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
	"github.com/marcosQuesada/mesh/router/handler"
	"github.com/marcosQuesada/mesh/router/request"
	"sync"
)

func TestBasicRouterHandling(t *testing.T) {
	r := &defaultRouter{
		handlers: make(map[message.MsgType]handler.Handler),
		exit:     make(chan bool, 1),
		mutex: &sync.Mutex{},
	}

	msg := &message.Ping{}
	r.registerHandler(msg.MessageType(), fakePingHandler)
	c1 := &peer.NopPeer{}
	result := r.handle(c1, msg)
	if result.MessageType() != 4 {
		t.Error("unexpected response ", result)
	}
}

func TestRouterAccept(t *testing.T) {
	r := &defaultRouter{
		handlers:        make(map[message.MsgType]handler.Handler),
		notifiers:       make(map[message.MsgType]bool),
		transactionals:  make(map[message.MsgType]bool),
		exit:            make(chan bool, 1),
		requestListener: request.NewRequestListener(),
		mutex: &sync.Mutex{},
	}

	hello := &message.Hello{}
	welcome := &message.Welcome{}
	ping := &message.Ping{}
	pong := &message.Pong{}
	r.registerHandler(hello.MessageType(), fakeHelloHandler)
	r.registerHandler(ping.MessageType(), fakePingHandler)
	r.registerNotifier(hello.MessageType(), false)
	r.registerNotifier(ping.MessageType(), true)
	r.registerTransactional(hello.MessageType(), false)
	r.registerTransactional(ping.MessageType(), true)
	r.registerTransactional(pong.MessageType(), true)
	r.registerTransactional(welcome.MessageType(), false)

	nodeA := node.Node{Host: "A", Port: 1}
	nodeB := node.Node{Host: "B", Port: 2}
	a, b := net.Pipe()

	c1 := peer.NewAcceptor(a, nodeA)
	c1.Identify(nodeB)
	go c1.Run()

	c1Mirror := peer.NewAcceptor(b, nodeB)
	c1Mirror.Identify(nodeA)
	go c1Mirror.Run()

	go func() {
		for {
			select {
			case <-c1.ResetWatcherChan():
				continue
			case <-c1Mirror.ResetWatcherChan():
				continue
			}
		}
	}()

	r.Accept(c1)

	c1Mirror.SayHello()
	result := <-c1Mirror.ReceiveChan()

	if result.MessageType() != 1 {
		t.Error("Unexpected response type, expected 1 got", result.MessageType())
	}

	id := message.NewId()
	msg := message.Ping{
		Id:   id,
		From: nodeA,
		To:   nodeB,
	}
	c1Mirror.Send(msg)

	result = <-c1Mirror.ReceiveChan()
	if result.MessageType() != 4 {
		t.Error("Unexpected response type, expected 1 got", result.MessageType())
	}

	pong = result.(*message.Pong)
	if pong.Id != id {
		t.Error("Unexpected result Id")
	}
	if pong.From != nodeB {
		t.Error("Unexpected result From")
	}

	r.Exit()
	c1.Exit()
	c1Mirror.Exit()
}

func fakeHelloHandler(c peer.NodePeer, msg message.Message) (message.Message, error) {
	hello := msg.(*message.Hello)
	return &message.Welcome{hello.Id, hello.To, hello.From}, nil
}

func fakePingHandler(c peer.NodePeer, msg message.Message) (message.Message, error) {
	ping := msg.(*message.Ping)
	return &message.Pong{ping.Id, ping.To, ping.From}, nil
}

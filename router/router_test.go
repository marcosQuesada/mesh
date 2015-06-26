package router

import (
	"net"
	"testing"

	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/peer"
)

func TestBasicRouterHandling(t *testing.T) {
	r := &defaultRouter{
		handlers: make(map[message.MsgType]Handler),
		exit:     make(chan bool, 1),
	}

	msg := &message.Ping{}
	r.RegisterHandler(msg.MessageType(), fakePingHandler)

	result := r.Handle(msg)
	if result.MessageType() != 4 {
		t.Error("unexpected response ", result)
	}
}

func TestRouterAccept(t *testing.T) {
	r := &defaultRouter{
		handlers: make(map[message.MsgType]Handler),
		exit:     make(chan bool, 1),
	}

	hello := &message.Hello{}
	ping := &message.Ping{}
	r.RegisterHandler(hello.MessageType(), fakeHelloHandler)
	r.RegisterHandler(ping.MessageType(), fakePingHandler)

	nodeA := node.Node{Host: "A", Port: 1}
	nodeB := node.Node{Host: "B", Port: 2}
	a, b := net.Pipe()

	c1 := peer.NewAcceptor(a, nodeA)
	c1.Identify(nodeB)
	go c1.Run()

	c1Mirror := peer.NewAcceptor(b, nodeB)
	c1Mirror.Identify(nodeA)
	go c1Mirror.Run()

	r.Accept(c1)

	c1Mirror.SayHello()
	result := <-c1Mirror.ReceiveChan()

	if result.MessageType() != 1 {
		t.Error("Unexpected response type, expected 1 got", result.MessageType())
	}

	msg := message.Ping{
		Id:   999,
		From: nodeA,
		To:   nodeB,
	}
	c1Mirror.Send(msg)

	result = <-c1Mirror.PongChan()
	if result.MessageType() != 4 {
		t.Error("Unexpected response type, expected 1 got", result.MessageType())
	}

	pong := result.(*message.Pong)
	if pong.Id != 999 {
		t.Error("Unexpected result Id")
	}
	if pong.From != nodeB {
		t.Error("Unexpected result From")
	}

	r.Exit()
	c1.Exit()
	c1Mirror.Exit()
}

func fakeHelloHandler(msg message.Message) (message.Message, error) {
	hello := msg.(*message.Hello)
	return &message.Welcome{hello.Id, hello.To, hello.From, hello.Details}, nil
}

func fakePingHandler(msg message.Message) (message.Message, error) {
	ping := msg.(*message.Ping)
	return &message.Pong{ping.Id, ping.To, ping.From}, nil
}

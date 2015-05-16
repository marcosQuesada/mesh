package server

import (
	"fmt"
)

type Broker interface {
	Accept(Peer, *Hello) Message
	Ping(Peer, *Ping) Message
	GoodBye(Peer, *GoodBye) Message
}

type defaultBroker struct {
	PeerHandler
}

func NewBroker() *defaultBroker {
	return &defaultBroker{
		PeerHandler: DefaultPeerHandler(),
	}
}

func (b *defaultBroker) Accept(p Peer, h *Hello) Message {
	fmt.Println("Broker Accept: ", h)
	err := b.PeerHandler.Accept(p)
	if err != nil {
		return &Abort{Id: 123123, Details: map[string]interface{}{"foo_bar": 1231}}
	}

	return &Welcome{Id: 123123, Details: map[string]interface{}{"foo_bar": 1231}}
}

func (b *defaultBroker) Ping(p Peer, pi *Ping) Message {
	/*	if err := b.PeerHandler.Accept(h); !err {
		return &Abort{}
	}*/

	return &Pong{}
}

func (b *defaultBroker) GoodBye(p Peer, g *GoodBye) Message {
	fmt.Println("Broker GoodBye: ", g)
	err := b.PeerHandler.Remove(p)
	if err != nil {
		return &Abort{Id: 123123, Details: map[string]interface{}{"foo_bar": 1231}}
	}
	return &GoodBye{Id: 123123, Details: map[string]interface{}{"foo_bar": 1231}}
}

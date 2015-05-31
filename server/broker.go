package server

import (
	"log"
)

type Broker interface {
	Accept(PeerClient, *Hello) Message
	Ping(PeerClient, *Ping) Message
	GoodBye(PeerClient, *GoodBye) Message
}

type defaultBroker struct {
	from Node
	//PeerHandler
}

func NewBroker(from Node) *defaultBroker {
	return &defaultBroker{
		from: from,
	}
}

func (b *defaultBroker) Accept(p PeerClient, h *Hello) Message {
	log.Println("Broker Accept: ", h)
	/*	err := b.PeerHandler.Accept(p)
		if err != nil {
			return &Abort{Id: h.Id, From: p.Id(), Details: map[string]interface{}{"foo_bar": 1231}}
		}*/

	return &Welcome{Id: h.Id, From: b.from, Details: map[string]interface{}{"foo_bar": 1231}}
}

func (b *defaultBroker) Ping(p PeerClient, pi *Ping) Message {
	return &Pong{Id: pi.Id, From: b.from, Details: map[string]interface{}{"foo_bar": 1231}}
}

func (b *defaultBroker) GoodBye(p PeerClient, g *GoodBye) Message {
	/*	log.Println("Broker GoodBye: ", g)
		err := b.PeerHandler.Remove(p)
		if err != nil {
			return &Abort{Id: g.Id, From: p.Id(), Details: map[string]interface{}{"foo_bar": 1231}}
		}*/
	return &GoodBye{Id: g.Id, From: b.from, Details: map[string]interface{}{"foo_bar": 1231}}
}

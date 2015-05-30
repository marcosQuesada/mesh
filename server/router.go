package server

import (
	"log"
)

type Router interface {
	Accept(p Peer)
}

type defaultRouter struct {
	Broker
	node *Node
	exit chan bool
}

func NewRouter(n *Node) *defaultRouter {
	return &defaultRouter{
		Broker: NewBroker(n),
		node:   n,
		exit:   make(chan bool),
	}
}

func (r *defaultRouter) Accept(p Peer) {
	defer close(r.exit)
	log.Println("Router accepting peer: ", p.Id())
	for {
		select {
		case <-r.exit:
			return
		case msg := <-p.ReadMessage():
			switch msg.(type) {
			case *Hello:
				log.Println("Router Hello", msg.(*Hello))
				result := r.Broker.Accept(p, msg.(*Hello))
				p.Send(result)

			case *Welcome:
				log.Println("Router Welcome: ", msg.(*Welcome))
				a := msg.(*Welcome)
				p.Send(&Error{Id: a.Id, Details: a.Details})

			case *Abort:
				log.Println("Router Abort: ", msg.(*Abort))
				a := msg.(*Abort)
				p.Send(&Error{Id: a.Id, Details: a.Details})

			case *Ping:
				result := r.Broker.Ping(p, msg.(*Ping))
				p.Send(result)
				log.Println("Router Ping: ", msg.(*Ping))

			case *Pong:
				log.Println("Router Pong: ", msg.(*Pong))
				pong := msg.(*Pong)
				p.Send(&Error{Id: pong.Id, Details: pong.Details})

			case *GoodBye:
				result := r.Broker.GoodBye(p, msg.(*GoodBye))
				p.Send(result)

			default:

			}
		default:

		}
	}
}

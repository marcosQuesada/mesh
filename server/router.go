package server

import (
	"log"
)

type Router interface {
	Accept(PeerClient)
}

type defaultRouter struct {
	node          Node
	exit          chan bool
	clientHandler ClientHandler
}

func NewRouter(n Node, clh ClientHandler) *defaultRouter {
	return &defaultRouter{
		node:          n,
		exit:          make(chan bool),
		clientHandler: clh,
	}
}

func (r *defaultRouter) Accept(p PeerClient) {
	for {
		select {
		case <-r.exit:
			return
		case msg := <-p.ReceiveChan():
			log.Println("Msg Received ", msg)
			switch msg.(type) {
			case *Hello:
				p.Identify(msg.(*Hello).From)
				err := r.clientHandler.Accept(p)
				if err != nil {
					p.Send(&Abort{Id: msg.(*Hello).Id, From: r.node, Details: map[string]interface{}{"foo_bar": 1231}})
					p.Exit()

					return
				}
				p.Send(&Welcome{Id: msg.(*Hello).Id, From: r.node, Details: map[string]interface{}{"foo_bar": 1231}})
			case *Welcome:
				log.Println("Router Welcome: ", msg.(*Welcome))
				a := msg.(*Welcome)
				p.Send(&Error{Id: a.Id, Details: a.Details})

			case *Abort:
				log.Println("Router Abort: ", msg.(*Abort))
				a := msg.(*Abort)
				p.Send(&Error{Id: a.Id, Details: a.Details})

			case *Ping:
				//result := r.Broker.Ping(p, msg.(*Ping))
				//p.Send(result)
				log.Println("Router Ping: ", msg.(*Ping))

			case *Pong:
				log.Println("Router Pong: ", msg.(*Pong))
				pong := msg.(*Pong)
				p.Send(&Error{Id: pong.Id, Details: pong.Details})

			case *GoodBye:
				//result := r.Broker.GoodBye(p, msg.(*GoodBye))
				//p.Send(result)

			default:

			}
		default:

		}
	}
}

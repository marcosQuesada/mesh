package server

import (
	"fmt"
)

type Router interface {
	Accept(p Peer)
}

type defaultRouter struct {
	exit chan bool
}

func NewRouter() *defaultRouter {
	return &defaultRouter{
		exit: make(chan bool),
	}
}

func (r *defaultRouter) Accept(p Peer) {
	fmt.Println("Router accepting peer: ", p.Id())
	for {
		select {
		case <-r.exit:
			return
		case msg := <-p.ReadMessage():
			switch msg.(type) {
			case *Hello:
				fmt.Println("Router Hello", msg.(*Hello))
			case *Welcome:
				fmt.Println("Router Welcome: ", msg.(*Welcome))
			case *Abort:
				fmt.Println("Router Abort: ", msg.(*Abort))
			case *Ping:
				fmt.Println("Router Ping: ", msg.(*Ping))
			case *Pong:
				fmt.Println("Router Pong: ", msg.(*Pong))
			default:

			}
		default:

		}
	}
}

/*

	defer close(p.exit)
	//var first bool = true //send ID on first message

	for {
		select {
		case m := <-p.rcvChan:
			msg := *m
			switch msg.MessageType() {
			case HELLO:
				fmt.Printf("Peer Hello received", msg.(*Hello))
			case WELCOME:
				fmt.Printf("Peer Welcome received", msg.(*Welcome))
			case ABORT:
				fmt.Printf("Peer Abort received", msg.(*Abort))
			default:
				fmt.Printf("unexpected type %T", msg)
			}
		case <-p.ticker.C:
			if !p.Connected() {
				fmt.Println("Connecting to ", p.remote)
				p.Connect(p.remote)
				defer p.Exit()
			}*/

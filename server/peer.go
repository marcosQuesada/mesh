package server

//A Peer is a representation of a remote node

//It handles connection state supervision

import (
	"log"
	"time"
)

type State string

const (
	PeerStateNew        = State("new")
	PeerStateConnecting = State("connecting")
	PeerStateConnected  = State("connected")
	PeerStateExiting    = State("exiting")
)

type Peer struct {
	Link
	local_node *Node
	remote     *Node
	state      State
	ticker     *time.Ticker
	exit       chan bool
	rcvChan    chan Message
	sndChan    chan Message
}

func NewPeer(lp *Node, n *Node, checkInterval int) *Peer {
	return &Peer{
		Link:       NewDialLink(),
		local_node: lp,
		remote:     n,
		state:      PeerStateNew,
		exit:       make(chan bool),
		ticker:     time.NewTicker(time.Duration(checkInterval) * time.Millisecond)}
}

func (p *Peer) Run() {
	go func() {
		defer close(p.exit)
		var first bool = true //send ID on first message

		for {
			select {
			case <-p.ticker.C:
				if !p.Connected() {
					p.Connect(p.remote)
					defer p.Exit()
				} else {
					message := "Hi from Peer " + p.local_node.Address() + "\n"
					if first {
						message = p.local_node.Address() + "\n"
						first = false
					}

					err := p.Send([]byte(message))
					if err != nil {
						log.Println("Error Writting on socket ", err)
					}
				}
			case <-p.exit:
				log.Println("Exiting from Peer ", p.remote)
				return
			default:
			}
		}
	}()
}

func (p *Peer) Exit() {
	log.Println("Exitting Peer ", p.remote)
	p.exit <- true
}

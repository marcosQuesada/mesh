package server

//A peer is a representation of a remote node

//It handles connection state supervision

import (
	"fmt"
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
	remote *Node
	state  State
	ticker *time.Ticker
	exit   chan bool
}

func NewPeer(n *Node, checkInterval int) *Peer {
	return &Peer{
		remote: n,
		state:  PeerStateNew,
		exit:   make(chan bool),
		ticker: time.NewTicker(time.Duration(checkInterval) * time.Millisecond)}
}

func (p *Peer) Run() {
	go func() {
		defer close(p.exit)
		for {
			select {
			case <-p.ticker.C:
				fmt.Println("Running Peer ", p.remote, "Ticker check")
				//Check remote peer
			case <-p.exit:
				fmt.Println("Exiting from Peer ", p.remote)
				return
			default:
			}
		}
	}()
}

func (p *Peer) Exit() {
	fmt.Println("Exitting Peer ", p.remote)
	p.exit <- true
}

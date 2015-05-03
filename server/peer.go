package server

//A peer is a representation of a remote node

//It handles connection state supervision

import (
	"fmt"
	"net"
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
	remote     *Node
	local_node *Node
	state      State
	ticker     *time.Ticker
	exit       chan bool
	conn       net.Conn
}

func NewPeer(lp *Node, n *Node, checkInterval int) *Peer {
	return &Peer{
		local_node: lp,
		remote:     n,
		state:      PeerStateNew,
		exit:       make(chan bool),
		ticker:     time.NewTicker(time.Duration(checkInterval) * time.Millisecond)}
}

func (p *Peer) Run() {
	go func() {
		defer close(p.exit)
		for {
			select {
			case <-p.ticker.C:
				if p.conn == nil {
					conn, err := net.Dial("tcp", p.remote.Address())
					if err != nil {
						fmt.Println("Error starting socket client to: ", p.remote, "err: ", err)
						continue
					}
					fmt.Println("Peer connected to ", p.remote.Address())
					p.conn = conn
					defer p.conn.Close()
				} else {
					_, err := p.conn.Write([]byte("Hi from Peer " + p.local_node.Address() + "\n"))
					if err != nil {
						p.conn = nil
						fmt.Println("Error Writting on socket ", err)
					}
				}
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

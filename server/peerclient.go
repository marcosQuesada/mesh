package server

//A PeerClient is a representation of a remote node

//It handles connection state supervision

import (
	"fmt"
	"net"
	"time"
)

type State string

const (
	PeerClientStateNew        = State("new")
	PeerClientStateConnecting = State("connecting")
	PeerClientStateConnected  = State("connected")
	PeerClientStateExiting    = State("exiting")
)

type PeerClient struct {
	local_node *Node
	remote     *Node
	state      State
	ticker     *time.Ticker
	exit       chan bool
	conn       net.Conn
}

func NewPeerClient(lp *Node, n *Node, checkInterval int) *PeerClient {
	return &PeerClient{
		local_node: lp,
		remote:     n,
		state:      PeerClientStateNew,
		exit:       make(chan bool),
		ticker:     time.NewTicker(time.Duration(checkInterval) * time.Millisecond)}
}

func (p *PeerClient) Run() {
	go func() {
		defer close(p.exit)
		var first bool = true //send ID on first message
		for {
			select {
			case <-p.ticker.C:
				if p.conn == nil {
					conn, err := net.Dial("tcp", p.remote.Address())
					if err != nil {
						fmt.Println("Error starting socket client to: ", p.remote, "err: ", err)
						continue
					}
					//On success persist connection
					p.conn = conn
					defer p.conn.Close()
				} else {
					message := "Hi from PeerClient " + p.local_node.Address() + "\n"
					if first {
						message = p.local_node.Address() + "\n"
						first = false
					}

					_, err := p.conn.Write([]byte(message))
					if err != nil {
						p.conn = nil
						fmt.Println("Error Writting on socket ", err)
					}
				}
			case <-p.exit:
				fmt.Println("Exiting from PeerClient ", p.remote)
				return
			default:
			}
		}
	}()
}

func (p *PeerClient) Exit() {
	fmt.Println("Exitting PeerClient ", p.remote)
	p.exit <- true
}

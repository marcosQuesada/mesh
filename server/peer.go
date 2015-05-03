package server

import (
	"bufio"
	"fmt"
	"net"
)

type Peer interface {
	Connect(*Node)
	Send(message []byte) error
	Receive() []byte
	Connected() bool
	Exit()
}

type DialPeer struct {
	conn net.Conn
}

func NewDialPeer() *DialPeer {
	return &DialPeer{}
}

func (p *DialPeer) Connect(n *Node) {
	conn, err := net.Dial("tcp", n.Address())
	if err != nil {
		fmt.Println("Error starting socket client to: ", n.Address(), "err: ", err)
		return
	}

	p.conn = conn
}

func (p *DialPeer) Send(message []byte) error {
	_, err := p.conn.Write([]byte(message))
	if err != nil {
		p.conn = nil
		fmt.Println("Error Writting on socket ", err)
		return err
	}

	return nil
}

func (p *DialPeer) Receive() []byte {

	return nil
}

func (p *DialPeer) Connected() bool {
	return p.conn != nil
}

func (p *DialPeer) Exit() {
	p.conn.Close()
	p.conn = nil
}

type ServerPeer struct {
	conn     net.Conn
	listener net.Listener
}

func NewServerPeer(l net.Listener) *ServerPeer {
	return &ServerPeer{
		listener: l,
	}
}

func (p *ServerPeer) Connect(n *Node) {
	conn, err := p.listener.Accept()
	if err != nil {
		fmt.Println("Error starting socket client to: ", n.Address(), "err: ", err)
		return
	}

	p.conn = conn
}

func (p *ServerPeer) Connected() bool {
	return p.conn != nil
}

func (p *ServerPeer) Send(message []byte) error {
	_, err := p.conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error Writting on socket ", err)

		return err
	}

	return nil
}

func (p *ServerPeer) Receive() []byte {
	message, err := bufio.NewReader(p.conn).ReadBytes('\n')
	if err != nil {
		fmt.Print("Error Receiving on server, err ", err)
		return nil
	}
	return message
}

func (p *ServerPeer) Exit() {
	p.conn.Close()
}

package server

import (
	"bufio"
	"fmt"
	"net"
)

type Peer interface {
	Id() *Node
	Identify(*Node)
	Connect(*Node)
	Send(message []byte) error
	Receive() ([]byte, error)
	Connected() bool
	Exit()
}

type DialPeer struct {
	conn net.Conn
	id   *Node
}

func NewDialPeer() *DialPeer {
	return &DialPeer{}
}

func (p *DialPeer) Connect(n *Node) {
	p.id = n
	conn, err := net.Dial("tcp", n.Address())
	if err != nil {
		fmt.Println("Error starting socket client to: ", n.Address(), "err: ", err)
		return
	}

	p.conn = conn
}

func (p *DialPeer) Identify(n *Node) {
	p.id = n
}

func (p *DialPeer) Id() *Node {
	return p.id
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

func (p *DialPeer) Receive() ([]byte, error) {
	message, err := bufio.NewReader(p.conn).ReadBytes('\n')
	if err != nil {
		fmt.Print("Error Receiving on server, err ", err)
		return nil, err
	}
	return message, nil
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
	id       *Node
}

func NewServerPeer(l net.Listener) *ServerPeer {
	return &ServerPeer{
		listener: l,
	}
}

func (p *ServerPeer) Identify(n *Node) {
	p.id = n
}

func (p *ServerPeer) Id() *Node {
	return p.id
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

func (p *ServerPeer) Receive() ([]byte, error) {
	message, err := bufio.NewReader(p.conn).ReadBytes('\n')
	if err != nil {
		fmt.Print("Error Receiving on server, err ", err)
		return nil, err
	}
	return message, nil
}

func (p *ServerPeer) Exit() {
	p.conn.Close()
}

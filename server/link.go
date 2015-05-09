package server

import (
	"bufio"
	"log"
	"net"
)

type Link interface {
	Id() *Node
	Identify(*Node)
	Connect(*Node)
	Send(message []byte) error
	Receive() ([]byte, error)
	Connected() bool
	Exit()
}

//Client Link Definitions
type DialLink struct {
	conn net.Conn
	id   *Node
}

func NewDialLink() *DialLink {
	return &DialLink{}
}

func (p *DialLink) Connect(n *Node) {
	p.id = n
	conn, err := net.Dial("tcp", n.Address())
	if err != nil {
		log.Println("Error starting socket client to: ", n.Address(), "err: ", err)
		return
	}

	p.conn = conn
}

func (p *DialLink) Identify(n *Node) {
	p.id = n
}

func (p *DialLink) Id() *Node {
	return p.id
}

func (p *DialLink) Send(message []byte) error {
	_, err := p.conn.Write([]byte(message))
	if err != nil {
		p.conn = nil
		log.Println("Error Writting on socket ", err)
		return err
	}

	return nil
}

func (p *DialLink) Receive() ([]byte, error) {
	message, err := bufio.NewReader(p.conn).ReadBytes('\n')
	if err != nil {
		log.Print("Error Receiving on server, err ", err)
		return nil, err
	}
	return message, nil
}

func (p *DialLink) Connected() bool {
	return p.conn != nil
}

func (p *DialLink) Exit() {
	p.conn.Close()
	p.conn = nil
}

//Server Link Definitions
type ServerLink struct {
	conn     net.Conn
	listener net.Listener
	id       *Node
}

func NewServerLink(l net.Listener) *ServerLink {
	return &ServerLink{
		listener: l,
	}
}

func (p *ServerLink) Identify(n *Node) {
	p.id = n
}

func (p *ServerLink) Id() *Node {
	return p.id
}

func (p *ServerLink) Connect(n *Node) {
	conn, err := p.listener.Accept()
	if err != nil {
		log.Println("Error starting socket client to: ", n.Address(), "err: ", err)
		return
	}

	p.conn = conn
}

func (p *ServerLink) Connected() bool {
	return p.conn != nil
}

func (p *ServerLink) Send(message []byte) error {
	_, err := p.conn.Write([]byte(message))
	if err != nil {
		log.Println("Error Writting on socket ", err)

		return err
	}

	return nil
}

func (p *ServerLink) Receive() ([]byte, error) {
	message, err := bufio.NewReader(p.conn).ReadBytes('\n')
	if err != nil {
		log.Print("Error Receiving on server, err ", err)
		return nil, err
	}
	return message, nil
}

func (p *ServerLink) Exit() {
	p.conn.Close()
}

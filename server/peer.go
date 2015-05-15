package server

//A Peer is a representation of a remote node

//It handles connection state supervision

import (
	"fmt"
	"github.com/nu7hatch/gouuid"
	"log"
	"net"
)

type Peer interface {
	Receive() (Message, error)
	Id() ID
}

type ID *uuid.UUID

type SocketPeer struct {
	conn       net.Conn
	id         ID
	serializer Serializer
	rcvChan    chan Message
	exitChan   chan bool
}

func NewSocketPeer(conn net.Conn) *SocketPeer {
	id, err := uuid.NewV4()
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}

	return &SocketPeer{
		id:         id,
		conn:       conn,
		serializer: &JsonSerializer{},
		rcvChan:    make(chan Message),
		exitChan:   make(chan bool),
	}
}

func (p *SocketPeer) Id() ID {
	return p.id
}
func (p *SocketPeer) Receive() (msg Message, err error) {
	var rawMsg []byte
	n, err := p.conn.Read(rawMsg)
	if err != nil {
		log.Print("Error Receiving on server, err ", err)
		return nil, err
	}

	if n > 0 {
		msg, err = p.serializer.Deserialize(rawMsg)
	}

	return
}

package server

//A Peer is a representation of a remote node

//It handles connection state supervision

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"io"
	"log"
	"net"
)

type Peer interface {
	Id() ID
	Receive() (Message, error)
	Send(Message) error
	ReadMessage() chan Message
}

type ID string

type SocketPeer struct {
	Conn       net.Conn
	id         ID
	serializer Serializer
	RcvChan    chan Message
	exitChan   chan bool
}

func NewSocketPeer(conn net.Conn) *SocketPeer {
	//connection origin
	fmt.Println("Socket Peer conn origin: ", conn.RemoteAddr())
	id, err := uuid.NewV4()
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}
	return &SocketPeer{
		id:         ID(id.String()),
		Conn:       conn,
		serializer: &JsonSerializer{}, //@TODO: Must be plugable!
		RcvChan:    make(chan Message),
		exitChan:   make(chan bool),
	}
}

func (p *SocketPeer) Id() ID {
	return p.id
}

func (p *SocketPeer) Send(msg Message) error {
	rawMsg, err := p.serializer.Serialize(msg)
	if err != nil {
		return err
	}

	var n int
	rawMsg = append(rawMsg, '\n')
	n, err = p.Conn.Write(rawMsg)
	if err != nil || n == 0 {
		return fmt.Errorf("Error writting ", err, n)
	}

	return nil
}

func (p *SocketPeer) Receive() (msg Message, err error) {
	b := bufio.NewReader(p.Conn)
	buffer, err := b.ReadBytes('\n')
	if err != nil {
		if err != io.EOF {
			log.Print("Error Receiving on server, err ", err)
			p.Conn.Close()
		}

		return nil, err
	}
	c := bytes.Trim(buffer, "\n")
	msg, err = p.serializer.Deserialize(c)

	return
}

func (p *SocketPeer) ReadMessage() chan Message {
	return p.RcvChan
}

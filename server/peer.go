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
	"time"
)

type Peer interface {
	Id() ID
	Remote() net.Addr //.String()
	Receive() (Message, error)
	Send(Message) error
	ReadMessage() chan Message
}

type ID string

type SocketPeer struct {
	Conn       net.Conn
	id         ID
	Node       Node
	serializer Serializer
	RcvChan    chan Message
	exitChan   chan bool
}

func NewJSONSocketPeer(conn net.Conn) *SocketPeer {
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

func (p *SocketPeer) Remote() net.Addr {
	return p.Conn.RemoteAddr()
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

type response struct {
	Msg Message
	Err error
}

func (p *SocketPeer) Receive() (msg Message, err error) {
	r := make(chan response)

	go func(rsp chan response) {
		b := bufio.NewReader(p.Conn)
		buffer, err := b.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Print("Error Receiving on server, err ", err)
				p.Conn.Close()
			}
			r <- response{Err: err}
		}
		c := bytes.Trim(buffer, "\n")
		msg, err = p.serializer.Deserialize(c)
		r <- response{Msg: msg, Err: err}
		close(r)
	}(r)

	timeout := time.NewTimer(time.Second * 2)

	select {
	case m := <-r:
		return m.Msg, m.Err
	case <-timeout.C:
		return nil, fmt.Errorf("Timeout receiving peer response")
	}
	return
}

func (p *SocketPeer) ReadMessage() chan Message {
	return p.RcvChan
}

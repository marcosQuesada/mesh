package peer

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/marcosQuesada/mesh/message"
	"github.com/marcosQuesada/mesh/node"
	"github.com/marcosQuesada/mesh/serializer"
	"github.com/nu7hatch/gouuid"
	"io"
	"log"
	"net"
	"time"
)

type Link interface {
	Id() ID
	Remote() net.Addr //.String()
	Receive() (message.Message, error)
	ReceiveTimeout() (message.Message, error)
	Send(message.Message) error
}

type ID string

type SocketLink struct {
	Conn       net.Conn
	id         ID
	Node       node.Node
	serializer serializer.Serializer
}

func NewJSONSocketLink(conn net.Conn) *SocketLink {
	id, err := uuid.NewV4()
	if err != nil {
		log.Println("error:", err)
		return nil
	}
	return &SocketLink{
		id:         ID(id.String()),
		Conn:       conn,
		serializer: &serializer.JsonSerializer{}, //@TODO: Must be plugable!
	}
}

func (p *SocketLink) Id() ID {
	return p.id
}

func (p *SocketLink) Remote() net.Addr {
	return p.Conn.RemoteAddr()
}

func (p *SocketLink) Send(msg message.Message) error {
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
	Msg message.Message
	Err error
}

func (p *SocketLink) Receive() (msg message.Message, err error) {
	b := bufio.NewReader(p.Conn)
	buffer, err := b.ReadBytes('\n')
	if err != nil {
		if err != io.ErrClosedPipe && err != io.EOF {
			log.Print("Socket Reader err: ", err)
		}
		p.Conn.Close()

		return nil, err
	}
	c := bytes.Trim(buffer, "\n")
	msg, err = p.serializer.Deserialize(c)

	return
}

func (p *SocketLink) ReceiveTimeout() (msg message.Message, err error) {
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
		return nil, fmt.Errorf("Timeout receiving Link response")
	}

	return
}

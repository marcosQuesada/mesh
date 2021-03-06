package peer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/marcosQuesada/mesh/pkg/message"
	"github.com/marcosQuesada/mesh/pkg/node"
	"github.com/marcosQuesada/mesh/pkg/serializer"
	"github.com/nu7hatch/gouuid"
)

type Link interface {
	Id() ID
	Remote() net.Addr //.String()
	Receive() (message.Message, error)
	Send(message.Message) error
}

type ID string

type SocketLink struct {
	Conn       net.Conn
	id         ID
	Node       node.Node
	serializer serializer.Serializer
	mutex      sync.Mutex
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

	p.mutex.Lock()
	defer p.mutex.Unlock()
	var n int
	rawMsg = append(rawMsg, '\n')
	n, err = p.Conn.Write(rawMsg)
	if err != nil || n == 0 {
		return fmt.Errorf("Error writting %v %d", err, n)
	}

	return nil
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

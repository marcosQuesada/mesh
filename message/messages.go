package message

import (
	n "github.com/marcosQuesada/mesh/node"
	"github.com/nu7hatch/gouuid"
	"log"
)

type Status string

type ID string

func NewId() ID{
	idV4, err := uuid.NewV4()
	if err != nil {
		log.Println("error generating v4 ID:", err)
	}

	id, err := uuid.NewV5(idV4, []byte("message"))
	if err != nil {
		log.Println("error generating v5 ID:", err)
	}

	return ID(id.String())
}

type Message interface {
	MessageType() MsgType
	Origin() n.Node
	Destination() n.Node
}

type MsgType int

func (mt MsgType) New() Message {
	switch mt {
	case HELLO:
		return &Hello{}
	case WELCOME:
		return &Welcome{}
	case ABORT:
		return &Abort{}
	case PING:
		return &Ping{}
	case PONG:
		return &Pong{}
	case GOODBYE:
		return &GoodBye{}
	case DONE:
		return &Done{}
	case ERROR:
		return &Error{}
	case COMMAND:
		return &Command{}
	}

	return nil
}

const (
	HELLO   = MsgType(0)
	WELCOME = MsgType(1)
	ABORT   = MsgType(2)
	PING    = MsgType(3)
	PONG    = MsgType(4)
	GOODBYE = MsgType(5)
	ACK 	= MsgType(6)
	DONE    = MsgType(90)
	ERROR   = MsgType(99)
	COMMAND = MsgType(50)
)

// First connection message
type Hello struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (h Hello) MessageType() MsgType {
	return HELLO
}

func (h Hello) Origin() n.Node {
	return h.From
}

func (h Hello) Destination() n.Node {
	return h.To
}

// Hello Accepted
type Welcome struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (w Welcome) MessageType() MsgType {
	return WELCOME
}

func (h Welcome) Origin() n.Node {
	return h.From
}

func (h Welcome) Destination() n.Node {
	return h.To
}

// Hello Rejected
type Abort struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (a Abort) MessageType() MsgType {
	return ABORT
}

func (h Abort) Origin() n.Node {
	return h.From
}

func (h Abort) Destination() n.Node {
	return h.To
}

// Ping request to a remote node
type Ping struct {
	Id   ID
	From n.Node
	To   n.Node
}

func (p Ping) MessageType() MsgType {
	return PING
}

func (h Ping) Origin() n.Node {
	return h.From
}

func (h Ping) Destination() n.Node {
	return h.To
}

// Pong response as a ping request
type Pong struct {
	Id   ID
	From n.Node
	To   n.Node
}

func (p Pong) MessageType() MsgType {
	return PONG
}

func (h Pong) Origin() n.Node {
	return h.From
}

func (h Pong) Destination() n.Node {
	return h.To
}

type GoodBye struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (g GoodBye) MessageType() MsgType {
	return GOODBYE
}

func (h GoodBye) Origin() n.Node {
	return h.From
}

func (h GoodBye) Destination() n.Node {
	return h.To
}

type Done struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (g Done) MessageType() MsgType {
	return DONE
}

func (h Done) Origin() n.Node {
	return h.From
}

func (h Done) Destination() n.Node {
	return h.To
}

type Error struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (w Error) MessageType() MsgType {
	return ERROR
}

func (h Error) Origin() n.Node {
	return h.From
}

func (h Error) Destination() n.Node {
	return h.To
}

type Command struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (w Command) MessageType() MsgType {
	return COMMAND
}

func (h Command) Origin() n.Node {
	return h.From
}

func (h Command) Destination() n.Node {
	return h.To
}


type Ack struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (w Ack) MessageType() MsgType {
	return COMMAND
}

func (h Ack) Origin() n.Node {
	return h.From
}

func (h Ack) Destination() n.Node {
	return h.To
}



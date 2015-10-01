package message

import (
	n "github.com/marcosQuesada/mesh/node"
	"github.com/nu7hatch/gouuid"
	"log"
)

type Status string

type ID string

func NewId() ID {
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
	ID() ID
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
	case RESPONSE:
		return &Response{}
	case RAFTCOMMAND:
		return &RaftCommand{}
	case ACK:
		return &Ack{}
	}

	return nil
}

const (
	HELLO       = MsgType(0)
	WELCOME     = MsgType(1)
	ABORT       = MsgType(2)
	PING        = MsgType(3)
	PONG        = MsgType(4)
	GOODBYE     = MsgType(5)
	ACK         = MsgType(6)
	DONE        = MsgType(90)
	ERROR       = MsgType(99)
	COMMAND     = MsgType(50)
	RESPONSE    = MsgType(51)
	RAFTCOMMAND = MsgType(52)
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

func (h Hello) ID() ID {
	return h.Id
}

// Hello Accepted
type Welcome struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (h Welcome) MessageType() MsgType {
	return WELCOME
}

func (h Welcome) Origin() n.Node {
	return h.From
}

func (h Welcome) Destination() n.Node {
	return h.To
}

func (h Welcome) ID() ID {
	return h.Id
}

// Hello Rejected
type Abort struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (h Abort) MessageType() MsgType {
	return ABORT
}

func (h Abort) Origin() n.Node {
	return h.From
}

func (h Abort) Destination() n.Node {
	return h.To
}

func (h Abort) ID() ID {
	return h.Id
}

// Ping request to a remote node
type Ping struct {
	Id   ID
	From n.Node
	To   n.Node
}

func (h Ping) MessageType() MsgType {
	return PING
}

func (h Ping) Origin() n.Node {
	return h.From
}

func (h Ping) Destination() n.Node {
	return h.To
}

func (h Ping) ID() ID {
	return h.Id
}

// Pong response as a ping request
type Pong struct {
	Id   ID
	From n.Node
	To   n.Node
}

func (h Pong) MessageType() MsgType {
	return PONG
}

func (h Pong) Origin() n.Node {
	return h.From
}

func (h Pong) Destination() n.Node {
	return h.To
}

func (h Pong) ID() ID {
	return h.Id
}

type GoodBye struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (h GoodBye) MessageType() MsgType {
	return GOODBYE
}

func (h GoodBye) Origin() n.Node {
	return h.From
}

func (h GoodBye) Destination() n.Node {
	return h.To
}

func (h GoodBye) ID() ID {
	return h.Id
}

type Done struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (h Done) MessageType() MsgType {
	return DONE
}

func (h Done) Origin() n.Node {
	return h.From
}

func (h Done) Destination() n.Node {
	return h.To
}

func (h Done) ID() ID {
	return h.Id
}

type Error struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (h Error) MessageType() MsgType {
	return ERROR
}

func (h Error) Origin() n.Node {
	return h.From
}

func (h Error) Destination() n.Node {
	return h.To
}

func (h Error) ID() ID {
	return h.Id
}

type Command struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
	Command interface{} //@TODO: Provisional
}

func (h Command) MessageType() MsgType {
	return COMMAND
}

func (h Command) Origin() n.Node {
	return h.From
}

func (h Command) Destination() n.Node {
	return h.To
}

func (h Command) ID() ID {
	return h.Id
}

type Response struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
	Result  interface{}
}

func (h Response) MessageType() MsgType {
	return RESPONSE
}

func (h Response) Origin() n.Node {
	return h.From
}

func (h Response) Destination() n.Node {
	return h.To
}

func (h Response) ID() ID {
	return h.Id
}

type Ack struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
}

func (h Ack) MessageType() MsgType {
	return ACK
}

func (h Ack) Origin() n.Node {
	return h.From
}

func (h Ack) Destination() n.Node {
	return h.To
}

func (h Ack) ID() ID {
	return h.Id
}

type RaftCommand struct {
	Id      ID
	From    n.Node
	To      n.Node
	Details map[string]interface{}
	Command interface{} //@TODO: Provisional
}

func (h RaftCommand) MessageType() MsgType {
	return RAFTCOMMAND
}

func (h RaftCommand) Origin() n.Node {
	return h.From
}

func (h RaftCommand) Destination() n.Node {
	return h.To
}

func (h RaftCommand) ID() ID {
	return h.Id
}
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
	case ERROR:
		return &Error{}
	case COMMAND:
		return &Command{}
	case RESPONSE:
		return &Response{}
	case RAFTVOTEREQUEST:
		return &RaftVoteRequest{Candidate: n.Node{}}
	case RAFTVOTERESPONSE:
		return &RaftVoteResponse{} //Vote: n.Node{}
	case RAFTHEARTBEATREQUEST:
		return &RaftHeartBeatRequest{Leader: n.Node{}}
	case ACK:
		return &Ack{}
	}

	return nil
}

const (
	HELLO                = MsgType(0)
	WELCOME              = MsgType(1)
	ABORT                = MsgType(2)
	ACK                  = MsgType(6)
	PING                 = MsgType(3)
	PONG                 = MsgType(4)
	ERROR                = MsgType(99)
	COMMAND              = MsgType(50)
	RESPONSE             = MsgType(51)
	RAFTVOTEREQUEST      = MsgType(52)
	RAFTVOTERESPONSE     = MsgType(53)
	RAFTHEARTBEATREQUEST = MsgType(54)
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
	Id   ID
	From n.Node
	To   n.Node
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
	Id   ID
	From n.Node
	To   n.Node
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

type Ack struct {
	Id   ID
	From n.Node
	To   n.Node
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

type Error struct {
	Id    ID
	From  n.Node
	To    n.Node
	Error interface{} //@TODO: Provisional!
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
	Id     ID
	From   n.Node
	To     n.Node
	Result interface{}
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

type RaftVoteRequest struct {
	Id           ID
	From         n.Node
	To           n.Node
	Candidate    n.Node
	Term         int
	LastLogIndex ID
	LastLogTerm  ID
}

func (h RaftVoteRequest) MessageType() MsgType {
	return RAFTVOTEREQUEST
}

func (h RaftVoteRequest) Origin() n.Node {
	return h.From
}

func (h RaftVoteRequest) Destination() n.Node {
	return h.To
}

func (h RaftVoteRequest) ID() ID {
	return h.Id
}

type RaftVoteResponse struct {
	Id          ID
	From        n.Node
	To          n.Node
	Vote        string //@TODO: PROVISIONAL
	Term        int
	VoteGranted bool
}

func (h RaftVoteResponse) MessageType() MsgType {
	return RAFTVOTERESPONSE
}

func (h RaftVoteResponse) Origin() n.Node {
	return h.From
}

func (h RaftVoteResponse) Destination() n.Node {
	return h.To
}

func (h RaftVoteResponse) ID() ID {
	return h.Id
}

type RaftHeartBeatRequest struct {
	Id     ID
	From   n.Node
	To     n.Node
	Leader n.Node
}

func (h RaftHeartBeatRequest) MessageType() MsgType {
	return RAFTHEARTBEATREQUEST
}

func (h RaftHeartBeatRequest) Origin() n.Node {
	return h.From
}

func (h RaftHeartBeatRequest) Destination() n.Node {
	return h.To
}

func (h RaftHeartBeatRequest) ID() ID {
	return h.Id
}

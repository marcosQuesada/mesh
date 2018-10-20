package message

import (
	"github.com/marcosQuesada/mesh/pkg/node"
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
	Origin() *node.Node
	Destination() *node.Node
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
		return &RaftVoteRequest{Candidate: &node.Node{}}
	case RAFTVOTERESPONSE:
		return &RaftVoteResponse{} //Vote: node.Node{}
	case RAFTHEARTBEATREQUEST:
		return &RaftHeartBeatRequest{Leader: &node.Node{}}
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
	From    *node.Node
	To      *node.Node
	Details map[string]interface{}
}

func (h Hello) MessageType() MsgType {
	return HELLO
}

func (h Hello) Origin() *node.Node {
	return h.From
}

func (h Hello) Destination() *node.Node {
	return h.To
}

func (h Hello) ID() ID {
	return h.Id
}

// Hello Accepted
type Welcome struct {
	Id   ID
	From *node.Node
	To   *node.Node
}

func (h Welcome) MessageType() MsgType {
	return WELCOME
}

func (h Welcome) Origin() *node.Node {
	return h.From
}

func (h Welcome) Destination() *node.Node {
	return h.To
}

func (h Welcome) ID() ID {
	return h.Id
}

// Hello Rejected
type Abort struct {
	Id   ID
	From *node.Node
	To   *node.Node
}

func (h Abort) MessageType() MsgType {
	return ABORT
}

func (h Abort) Origin() *node.Node {
	return h.From
}

func (h Abort) Destination() *node.Node {
	return h.To
}

func (h Abort) ID() ID {
	return h.Id
}

// Ping request to a remote node
type Ping struct {
	Id   ID
	From *node.Node
	To   *node.Node
}

type Ack struct {
	Id   ID
	From *node.Node
	To   *node.Node
}

func (h Ack) MessageType() MsgType {
	return ACK
}

func (h Ack) Origin() *node.Node {
	return h.From
}

func (h Ack) Destination() *node.Node {
	return h.To
}

func (h Ack) ID() ID {
	return h.Id
}

func (h Ping) MessageType() MsgType {
	return PING
}

func (h Ping) Origin() *node.Node {
	return h.From
}

func (h Ping) Destination() *node.Node {
	return h.To
}

func (h Ping) ID() ID {
	return h.Id
}

// Pong response as a ping request
type Pong struct {
	Id   ID
	From *node.Node
	To   *node.Node
}

func (h Pong) MessageType() MsgType {
	return PONG
}

func (h Pong) Origin() *node.Node {
	return h.From
}

func (h Pong) Destination() *node.Node {
	return h.To
}

func (h Pong) ID() ID {
	return h.Id
}

type Error struct {
	Id   ID
	From *node.Node
	To   *node.Node
	Err  error
}

func (h Error) MessageType() MsgType {
	return ERROR
}

func (h Error) Origin() *node.Node {
	return h.From
}

func (h Error) Destination() *node.Node {
	return h.To
}

func (h Error) ID() ID {
	return h.Id
}

type Command struct {
	Id      ID
	From    *node.Node
	To      *node.Node
	Command interface{} //@TODO: Provisional
}

func (h Command) MessageType() MsgType {
	return COMMAND
}

func (h Command) Origin() *node.Node {
	return h.From
}

func (h Command) Destination() *node.Node {
	return h.To
}

func (h Command) ID() ID {
	return h.Id
}

type Response struct {
	Id     ID
	From   *node.Node
	To     *node.Node
	Result interface{}
}

func (h Response) MessageType() MsgType {
	return RESPONSE
}

func (h Response) Origin() *node.Node {
	return h.From
}

func (h Response) Destination() *node.Node {
	return h.To
}

func (h Response) ID() ID {
	return h.Id
}

type RaftVoteRequest struct {
	Id           ID
	From         *node.Node
	To           *node.Node
	Candidate    *node.Node
	Term         int
	LastLogIndex ID
	LastLogTerm  ID
}

func (h RaftVoteRequest) MessageType() MsgType {
	return RAFTVOTEREQUEST
}

func (h RaftVoteRequest) Origin() *node.Node {
	return h.From
}

func (h RaftVoteRequest) Destination() *node.Node {
	return h.To
}

func (h RaftVoteRequest) ID() ID {
	return h.Id
}

type RaftVoteResponse struct {
	Id          ID
	From        *node.Node
	To          *node.Node
	Term        int
	VoteGranted bool
}

func (h RaftVoteResponse) MessageType() MsgType {
	return RAFTVOTERESPONSE
}

func (h RaftVoteResponse) Origin() *node.Node {
	return h.From
}

func (h RaftVoteResponse) Destination() *node.Node {
	return h.To
}

func (h RaftVoteResponse) ID() ID {
	return h.Id
}

type RaftHeartBeatRequest struct {
	Id     ID
	From   *node.Node
	To     *node.Node
	Leader *node.Node
}

func (h RaftHeartBeatRequest) MessageType() MsgType {
	return RAFTHEARTBEATREQUEST
}

func (h RaftHeartBeatRequest) Origin() *node.Node {
	return h.From
}

func (h RaftHeartBeatRequest) Destination() *node.Node {
	return h.To
}

func (h RaftHeartBeatRequest) ID() ID {
	return h.Id
}

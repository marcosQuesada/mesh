package message

import (
	n "github.com/marcosQuesada/mesh/node"
)

type Message interface {
	MessageType() MsgType
}

type MsgType int

func (mt MsgType) New() Message {
	switch mt {
	case HELLO:
		return &Hello{From: n.Node{}}
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
	case ERROR:
		return &Error{}
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
	ERROR   = MsgType(99)
)

// First connection message
type Hello struct {
	Id      int
	From    n.Node
	Details map[string]interface{}
}

func (h Hello) MessageType() MsgType {
	return HELLO
}

// Hello Accepted
type Welcome struct {
	Id      int
	From    n.Node
	Details map[string]interface{}
}

func (w Welcome) MessageType() MsgType {
	return WELCOME
}

// Hello Rejected
type Abort struct {
	Id      int
	From    n.Node
	Details map[string]interface{}
}

func (a Abort) MessageType() MsgType {
	return ABORT
}

// Ping request to a remote node
type Ping struct {
	Id      int
	From    n.Node
	Details map[string]interface{}
}

func (p Ping) MessageType() MsgType {
	return PING
}

// Pong response as a ping request
type Pong struct {
	Id      int
	From    n.Node
	Details map[string]interface{}
}

func (p Pong) MessageType() MsgType {
	return PONG
}

type GoodBye struct {
	Id      int
	From    n.Node
	Details map[string]interface{}
}

func (g GoodBye) MessageType() MsgType {
	return GOODBYE
}

type Error struct {
	Id      int
	From    n.Node
	Details map[string]interface{}
}

func (w Error) MessageType() MsgType {
	return ERROR
}

type Status string

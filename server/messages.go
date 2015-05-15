package server

type Message interface {
	MessageType() messageType
}

type messageType int

func (mt messageType) New() Message {
	switch mt {
	case HELLO:
		return &Hello{}
	case WELCOME:
		return &Welcome{}
	}

	return nil
}

const (
	HELLO   = messageType(0)
	WELCOME = messageType(1)
	ABORT   = messageType(2)
	PING    = messageType(3)
	PONG    = messageType(4)
)

// First connection message
type Hello struct {
	Id      int
	Details map[string]interface{}
}

func (h Hello) MessageType() messageType {
	return HELLO
}

// Hello Accepted
type Welcome struct {
	Id int
}

func (w Welcome) MessageType() messageType {
	return WELCOME
}

// Hello Rejected
type Abort struct {
	Id int
}

func (a Abort) MessageType() messageType {
	return ABORT
}

// Ping request to a remote node
type Ping struct {
	Id int
}

func (p Ping) MessageType() messageType {
	return PING
}

// Pong response as a ping request
type Pong struct {
	Id int
}

func (p Pong) MessageType() messageType {
	return PONG
}
